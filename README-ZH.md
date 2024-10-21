# idalloc
中文 | [English](README.md)
## 1. idalloc是什么
idalloc 是一个Go语言开发的框架，他提供了一个生成趋势递增id的server。

## 2. 为什么需要一个分布式ID生成服务？
在大多数场景下，我们可以用MySQL的自增主键生成id即可，但是有些特殊的场景是需要在数据库外部生成id：
- 写数据库之前就要用id去做一些业务处理。
- 数据存储于多种数据库中，例如文件上传时，文件元信息存MySQL，文件内容采用对象存储，可以外部生成一个id把两部分数据做统一。
- 跨多个系统共用的id，如分布式事务id。
- 在一些高并发的场景下，MySQL受限写速度的问题，不适用于高并发场景下的id生成。

## 3. 业界各种ID生成方案的对比
- UUID：
  - 优点：本地生成，性能高。
  - 缺点：存在MAC地址泄漏风险，集群环境中不能保证唯一性，ID长度较长且无序，占用较多存储空间。
- MySQL自增主键或计数表：
  - 优点：实现简单。
  - 缺点：写入性能有限，不适合高并发场景。
- Redis的INCR命令：
  - 优点：性能相对MySQL方案要高很多。
  - 缺点：存在数据丢失风险，需额外出方案保证数据的持久化。
- Snowflake：
  - 优点：本地生成，性能高。
  - 缺点：强依赖系统时间，一旦系统时间回调，会带来灾难性的影响，过于依赖不可控的系统时间不是好的设计；还需要为每个实例维护一份唯一编号。
- 美团Leaf：
  - 优点：与idalloc思路不谋而合，整体方案大同小异
  - 缺点：用Java实现，较为笨重。
- idalloc：
  - 优点：
    - 趋势递增：生成的ID类似MySQL的自增主键，具有趋势递增特性。
    - 高性能：支持高并发，批量生成、提前预生成ID。
    - 高可用性：ID存储基于MySQL和Redis组合，允许短时间的数据库宕机而不影响业务运行；同时预留了服务宕机后的备用号段（降级随机数的方案，可以理解为低配版UUID）。
    - 可扩展性强：支持多个业务线共用，满足大规模业务需求。
    - 轻量化：基于Go语言实现，资源占用少，轻负载时只需要10+MB左右的内存。
    - 灵活性高：提供丰富的配置选项，可以根据不同场景进行调优。
  - 缺点：
    - 生成的id只能保证趋势递增，不能保证严格递增。
    - 在系统重启时，会出现id段空洞的现象，浪费部分id。
    - 最多生成2^63个id，引入时需要考虑是否足够。

## 4. idalloc的实现方案
### 4.1. 整体架构
idalloc的架构结合了Redis和MySQL进行ID的存储与恢复，同时包含了一个进程内ID池和异步请求通道。通过这些机制保证了ID的快速生成与高效同步。下图展示了idalloc的整体架构：
<p align="center">
<img src="https://github.com/daemon-coder/idalloc/blob/main/docs/images/idalloc-arch.png?raw=true">
</p>

### 4.2. 批量申请
idalloc通过Redis的INCR命令每次批量申请1万个ID（该值可配置），从而减少高并发场景下对Redis的频繁请求，提升性能。
### 4.3. 预申请机制
为了进一步优化性能，idalloc实现了预申请机制。当现有ID池中的ID快要耗尽时，异步申请线程会立即触发下一轮的ID申请，并在交付时阻塞。
优点：
- 避免了ID池耗尽时阻塞Redis请求的问题，进一步提升了性能。
缺点：
- 如果进程重启，预申请的ID号段会作废，但这对业务没有影响，属于可接受的缺陷。
### 4.4. 数据库更新的原子性保障
在Redis同步MySQL或从MySQL恢复Redis的数据时，idalloc通过保证操作的原子性来避免并发冲突。具体的存储格式如下：
**MySQL表结构：**
```sql
CREATE TABLE IF NOT EXISTS `tbl_alloc_info` (
    `service_name`        VARCHAR(64)     NOT NULL PRIMARY KEY,
    `last_alloc_value`    BIGINT UNSIGNED NOT NULL DEFAULT '0',
    `data_version`        BIGINT UNSIGNED NOT NULL DEFAULT '0'
) ENGINE = InnoDB CHARACTER SET = utf8mb4;
```
**Redis存储格式：**
- 键：`idalloc:alloc_info_$service_name`
- 类型：`Hash`
- 字段：`lastAllocValue`, `dataVersion`

**Redis同步到MySQL：**
更新MySQL时，会判断添加版本筛选条件：将要更新的data_version>数据库里的data_version，从而保证并发更新也不会出现旧数据覆盖新数据的情况。

**MySQL恢复到Redis：**
通过Lua脚本执行以下命令：判断Redis中的data_version是否大于mysql中的，是则更新Redis。从而原子性在保证了数据只会更新为更新的版本。

## 5. 使用示例
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
