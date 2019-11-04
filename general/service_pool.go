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
	result.creater = make(map[reflect.Type]func() interface{})
	result.instanceList = make(map[reflect.Type]*list.List)
	result.instanceListMax = 200
	return result
}

// ServicePool .
type ServicePool struct {
	creater           map[reflect.Type]func() interface{}
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

func (pool *ServicePool) bind(obj interface{}, f func() interface{}) {
	pool.creater[reflect.TypeOf(obj)] = f
}

func (pool *ServicePool) malloc(t reflect.Type) interface{} {
	return pool.creater[t]()
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
	allFields(obj, func(value reflect.Value) {
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
