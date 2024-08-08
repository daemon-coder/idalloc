package middleware

import (
	"time"

	"github.com/kataras/iris/v12"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewAccessLogMiddleware() iris.Handler {
	return func(ctx iris.Context) {
		start := time.Now()
		defer func() {
			end := time.Now()
			latency := end.Sub(start)
			fields := []zapcore.Field{
				zap.Int("status", ctx.GetStatusCode()),
				zap.String("method", ctx.Method()),
				zap.String("path", ctx.Request().URL.Path),
				zap.String("query", ctx.Request().URL.RawQuery),
				zap.String("ua", ctx.Request().UserAgent()),
				zap.Duration("latency", latency),
			}

			if ctx.GetErr() != nil {
				fields = append(fields, zap.Error(ctx.GetErr()))
				log.Logger.Info("AccessError", fields...)
			} else {
				log.Logger.Info("AccessLog", fields...)
			}
		}()
		ctx.Next()
	}
}
