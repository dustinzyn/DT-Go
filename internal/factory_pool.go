package internal

import (
	"fmt"
	"reflect"
)

type FactoryPool struct {
	private bool
	creater map[reflect.Type]interface{}
}

func newFactoryPool(private bool) *FactoryPool {
	result := new(FactoryPool)
	result.private = private
	result.creater = make(map[reflect.Type]interface{})
	return result
}

// app .
func (pool *FactoryPool) app() *Application {
	if pool.private {
		return privateApp
	} else {
		return publicApp
	}
}

// get .
func (pool *FactoryPool) get(t reflect.Type) (ok bool, result reflect.Value) {
	fun, ok := pool.creater[t]
	if !ok {
		return false, reflect.ValueOf(nil)
	}

	values := reflect.ValueOf(fun).Call([]reflect.Value{})
	if len(values) == 0 {
		panic(fmt.Sprintf("BindFactory func return to empty, %v", t))
	}

	return true, values[0]
}

// bind
func (pool *FactoryPool) bind(outType reflect.Type, f interface{}) {
	pool.creater[outType] = f
}

// allType .
func (pool *FactoryPool) allType() (list []reflect.Type) {
	for t := range pool.creater {
		list = append(list, t)
	}
	return
}

// diFactory .
func (pool *FactoryPool) diFactory(factory interface{}, instance *serviceElement) {
	allFields(factory, func(value reflect.Value) {
		pool.diFactoryFromValue(value, instance)
	})
}

func (pool *FactoryPool) diFactoryFromValue(value reflect.Value, instance *serviceElement) bool {
	//如果是指针的成员变量
	if value.Kind() == reflect.Ptr && value.IsZero() {
		ok, newfield := pool.get(value.Type())
		if !ok {
			return false
		}
		if !value.CanSet() {
			panic(fmt.Sprintf("This use factory object must be a capital variable: %v" + value.Type().String()))
		}
		//创建实例并且注入基础设施组件和资源库
		value.Set(newfield)
		allFieldsFromValue(newfield, func(fieldValue reflect.Value) {
			kind := fieldValue.Kind()
			if kind == reflect.Interface && workerType.AssignableTo(fieldValue.Type()) && fieldValue.CanSet() {
				//如果是运行时对象
				instance.workers = append(instance.workers, fieldValue)
				return
			}

			pool.app().repoPool.diRepoFromValue(fieldValue, instance)
			pool.app().comPool.diInfraFromValue(fieldValue)

			if fieldValue.IsNil() {
				return
			}
			if br, ok := fieldValue.Interface().(BeginRequest); ok {
				instance.calls = append(instance.calls, br)
			}
		})
		// globalApp.comPool.diInfra(factoryObj)
		// globalApp.rpool.diRepo(factoryObj)
		return true
	}

	//如果是接口的成员变量
	if value.Kind() == reflect.Interface && value.IsZero() {
		typeList := pool.allType()
		for index := 0; index < len(typeList); index++ {
			if !typeList[index].Implements(value.Type()) {
				continue
			}
			ok, newfield := pool.get(typeList[index])
			if !ok {
				continue
			}
			if !value.CanSet() {
				panic(fmt.Sprintf("This use factory object must be a capital variable: %v", value.Type().String()))
			}
			//创建实例并且注入基础设施组件和资源库
			value.Set(newfield)
			allFieldsFromValue(newfield, func(fieldValue reflect.Value) {
				kind := fieldValue.Kind()
				if kind == reflect.Interface && workerType.AssignableTo(fieldValue.Type()) && fieldValue.CanSet() {
					//如果是运行时对象
					instance.workers = append(instance.workers, fieldValue)
					return
				}

				pool.app().repoPool.diRepoFromValue(fieldValue, instance)
				pool.app().comPool.diInfraFromValue(fieldValue)

				if fieldValue.IsNil() {
					return
				}
				if br, ok := fieldValue.Interface().(BeginRequest); ok {
					instance.calls = append(instance.calls, br)
				}
			})
			return true
		}
	}
	return false
}
