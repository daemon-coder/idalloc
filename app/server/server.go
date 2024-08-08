package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daemon-coder/idalloc/definition"
	db "github.com/daemon-coder/idalloc/infrastructure/db_infra"
	iris "github.com/daemon-coder/idalloc/infrastructure/iris_infra"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	redis "github.com/daemon-coder/idalloc/infrastructure/redis_infra"
	"github.com/daemon-coder/idalloc/service"
)

type Server struct {
	Config  *definition.Config
	Context context.Context
	Cancel  context.CancelFunc
	Stopped chan struct{}

	IrisApp           *iris.IrisApp
	AllocHandler      *service.AllocHandler
	RedisAllocHandler *service.RedisAllocHandler
}

func NewServer(config *definition.Config) *Server {
	// TODO validate config
	definition.Cfg = config
	db.DBClient = config.DB
	redis.RedisClient = config.Redis

	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Config:  config,
		Context: ctx,
		Cancel:  cancel,
		Stopped: make(chan struct{}),

		IrisApp:           iris.NewIrisApp(config, AddRoute),
		AllocHandler:      service.InitAllocHandler(),
		RedisAllocHandler: service.InitRedisAllocHandler(config),
	}
}

func (s *Server) Run() {
	s.RedisAllocHandler.Start()
	s.AllocHandler.Start()
	s.IrisApp.Start(s.Config)

	s.HandleSignal()
	s.ShutdownWait(20 * time.Second)
}

func (s *Server) ShutdownWait(wait time.Duration) {
	go s.Shutdown()

	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-s.Stopped:
	}
}

func (s *Server) Shutdown() {
	// the order is important, the iris should be shutdown first
	s.IrisApp.Shutdown()
	<-s.IrisApp.Stopped
	s.AllocHandler.Shutdown()
	s.RedisAllocHandler.Shutdown()

	close(s.Stopped)
}

func (s *Server) HandleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	for {
		sig := <-signalChan
		log.GetLogger().Infow("ReceiveSignal", "signal", sig.String())
		switch sig {
		case syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2:
			log.GetLogger().Infow("IgnoreSignal", "signal", sig.String())
			continue
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			s.Cancel()
			return
		}
	}
}
