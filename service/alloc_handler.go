package service

import (
	"context"
	"sync"
	"time"

	e "github.com/daemon-coder/idalloc/definition/errors"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	threadLocal "github.com/daemon-coder/idalloc/infrastructure/threadlocal_infra"
	"github.com/daemon-coder/idalloc/repository"
)


var DefaultAllocHandler *AllocHandler

type AllocHandler struct {
	sync.Mutex
	ctx      context.Context
	wg       *sync.WaitGroup
	cancel   context.CancelFunc
	Stopped  chan struct{}
	handlers map[string]*ServiceAllocHandler
}

type ServiceAllocHandler struct {
	sync.Mutex
	ctx            context.Context
	wg             *sync.WaitGroup
	serviceName    string
	allocResult    *AllocResult
	AsyncAllocChan chan *AllocResult
}

type AllocResult struct {
	LastAllocValue int64 `json:"lastAllocValue"`
	MaxValue       int64 `json:"maxValue"`
}

func InitAllocHandler() *AllocHandler {
	ctx, cancel := context.WithCancel(context.Background())
	DefaultAllocHandler = &AllocHandler{
		wg:			&sync.WaitGroup{},
		ctx:		ctx,
		cancel:		cancel,
		Stopped:	make(chan struct{}),
		handlers:	make(map[string]*ServiceAllocHandler),
	}
	return DefaultAllocHandler
}

// Start: recover redis from db and init all service alloc handlers
func (a *AllocHandler) Start() {
	DefaultRedisAllocHandler.RecoverRedisFromDB()
	allocInfoList := repository.GetAllFromDB()
	for _, allocInfo := range allocInfoList {
		a.GetServiceAllocHandler(*allocInfo.ServiceName)
	}
}

// Shutdown: shutdown all alloc handlers
func (a *AllocHandler) Shutdown() {
	log.GetLogger().Info("AsyncAllocHandlerShutdownStart")
	a.cancel()
	a.wg.Wait()
	close(a.Stopped)
	log.GetLogger().Info("AsyncAllocHandlerShutdownFinish")
}

func (a *AllocHandler) Alloc(serviceName string, count int64) []int64 {
	serviceHandler := a.GetServiceAllocHandler(serviceName)
	return serviceHandler.Alloc(count)
}

func (a *AllocHandler) GetServiceAllocHandler(serviceName string) *ServiceAllocHandler {
	a.Lock()
	defer a.Unlock()

	handler, ok := a.handlers[serviceName]
	if ok {
		return handler
	}
	handler = a.NewServiceAllocHandler(serviceName)
	a.handlers[serviceName] = handler
	return handler
}

func (a *AllocHandler) NewServiceAllocHandler(serviceName string) *ServiceAllocHandler {
	result := &ServiceAllocHandler{
		ctx:			a.ctx,
		wg:				a.wg,
		serviceName:	serviceName,
		AsyncAllocChan:	make(chan *AllocResult),
	}
	result.allocResult = DefaultRedisAllocHandler.Alloc(serviceName)
	result.StartAsyncAlloc()
	return result
}

func (a *ServiceAllocHandler) Alloc(count int64) (result []int64) {
	result = make([]int64, 0, count)
	a.Lock()
	defer a.Unlock()

	targetValue := a.allocResult.LastAllocValue + count
	if targetValue <= a.allocResult.MaxValue {
		for i := a.allocResult.LastAllocValue + 1; i <= targetValue; i++ {
			result = append(result, i)
		}
		a.allocResult.LastAllocValue = targetValue
		return
	} else {
		// alloc from AsyncAllocChan
		var newAllocResult *AllocResult
		timeout := time.NewTimer(5 * time.Second)
		defer timeout.Stop()
		select {
		case newAllocResult = <-a.AsyncAllocChan:
			if newAllocResult == nil {
				e.Panic(e.NewServerError(e.WithMsg("ServiceStopped")))
			}
			log.GetLogger().Infow("AllocFromAsyncAllocChan", "allocResult", newAllocResult)
		case <-timeout.C:
			e.Panic(e.NewServerError(e.WithMsg("ServiceBusy")))
		}
		
		for i := a.allocResult.LastAllocValue + 1; i <= a.allocResult.MaxValue; i++ {
			result = append(result, i)
		}
		targetValue := count - int64(len(result)) + newAllocResult.LastAllocValue
		for i := newAllocResult.LastAllocValue + 1; i <= targetValue; i++ {
			result = append(result, i)
		}
		newAllocResult.LastAllocValue = targetValue
		a.allocResult = newAllocResult
		return
	}
}

func (a *ServiceAllocHandler) StartAsyncAlloc() {
	a.wg.Add(1)
	go threadLocal.SetTraceIdWithCallBack("AsyncAllocHandler-" + a.serviceName, func() {
		log.GetLogger().Info("Start")
		defer log.GetLogger().Info("Stopped")
		defer a.wg.Done()

		for {
			select {
			case <-a.ctx.Done():
				return
			default:
			}

			allocResult, err := DefaultRedisAllocHandler.AllocWithoutPanic(a.serviceName)
			if err != nil {
				continue
			}
			select {
			case <-a.ctx.Done():
				return
			case a.AsyncAllocChan <- allocResult:
			}
		}
	})
}
