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
	redisClient.Ping(context.Background())

	dsnPattern := "%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=%dms&readTimeout=%dms&writeTimeout=%dms&parseTime=true&loc=Local"
	dsn := fmt.Sprintf(
		dsnPattern,
		"root",
		"root",
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
	mysqlClient.Ping()

	idallocServer := server.NewServer(&definition.Config{
		Redis: redisClient,
		DB:    mysqlClient,
		Server: definition.Server{
			Port:                  8080,
			UsePprof:              true,
			UsePrometheus:         true,
			PrometheusServiceName: "idalloc",
			LogLevel:              "info",
		},
		RateLimit: definition.RateLimit{
			Enable: true,
			Qps:    10000,
		},
		SyncRedisAndDBChanSize:    10000,
		SyncRedisAndDBThreadNum:   10,
		WriteDBEveryNVersion:      10,
		RecoverRedisEveryNVersion: 100,
	})
	idallocServer.Run()
}
