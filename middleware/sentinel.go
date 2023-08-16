package middleware

import (
	"net/http"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

// NewSentinel returns new iris.HandlerFunc
// Default resource name is {method}:{path}, such as "GET:/api/users/:id"
// Default block fallback is returning 429 code
// Define your own behavior by setting options
func NewSentinel(opts ...Option) func(*context.Context) {
	options := evaluateOptions(opts)
	return func(ctx hive.Context) {
		resourceName := ctx.Method() + ":" + ctx.RequestPath(true)

		if options.resourceExtract != nil {
			resourceName = options.resourceExtract(ctx)
		}

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeWeb),
			sentinel.WithTrafficType(base.Inbound),
		)

		if err != nil {
			if options.blockFallback != nil {
				options.blockFallback(ctx)
			} else {
				language := utils.ParseXLanguage(ctx.GetHeader("x-language"))
				err := errors.New(language, errors.TooManyRequestsErr, "", nil)
				ctx.StopWithJSON(http.StatusTooManyRequests, iris.Map{
					"code":        err.Code(),
					"message":     err.Message(),
					"cause":       err.Cause(),
					"detail":      err.Detail(),
					"description": err.Description(),
					"solution":    err.Solution(),
				})
			}
			return
		}

		defer entry.Exit()
		ctx.Next()
	}
}
