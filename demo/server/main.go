package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/daemon-coder/idalloc/app/server"
	"github.com/daemon-coder/idalloc/definition"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	dsnPattern := "%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=%dms&readTimeout=%dms&writeTimeout=%dms&parseTime=true&loc=Local"
	dsn := fmt.Sprintf(
		dsnPattern,
		"user",
		"password",
		"127.0.0.1",
		3306,
		"idalloc",
		"utf8mb4",
		10000,
		10000,
		10000,
	)
	mysqlClient, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	if err = mysqlClient.Ping(); err != nil {
		panic(err)
	}

	idallocServer := server.NewServer(&definition.Config{
		AppName: "idalloc",
		Redis: redisClient,
		DB: mysqlClient,
		ServerPort: 8080,
		UsePprof: true,
		UsePrometheus: false,
		LogLevel: "INFO",
		RateLimit: definition.RateLimit{
			Enable: true,
			Qps:    100000,
		},
		SyncRedisAndDBChanSize:    10000,
		SyncRedisAndDBThreadNum:   10,
		RedisBatchAllocNum:        10000,
		WriteDBEveryNVersion:      10,
		RecoverRedisEveryNVersion: 100,
	})
	idallocServer.Run()
}
