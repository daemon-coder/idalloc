package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/daemon-coder/idalloc/definition/entity"
	"github.com/daemon-coder/idalloc/definition/errors"
	redis "github.com/daemon-coder/idalloc/infrastructure/redis_infra"
	"github.com/daemon-coder/idalloc/util"
	goRedis "github.com/redis/go-redis/v9"
)

const (
	REDIS_KEY_PREFIX				= "idalloc:"
	ALLOC_INFO_KEY_PATTERN			= "alloc_info_%s"
	LAST_ALLOC_VALUE				= "lastAllocValue"
	DATA_VERSION					= "dataVersion"
	LOCK_KEY_PATTERN				= "lock_%s"
)

func GetAllocInfoRedisKey(serviceName string) string {
	return REDIS_KEY_PREFIX + fmt.Sprintf(ALLOC_INFO_KEY_PATTERN, serviceName)
}

func RedisIncr(serviceName string, increment int64) *entity.AllocInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	defer cancel()
	pipeline := redis.RedisClient.Pipeline()
	lastAllocValueCmd := pipeline.HIncrBy(ctx, GetAllocInfoRedisKey(serviceName), LAST_ALLOC_VALUE, increment)
	dataVersionCmd := pipeline.HIncrBy(ctx, GetAllocInfoRedisKey(serviceName), DATA_VERSION, 1)
	_, err := pipeline.Exec(ctx)
	if err != nil {
		errors.Panic(err)
	}
	lastAllocValue, err := lastAllocValueCmd.Result()
	if err != nil {
		errors.Panic(err)
	}
	dataVersion, err := dataVersionCmd.Result()
	if err != nil {
		errors.Panic(err)
	}
	return &entity.AllocInfo{
		ServiceName: util.Ptr(serviceName),
		LastAllocValue: util.Ptr(lastAllocValue),
		DataVersion: util.Ptr(dataVersion),
	}
}

func RedisSet(serviceName string, lastAllocValue, dataVersion int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	defer cancel()
	data := map[string]interface{} {
		LAST_ALLOC_VALUE: lastAllocValue,
		DATA_VERSION: dataVersion,
	}
	redisResult := redis.RedisClient.HMSet(ctx, GetAllocInfoRedisKey(serviceName), data)
	err := redisResult.Err()
	if err != nil {
		errors.Panic(err)
	}
}

func RedisCheckVersionAndSet(serviceName string, lastAllocValue, dataVersion int64) {
	// TODO
}

func RedisGet(serviceName string) *entity.AllocInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
	defer cancel()
	redisCmd := redis.RedisClient.HMGet(ctx, GetAllocInfoRedisKey(serviceName), LAST_ALLOC_VALUE, DATA_VERSION)
	values, err := redisCmd.Result()
	if err == goRedis.Nil {
		return nil
	} else if err != nil {
		errors.Panic(redisCmd.Err())
	} else if len(values) != 2 {
		errors.Panic(errors.NewCriticalError(errors.WithMsg("RedisAllocInfoDirty. serviceName:" + serviceName)))
	}
	if values[0] == nil || values[1] == nil {
		return nil
	}

	lastAllocValue, err := strconv.ParseInt(values[0].(string), 10, 64)
	if err != nil {
		msg := fmt.Sprintf("RedisAllocInfoDirty. serviceName:%s lastAllocValue:%s", serviceName, values[0])
		errors.Panic(errors.NewCriticalError(errors.WithMsg(msg)))
	}
	dataVersion, err := strconv.ParseInt(values[1].(string), 10, 64)
	if err != nil {
		msg := fmt.Sprintf("RedisAllocInfoDirty. serviceName:%s dataVersion:%s", serviceName, values[1])
		errors.Panic(errors.NewCriticalError(errors.WithMsg(msg)))
	}

	return &entity.AllocInfo{
		ServiceName: util.Ptr(serviceName),
		LastAllocValue: util.Ptr(lastAllocValue),
		DataVersion: util.Ptr(dataVersion),
	}
}

func GetLockRedisKey(key string) string {
	return REDIS_KEY_PREFIX + fmt.Sprintf(LOCK_KEY_PATTERN, key)
}

func WithRedisLock(key string, expire time.Duration, fn func()) {
	ctx, cancel := context.WithTimeout(context.Background(), expire)
	defer cancel()
	redisKey := GetLockRedisKey(key)
	redisResult := redis.RedisClient.SetNX(ctx, redisKey, 1, expire)
	ok, err := redisResult.Result()
	if err != nil {
		errors.Panic(err)
	}
	if !ok {
		errors.Panic(errors.NewBusinessError(errors.WithMsg("LockFailed. key:" + key)))
	}
	defer redis.RedisClient.Del(ctx, redisKey)
	fn()
}
