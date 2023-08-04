package middleware

import (
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"

	uuid "github.com/iris-contrib/go.uuid"
	"github.com/kataras/iris/v12/context"
)

// NewTrace .
func NewTrace(traceIDName string) func(*context.Context) {
	return func(ctx *context.Context) {
		bus := hive.ToWorker(ctx).Bus()
		language := utils.ParseXLanguage(ctx.GetHeader("x-language"))
		bus.Add("language", language)
		traceID := bus.Get(traceIDName)
		for {
			if traceID != "" {
				break
			}
			uuidv4, e := uuid.NewV4()
			if e != nil {
				break
			}
			traceID = strings.ReplaceAll(uuidv4.String(), "-", "")
			bus.Add(traceIDName, traceID)
			break
		}
		ctx.Next()
	}
}

func init() {
}
