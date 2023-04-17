package requests

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/sync/singleflight"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/GoCommon/api"
)

var (
	// DefaultH2CClient .
	DefaultH2CClient *http.Client
	h2cclientGroup   singleflight.Group
	// DefaultHTTPClient .
	DefaultHTTPClient *http.Client
	httpclientGroup   singleflight.Group
	Oauth2HTTPClient  *http.Client
)

func init() {
	InitH2cClient(10 * time.Second)
	InitHTTPClient(10 * time.Second)
}

// InitHTTPClient .
func InitHTTPClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	t := 2 * time.Second
	if len(connectTimeout) > 0 {
		t = connectTimeout[0]
	}

	tran := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   t,
			KeepAlive: 15 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          512,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   100,
	}

	DefaultHTTPClient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

// InitOauthHTTPClient .
func InitOauthHTTPClient() {
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
	clientID, clientSecret := api.NewOAuth2().GetSelfClientInfo()
	credConf := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{},
		TokenURL:     getTokenEndpoint(),
	}
	Oauth2HTTPClient = credConf.Client(ctx)
	return
}

// InitH2cClient .
func InitH2cClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	tran := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			t := 2 * time.Second
			if len(connectTimeout) > 0 {
				t = connectTimeout[0]
			}
			fun := timeoutDialer(t)
			return fun(network, addr)
		},
	}

	DefaultH2CClient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

// timeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func timeoutDialer(cTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		return conn, err
	}
}

// InstallHTTPClient .
func InstallHTTPClient(client *http.Client) {
	DefaultHTTPClient = client
}

// InstallH2CClient .
func InstallH2CClient(client *http.Client) {
	DefaultH2CClient = client
}

func getHydraPublicURL() url.URL {
	schema := os.Getenv("HYDRA_PUBLIC_PROTOCOL")
	host := os.Getenv("HYDRA_PUBLIC_HOST")
	port := os.Getenv("HYDRA_PUBLIC_PORT")

	url := url.URL{
		Scheme: schema,
		Host:   fmt.Sprintf("%v:%v", host, port),
	}
	return url
}

func getTokenEndpoint() string {
	url := getHydraPublicURL()
	url.Path = "/oauth2/token"
	return url.String()
}
