package general

import (
	"reflect"
)

func newComponentPool() *ComponentPool {
	result := new(ComponentPool)
	result.creater = make(map[reflect.Type]interface{})
	result.singlemap = make(map[reflect.Type]interface{})
	return result
}

// ComponentPool .
type ComponentPool struct {
	creater   map[reflect.Type]interface{}
	singlemap map[reflect.Type]interface{}
}

// bind .
func (pool *ComponentPool) bind(single bool, t reflect.Type, com interface{}) {
	if single {
		pool.singlemap[t] = com
		return
	}
	pool.creater[t] = com
}

// get .
func (pool *ComponentPool) get(rt *appRuntime, ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}
	if rtCom, ok := rt.coms[ptr.Type()]; ok {
		ptr.Set(reflect.ValueOf(rtCom))
		return true
	}

	newcom := pool.malloc(ptr.Type())
	if newcom != nil {
		ptr.Set(reflect.ValueOf(newcom))
		rt.coms[ptr.Type()] = newcom
		br, ok := newcom.(BeginRequest)
		if ok {
			br.BeginRequest(rt)
		}
		return true
	}
	return false
}

func (pool *ComponentPool) single(t reflect.Type) interface{} {
	return pool.singlemap[t]
}

func (pool *ComponentPool) malloc(t reflect.Type) interface{} {
	if pool.creater[t] == nil {
		return nil
	}
	values := reflect.ValueOf(pool.creater[t]).Call([]reflect.Value{})
	if len(values) == 0 {
		panic("Component func return to empty")
	}

	newCom := values[0].Interface()
	return newCom
}
