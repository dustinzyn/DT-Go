# Introduction
Hive 是一个基于六边形架构，结合领域模型范式的框架。

# 功能特性
* 集成 Iris
* HTTP/H2C Server & Client
* AOP Worker & 无侵入 Context
* 可扩展组件 Infrastructure
* 依赖注入 & 依赖倒置 & 开闭原则
* DDD & 六边形架构
* 领域事件 & 消息队列组件
* CQS & 聚合根
* CRUD & PO Generate
* 一级缓存 & 二级缓存 & 防击穿

# 接口介绍
``` golang
// main 应用安装接口
type Application interface {
    //安装DB
    InstallDB(f func() interface{})
    //安装redis
    InstallRedis(f func() (client redis.Cmdable))
    //安装路由中间件
    InstallMiddleware(handler iris.Handler)
    //安装总线中间件,参考Http2 example
    InstallBusMiddleware(handle ...BusHandler)
    //安装全局Party http://domian/relativePath/controllerParty
    InstallParty(relativePath string)
    //创建 Runner
    NewRunner(addr string, configurators ...host.Configurator) iris.Runner
    NewH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
    NewAutoTLSRunner(addr string, domain string, email string, configurators ...host.Configurator) iris.Runner
    NewTLSRunner(addr string, certFile, keyFile string, configurators ...host.Configurator) iris.Runner
    //返回iris应用
    Iris() *iris.Application
    //日志
    Logger() *golog.Logger
    //启动
    Run(serve iris.Runner, c iris.Configuration)
    //安装其他, 如mongodb、es 等
    InstallCustom(f func() interface{})
    //启动回调: Prepare之后，Run之前.
    BindBooting(f func(bootManager freedom.BootManager))
    //安装序列化，未安装默认使用官方json
    InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
}

/*
    Worker 请求运行时对象，一个请求创建一个运行时对象，可以直接注入到controller、service、factory、repository, 无需侵入的传递。
*/
type Worker interface {
    //获取iris的上下文
    IrisContext() freedom.Context
    //获取带上下文的日志实例。
    Logger() Logger
    //设置带上下文的日志实例。
    SetLogger(Logger)
    //获取一级缓存实例，请求结束，该缓存生命周期结束。
    Store() *memstore.Store
    //获取总线，读写上下游透传的数据
    Bus() *Bus
    //获取标准上下文
    Context() stdContext.Context
    //With标准上下文
    WithContext(stdContext.Context)
    //该worker起始的时间
    StartTime() time.Time
    //延迟回收对象
    DelayReclaiming()
}

// Initiator 实例初始化接口，在Prepare使用。
type Initiator interface {
    //创建 iris.Party，可以指定中间件。
    CreateParty(relativePath string, handlers ...context.Handler) iris.Party
   //绑定控制器到 iris.Party。
    BindControllerWithParty(party iris.Party, controller interface{})
    //绑定控制器到路径，可以指定中间件。
    BindController(relativePath string, controller interface{}, handlers ...context.Handler)

    //绑定创建服务函数，绑定后客户可以依赖注入该类型使用。
    BindService(f interface{})
    //绑定创建工厂函数，绑定后客户可以依赖注入该类型使用。
    BindFactory(f interface{})
    //绑定创建Repository函数，绑定后客户可以依赖注入该类型使用。
    BindRepository(f interface{})
    //绑定创建组件函数，绑定后客户可以依赖注入该类型使用。 如果组件是单例 com是对象， 如果组件是多例com是创建函数。
    BindInfra(single bool, com interface{})

    //注入实例到控制器，适配iris的注入方式。
    InjectController(f interface{})
    //配合InjectController
    FetchInfra(ctx iris.Context, com interface{})
    //配合InjectController
    FetchService(ctx iris.Context, service interface{})



    //启动回调. Prepare之后，Run之前.
    BindBooting(f func(bootManager BootManager))
    //监听事件. 监听1个topic的事件，由指定控制器消费.
    ListenEvent(topic string, controller string})
    Iris() *iris.Application
}
```
# Application生命周期
| 作用  | API |
| --- | --- |
| 注册全局中间件 | Application.InstallMiddleware |
| 安装DB | Application.InstallDB |
| 安装Redis | Application.InstallRedis |
| 单例组件方法(需要重写方法) | infra.Booting |
| 回调已注册的匿名函数 | Initiator.BindBooting |
| 局部初始化 | hive.Prepare |
| 开启监听服务 | http.Run |
| 回调已注册的匿名函数 | infra.RegisterShutdown |
| 程序关闭 | Application.Close |

# 请求生命周期
每一个请求开始都会创建若干依赖对象，worker、controller、service、factory、repository、infra等。每一个请求独立使用这些对象，不会多请求并发的读写共享对象。框架已经做了池化处理，效率上也有保障。请求结束会回收这些对象。 如果过程中使用了go func(){//访问相关对象}，请在之前调用 Worker.DelayReclaiming().
