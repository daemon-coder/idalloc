package middleware

import (
	"net/http"

	"github.com/daemon-coder/idalloc/definition"
	e "github.com/daemon-coder/idalloc/definition/errors"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	"github.com/kataras/iris/v12"
	"golang.org/x/time/rate"
)

func NewRateLimitMiddleware() iris.Handler {
	qps := definition.Cfg.RateLimit.Qps
	limiter := rate.NewLimiter(rate.Limit(qps), qps)
	return func(ctx iris.Context) {
		if !definition.Cfg.RateLimit.Enable {
			ctx.Next()
		} else {
			if limiter.Allow() {
				ctx.Next()
			} else {
				log.GetLogger().Warnw("RateLimit", "qps", qps)
				result := definition.NewResultFromError(e.NewRateLimitError())
				ctx.StopWithJSON(http.StatusTooManyRequests, result)
			}
		}
	}
}
