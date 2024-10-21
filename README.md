# idalloc
English | [中文](README-ZH.md)
## 1. What is idalloc?
idalloc is a framework developed in Go that provides a server for generating trending incremental IDs.

## 2. Why do you need a distributed ID generation service?
In most scenarios, MySQL’s auto-increment primary keys can be used to generate IDs. However, there are certain cases where generating IDs outside of the database is necessary:
- IDs are needed before writing to the database for business processes.
- Data is stored in multiple databases. For example, when uploading a file, the metadata is stored in MySQL, and the file content is stored in object storage. An externally generated ID can unify both parts.
- IDs are shared across multiple systems, such as for distributed transaction IDs.
- In high-concurrency scenarios, MySQL's write limitations make it unsuitable for generating IDs at scale.

## 3. Comparison of industry ID generation solutions
- **UUID**:
  - Advantages: Locally generated, high performance.
  - Disadvantages: Risk of MAC address leakage, does not guarantee uniqueness in a cluster, long and unordered ID structure, occupies a lot of storage space.
- **MySQL auto-increment primary key or counter table**:
  - Advantages: Simple to implement.
  - Disadvantages: Limited write performance, unsuitable for high-concurrency scenarios.
- **Redis INCR command**:
  - Advantages: Higher performance than MySQL solutions.
  - Disadvantages: Risk of data loss, requires additional mechanisms for data persistence.
- **Snowflake**:
  - Advantages: Locally generated, high performance.
  - Disadvantages: Heavily reliant on system time, system clock rollback can cause catastrophic issues, relying on uncontrollable system time is poor design; each instance requires a unique identifier.
- **Meituan Leaf**:
  - Advantages: Similar to idalloc in concept, overall design is largely the same.
  - Disadvantages: Implemented in Java, which is bulkier.
- **idalloc**:
  - Advantages:
    - **Trending Incremental**: Generated IDs are similar to MySQL auto-increment primary keys, with trending incremental characteristics.
    - **High Performance**: Supports high concurrency, batch generation, and pre-generating IDs.
    - **High Availability**: Combines MySQL and Redis for ID storage, allowing short database outages without impacting operations. It also provides backup ranges of IDs in case of service failure (fallback to random numbers, a low-end version of UUID).
    - **Scalability**: Suitable for multiple business lines and large-scale applications.
    - **Lightweight**: Implemented in Go, consumes minimal resources, requiring only around 10+MB of memory under light load.
    - **Flexible**: Offers extensive configuration options to optimize for different scenarios.
  - Disadvantages:
    - Generated IDs are only guaranteed to trend upwards, not strictly sequential.
    - ID gaps may appear after a system restart, wasting some IDs.
    - It can generate up to 2^63 IDs, which may need consideration for long-term sufficiency.

## 4. idalloc Implementation Design
### 4.1. Architecture Overview
idalloc combines Redis and MySQL for ID storage and recovery. It also includes an in-process ID pool and asynchronous request channel to ensure fast ID generation and efficient synchronization. The following diagram shows the overall architecture of idalloc:
<p align="center">
<img src="https://github.com/daemon-coder/idalloc/blob/main/docs/images/idalloc-arch.png">
</p>

### 4.2. Batch Allocation
idalloc uses Redis's INCR command to request 10,000 IDs in bulk (this value is configurable), reducing frequent requests to Redis in high-concurrency scenarios and improving performance.

### 4.3. Pre-allocation Mechanism
To further optimize performance, idalloc implements a pre-allocation mechanism. When the available IDs in the current pool are nearly exhausted, an asynchronous thread triggers the next ID request in advance, blocking only during delivery.

Advantages:
- Prevents blocking Redis requests when the ID pool is exhausted, further enhancing performance.
Disadvantages:
- Pre-allocated ID ranges are invalidated if the process restarts, but this is an acceptable trade-off with no business impact.

### 4.4. Ensuring Atomicity of Database Updates
When syncing Redis with MySQL or recovering Redis data from MySQL, idalloc ensures atomicity to prevent concurrency conflicts. The storage format is as follows:

**MySQL Table Structure:**
```sql
CREATE TABLE IF NOT EXISTS `tbl_alloc_info` (
    `service_name`        VARCHAR(64)     NOT NULL PRIMARY KEY,
    `last_alloc_value`    BIGINT UNSIGNED NOT NULL DEFAULT '0',
    `data_version`        BIGINT UNSIGNED NOT NULL DEFAULT '0'
) ENGINE = InnoDB CHARACTER SET = utf8mb4;
```
**Redis Storage Format:**
- Key: `idalloc:alloc_info_$service_name`
- Type: `Hash`
- Fields: `lastAllocValue`, `dataVersion`

**Syncing Redis to MySQL**:
When updating MySQL, a version condition is added: the `data_version` to be updated must be greater than the current `data_version` in the database. This ensures concurrent updates do not overwrite newer data with older data.

**Restoring MySQL to Redis**:
A Lua script checks if the `data_version` in Redis is greater than the version in MySQL before updating Redis, ensuring that only newer versions are updated.

## 5. Usage Example
```go
package main

import (
        "github.com/daemon-coder/idalloc/app/server"
        "github.com/daemon-coder/idalloc/definition"
        "github.com/daemon-coder/idalloc/infrastructure/iris_infra/middleware"
        _ "github.com/go-sql-driver/mysql"
        db "myapp/infrastructure/db_infra"
        redis "myapp/infrastructure/redis_infra"
)

func main() {
        // TraceIDHeaderKey：调用idalloc服务时，传递的traceid的请求头，默认为：X-Trace-Id
        middleware.TraceIDHeaderKey = "X-Request-Id"
        idallocServer := server.NewServer(&definition.Config{
                AppName:       "idalloc",
                Redis:         redis.GetClient(),  // Redis连接
                DB:            db.GetDB(),         // MySQL连接
                ServerPort:    8080,
                UsePprof:      true,
                UsePrometheus: true,
                LogLevel:      "INFO",
                RateLimit: definition.RateLimit{
                        Enable: true,
                        Qps:    10000,
                },
                SyncRedisAndDBChanSize:     10000,  // 同步Redis和DB的任务队列大小
                SyncRedisAndDBThreadNum:    10,     // 同步Redis和DB的线程数
                RedisBatchAllocNum:         10000,  // Redis每次分配ID的数量
                WriteDBEveryNVersion:       10,     // Redis更新多少次，才会同步一次到MySQL
                RecoverRedisEveryNVersion:  100,    // Redis更新多少次，才会判断是否从MySQL中恢复到Redis
        })
        idallocServer.Run()
}
```