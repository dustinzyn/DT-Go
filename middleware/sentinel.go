package middleware

import (
	"net/http"

	hive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/core/base"
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/utils"
	sentinel "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/rate/sentinel/api"
)

// NewSentinel returns new iris.HandlerFunc
// Default resource name is {method}:{path}, such as "GET:/api/users/{param1:int}"
// Default block fallback is returning 429 code
// Define your own behavior by setting options
func NewSentinel(opts ...Option) func(hive.Context) {
	options := evaluateOptions(opts)
	return func(ctx hive.Context) {
		resourceName := ctx.Method() + ":" + ctx.GetCurrentRoute().Path()

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
				ctx.StopWithJSON(http.StatusTooManyRequests, map[string]interface{}{
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
