package internal

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/kataras/iris/v12/context"
)

// newServicePool create a ServiceElementPool
func newServicePool() *servicePool {
	result := new(servicePool)
	result.pool = make(map[reflect.Type]*sync.Pool)
	return result
}

// servicePool is domain service pool,a domain service type owns a pool
// pool key: the reflect.Type of domain service
// pool value: *sync.Pool
type servicePool struct {
	pool map[reflect.Type]*sync.Pool
}

type serviceElement struct {
	workers       []reflect.Value
	calls         []BeginRequest
	serviceObject interface{}
}

// get .
func (pool *servicePool) get(rt *worker, service interface{}) {
	svalue := reflect.ValueOf(service)
	if svalue.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("[Freedom] No dependency injection was found for the service object,%v", svalue.Type()))
	}
	ptr := svalue.Elem()

	var newService interface{}
	newService = pool.malloc(ptr.Type())
	if newService == nil {
		panic(fmt.Sprintf("[Freedom] No dependency injection was found for the service object,%v", ptr.Type()))
	}

	ptr.Set(reflect.ValueOf(newService.(serviceElement).serviceObject))
	pool.beginRequest(rt, newService)
	rt.freeServices = append(rt.freeServices, newService)
}

// create .
func (pool *servicePool) create(rt Worker, service reflect.Type) (interface{}, error) {
	newService := pool.malloc(service)
	if newService == nil {
		return nil, fmt.Errorf("[Freedom] No dependency injection was found for the service object,%v", service)
	}

	pool.beginRequest(rt, newService)
	return newService, nil
}

// freeHandle .
func (pool *servicePool) freeHandle() context.Handler {
	return func(ctx *context.Context) {
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
func (pool *servicePool) free(service interface{}) {
	t := reflect.TypeOf(service.(serviceElement).serviceObject)
	syncpool, ok := pool.pool[t]
	if !ok {
		return
	}

	syncpool.Put(service)
}

func (pool *servicePool) bind(t reflect.Type, f interface{}) {
	pool.pool[t] = &sync.Pool{
		New: func() interface{} {
			values := reflect.ValueOf(f).Call([]reflect.Value{})
			if len(values) == 0 {
				panic(fmt.Sprintf("[Freedom] BindService: func return to empty, %v", reflect.TypeOf(f)))
			}

			newService := values[0].Interface()
			result := serviceElement{serviceObject: newService, calls: []BeginRequest{}, workers: []reflect.Value{}}
			allFields(newService, func(fieldValue reflect.Value) {
				kind := fieldValue.Kind()
				if kind == reflect.Interface && workerType.AssignableTo(fieldValue.Type()) && fieldValue.CanSet() {
					//如果是运行时对象
					result.workers = append(result.workers, fieldValue)
					return
				}
				globalApp.rpool.diRepoFromValue(fieldValue, &result)
				globalApp.comPool.diInfraFromValue(fieldValue)
				globalApp.factoryPool.diFactoryFromValue(fieldValue, &result)

				if fieldValue.IsNil() {
					return
				}
				if br, ok := fieldValue.Interface().(BeginRequest); ok {
					result.calls = append(result.calls, br)
				}
			})

			if br, ok := newService.(BeginRequest); ok {
				result.calls = append(result.calls, br)
			}
			return result
		},
	}
}

// malloc return a service object from pool or create a new service obj
func (pool *servicePool) malloc(t reflect.Type) interface{} {
	// 判断此 领域服务类型是否存在 pool
	syncpool, ok := pool.pool[t]
	if !ok {
		return nil
	}
	// Get 其实是在 BindService 时注入的 生成 service 对象的函数
	newService := syncpool.Get()
	if newService == nil {
		panic(fmt.Sprintf("[Freedom] BindService: func return to empty, %v", t))
	}
	return newService
}

func (pool *servicePool) allType() (list []reflect.Type) {
	for t := range pool.pool {
		list = append(list, t)
	}
	return
}

func (pool *servicePool) factoryCall(rt Worker, wokrerValue reflect.Value, factoryFieldValue reflect.Value) {
	allFieldsFromValue(factoryFieldValue, func(factoryValue reflect.Value) {
		factoryKind := factoryValue.Kind()
		if factoryKind == reflect.Interface && wokrerValue.Type().AssignableTo(factoryValue.Type()) && factoryValue.CanSet() {
			//如果是运行时对象
			factoryValue.Set(wokrerValue)
			return
		}
		if !factoryValue.CanInterface() {
			return
		}
		fvi := factoryValue.Interface()
		allFieldsFromValue(factoryValue, func(value reflect.Value) {
			if !value.CanInterface() {
				return
			}
			infra := value.Interface()
			br, ok := infra.(BeginRequest)
			if ok {
				br.BeginRequest(rt)
			}
		})
		br, ok := fvi.(BeginRequest)
		if ok {
			br.BeginRequest(rt)
		}
		return
	})
}

// beginRequest .
func (pool *servicePool) beginRequest(worker Worker, obj interface{}) {
	workerValue := reflect.ValueOf(worker)
	instance := obj.(serviceElement)

	for i := 0; i < len(instance.workers); i++ {
		instance.workers[i].Set(workerValue)
	}
	for i := 0; i < len(instance.calls); i++ {
		instance.calls[i].BeginRequest(worker)
	}
}
