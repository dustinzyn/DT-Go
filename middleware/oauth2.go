package middleware

import (
	"Hive"
	"Hive/errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/GoCommon/api"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

func NewAuthentication() context.Handler {
	return func(ctx hive.Context) {
		var err *api.Error
		oauth2 := api.NewOAuth2()
		token, err := parseBearerToken(ctx.Request())
		if err != nil {
			errorResponse(err, ctx)
			return
		}
		result, introErr := oauth2.Introspection(token, nil)
		if !result.Active {
			err = errors.UnauthorizationError(&api.ErrorInfo{Cause: "access token expired"})
			errorResponse(err, ctx)
			return
		}
		ctx.Values().Set("userId", result.Subject)
		ctx.Values().Set("userToken", token)
		ctx.Values().Set("ip", result.Extra["login_ip"])
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
		ctx.Values().Set("clientType", cType)
		ctx.Values().Set("udid", udid)
		ctx.Values().Set("visitorType", visitorType)
		if err != nil {
			err = errors.InternalServerError(&api.ErrorInfo{Cause: "introspection failed", Detail: map[string]string{"reason": introErr.Error()}})
			errorResponse(err, ctx)
			return
		}

		if result.ClientID != result.Subject {
			userRoles, csfLevel, name, roleErr := getUserRoles(result.Subject)
			if roleErr != nil {
				err = errors.InternalServerError(&api.ErrorInfo{Cause: "introspection failed", Detail: map[string]string{"reason": roleErr.Error()}})
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

// parseBearerToken 解析token
func parseBearerToken(req *http.Request) (token string, err *api.Error) {
	hdr := req.Header.Get("Authorization")
	if hdr == "" {
		err = errors.UnauthorizationError(&api.ErrorInfo{Cause: "access_token empty", Detail: map[string]string{"original_data": hdr}})
		return
	}

	// Example: Bearer xxxx
	tokenList := strings.SplitN(hdr, " ", 2)
	if len(tokenList) != 2 || strings.ToLower(tokenList[0]) != "bearer" {
		err = errors.UnauthorizationError(&api.ErrorInfo{Cause: "access_token invalid", Detail: map[string]string{"original_data": hdr}})
		return
	}
	return tokenList[1], nil
}

// errorResponse .
func errorResponse(err *api.Error, ctx hive.Context) {
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
