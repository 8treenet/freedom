package general

import (
	"reflect"
	"sync"

	"github.com/kataras/iris/context"
)

func newInfraPool() *InfraPool {
	result := new(InfraPool)
	result.singlemap = make(map[reflect.Type]interface{})
	result.instancePool = make(map[reflect.Type]*sync.Pool)
	return result
}

// InfraPool .
type InfraPool struct {
	instancePool map[reflect.Type]*sync.Pool
	singlemap    map[reflect.Type]interface{}
}

// bind .
func (pool *InfraPool) bind(single bool, t reflect.Type, com interface{}) {
	if single {
		pool.singlemap[t] = com
		return
	}
	pool.instancePool[t] = &sync.Pool{
		New: func() interface{} {

			values := reflect.ValueOf(com).Call([]reflect.Value{})
			if len(values) == 0 {
				panic("Infra func return to empty")
			}
			newCom := values[0].Interface()
			return newCom
		},
	}
}

// get .
func (pool *InfraPool) get(rt *appRuntime, ptr reflect.Value) bool {
	if scom := pool.single(ptr.Type()); scom != nil {
		ptr.Set(reflect.ValueOf(scom))
		return true
	}
	if rtCom, ok := rt.coms[ptr.Type()]; ok {
		ptr.Set(reflect.ValueOf(rtCom))
		return true
	}

	syncpool, ok := pool.instancePool[ptr.Type()]
	if !ok {
		return false
	}

	newcom := syncpool.Get()
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

// booting .
func (pool *InfraPool) singleBooting(app *Application) {
	type boot interface {
		Booting(SingleBoot)
	}
	for _, com := range pool.singlemap {
		bootimpl, ok := com.(boot)
		if !ok {
			continue
		}
		bootimpl.Booting(app)
	}
}

func (pool *InfraPool) single(t reflect.Type) interface{} {
	return pool.singlemap[t]
}

// freeHandle .
func (pool *InfraPool) freeHandle() context.Handler {
	return func(ctx context.Context) {
		ctx.Next()
		rt := ctx.Values().Get(runtimeKey).(*appRuntime)
		for objType, obj := range rt.coms {
			pool.free(objType, obj)
		}
	}
}

// free .
func (pool *InfraPool) free(objType reflect.Type, obj interface{}) {
	syncpool, ok := pool.instancePool[objType]
	if !ok {
		return
	}
	syncpool.Put(obj)
}

func (pool *InfraPool) diInfra(rt *appRuntime, obj interface{}) {
	allFields(obj, func(value reflect.Value) {
		globalApp.comPool.get(rt, value)
	})
}
