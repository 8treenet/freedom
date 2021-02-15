package internal

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/kataras/iris/v12/context"
)

func newInfraPool() *InfraPool {
	result := new(InfraPool)
	result.singletons = make(map[reflect.Type]interface{})
	result.instances = make(map[reflect.Type]*sync.Pool)
	return result
}

// InfraPool represents a pool of zero or more infrastructure components.
type InfraPool struct {
	instances    map[reflect.Type]*sync.Pool
	singletons   map[reflect.Type]interface{}
	shutdownList []func()
}

func (pool *InfraPool) bindSingleton(t reflect.Type, com interface{}) {
	type setSingle interface {
		setSingle()
	}

	pool.singletons[t] = com
	if call, ok := com.(setSingle); ok {
		call.setSingle()
	}
}

func (pool *InfraPool) bindInstance(t reflect.Type, com interface{}) {
	pool.instances[t] = &sync.Pool{
		New: func() interface{} {
			values := reflect.ValueOf(com).Call([]reflect.Value{})
			if len(values) == 0 {
				panic(fmt.Sprintf("[Freedom] BindInfra: Infra func return to empty, %v", reflect.TypeOf(com)))
			}
			newCom := values[0].Interface()
			return newCom
		},
	}
}

// bind accepts a bool indicating should treat the component as singleton, a
// reflect.Type representing the component's underlying type, and a specified
// component instance. bind registers the component into the *InfraPool with the
// reflect.Type.
func (pool *InfraPool) bind(isSingleton bool, t reflect.Type, com interface{}) {
	if isSingleton {
		pool.bindSingleton(t, com)
		return
	}

	pool.bindInstance(t, com)
}

func (pool *InfraPool) getSingleton(t reflect.Type) interface{} {
	if t.Kind() != reflect.Interface {
		return pool.singletons[t]
	}
	for objType, ObjValue := range pool.singletons {
		if objType.Implements(t) {
			return ObjValue
		}
	}
	return nil
}

func (pool *InfraPool) getInstanceOrFalse(t reflect.Type) (*sync.Pool, bool) {
	if t.Kind() != reflect.Interface {
		pool, ok := pool.instances[t]
		return pool, ok
	}

	for objType, ObjValue := range pool.instances {
		if objType.Implements(t) {
			return ObjValue, true
		}
	}
	return nil, false
}

// get accepts a *worker and a reflect.Value representing what the component
// should be retrieved. get looks-up the component from the *InfraPool and fill
// out the reflect.Value with the component found in *InfraPool. get returns
// true if the component found, and false otherwise.
func (pool *InfraPool) get(rt *worker, ptr reflect.Value) bool {
	if scom := pool.getSingleton(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}

	syncpool, ok := pool.getInstanceOrFalse(ptr.Type())
	if !ok {
		return false
	}

	newcom := syncpool.Get()
	if newcom != nil {
		newValue := reflect.ValueOf(newcom)
		ptr.Set(newValue)
		rt.freeComs = append(rt.freeComs, newcom)
		br, ok := newcom.(BeginRequest)
		if ok {
			br.BeginRequest(rt)
		}
		return true
	}
	return false
}

// get .
func (pool *InfraPool) getByInternal(ptr reflect.Value) bool {
	if scom := pool.getSingleton(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}

	syncpool, ok := pool.getInstanceOrFalse(ptr.Type())
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

// booting .
func (pool *InfraPool) singleBooting(app *Application) {
	type boot interface {
		Booting(SingleBoot)
	}
	for _, com := range pool.singletons {
		bootimpl, ok := com.(boot)
		if !ok {
			continue
		}
		bootimpl.Booting(app)
	}
}

func (pool *InfraPool) registerShutdown(f func()) {
	pool.shutdownList = append(pool.shutdownList, f)
}

// shutdown .
func (pool *InfraPool) shutdown() {
	for i := 0; i < len(pool.shutdownList); i++ {
		pool.shutdownList[i]()
	}
}

// freeHandle .
func (pool *InfraPool) freeHandle() context.Handler {
	return func(ctx context.Context) {
		ctx.Next()
		rt := ctx.Values().Get(WorkerKey).(*worker)
		if rt.IsDeferRecycle() {
			return
		}
		for _, obj := range rt.freeComs {
			pool.free(obj)
		}
	}
}

// free .
func (pool *InfraPool) free(obj interface{}) {
	t := reflect.TypeOf(obj)
	syncpool, ok := pool.instances[t]
	if !ok {
		return
	}
	syncpool.Put(obj)
}

func (pool *InfraPool) diInfra(obj interface{}) {
	allFields(obj, func(value reflect.Value) {
		pool.diInfraFromValue(value)
	})
}

func (pool *InfraPool) diInfraFromValue(value reflect.Value) {
	globalApp.infraPool.getByInternal(value)
}

// GetSingleInfra .
func (pool *InfraPool) GetSingleInfra(ptr reflect.Value) bool {
	if singleton := pool.getSingleton(ptr.Type()); singleton != nil {
		ptr.Set(reflect.ValueOf(singleton))
		return true
	}
	return false
}
