package internal

import (
	"fmt"
	"reflect"
)

type other struct {
	private  bool
	installs []func() interface{}
	pool     map[reflect.Type]reflect.Value
}

func NewOther(private bool) *other {
	return &other{
		private: private,
		pool:    make(map[reflect.Type]reflect.Value),
	}
}

func (o *other) add(f func() interface{}) {
	o.installs = append(o.installs, f)
}

func (o *other) booting() {
	for i := 0; i < len(o.installs); i++ {
		install := o.installs[i]()
		t := reflect.ValueOf(install).Type()
		for t.Kind() != reflect.Struct {
			t = t.Elem()
		}

		o.pool[t] = reflect.ValueOf(install)
	}
}

func (o *other) get(object interface{}) {
	value := reflect.ValueOf(object)
	vtype := value.Type()
	for vtype.Kind() != reflect.Struct {
		vtype = vtype.Elem()
	}

	poolValue, ok := o.pool[vtype]
	if !ok {
		panic(fmt.Sprintf("Repository.Other: Does not exist, %v", vtype))
	}
	value.Elem().Set(poolValue)
}
