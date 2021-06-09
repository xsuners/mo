##  通信框架的封装

###  xws
- heartbeat
- stop validate
- more options


### TODO
- 连接保活
- 健康检查
- 限流
- 熔断
- 服务发现
- 负载均衡
- 调用链监控
- 指标监控
- 统一的日志监控
- 代理功能完善
    - tcp
        - grpc
        - nats
    - ws
        - grpc
        - nats
    - http
        - grpc
        - nats
    - grpc
        - grpc
        - nats
- 加解密
- 解压缩


### more todos
- 思考是否可以把网络库再抽象,拆分成服务接入层,网络接入层
- message 抽象,做成可由服务使用方自定义
