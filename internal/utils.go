package internal

import (
	"errors"
	"reflect"
)

func parsePoolFunc(f interface{}) (outType reflect.Type, e error) {
	ftype := reflect.TypeOf(f)
	if ftype.Kind() != reflect.Func {
		e = errors.New("It's not a func")
		return
	}
	if ftype.NumOut() != 1 {
		e = errors.New("Return must be an object pointer")
		return
	}
	outType = ftype.Out(0)
	if outType.Kind() != reflect.Ptr {
		e = errors.New("Return must be an object pointer")
		return
	}
	return
}

// allFields
func allFields(dest interface{}, call func(reflect.Value)) {
	destVal := indirect(reflect.ValueOf(dest))
	destType := destVal.Type()
	if destType.Kind() != reflect.Struct && destType.Kind() != reflect.Interface {
		return
	}

	for index := 0; index < destVal.NumField(); index++ {
		if destType.Field(index).Anonymous {
			allFields(destVal.Field(index).Addr().Interface(), call)
			continue
		}
		val := destVal.Field(index)
		kind := val.Kind()
		if kind != reflect.Ptr && kind != reflect.Interface {
			continue
		}
		call(val)
	}
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr || reflectValue.Kind() == reflect.Interface {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

// allFieldsFromValue
func allFieldsFromValue(val reflect.Value, call func(reflect.Value)) {
	destVal := indirect(val)
	destType := destVal.Type()
	if destType.Kind() != reflect.Struct && destType.Kind() != reflect.Interface {
		return
	}
	for index := 0; index < destVal.NumField(); index++ {
		if destType.Field(index).Anonymous {
			allFieldsFromValue(destVal.Field(index).Addr(), call)
			continue
		}
		val := destVal.Field(index)
		kind := val.Kind()
		if kind != reflect.Ptr && kind != reflect.Interface {
			continue
		}
		call(val)
	}
}

func fetchValue(dest, src interface{}) bool {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return false
	}
	value = value.Elem()
	srvValue := reflect.ValueOf(src)
	if value.Type() == srvValue.Type() {
		value.Set(srvValue)
		return true
	}
	return false
}


func parseCallServiceFunc(f interface{}) (inType reflect.Type, e error) {
	ftype := reflect.TypeOf(f)
	if ftype.Kind() != reflect.Func {
		e = errors.New("It's not a func")
		return
	}
	if ftype.NumIn() != 1 {
		e = errors.New("The pointer parameter must be a service object")
		return
	}
	inType = ftype.In(0)
	if inType.Kind() != reflect.Ptr {
		e = errors.New("The pointer parameter must be a service object")
		return
	}
	return
}