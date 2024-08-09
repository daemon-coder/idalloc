package service

import (
	"context"
	"strconv"
	"sync"

	def "github.com/daemon-coder/idalloc/definition"
	"github.com/daemon-coder/idalloc/definition/entity"
	"github.com/daemon-coder/idalloc/definition/errors"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	threadLocal "github.com/daemon-coder/idalloc/infrastructure/threadlocal_infra"
	"github.com/daemon-coder/idalloc/repository"
)

type RedisAllocHandler struct {
	SyncRedisAndDBChan        chan *entity.AllocInfo
	redisToDBThreadNum        int
	writeDBEveryNVersion      int64
	recoverRedisEveryNVersion int64

	Stopped chan struct{}
	wg      *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

var DefaultRedisAllocHandler *RedisAllocHandler

func InitRedisAllocHandler(config *def.Config) *RedisAllocHandler {
	ctx, cancel := context.WithCancel(context.Background())
	DefaultRedisAllocHandler = &RedisAllocHandler{
		SyncRedisAndDBChan:        make(chan *entity.AllocInfo, config.SyncRedisAndDBChanSize),
		redisToDBThreadNum:        config.SyncRedisAndDBThreadNum,
		writeDBEveryNVersion:      config.WriteDBEveryNVersion,
		recoverRedisEveryNVersion: config.RecoverRedisEveryNVersion,
		Stopped:                   make(chan struct{}),
		wg:                        &sync.WaitGroup{},
		ctx:                       ctx,
		cancel:                    cancel,
	}
	return DefaultRedisAllocHandler
}

func (r *RedisAllocHandler) Start() {
	for i := 0; i < r.redisToDBThreadNum; i++ {
		r.wg.Add(1)
		go threadLocal.SetTraceIdWithCallBack("SyncRedisAndDB-"+strconv.Itoa(i), func() {
			log.GetLogger().Info("Start")
			defer log.GetLogger().Info("Stopped")
			defer r.wg.Done()

			for {
				select {
				case <-r.ctx.Done():
					for allocInfo := range r.SyncRedisAndDBChan {
						r.SyncRedisAndDB(allocInfo)
					}
					return
				case allocInfo := <-r.SyncRedisAndDBChan:
					if allocInfo == nil {
						continue
					}
					r.SyncRedisAndDB(allocInfo)
				}
			}
		})
	}
}

func (r *RedisAllocHandler) Shutdown() {
	log.GetLogger().Info("RedisAllocHandlerShutdownStart")
	close(r.SyncRedisAndDBChan)
	r.cancel()
	r.wg.Wait()
	close(r.Stopped)
	log.GetLogger().Info("RedisAllocHandlerShutdownFinish")
}

func (r *RedisAllocHandler) AllocWithoutPanic(serviceName string) (result *AllocResult, err error) {
	defer errors.PanicRecover(func(recoverErr errors.BaseError) {
		err = recoverErr
	})
	result = r.Alloc(serviceName)
	return
}

func (r *RedisAllocHandler) Alloc(serviceName string) *AllocResult {
	newAllocInfo := repository.RedisIncr(serviceName, def.RedisBatchAllocNum)
	// Synchronize the data changes in Redis to the database every 10 times.
	if r.NeedRecoverRedis(*newAllocInfo.DataVersion) || r.NeedWriteDB(*newAllocInfo.DataVersion) {
		r.SyncRedisAndDBChan <- newAllocInfo
	}
	return &AllocResult{
		LastAllocValue: *newAllocInfo.LastAllocValue - def.RedisBatchAllocNum,
		MaxValue:       *newAllocInfo.LastAllocValue,
	}
}

func (r *RedisAllocHandler) SyncRedisAndDB(allocInfo *entity.AllocInfo) {
	defer errors.PanicRecover(func(err errors.BaseError) {
		log.GetLogger().Warnw("SaveToDBPanic", "err", err)
	})

	if r.NeedRecoverRedis(*allocInfo.DataVersion) {
		r.RecoverRedisFromDB(*allocInfo.ServiceName)
	}

	if r.NeedWriteDB(*allocInfo.DataVersion) {
		repository.InsertOrUpdateAllocInfoToDB(allocInfo)
	}
}

func (r *RedisAllocHandler) RecoverRedisFromDB(serviceNames ...string) {
	var allocInfos []*entity.AllocInfo
	if len(serviceNames) == 0 {
		allocInfos = repository.GetAllFromDB()
	} else {
		allocInfos = repository.GetAllocInfoFromDB(serviceNames...)
	}
	for _, allocInfo := range allocInfos {
		repository.RedisCompareVersionAndSet(
			*allocInfo.ServiceName,
			*allocInfo.LastAllocValue,
			*allocInfo.DataVersion,
		)
	}
}

// NeedRecoverRedis: Synchronize the data from Redis to the database, and perform sampling checks to ensure
// that the Redis data version is not behind the database (to minimize the risk of data loss in Redis).
func (r *RedisAllocHandler) NeedWriteDB(version int64) bool {
	return version == 1 || version%r.writeDBEveryNVersion == 0
}

// NeedRecoverRedis: Synchronize the data from database to Redis
func (r *RedisAllocHandler) NeedRecoverRedis(version int64) bool {
	return version == 1 || version%r.recoverRedisEveryNVersion == 0
}
