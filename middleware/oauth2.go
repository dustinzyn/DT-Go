// 请求 token 内省中间件
package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	dt "DT-Go"
	"DT-Go/errors"
	"DT-Go/infra/dhttp"
	"DT-Go/internal"
	"DT-Go/utils"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

type Introspection struct {
	Active            bool                   `json:"active"`                       // Active is a boolean indicator of whether or not the presented token is currently active.
	Audience          []string               `json:"aud,omitempty"`                // Audience contains a list of the token's intended audiences.
	ClientID          string                 `json:"client_id,omitempty"`          // ClientID is aclient identifier for the OAuth 2.0 client that requested this token.
	ExpiresAt         int64                  `json:"exp,omitempty"`                // Expires at is an integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token will expire.
	Extra             map[string]interface{} `json:"ext,omitempty"`                // Extra is arbitrary data set by the session.
	IssuedAt          int64                  `json:"iat,omitempty"`                // Issued at is an integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token was originally issued.
	IssuerURL         string                 `json:"iss,omitempty"`                // IssuerURL is a string representing the issuer of this token
	NotBefore         int64                  `json:"nbf,omitempty"`                // NotBefore is an integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token is not to be used before.
	ObfuscatedSubject string                 `json:"obfuscated_subject,omitempty"` // ObfuscatedSubject is set when the subject identifier algorithm was set to "pairwise" during authorization. It is the `sub` value of the ID Token that was issued.
	Scope             string                 `json:"scope,omitempty"`              // Scope is a JSON string containing a space-separated list of scopes associated with this token.
	Subject           string                 `json:"sub,omitempty"`                // Subject of the token, as defined in JWT [RFC7519]. Usually a machine-readable identifier of the resource owner who authorized this token.
	TokenType         string                 `json:"token_type,omitempty"`         // TokenType is the introspected token's type, for example `access_token` or `refresh_token`.
	Username          string                 `json:"username,omitempty"`           // Username is a human-readable identifier for the resource owner who authorized this token.
}

func NewAuthentication() context.Handler {
	return func(ctx dt.Context) {
		language := utils.ParseXLanguage(ctx.GetHeader("x-language"))
		token, err := parseBearerToken(ctx.Request())
		if err != nil {
			errorResponse(err, ctx)
			return
		}
		result, introErr := introspection(token, nil)
		if !result.Active {
			err = errors.New(language, errors.UnauthorizedErr, "access token expired.", nil)
			errorResponse(err, ctx)
			return
		}
		if introErr != nil {
			err = errors.New(language, errors.InternalErr, "introspection failed.", map[string]string{"reason": introErr.Error()})
			errorResponse(err, ctx)
			return
		}
		worker := ctx.Values().Get(internal.WorkerKey).(internal.Worker)
		worker.Bus().Add("user_id", result.Subject)
		worker.Bus().Add("client_id", result.ClientID)
		worker.Bus().Add("bearer_token", token)
		worker.Bus().Add("language", language)
		var accountType string
		if result.Extra != nil {
			if str, ok := result.Extra["login_ip"].(string); ok {
				worker.Bus().Add("ip", str)
			}
			var cType, udid, visitorType string
			if result.ClientID != result.Subject {
				if str, ok := result.Extra["client_type"].(string); ok {
					cType = str
				}
				if str, ok := result.Extra["udid"].(string); ok {
					udid = str
				}
				if v, ok := result.Extra["visitor_type"].(string); ok {
					// realname anonymous
					visitorType = v
				}
				accountType = "user"
			}
			worker.Bus().Add("client_type", cType)
			worker.Bus().Add("udid", udid)
			worker.Bus().Add("visitor_type", visitorType)
		}
		worker.Bus().Add("account_type", accountType)
		ctx.Next()
	}
}

func getIntrospectEndpoint() string {
	cg := dt.NewConfiguration()
	url := url.URL{
		Scheme: cg.DS.HydraAdminProtocol,
		Host:   fmt.Sprintf("%v:%v", cg.DS.HydraAdminHost, cg.DS.HydraAdminPort),
	}
	url.Path = "/admin/oauth2/introspect"
	return url.String()
}

// introspection token内审
func introspection(token string, scopes []string) (result Introspection, err error) {
	data := fmt.Sprintf("token=%s", token)
	if len(scopes) > 0 {
		data += fmt.Sprintf("&scope=%s", strings.Join(scopes, " "))
	}
	introspectEndpoint := getIntrospectEndpoint()
	req := dhttp.NewHTTPRequest(introspectEndpoint)
	resp := req.Post().SetBody([]byte(data)).ToJSON(&result)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Introspection failed, status code is %d", resp.StatusCode)
		return
	}
	return
}

// parseBearerToken 解析token
func parseBearerToken(req *http.Request) (token string, err *errors.ErrorResp) {
	hdr := req.Header.Get("Authorization")
	if hdr == "" {
		err = errors.New(utils.ParseXLanguage(req.Header.Get("x-language")), errors.UnauthorizedErr, "access token is empty.", map[string]string{"original_data": hdr})
		return
	}

	// Example: Bearer xxxx
	tokenList := strings.SplitN(hdr, " ", 2)
	if len(tokenList) != 2 || strings.ToLower(tokenList[0]) != "bearer" {
		err = errors.New(utils.ParseXLanguage(req.Header.Get("x-language")), errors.UnauthorizedErr, "invalid token.", map[string]string{"original_data": hdr})
		return
	}
	return tokenList[1], nil
}

// errorResponse .
func errorResponse(err errors.APIError, ctx dt.Context) {
	codeStr := strconv.Itoa(err.Code())
	code, _ := strconv.Atoi(codeStr[:3])
	ctx.Values().Set("code", code)
	ctx.Values().Set("response", string(err.Marshal()))
	ctx.StatusCode(code)
	ctx.JSON(iris.Map{
		"code":        err.Code(),
		"message":     err.Message(),
		"cause":       err.Cause(),
		"detail":      err.Detail(),
		"description": err.Description(),
		"solution":    err.Solution(),
	})
	ctx.StopExecution()
}
