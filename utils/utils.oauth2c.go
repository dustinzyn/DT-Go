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

	"DT-Go/config"
	"DT-Go/infra/dhttp"
)

// InitOauthHTTPClient .
func InitOauthHTTPClient(svcName string, conf config.Configurations) {
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
	dhttp.Oauth2HTTPClient = credConf.Client(ctx)
	return
}

func tokenEndpoint() string {
	cg := config.NewConfiguration().DS
	url := url.URL{
		Scheme: cg.HydraPublicProtocol,
		Host:   fmt.Sprintf("%v:%v", cg.HydraPublicHost, cg.HydraPublicPort),
	}
	url.Path = "/oauth2/token"
	return url.String()
}

type AccountInfo struct {
	ClientID     string
	ClientSecret string
}

// clientInfo return clientID ad client secret.
func clientInfo(svcName string, conf config.Configurations) (clientID, secret string) {
	var result AccountInfo
	cgdb := config.NewConfiguration().DB
	db := ConnectDB(conf.DB)
	table := fmt.Sprintf("%s.account", cgdb.DBName)
	rows, err := db.Table(table).Select("client_id, client_secret").Where("name = ?", svcName).Rows()
	defer CloseRows(rows)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		if err := rows.Scan(&result.ClientID, &result.ClientSecret); err != nil {
			panic(err)
		}
	}
	return result.ClientID, result.ClientSecret
}
