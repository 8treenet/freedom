package internal

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/kataras/iris/v12/context"
)

func newInfraPool() *infraPool {
	result := new(infraPool)
	result.singlemap = make(map[reflect.Type]interface{})
	result.instancePool = make(map[reflect.Type]*sync.Pool)
	return result
}

// infraPool .
type infraPool struct {
	instancePool map[reflect.Type]*sync.Pool
	singlemap    map[reflect.Type]interface{}
	shutdownList []func()
}

// bind .
func (pool *infraPool) bind(single bool, t reflect.Type, com interface{}) {
	type setSingle interface {
		setSingle()
	}
	if single {
		pool.singlemap[t] = com
		if call, ok := com.(setSingle); ok {
			call.setSingle()
		}
		return
	}
	pool.instancePool[t] = &sync.Pool{
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

// get .
func (pool *infraPool) get(rt *worker, ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}

	syncpool, ok := pool.much(ptr.Type())
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
func (pool *infraPool) getByInternal(ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}

	syncpool, ok := pool.much(ptr.Type())
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

// singleBooting .
func (pool *infraPool) singleBooting(app *Application) {
	type boot interface {
		Booting(BootManager)
	}
	for _, com := range pool.singlemap {
		bootimpl, ok := com.(boot)
		if !ok {
			continue
		}
		bootimpl.Booting(app)
	}
}

func (pool *infraPool) registerShutdown(f func()) {
	pool.shutdownList = append(pool.shutdownList, f)
}

// shutdown .
func (pool *infraPool) shutdown() {
	for i := 0; i < len(pool.shutdownList); i++ {
		pool.shutdownList[i]()
	}
}

func (pool *infraPool) single(t reflect.Type) interface{} {
	if t.Kind() != reflect.Interface {
		return pool.singlemap[t]
	}
	for objType, ObjValue := range pool.singlemap {
		if objType.Implements(t) {
			return ObjValue
		}
	}
	return nil
}

func (pool *infraPool) much(t reflect.Type) (*sync.Pool, bool) {
	if t.Kind() != reflect.Interface {
		pool, ok := pool.instancePool[t]
		return pool, ok
	}

	for objType, ObjValue := range pool.instancePool {
		if objType.Implements(t) {
			return ObjValue, true
		}
	}
	return nil, false
}

// freeHandle .
func (pool *infraPool) freeHandle() context.Handler {
	return func(ctx *context.Context) {
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
func (pool *infraPool) free(obj interface{}) {
	t := reflect.TypeOf(obj)
	syncpool, ok := pool.instancePool[t]
	if !ok {
		return
	}
	syncpool.Put(obj)
}

func (pool *infraPool) diInfra(obj interface{}) {
	allFields(obj, func(value reflect.Value) {
		pool.diInfraFromValue(value)
	})
}

func (pool *infraPool) diInfraFromValue(value reflect.Value) {
	globalApp.comPool.getByInternal(value)
}

// FetchSingleInfra .
func (pool *infraPool) FetchSingleInfra(ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}
	return false
}
