# 简介
Hive 是一个基于六边形架构，结合领域模型范式（DDD）的框架。
> 参考 https://confluence.aishu.cn/pages/viewpage.action?pageId=166869672

## Hive名字由来
如上简介说明，Hive是一个框架，是基于六边形架构的框架。为什么取名Hive呢？

Hive的意思为“蜂巢”。
在大自然里，有各种各样的形状。蜂巢就是其中一个非常神奇的存在。衡量蜜蜂的经济活动是靠蜂蜜完成的。蜂蜜既是蜜蜂的食物也是筑造蜂巢的原料，所以对于蜜蜂来说蜂蜜更显得弥足珍贵了一丁点都不能浪费。所以蜜蜂在筑造蜂巢的时候必须考虑以下两个问题：1.空间尽可能大; 2.节约成本。

根据科学家后来的计算，在蜂蜡一定的情况下将蜂巢筑造成**六边形**可以使横截面的面积达到最大，并且结构非常的稳定。所以六边形结构也被称为“蜂窝状结构”。

而适配器的架构图正好也是六边形的，最开始想用HiveCore来做名字，寓意是以六边形架构（蜂巢）为核心的框架，后来考虑到名字的简洁性，最终取名为 *Hive*（像蜂窝状结构一样稳定）。

# 功能特性
* 集成 Iris
* HTTP/H2C/Oauth2C Server & Client
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
    BindBooting(f func(bootManager dhive.BootManager))
    //安装序列化，未安装默认使用官方json
    InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
}

/*
    Worker 请求运行时对象，一个请求创建一个运行时对象，可以直接注入到controller、service、factory、repository, 无需侵入的传递。
*/
type Worker interface {
    //获取iris的上下文
    IrisContext() dhive.Context
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
| 局部初始化 | dhive.Prepare |
| 开启监听服务 | http.Run |
| 回调已注册的匿名函数 | infra.RegisterShutdown |
| 程序关闭 | Application.Close |

# 请求生命周期
每一个请求开始都会创建若干依赖对象，worker、controller、service、factory、repository、infra等。每一个请求独立使用这些对象，不会多请求并发的读写共享对象。框架已经做了池化处理，效率上也有保障。请求结束会回收这些对象。 如果过程中使用了go func(){//访问相关对象}，请在之前调用 Worker.DelayReclaiming().

# 简单示例
## main.go
``` golang

import (
    "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DocCenter/conf"
    _ "HiveCore/adapter/controller" //引入输入适配器 http路由
    _ "HiveCore/adapter/repository" //引入输出适配器 repository资源库

    "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
    "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/infra/requests"
    "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/middleware"
)

func main() {
    app := dhive.NewApplication() //创建应用
    installDatabase(app)
    installRedis(app)
    installMiddleware(app)

    //创建http 监听
    addrRunner := app.NewRunner(conf.Get().App.Other["listen_addr"].(string))
    //创建http2.0 h2c 监听
    addrRunner = app.NewH2CRunner(conf.Get().App.Other["listen_addr"].(string))
    app.Run(addrRunner, *conf.Get().App)
}

func installMiddleware(app dhive.Application) {
    //Recover中间件
    app.InstallMiddleware(middleware.NewRecover())
    //Trace链路中间件
    app.InstallMiddleware(middleware.NewTrace("x-request-id"))
    //日志中间件，每个请求一个logger
    app.InstallMiddleware(middleware.NewRequestLogger("x-request-id"))
    //logRow中间件，每一行日志都会触发回调。如果返回true，将停止中间件遍历回调。
    app.Logger().Handle(middleware.DefaultLogRowHandle)

    //总线中间件，处理上下游透传的Header
    app.InstallBusMiddleware(middleware.NewBusFilter())
}

func installDatabase(app dhive.Application) {
    app.InstallDB(func() interface{} {
        //安装db的回调函数
        cfg := conf.Cfg.RWDB
		db := hiveutils.ConnProtonRWDB(cfg)
        return db
    })
}

func installRedis(app dhive.Application) {
    app.InstallRedis(func() (client redis.Cmdable) {
        cfg := conf.SvcConfig.Redis
		return hiveutils.ConnectRedis(*cfg)
    })
}
```
## 输入适配器 adapter/controller/default.go
``` golang
package controller

import (
	"HiveCore/domain"
	"HiveCore/infra"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
)

func init() {
	dhive.Prepare(func(initiator dhive.Initiator) {
		/*
		   普通方式绑定 Default控制器到路径 /
		   initiator.BindController("/", &DefaultController{})
		*/

		//中间件方式绑定， 只对本控制器生效，全局中间件请在main加入。
		if initiator.IsPrivate() {
			initiator.BindController("/", &Default{}, func(ctx dhive.Context) {
				worker := dhive.ToWorker(ctx)
				worker.Logger().Info("Hello middleware begin")
				ctx.Next()
				worker.Logger().Info("Hello middleware end")
			})
		}
	})
}

type Default struct {
	Sev    *domain.Default //依赖注入领域服务 Default
	Worker dhive.Worker     //依赖注入请求运行时 Worker，无需侵入的传递。
}

// Get handles the GET: / route.
func (c *Default) Get() dhive.Result {
    c.Worker.Logger().Infof("我是控制器")
    remote := c.Sev.RemoteInfo() //调用服务方法
    //返回JSON对象
    return &infra.JSONResponse{Object: remote}
}

// GetHello handles the GET: /hello route.
func (c *Default) GetHello() string {
	return "hello"
}

// PutHello handles the PUT: /hello route.
func (c *Default) PutHello() dhive.Result {
	return &infra.JSONResponse{Object: "putHello"}
}

// PostHello handles the POST: /hello route.
func (c *Default) PostHello() dhive.Result {
	return &infra.JSONResponse{Object: "postHello"}
}

func (m *Default) BeforeActivation(b dhive.BeforeActivation) {
	b.Handle("ANY", "/custom", "CustomHello")
	//b.Handle("GET", "/custom", "CustomHello")
	//b.Handle("PUT", "/custom", "CustomHello")
	//b.Handle("POST", "/custom", "CustomHello")
}

// PostHello handles the POST: /hello route.
func (c *Default) CustomHello() dhive.Result {
	method := c.Worker.IrisContext().Request().Method
	c.Worker.Logger().Info("CustomHello", dhive.LogFields{"method": method})
	return &infra.JSONResponse{Object: method + "CustomHello"}
}

// GetUserBy handles the GET: /user/{username:string} route.
func (c *Default) GetUserBy(username string) string {
	return username
}

// GetAgeByUserBy handles the GET: /age/{age:int}/user/{user:string} route.
func (c *Default) GetAgeByUserBy(age int, user string) dhive.Result {
	var result struct {
		User string
		Age  int
	}
	result.Age = age
	result.User = user

	return &infra.JSONResponse{Object: result}
}

```
## 领域服务 domain/default.go
``` golang
package domain

import (
	"HiveCore/adapter/repository"

    "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
)

func init() {
	dhive.Prepare(func(initiator dhive.Initiator) {
            //绑定 Default Service
            initiator.BindService(func() *Default {
                return &Default{}
            })
            initiator.InjectController(func(ctx dhive.Context) (service *Default) {
                //Default 注入到控制器
                initiator.GetService(ctx, &service)
                return
            })
	})
}

// Default .
type Default struct {
	Worker    dhive.Worker    //依赖注入请求运行时,无需侵入的传递。
	DefRepo   *repository.Default   //依赖注入资源库对象  DI方式
	DefRepoIF repository.DefaultRepoInterface  //也可以注入资源库接口 DIP方式
}

// RemoteInfo .
func (s *Default) RemoteInfo() (result struct {
	Ip string
	Ua string
}) {
        s.Worker.Logger().Infof("我是service")
        //调用资源库的方法
        result.Ip = s.DefRepo.GetIP()
        result.Ua = s.DefRepoIF.GetUA()
        return
}
```

## 领域服务依赖倒置接口层
``` golang
package dependency

// DefaultRepoInterface .
type DefaultRepoInterface interface {
	GetUA() string
}

```

## 输出适配器 adapter/repository/default.go
``` golang
package repository

import (
	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"

    "gorm.io/gorm"
)

func init() {
	dhive.Prepare(func(initiator dhive.Initiator) {
		initiator.BindRepository(func() *Default {
			return &Default{}
		})
	})
}

// Default .
type Default struct {
	dhive.Repository
}

// GetIP .
func (repo *Default) GetIP() string {
	//repo.DB().Find()
	repo.Worker().Logger().Info("I'm Repository GetIP")
	return repo.Worker().IrisContext().RemoteAddr()
}

// GetUA - implement DefaultRepoInterface interface
func (repo *Default) GetUA() string {
	repo.Worker().Logger().Info("I'm Repository GetUA")
	return repo.Worker().IrisContext().Request().UserAgent()
}

```