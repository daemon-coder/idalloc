package iris_infra

import (
	"time"

	"github.com/daemon-coder/idalloc/definition"
	threadLocal "github.com/daemon-coder/idalloc/infrastructure/threadlocal_infra"
	"github.com/kataras/iris/v12/context"
)

type TransportHandler func(*context.Context) definition.Result

func JsonWrapper(transportHandler TransportHandler) context.Handler {
	return func(ctx *context.Context) {
		result := transportHandler(ctx)
		if result.TraceId == "" {
			result.TraceId = threadLocal.GetTraceId()
		}
		if result.Ts == 0 {
			result.Ts = time.Now().UnixMilli()
		}
		ctx.Header("Content-Type", "application/json; charset=utf-8")
		ctx.JSON(result)
	}
}
