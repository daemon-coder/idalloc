package server

import (
	"context"
	"time"

	def "github.com/daemon-coder/idalloc/definition"
	e "github.com/daemon-coder/idalloc/definition/errors"
)

func CheckConfig(config *def.Config) {
	if config.AppName == "" {
		config.AppName = def.DEFAULT_APP_NAME
	}

	if config.ServerPort <= 0 {
		config.ServerPort = def.DEFAULT_SERVER_PORT
	}

	if config.LogLevel == "" {
		config.LogLevel = def.DEFAULT_LOG_LEVEL
	}

	if config.Redis == nil {
		e.Panic(e.NewCriticalError(e.WithMsg("config invalid. redis is nil")))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	cmd := config.Redis.Ping(ctx)
	if cmd.Err() != nil {
		e.Panic(e.NewCriticalError(e.WithMsg("config invalid. redis ping failed")))
	}

	if config.DB == nil {
		e.Panic(e.NewCriticalError(e.WithMsg("config invalid. db is nil")))
	}
	err := config.DB.Ping()
	if err != nil {
		e.Panic(e.NewCriticalError(e.WithMsg("config invalid. db ping failed")))
	}

	if config.RedisKeyPrefix != "" {
		def.RedisKeyPrefix = config.RedisKeyPrefix
	}

	if config.SyncRedisAndDBChanSize <= 0 {
		config.SyncRedisAndDBChanSize = def.DEFAULT_SYNC_REDIS_AND_DB_CHAN_SIZE
	}

	if config.SyncRedisAndDBThreadNum <= 0 {
		config.SyncRedisAndDBThreadNum = def.DEFAULT_SYNC_REDIS_AND_DB_THREAD_NUM
	}

	if config.RedisBatchAllocNum < def.MAX_USER_BATCH_ALLOC_NUM {
		config.RedisBatchAllocNum = def.DEFAULT_REDIS_BATCH_ALLOC_NUM
	}
	def.RedisBatchAllocNum = config.RedisBatchAllocNum

	if config.WriteDBEveryNVersion <= 0 {
		config.WriteDBEveryNVersion = def.DEFAULT_WRITE_DB_EVERY_N_VERSION
	}

	if config.RecoverRedisEveryNVersion <= 0 {
		config.RecoverRedisEveryNVersion = def.DEFAULT_RECOVER_REDIS_EVERY_N_VERSION
	}
}
