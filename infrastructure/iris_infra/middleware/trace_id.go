package middleware

import (
	"strings"

	threadLocal "github.com/daemon-coder/idalloc/infrastructure/threadlocal_infra"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12/context"
)

const TraceIDHeaderKey = "X-Trace-Id"

type Generator func(ctx *context.Context) string

var DefaultGenerator Generator = func(ctx *context.Context) string {
	id := ctx.ResponseWriter().Header().Get(TraceIDHeaderKey)
	if id != "" {
		return id
	}

	id = ctx.GetHeader(TraceIDHeaderKey)
	if id == "" {
		uid, err := uuid.NewRandom()
		if err != nil {
			ctx.StopWithStatus(500)
			return ""
		}
		id = strings.ReplaceAll(uid.String(), "-", "")
	}

	ctx.Header(TraceIDHeaderKey, id)
	return id
}

func NewTraceIdMiddleware(generators ...Generator) context.Handler {
	gen := DefaultGenerator
	if len(generators) > 0 {
		gen = generators[0]
	}

	return func(ctx *context.Context) {
		if threadLocal.GetTraceId() != "" {
			ctx.Next()
			return
		}

		id := gen(ctx)
		if ctx.IsStopped() {
			return
		}

		threadLocal.SetTraceIdWithCallBack(id, func() {
			ctx.Next()
		})
	}
}
