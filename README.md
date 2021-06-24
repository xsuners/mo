##  MO

```
   ____ ___   ____
  / __ `__ \ / __ \
 / / / / / // /_/ /
/_/ /_/ /_/ \____/

```

### Features
- 模块化
- 自动化
- 一致性



### 微服务框架设计


**框架结构**
```
metadata
config
cli
database
    mongo
    sql
    tidb
    dgraph
cache
    memcache
    badger
    redis
timer
    wheel
    rate
net
    ws      // WS
    grpc    // RPC
    aero    // HTTP
    tao     // TCP
    quic    // QUIC
mq
    nats
    nsq
    kafka
log         // 支持context
    zap
container
    deque
    aqm
sync
    errgroup
    pipeline
    observer
    wpool       // 线程池
    opool       // 对象池
    cpool       // 连接池
naming
    resolver
    balancer
```


### 封装
*基础封装*
```
tracing
logging
metrics

异常处理
超时控制
断线重连
消息去重
超时重发
```


*服务端*
```
服务注册
请求隔离
```


*客户端*
```
服务发现
读写均衡
负载均衡
限流
熔断
降级
```

子系统选型
```
运维子系统
    注册中心
        consul


消息子系统
    nats


存储子系统
    缓存系统
        memcached
    内存存储
        redis
    对象存储
        minio
        s3
        oss
    结构存储
        mysql
        dgraph
    文档存储
        mongodb


监控子系统
    日志日志
        loki
    指标监控
        prometheus
    调用监控
        jaeger

```


### 愿景
好的框架都相似


业务
框架
中间件


存储
通信
计算



支持全双工
    grpc
    tcp
    udp
    quic
