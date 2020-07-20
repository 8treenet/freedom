package internal

import (
	"reflect"
	"sync"

	"github.com/kataras/iris/v12/context"
)

const (
	WorkerKey = "freedom-worker-ctx"
)

func newServicePool() *ServicePool {
	result := new(ServicePool)
	result.instancePool = make(map[reflect.Type]*sync.Pool)
	return result
}

// ServicePool .
type ServicePool struct {
	instancePool map[reflect.Type]*sync.Pool
}

// get .
func (pool *ServicePool) get(rt *worker, service interface{}) {
	svalue := reflect.ValueOf(service)
	ptr := svalue.Elem() // Level 1

	var newService interface{}
	newService = pool.malloc(ptr.Type())
	if newService == nil {
		panic("newService is nil")
	}

	ptr.Set(reflect.ValueOf(newService))
	pool.objBeginRequest(rt, newService)
	rt.freeServices = append(rt.freeServices, newService)
}

// freeHandle .
func (pool *ServicePool) freeHandle() context.Handler {
	return func(ctx context.Context) {
		ctx.Next()
		rt := ctx.Values().Get(WorkerKey).(*worker)
		if rt.IsDeferRecycle() {
			return
		}
		for index := 0; index < len(rt.freeServices); index++ {
			pool.free(rt.freeServices[index])
		}
	}
}

// free .
func (pool *ServicePool) free(obj interface{}) {
	t := reflect.TypeOf(obj)
	syncpool, ok := pool.instancePool[t]
	if !ok {
		return
	}

	syncpool.Put(obj)
}

func (pool *ServicePool) bind(t reflect.Type, f interface{}) {
	pool.instancePool[t] = &sync.Pool{
		New: func() interface{} {
			values := reflect.ValueOf(f).Call([]reflect.Value{})
			if len(values) == 0 {
				panic("BindService func return to empty")
			}

			newService := values[0].Interface()
			globalApp.rpool.diRepo(newService)
			globalApp.comPool.diInfra(newService)
			return newService
		},
	}
}

func (pool *ServicePool) malloc(t reflect.Type) interface{} {
	syncpool, ok := pool.instancePool[t]
	if !ok {
		return nil
	}
	newSercice := syncpool.Get()
	if newSercice == nil {
		panic("BindService func return to empty")
	}
	return newSercice
}

func (pool *ServicePool) allType() (list []reflect.Type) {
	for t := range pool.instancePool {
		list = append(list, t)
	}
	return
}

func (pool *ServicePool) objBeginRequest(rt *worker, obj interface{}) {
	rtValue := reflect.ValueOf(rt)
	allFields(obj, func(value reflect.Value) {
		kind := value.Kind()
		if kind == reflect.Interface && rtValue.Type().AssignableTo(value.Type()) && value.CanSet() {
			value.Set(rtValue)
			return
		}
		if !value.CanInterface() {
			return
		}
		vi := value.Interface()
		allFields(vi, func(value reflect.Value) {
			if !value.CanInterface() {
				return
			}
			infra := value.Interface()
			br, ok := infra.(BeginRequest)
			if ok {
				br.BeginRequest(rt)
			}
		})
		br, ok := vi.(BeginRequest)
		if ok {
			//globalApp.comPool.diInfra(rt, vi)
			br.BeginRequest(rt)
		}
	})

	br, ok := obj.(BeginRequest)
	if ok {
		br.BeginRequest(rt)
	}

	/*
		structFields(obj, func(sf reflect.StructField, val reflect.Value) {
			kind := val.Kind()
			if kind == reflect.Interface && rtValue.Type().AssignableTo(val.Type()) && val.CanSet() {
				val.Set(rtValue)
			}
		})

		structFields(obj, func(sf reflect.StructField, val reflect.Value) {
			kind := val.Kind()
			if kind == reflect.Ptr || kind == reflect.Interface {
				if fun := val.MethodByName("BeginRequest"); fun.IsValid() {
					fun.Call([]reflect.Value{})
				}
			}
		})
	*/
}
