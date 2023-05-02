package middleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/requests"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/internal"
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
		result, introErr := introspection(token, []string{"all"})
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

		if result.ClientID != result.Subject {
			userRoles, csfLevel, name, roleErr := getUserRoles(result.Subject)
			if roleErr != nil {
				err = errors.InternalServerError(&errors.ErrorInfo{Cause: "introspection failed", Detail: map[string]string{"reason": roleErr.Error()}})
				errorResponse(err, ctx)
				return
			}
			ctx.Values().Set("userRoles", userRoles)
			ctx.Values().Set("csfLevel", csfLevel)
			ctx.Values().Set("name", name)
		} else {
			return
		}
		ctx.Next()
	}
}

func getHydraAdminURL() url.URL {
	schema := os.Getenv("HYDRA_ADMIN_PROTOCOL")
	host := os.Getenv("HYDRA_ADMIN_HOST")
	port := os.Getenv("HYDRA_ADMIN_PORT")
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

func introspection(token string, scopes []string) (result Introspection, err error) {
	data := fmt.Sprintf("token=%s", token)
	if len(scopes) > 0 {
		data += fmt.Sprintf("&scope=%s", strings.Join(scopes, " "))
	}
	introspectEndpoint := getIntrospectEndpoint()
	req := requests.NewHTTPRequest(introspectEndpoint)
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

func getUserMgntPrivateURL() url.URL {
	schema := os.Getenv("USER_MANAGEMENT_PRIVATE_PROTOCOL")
	host := os.Getenv("USER_MANAGEMENT_PRIVATE_HOST")
	port := os.Getenv("USER_MANAGEMENT_PRIVATE_PORT")

	url := url.URL{
		Scheme: schema,
		Host:   fmt.Sprintf("%v:%v", host, port),
	}
	return url
}

func getOwnersEndpoint(userId string) string {
	url := getUserMgntPrivateURL()
	url.Path = fmt.Sprintf("/api/user-management/v1/users/%v/roles,csf_level,name", userId)
	return url.String()
}

// getUserRoles 获取用户角色、密级、显示名
func getUserRoles(userId string) (userRoles []string, csfLevel float64, name string, err error) {
	var resp *http.Response
	var respBodyByte []byte
	var respData map[string]interface{}
	url := getOwnersEndpoint(userId)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("ERROR: GetUserRoles: Response = %v", resp.StatusCode)
		return
	}
	respBodyByte, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(respBodyByte, &respData)
	if err != nil {
		return
	}
	for _, v := range respData["roles"].([]interface{}) {
		userRoles = append(userRoles, v.(string))
	}
	csfLevel = respData["csf_level"].(float64)
	name = respData["name"].(string)
	return
}
