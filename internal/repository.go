package internal

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	redis "github.com/go-redis/redis/v8"
	iris "github.com/kataras/iris/v12"

	"DT-Go/infra/dhttp"
)

type Repository struct {
	worker Worker
}

// BeginRequest .
func (repo *Repository) BeginRequest(rt Worker) {
	repo.worker = rt
}

// FetchDB .
func (repo *Repository) FetchDB(db interface{}) error {
	resultDB := repo.app().Database.db

	transactionData := repo.worker.Store().Get("local_transaction_db")
	if transactionData != nil {
		resultDB = transactionData
	}
	if resultDB == nil {
		return errors.New("DB not found, please install")
	}
	// db必须为指针类型
	if !fetchValue(db, resultDB) {
		return errors.New("DB not found, please install")
	}
	// db = resultDB
	return nil
}

// FetchSourceDB .
func (repo *Repository) FetchSourceDB(db interface{}) error {
	resultDB := repo.app().Database.db
	if resultDB == nil {
		return errors.New("DB not found, please install")
	}
	if !fetchValue(db, resultDB) {
		return errors.New("DB not found, please install")
	}
	return nil
}

// Redis .
func (repo *Repository) Redis() redis.Cmdable {
	return repo.app().Cache.client
}

// NewHTTPRequest transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (repo *Repository) NewHTTPRequest(url string, transferBus ...bool) dhttp.Request {
	req := dhttp.NewHTTPRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	if repo.worker == nil {
		//The singleton object does not have a Worker component
		return req
	}
	req.SetHeader(repo.worker.Bus().Header)
	return req
}

// NewH2CRequest transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (repo *Repository) NewH2CRequest(url string, transferBus ...bool) dhttp.Request {
	req := dhttp.NewH2CRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	if repo.worker == nil {
		//The singleton object does not have a Worker component
		return req
	}
	req.SetHeader(repo.worker.Bus().Header)
	return req
}

// NewOAuth2Request transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (repo *Repository) NewOAuth2Request(url string, transferBus ...bool) dhttp.Request {
	req := dhttp.NewOauth2Request(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	if repo.worker == nil {
		//The singleton object does not have a Worker component
		return req
	}
	req.SetHeader(repo.worker.Bus().Header)
	return req
}

// // NewThriftClient .
// func (repo *Repository) NewThriftClient(config *dthrift.ThriftPoolConfig) *dthrift.ThriftPoolAgent {
// 	return dthrift.NewThriftPoolAgent(config)
// }

// // SingleFlight .
// func (repo *Repository) SingleFlight(key string, value, takeObject interface{}, fn func() (interface{}, error)) error {
// 	takeValue := reflect.ValueOf(takeObject)
// 	if takeValue.Kind() != reflect.Ptr {
// 		panic("'takeObject' must be a pointer")
// 	}
// 	takeValue = takeValue.Elem()
// 	if !takeValue.CanSet() {
// 		panic("'takeObject' cannot be set")
// 	}
// 	v, err, _ := globalApp.singleFlight.Do(key+"-"+fmt.Sprint(value), fn)
// 	if err != nil {
// 		return err
// 	}

// 	newValue := reflect.ValueOf(v)
// 	if takeValue.Type() != newValue.Type() {
// 		panic("'takeObject' type error")
// 	}
// 	takeValue.Set(reflect.ValueOf(v))
// 	return nil
// }

// InjectBaseEntity .
func (repo *Repository) InjectBaseEntity(entity Entity) {
	injectBaseEntity(repo.worker, entity)
	return
}

// InjectBaseEntitys .
func (repo *Repository) InjectBaseEntitys(entitys interface{}) {
	entitysValue := reflect.ValueOf(entitys)
	if entitysValue.Kind() != reflect.Slice {
		panic(fmt.Sprintf("InjectBaseEntitys: It's not a slice, %v", entitysValue.Type()))
	}
	for i := 0; i < entitysValue.Len(); i++ {
		iface := entitysValue.Index(i).Interface()
		if _, ok := iface.(Entity); !ok {
			panic(fmt.Sprintf("InjectBaseEntitys: This is not an entity, %v", entitysValue.Type()))
		}
		injectBaseEntity(repo.worker, iface)
	}
	return
}

// Other .
func (repo *Repository) Other(obj interface{}) {
	repo.app().other.get(obj)
	return
}

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := int64(5)
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		sec = v.(int64)
	}
	dhttp.InitHTTPClient(time.Duration(sec) * time.Second)
	dhttp.InitH2cClient(time.Duration(sec) * time.Second)
}

// Worker .
func (repo *Repository) Worker() Worker {
	return repo.worker
}

// app returns an application
func (repo *Repository) app() *Application {
	if repo.worker.IsPrivate() {
		return privateApp
	} else {
		return publicApp
	}
}
