package middleware

import (
	"Hive"

	"github.com/kataras/iris/v12/context"
)

func NewCors() context.Handler {
	return func(ctx hive.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == "OPTIONS" {
			ctx.Header("Access-Control-Allow-Headers", "Access-Control-Allow-Origin,Content-Type")
			ctx.Header("Access-Control-Max-Age", "86400")
			ctx.StatusCode(204)
			return
		}
		ctx.Next()
	}
}
