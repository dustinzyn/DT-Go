package internal

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/kataras/iris/v12/context"
)

type InfraPool struct {
	private      bool
	instancePool map[reflect.Type]*sync.Pool
	singlemap    map[reflect.Type]interface{}
	shutdownList []func()
}

func newInfraPool(private bool) *InfraPool {
	result := new(InfraPool)
	result.private = private
	result.instancePool = make(map[reflect.Type]*sync.Pool)
	result.singlemap = make(map[reflect.Type]interface{})
	return result
}

// app .
func (pool *InfraPool) app() *Application {
	if pool.private {
		return privateApp
	} else {
		return publicApp
	}
}

// bind .
func (pool *InfraPool) bind(single, private bool, t reflect.Type, com interface{}) {
	type setSingle interface {
		setSingle()
	}
	type setPrivate interface {
		setPrivate()
	}
	if call, ok := com.(setPrivate); ok {
		if private {
			call.setPrivate()
		}
	}
	if single {
		pool.singlemap[t] = com
		if call, ok := com.(setSingle); ok {
			call.setSingle()
		}

		return
	}
	pool.instancePool[t] = &sync.Pool{
		New: func() interface{} {
			values := reflect.ValueOf(com).Call([]reflect.Value{})
			if len(values) == 0 {
				panic(fmt.Sprintf("BindInfra: Infra func return to empty, %v", reflect.TypeOf(com)))
			}
			newCom := values[0].Interface()
			return newCom
		},
	}
}

// get .
func (pool *InfraPool) get(rt *worker, ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}

	syncpool, ok := pool.much(ptr.Type())
	if !ok {
		return false
	}

	newcom := syncpool.Get()
	if newcom != nil {
		newValue := reflect.ValueOf(newcom)
		ptr.Set(newValue)
		rt.coms = append(rt.coms, newcom)
		br, ok := newcom.(BeginRequest)
		if ok {
			br.BeginRequest(rt)
		}
		return true
	}
	return false
}

func (pool *InfraPool) single(t reflect.Type) interface{} {
	if t.Kind() != reflect.Interface {
		return pool.singlemap[t]
	}
	for objType, ObjValue := range pool.singlemap {
		if objType.Implements(t) {
			return ObjValue
		}
	}
	return nil
}

func (pool *InfraPool) much(t reflect.Type) (*sync.Pool, bool) {
	if t.Kind() != reflect.Interface {
		pool, ok := pool.instancePool[t]
		return pool, ok
	}

	for objType, ObjValue := range pool.instancePool {
		if objType.Implements(t) {
			return ObjValue, true
		}
	}
	return nil, false
}

// get .
func (pool *InfraPool) getByInternal(ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}

	syncpool, ok := pool.much(ptr.Type())
	if !ok {
		return false
	}

	newcom := syncpool.Get()
	if newcom != nil {
		newValue := reflect.ValueOf(newcom)
		ptr.Set(newValue)
		return true
	}
	return false
}

func (pool *InfraPool) diInfra(obj interface{}) {
	allFields(obj, func(value reflect.Value) {
		pool.diInfraFromValue(value)
	})
}

func (pool *InfraPool) diInfraFromValue(value reflect.Value) {
	pool.app().comPool.getByInternal(value)
}

// booting .
func (pool *InfraPool) singleBooting(app *Application) {
	type boot interface {
		Booting(SingleBoot)
	}
	for _, com := range pool.singlemap {
		bootimpl, ok := com.(boot)
		if !ok {
			continue
		}
		bootimpl.Booting(app)
	}
}

// shutdown .
func (pool *InfraPool) shutdown() {
	for i := 0; i < len(pool.shutdownList); i++ {
		pool.shutdownList[i]()
	}
}

// GetSingleInfra .
func (pool *InfraPool) GetSingleInfra(ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}
	return false
}

func (pool *InfraPool) registerShutdown(f func()) {
	pool.shutdownList = append(pool.shutdownList, f)
}

// freeHandle .
func (pool *InfraPool) freeHandle() context.Handler {
	return func(ctx *context.Context) {
		ctx.Next()
		rt := ctx.Values().Get(WorkerKey).(*worker)
		if rt.IsDeferRecycle() {
			return
		}
		for _, obj := range rt.coms {
			pool.free(obj)
		}
	}
}

// free .
func (pool *InfraPool) free(obj interface{}) {
	t := reflect.TypeOf(obj)
	syncpool, ok := pool.instancePool[t]
	if !ok {
		return
	}
	syncpool.Put(obj)
}
