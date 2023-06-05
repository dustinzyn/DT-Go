// 请求 token 内省中间件
package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/hivehttp"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/internal"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
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
	return func(ctx hive.Context) {
		var err *errors.Error
		token, err := parseBearerToken(ctx.Request())
		if err != nil {
			errorResponse(err, ctx)
			return
		}
		result, introErr := introspection(token, nil)
		if !result.Active {
			err = errors.UnauthorizationError(&errors.ErrorInfo{Cause: "access token expired"})
			errorResponse(err, ctx)
			return
		}
		worker := ctx.Values().Get(internal.WorkerKey).(internal.Worker)
		worker.Bus().Add("userId", result.Subject)
		worker.Bus().Add("clientId", result.ClientID)
		worker.Bus().Add("userToken", token)
		worker.Bus().Add("ip", result.Extra["login_ip"].(string))
		var cType, udid, visitorType string
		if result.ClientID != result.Subject {
			if str, ok := result.Extra["client_type"].(string); ok {
				cType = str
			}
			if str, ok := result.Extra["udid"].(string); ok {
				udid = str
			}
			if v, ok := result.Extra["visitor_type"].(string); ok {
				switch v {
				case "realname":
					visitorType = "authenticated_user"
				case "anonymous":
					visitorType = "anonymous_user"
				}
			}
		}
		worker.Bus().Add("clientType", cType)
		worker.Bus().Add("udid", udid)
		worker.Bus().Add("visitorType", visitorType)
		if err != nil {
			err = errors.InternalServerError(&errors.ErrorInfo{Cause: "introspection failed", Detail: map[string]string{"reason": introErr.Error()}})
			errorResponse(err, ctx)
			return
		}
		ctx.Next()
	}
}

func getHydraAdminURL() url.URL {
	schema := utils.GetEnv("HYDRA_ADMIN_PROTOCOL", "http")
	host := utils.GetEnv("HYDRA_ADMIN_HOST", "hydra-admin.anyshare.svc.cluster.local")
	port := utils.GetEnv("HYDRA_ADMIN_PORT", "4445")
	url := url.URL{
		Scheme: schema,
		Host:   fmt.Sprintf("%v:%v", host, port),
	}
	return url
}

func getIntrospectEndpoint() string {
	url := getHydraAdminURL()
	url.Path = "/oauth2/introspect"
	return url.String()
}

// introspection token内审
func introspection(token string, scopes []string) (result Introspection, err error) {
	data := fmt.Sprintf("token=%s", token)
	if len(scopes) > 0 {
		data += fmt.Sprintf("&scope=%s", strings.Join(scopes, " "))
	}
	introspectEndpoint := getIntrospectEndpoint()
	req := hivehttp.NewHTTPRequest(introspectEndpoint)
	resp := req.Post().SetBody([]byte(data)).ToJSON(&result)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Introspection failed, status code is %d", resp.StatusCode)
		return
	}
	return
}

// parseBearerToken 解析token
func parseBearerToken(req *http.Request) (token string, err *errors.Error) {
	hdr := req.Header.Get("Authorization")
	if hdr == "" {
		err = errors.UnauthorizationError(&errors.ErrorInfo{Cause: "access_token empty", Detail: map[string]string{"original_data": hdr}})
		return
	}

	// Example: Bearer xxxx
	tokenList := strings.SplitN(hdr, " ", 2)
	if len(tokenList) != 2 || strings.ToLower(tokenList[0]) != "bearer" {
		err = errors.UnauthorizationError(&errors.ErrorInfo{Cause: "access_token invalid", Detail: map[string]string{"original_data": hdr}})
		return
	}
	return tokenList[1], nil
}

// errorResponse .
func errorResponse(err *errors.Error, ctx hive.Context) {
	codeStr := strconv.Itoa(err.Code)
	code, _ := strconv.Atoi(codeStr[:3])
	ctx.Values().Set("code", code)
	repByte, _ := json.Marshal(err)
	ctx.Values().Set("response", string(repByte))
	ctx.StatusCode(code)
	ctx.JSON(iris.Map{
		"code":    err.Code,
		"message": err.Message,
		"cause":   err.Cause,
		"detail":  err.Detail,
	})
	ctx.StopExecution()
}
