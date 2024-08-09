package middleware

import (
	"net/http/httputil"
	"runtime/debug"

	"github.com/daemon-coder/idalloc/definition"
	e "github.com/daemon-coder/idalloc/definition/errors"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	"github.com/kataras/iris/v12"
)

func NewPanicRecoerMiddleware() iris.Handler {
	return func(ctx iris.Context) {
		defer e.PanicRecover(func(err e.BaseError) {
			httpRequest, _ := httputil.DumpRequest(ctx.Request(), false)
			log.LogError(err, "RecoveryFromPanic", "err", err, "request", string(httpRequest), "stack", string(debug.Stack()))

			errResponse := definition.NewResultFromError(err)
			ctx.Header("Content-Type", "application/json; charset=utf-8")
			ctx.JSON(errResponse)
		})
		ctx.Next()
	}
}
