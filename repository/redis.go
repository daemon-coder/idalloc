package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/daemon-coder/idalloc/definition"
	"github.com/daemon-coder/idalloc/definition/entity"
	"github.com/daemon-coder/idalloc/definition/errors"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	redis "github.com/daemon-coder/idalloc/infrastructure/redis_infra"
	"github.com/daemon-coder/idalloc/util"
	goRedis "github.com/redis/go-redis/v9"
)

const (
	ALLOC_INFO_KEY_PATTERN			= "alloc_info_%s"
	LAST_ALLOC_VALUE				= "lastAllocValue"
	DATA_VERSION					= "dataVersion"
	LOCK_KEY_PATTERN				= "lock_%s"
)

func GetAllocInfoRedisKey(serviceName string) string {
	return definition.RedisKeyPrefix + fmt.Sprintf(ALLOC_INFO_KEY_PATTERN, serviceName)
}

func RedisIncr(serviceName string, increment int64) *entity.AllocInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
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
	log.GetLogger().Infow("RedisIncr", "serviceName", serviceName, "increment", increment, "lastAllocValue", lastAllocValue, "dataVersion", dataVersion)
	return &entity.AllocInfo{
		ServiceName: util.Ptr(serviceName),
		LastAllocValue: util.Ptr(lastAllocValue),
		DataVersion: util.Ptr(dataVersion),
	}
}

// RedisCompareVersionAndSetCmd compare the data version in redis.
// if the version is behind the given version, set the given data to redis
//
// Output:
// Returns lastAllocValue after all operations
// Returns dataVersion after all operations
var RedisCompareVersionAndSetCmd = goRedis.NewScript(`
local key = KEYS[1]
local valueField = KEYS[2]
local versionField = KEYS[3]
local inputValue = tonumber(ARGV[1])
local inputVersion = tonumber(ARGV[2])

local values = redis.call("HMGET", key, valueField, versionField)
local valueInRedis = tonumber(values[1])
local versionInRedis = tonumber(values[2])
if versionInRedis == nil or valueInRedis == nil or versionInRedis < inputVersion then
	redis.call("HMSET", key, valueField, inputValue, versionField, inputVersion)
	return {inputValue, inputVersion}
end
return {valueInRedis, versionInRedis}
`)

func RedisCompareVersionAndSet(serviceName string, lastAllocValue, dataVersion int64) (curLastAllocValue int64, curDataVersion int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	keys := []string{
		GetAllocInfoRedisKey(serviceName),
		LAST_ALLOC_VALUE,
		DATA_VERSION,
	}
	argv := []interface{}{
		lastAllocValue,
		dataVersion,
	}
	result, err := RedisCompareVersionAndSetCmd.Run(ctx, redis.RedisClient, keys, argv...).Result()
	if err != nil {
		errors.Panic(err)
	}
	curLastAllocValue = result.([]interface{})[0].(int64)
	curDataVersion = result.([]interface{})[1].(int64)
	log.GetLogger().Infow(
		"RedisCompareVersionAndSet",
		"serviceName", serviceName,
		"valueInDB", lastAllocValue,
		"versionInDB", dataVersion,
		"curLastAllocValue", curLastAllocValue,
		"curDataVersion", curDataVersion,
	)
	return
}

func RedisSet(serviceName string, lastAllocValue, dataVersion int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
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

func RedisGet(serviceName string) *entity.AllocInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
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
	return definition.RedisKeyPrefix + fmt.Sprintf(LOCK_KEY_PATTERN, key)
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
