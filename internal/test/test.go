package test

import "reflect"

func Clone[T any](src T) T {
	v := reflect.ValueOf(src).Elem()
	clone := reflect.New(v.Type())
	clone.Elem().Set(v)
	return clone.Interface().(T)
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
