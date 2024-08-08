package definition

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

var Cfg *Config

type Config struct {
	Redis                     *redis.Client
	DB                        *sql.DB
	Server                    Server
	RateLimit                 RateLimit
	SyncRedisAndDBChanSize    int
	SyncRedisAndDBThreadNum   int
	WriteDBEveryNVersion      int64
	RecoverRedisEveryNVersion int64
}

type Server struct {
	Port                  int
	UsePprof              bool
	UsePrometheus         bool
	PrometheusServiceName string
	LogLevel              string
}

type RateLimit struct {
	Enable bool
	Qps    int
}
