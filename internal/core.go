package internal

import (
	redis "github.com/go-redis/redis/v8"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"gorm.io/gorm"
)

type Initiator interface {
	CreateParty(relativePath string, handlers ...context.Handler) iris.Party
	BindController(relativePath string, controller interface{}, handlers ...context.Handler)
	BindControllerByParty(party iris.Party, controller interface{})
	BindService(f interface{})
	InjectController(f interface{})
	BindRepository(f interface{})
	BindFactory(f interface{})
	GetService(ctx iris.Context, service interface{})
	// BindInfra if is a singleton, com is an object. if is multiton, com is a function
	BindInfra(single, private bool, com interface{})
	GetInfra(ctx iris.Context, com interface{})
	// Listen Event
	ListenEvent(eventName string, objMethod string, appointInfra ...interface{})
	Start(f func(starter Starter))
	Iris() *iris.Application
	IsPrivate() bool
}

type Starter interface {
	Iris() *iris.Application
	// Asynchronous cache warm-up
	AsyncCachePreheat(f func(repo *Repository))
	// Sync cache warm-up
	CachePreheat(f func(repo *Repository))
	GetSingleInfra(com interface{}) bool
	Db() *gorm.DB
	Redis() redis.Cmdable
}

// SingleBoot singlton startup component.
type SingleBoot interface {
	Iris() *iris.Application
	EventsPath(infra interface{}) map[string]string
	RegisterShutdown(func())
}

// BeginRequest
type BeginRequest interface {
	BeginRequest(Worker Worker)
}

var (
	prepares []func(Initiator)
	privatePrepares []func(Initiator)
	publicPrepares []func(Initiator)
	privateStarters []func(starter Starter)
	publicStarters []func(starter Starter)
)

// Prepare app.BindController or app.BindControllerByParty
func Prepare(f func(Initiator)) {
	prepares = append(prepares, f)
	// if private {
	// 	privatePrepares = append(privatePrepares, f)
	// } else {
	// 	publicPrepares = append(publicPrepares, f)
	// }
}
