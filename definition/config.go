package definition

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

var Cfg *Config

type Config struct {
	AppName       string
	ServerPort    int
	LogLevel      string
	UsePprof      bool
	UsePrometheus bool
	RateLimit     RateLimit

	DB                        *sql.DB
	Redis                     *redis.Client
	RedisKeyPrefix            string
	SyncRedisAndDBChanSize    int
	SyncRedisAndDBThreadNum   int
	RedisBatchAllocNum        int64
	WriteDBEveryNVersion      int64
	RecoverRedisEveryNVersion int64
}

type RateLimit struct {
	Enable bool
	Qps    int
}

const (
	DEFAULT_APP_NAME                      = "idalloc"
	DEFAULT_SERVER_PORT                   = 8080
	DEFAULT_LOG_LEVEL                     = "INFO"
	DEFAULT_SYNC_REDIS_AND_DB_CHAN_SIZE   = 10000
	DEFAULT_SYNC_REDIS_AND_DB_THREAD_NUM  = 10
	DEFAULT_REDIS_BATCH_ALLOC_NUM         = 10000
	DEFAULT_WRITE_DB_EVERY_N_VERSION      = 10
	DEFAULT_RECOVER_REDIS_EVERY_N_VERSION = 100
)

const (
	MAX_USER_BATCH_ALLOC_NUM = 100
	DEFAULT_USER_ALLOC_NUM   = 1
)

var (
	RedisKeyPrefix     string = DEFAULT_APP_NAME + ":"
	RedisBatchAllocNum int64  = DEFAULT_REDIS_BATCH_ALLOC_NUM
)
