package infra

import (
	"encoding/json"
	"strconv"
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/GoCommon/api"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
)

// JSONResponse
type JSONResponse struct {
	Error       error
	contentType string
	content     []byte
	Object      interface{}
}

// Dispatch overwrite response dispath
func (jrep JSONResponse) Dispatch(ctx context.Context) {
	if jrep.contentType == "" {
		jrep.contentType = "application/json"
	}
	if jrep.Error != nil {
		repErr, ok := jrep.Error.(*api.Error)
		if !ok {
			repErr = errors.InternalServerError(&api.ErrorInfo{Cause: jrep.Error.Error()})
		}

		codeStr := strconv.Itoa(repErr.Code)
		code, _ := strconv.Atoi(codeStr[:3])
		ctx.Values().Set("code", code)
		jrep.content, _ = json.Marshal(repErr)
		ctx.Values().Set("response", string(jrep.content))
		ctx.StatusCode(code)
		ctx.JSON(iris.Map{
			"code":    repErr.Code,
			"message": repErr.Message,
			"cause":   repErr.Cause,
			"detail":  repErr.Detail,
		})
		ctx.StopExecution()
	} else {
		ctx.Values().Set("code", 200)
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
