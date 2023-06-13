package internal

import (
	"net/http"
	"reflect"

	redis "github.com/go-redis/redis/v8"
	redismock "github.com/go-redis/redismock/v8"
	"github.com/kataras/iris/v12/context"
)

var _ UnitTest = (*UnitTestImpl)(nil)

// UnitTest .
type UnitTest interface {
	App() *Application
	GetService(service interface{})
	GetRepository(repository interface{})
	GetFactory(factory interface{})
	InstallDB(f func() (db interface{}))
	InstallDBTable(f func() (tables map[string]interface{}))
	InstallRedis(f func() (client redis.Cmdable))
	SetRedisMock(mock redismock.ClientMock)
	RedisMock() redismock.ClientMock
	Run()
	SetRequest(request *http.Request)
	InjectBaseEntity(entity interface{})
}

// UnitTestImpl .
type UnitTestImpl struct {
	rt        *worker
	request   *http.Request
	Private   bool
	redisMock redismock.ClientMock
}

// App .
func (u *UnitTestImpl) App() *Application {
	if u.Private {
		return privateApp
	} else {
		return publicApp
	}
}

// RedisMock .
func (u *UnitTestImpl) RedisMock() redismock.ClientMock {
	return u.redisMock
}

// SetRedisMock
func (u *UnitTestImpl) SetRedisMock(mock redismock.ClientMock) {
	u.redisMock = mock
}

// GetService .
func (u *UnitTestImpl) GetService(service interface{}) {
	u.App().GetService(u.rt.IrisContext(), service)
}

// GetRepository .
func (u *UnitTestImpl) GetRepository(repository interface{}) {
	instance := serviceElement{calls: []BeginRequest{}, workers: []reflect.Value{}}
	value := reflect.ValueOf(repository).Elem()
	ok := u.App().repoPool.diRepoFromValue(value, &instance)
	if !ok {
		u.App().IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		u.App().IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable, %v" + value.Type().String())
	}

	if br, ok := value.Interface().(BeginRequest); ok {
		instance.calls = append(instance.calls, br)
	}
	u.App().pool.beginRequest(u.rt, instance)
}

// GetFactory .
func (u *UnitTestImpl) GetFactory(factory interface{}) {
	instance := serviceElement{calls: []BeginRequest{}, workers: []reflect.Value{}}
	value := reflect.ValueOf(factory).Elem()
	ok := u.App().factoryPool.diFactoryFromValue(value, &instance)
	if !ok {
		u.App().IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		u.App().IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable, %v" + value.Type().String())
	}

	u.App().pool.beginRequest(u.rt, instance)
}

// InstallDB .
func (u *UnitTestImpl) InstallDB(f func() (db interface{})) {
	u.App().InstallDB(f)
}

// InstallDBTable .
func (u *UnitTestImpl) InstallDBTable(f func() (tables map[string]interface{})) {
	u.App().InstallDBTable(f)
}

// InstallRedis .
func (u *UnitTestImpl) InstallRedis(f func() (client redis.Cmdable)) {
	u.App().InstallRedis(f)
}

// Run .
func (u *UnitTestImpl) Run() {
	for index := 0; index < len(prepares); index++ {
		prepares[index](u.App())
	}
	u.rt = u.newRuntime()
	logLevel := "debug"
	u.App().IrisApp.Logger().SetLevel(logLevel)
	u.App().installDB()
	u.App().installDBTable()
	u.App().comPool.singleBooting(u.App())
}

func (u *UnitTestImpl) newRuntime() *worker {
	ctx := context.NewContext(u.App().IrisApp)
	if u.request == nil {
		u.request = new(http.Request)
	}
	ctx.BeginRequest(nil, u.request)
	rt := newWorker(ctx, false)
	ctx.Values().Set(WorkerKey, rt)
	return rt
}

// SetRequest .
func (u *UnitTestImpl) SetRequest(request *http.Request) {
	u.request = request
}

// InjectBaseEntity .
func (u *UnitTestImpl) InjectBaseEntity(entity interface{}) {
	injectBaseEntity(u.rt, entity)
	return
}
