// 提供带服务应用账户凭据的HTTPClient
// 需要用到服务注册的应用账户, 故需独立提供安装入口
package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/config"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/hivehttp"
)

// InitOauthHTTPClient .
func InitOauthHTTPClient(svcName string, conf config.DBConfiguration) {
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		MaxIdleConnsPerHost:   100,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{Transport: tr}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
	clientID, clientSecret := clientInfo(svcName, conf)
	credConf := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{},
		TokenURL:     tokenEndpoint(),
	}
	hivehttp.Oauth2HTTPClient = credConf.Client(ctx)
	return
}

func hydraPublicURL() url.URL {
	schema := GetEnv("HYDRA_PUBLIC_PROTOCOL", "http")
	host := GetEnv("HYDRA_PUBLIC_HOST", "hydra-public.anyshare.svc.cluster.local")
	port := GetEnv("HYDRA_PUBLIC_PORT", "4444")

	url := url.URL{
		Scheme: schema,
		Host:   fmt.Sprintf("%v:%v", host, port),
	}
	return url
}

func tokenEndpoint() string {
	url := hydraPublicURL()
	url.Path = "/oauth2/token"
	return url.String()
}

type AccountInfo struct {
	ClientID     string `gorm:"column:client_id"`
	ClientSecret string `gorm:"column:client_secret"`
}

// clientInfo return clientID ad client secret.
func clientInfo(svcName string, conf config.DBConfiguration) (clientID, secret string) {
	var result AccountInfo
	db := ConnectDB(&conf)
	err := db.Model(&Account{}).Where(&Account{Name: svcName}).Scan(&result).Error
	if err != nil {
		panic(err)
	}
	return result.ClientID, result.ClientSecret
}
