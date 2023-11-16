package middleware

import (
	"net/http"

	dt "DT-Go"
	"DT-Go/errors"
	sentinel "DT-Go/infra/rate/sentinel/api"
	"DT-Go/infra/rate/sentinel/core/base"
	"DT-Go/utils"
)

// NewSentinel returns new iris.HandlerFunc
// Default resource name is {method}:{path}, such as "GET:/api/users/{param1:int}"
// Default block fallback is returning 429 code
// Define your own behavior by setting options
func NewSentinel(opts ...Option) func(dt.Context) {
	options := evaluateOptions(opts)
	return func(ctx dt.Context) {
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
