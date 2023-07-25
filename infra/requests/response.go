package requests

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

// JSONResponse
type JSONResponse struct {
	Code        int         // 错误码 默认200不需要传递 Error不为空时也不需要传递
	Error       error       // *errors.Error
	contentType string      // Content-Type 默认application/json
	content     []byte      // 用于将response数据保存到context
	Object      interface{} // 返回的对象
}

// Dispatch overwrite response dispath
func (jrep JSONResponse) Dispatch(ctx *context.Context) {
	if jrep.contentType == "" {
		jrep.contentType = "application/json"
	}
	if jrep.Error != nil {
		repErr, ok := jrep.Error.(*errors.Error)
		if !ok {
			repErr = errors.InternalServerError(&errors.ErrorInfo{Cause: jrep.Error.Error()})
		}

		codeStr := strconv.Itoa(repErr.Code)
		code, _ := strconv.Atoi(codeStr[:3])
		ctx.Values().Set("code", code)
		jrep.content, _ = json.Marshal(repErr)
		ctx.Values().Set("response", string(jrep.content))
		ctx.StatusCode(code)
		ctx.JSON(iris.Map{
			"code":        repErr.Code,
			"message":     repErr.Message,
			"cause":       repErr.Cause,
			"detail":      repErr.Detail,
			"description": repErr.Description,
			"solution":    repErr.Solution,
		})
		ctx.StopExecution()
	} else {
		if jrep.Code != 0 {
			ctx.StatusCode(jrep.Code)
			ctx.Values().Set("code", jrep.Code)
		} else {
			ctx.Values().Set("code", http.StatusOK)
		}
		jrep.content, _ = json.Marshal(jrep.Object)
		ctx.Values().Set("response", string(jrep.content))
		if strings.HasPrefix(jrep.contentType, context.ContentJavascriptHeaderValue) {
			ctx.JSONP(jrep.Object)
		} else if strings.HasPrefix(jrep.contentType, context.ContentXMLHeaderValue) {
			ctx.XML(jrep.Object, context.XML{Indent: " "})
		} else {
			// defaults to json if content type is missing or its application/json.
			ctx.JSON(jrep.Object, context.JSON{Indent: " "})
		}
	}

	// hero.DispatchCommon(ctx, 200, jrep.contentType, jrep.content, nil, nil, true)
}
