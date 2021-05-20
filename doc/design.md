接口的应用
    1. grpc Service 的定义
    2. 三方框架log的定义
    3. opentracing
    4. database/sql driver


接口集
    参考 opentracing, opentelementry


拦截器 vs 过滤器 vs 钩子

设计思想
    - 自动化,解放编码,专注业务
    - 单一化,解决具体问题的办法有100种,但是好办法只有一种
    - 安全性,规范编码,充分测试
    - 易用性,极简接口设计
    - 易理解,极端的一致性使其极易理解
    - 拓展性,无论通信,存储,计算还是监控都可插拔的


