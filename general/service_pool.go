package general

import (
	"container/list"
	"reflect"
	"sync"

	"github.com/kataras/iris/context"
)

const (
	runtimeKey = "runtime-ctx"
)

func newServicePool() *ServicePool {
	result := new(ServicePool)
	result.creater = make(map[reflect.Type]interface{})
	result.instanceList = make(map[reflect.Type]*list.List)
	result.instanceListMax = 200
	return result
}

// ServicePool .
type ServicePool struct {
	creater           map[reflect.Type]interface{}
	instanceList      map[reflect.Type]*list.List
	instanceListMutex sync.Mutex
	instanceListMax   int
}

// get .
func (pool *ServicePool) get(rt *appRuntime, service interface{}) {
	svalue := reflect.ValueOf(service)
	ptr := svalue.Elem() // 1级

	var newService interface{}
	newService = pool.take(ptr.Type())
	if newService == nil {
		newService = pool.malloc(ptr.Type())
	}
	if newService == nil {
		panic("newService is nil")
	}

	ptr.Set(reflect.ValueOf(newService))
	pool.objBeginRequest(rt, newService)
	rt.services = append(rt.services, newService)
}

// freeHandle .
func (pool *ServicePool) freeHandle() context.Handler {
	return func(ctx context.Context) {
		ctx.Next()
		rt := ctx.Values().Get(runtimeKey).(*appRuntime)
		for index := 0; index < len(rt.services); index++ {
			pool.free(rt.services[index])
		}
	}
}

// free .
func (pool *ServicePool) free(obj interface{}) {
	t := reflect.TypeOf(obj)
	defer pool.instanceListMutex.Unlock()
	pool.instanceListMutex.Lock()
	l, ok := pool.instanceList[t]
	if !ok {
		l = list.New()
		pool.instanceList[t] = l
	}
	if l.Len() >= pool.instanceListMax {
		return
	}
	l.PushFront(obj)
}

func (pool *ServicePool) bind(t reflect.Type, f interface{}) {
	pool.creater[t] = f
}

func (pool *ServicePool) malloc(t reflect.Type) interface{} {
	values := reflect.ValueOf(pool.creater[t]).Call([]reflect.Value{})
	if len(values) == 0 {
		panic("BindService func return to empty")
	}

	newSercice := values[0].Interface()
	allFields(newSercice, func(value reflect.Value) {
		if value.Kind() == reflect.Ptr {
			ok, newfield := globalApp.rpool.get(value.Type())
			if !ok {
				return
			}
			if !value.CanSet() {
				globalApp.IrisApp.Logger().Fatal("The member variable must be publicly visible, Its type is " + value.Type().String())
			}
			value.Set(newfield)
		}

		if value.Kind() == reflect.Interface {
			typeList := globalApp.rpool.allType()
			for index := 0; index < len(typeList); index++ {
				if !typeList[index].Implements(value.Type()) {
					continue
				}
				ok, newfield := globalApp.rpool.get(typeList[index])
				if !ok {
					continue
				}
				if !value.CanSet() {
					globalApp.IrisApp.Logger().Fatal("The member variable must be publicly visible, Its type is " + value.Type().String())
				}
				value.Set(newfield)
				return
			}
		}
	})
	return newSercice
}

func (pool *ServicePool) take(t reflect.Type) interface{} {
	defer pool.instanceListMutex.Unlock()
	pool.instanceListMutex.Lock()

	l, ok := pool.instanceList[t]
	if !ok || l == nil {
		return nil
	}
	if l.Len() <= 0 {
		return nil
	}
	back := l.Back()
	l.Remove(back)
	return back.Value
}

func (pool *ServicePool) objBeginRequest(rt *appRuntime, obj interface{}) {
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
		br, ok := value.Interface().(BeginRequest)
		if ok {
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
