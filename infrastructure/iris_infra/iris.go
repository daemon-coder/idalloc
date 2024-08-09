package iris_infra

import (
	"context"
	"strconv"
	"time"

	"github.com/daemon-coder/idalloc/definition"
	"github.com/daemon-coder/idalloc/infrastructure/iris_infra/middleware"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/pprof"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type IrisApp struct {
	*iris.Application
	Stopped chan struct{}
}

func NewIrisApp(cfg *definition.Config, addRouteFn func(*IrisApp)) *IrisApp {
	app := &IrisApp{
		Application: iris.New(),
		Stopped:     make(chan struct{}),
	}
	app.Use(recover.New())
	app.Use(middleware.NewTraceIdMiddleware())
	app.Use(middleware.NewAccessLogMiddleware())
	app.Use(middleware.NewPanicRecoerMiddleware())
	app.Use(middleware.NewRateLimitMiddleware())

	addRouteFn(app)

	// pprof
	if cfg.UsePprof {
		pprofHandler := pprof.New()
		app.Any("/debug/pprof", pprofHandler)
		app.Any("/debug/pprof/{action:path}", pprofHandler)
	}

	// prometheus
	if cfg.UsePrometheus {
		m := middleware.NewPrometheusMiddleware(cfg.AppName)
		app.Use(m.ServeHTTP)
		app.Get("/metrics", iris.FromStd(promhttp.Handler()))
	}
	return app
}

func (app *IrisApp) Start(cfg *definition.Config) {
	go func() {
		defer close(app.Stopped)
		err := app.Run(
			iris.Addr(":"+strconv.Itoa(cfg.ServerPort)),
			iris.WithConfiguration(iris.Configuration{
				DisableInterruptHandler:           false,
				DisablePathCorrection:             false,
				EnablePathEscape:                  false,
				FireMethodNotAllowed:              false,
				DisableBodyConsumptionOnUnmarshal: false,
				DisableAutoFireStatusCode:         false,
				TimeFormat:                        time.DateTime,
				Charset:                           "UTF-8",
			}),
		)
		log.GetLogger().Infow("IrisShutDownFinish", "err", err)
	}()
}

func (app *IrisApp) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.GetLogger().Info("IrisShutdownGracefully")
	app.Application.Shutdown(ctx)
}
