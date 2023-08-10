// 提供服务应用账户的注册、权限配置
package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/config"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/mq"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton-rds-sdk-go/sqlx"
	redis "github.com/go-redis/redis/v8"
)

const (
	// 设置应用账户文档库权限Topic
	SetDoclibPermTopic string = "core.doc_share.doc_lib_perm.app.set"
	// 设置应用账户文档权限Topic
	SetDocPermTopic string = "core.doc_share.doc_perm.app.set"
)

// Account 服务应用账户
type Account struct {
	ID           int
	ClientID     string
	ClientSecret string
	Name         string
	Perm         int
	Created      int64
	Updated      int64
}

// TableName .
func (obj *Account) TableName() string {
	return "Account"
}

func InstallAPPAccount(svcName string, redisClient redis.Cmdable, db *sqlx.DB) {
	var clientID string
	var err error
	cgdb := config.NewConfiguration().DB
	ctx := context.Background()
	if redisClient != nil {
		lock := redisClient.SetNX(ctx, svcName, true, 5*time.Second)
		if !lock.Val() {
			return
		}
	}
	defer func() {
		if redisClient != nil {
			redisClient.Del(ctx, svcName)
		}
		if err != nil {
			panic(err)
		}
	}()

	// 查询是否已有账户
	account := Account{}
	sqlStr := "SELECT id, client_id FROM %v.account WHERE name = ?"
	sqlStr = fmt.Sprintf(sqlStr, cgdb.DBName)
	rows, err := db.Query(sqlStr, svcName)
	defer CloseRows(rows)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&account.ID, &account.ClientID); err != nil {
			return
		}
	}
	clientID = account.ClientID
	if account.ID == 0 {
		clientSecret := RandString(12)
		if clientID == "" {
			clientID, err = registerAPPAccount(svcName, clientSecret)
		}
		account.ClientID = clientID
		account.ClientSecret = clientSecret
		account.Name = svcName
		account.Perm = 0
		ct := NowTimestamp()
		account.Updated = ct
		account.Created = ct
		sqlStr = "INSERT INTO %v.account (client_id, client_secret, name, perm, updated, created) VALUES (?,?,?,?,?,?)"
		sqlStr = fmt.Sprintf(sqlStr, cgdb.DBName)
		_, err = db.Exec(sqlStr, clientID, clientSecret, svcName, 0, ct, ct)
	}
	if account.Perm == 1 {
		return
	}
	// 配置权限
	err = setAPPAccountPerm(clientID)
	// 更新状态
	sqlStr = "UPDATE %v.account SET perm = 1, updated = ? WHERE client_id = ?"
	sqlStr = fmt.Sprintf(sqlStr, cgdb.DBName)
	_, err = db.Exec(sqlStr, NowTimestamp(), account.ClientID)
}

func registerAPPAccount(name string, password string) (appID string, err error) {
	// 新增内部账户
	client := &http.Client{}
	reqBody := map[string]string{
		"name":     name,
		"type":     "internal",
		"password": password,
	}

	reqBodyByte, _ := json.Marshal(reqBody)

	cg := config.NewConfiguration().DS
	url := fmt.Sprintf("http://%v:%v/api/user-management/v1/apps", cg.UserMgntHost, cg.UserMgntPort)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(reqBodyByte))
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if resp.StatusCode != 201 {
		err = fmt.Errorf("Get internal account failed, status code is %d", resp.StatusCode)
		return
	}

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var res map[string]string
	err = json.Unmarshal(body, &res)
	appID = res["id"]
	return
}

type DocLibPermMsg struct {
	AppID      string              `json:"app_id"`
	DocLibType string              `json:"doc_lib_type"`
	Expires    string              `json:"expires_at"`
	Perm       map[string][]string `json:"perm"`
}

type DocPermMsg struct {
	AppID   string              `json:"app_id"`
	DocID   string              `json:"doc_id"`
	Expires string              `json:"expires_at"`
	Perm    map[string][]string `json:"perm"`
}

func setAPPAccountPerm(clientID string) (err error) {
	// 内部账户配置权限
	doclibPermByte, _ := json.Marshal(DocLibPermMsg{
		AppID:      clientID,
		DocLibType: "all_doc_lib",
		Expires:    "1970-01-01T08:00:00+08:00",
		Perm: map[string][]string{
			"allow": {"read", "modify", "create", "delete"},
		},
	})

	docPermByte, _ := json.Marshal(DocPermMsg{
		DocID:   "gns://",
		AppID:   clientID,
		Expires: "1970-01-01T08:00:00+08:00",
		Perm: map[string][]string{
			"deny":  {},
			"allow": {"display", "preview", "download", "modify", "create", "delete"},
		},
	})

	msqClient := mqclient.ProtonMQClientImpl{}
	msqClient.Begin()
	err = msqClient.Pub(SetDoclibPermTopic, doclibPermByte)
	if err != nil {
		return
	}

	err = msqClient.Pub(SetDocPermTopic, docPermByte)
	if err != nil {
		return
	}
	return
}
