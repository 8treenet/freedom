package internal

import (
	"reflect"
	"sync"

	"github.com/kataras/iris/v12/context"
)

// newServicePool create a ServicePool
func newServicePool() *ServicePool {
	result := new(ServicePool)
	result.instancePool = make(map[reflect.Type]*sync.Pool)
	return result
}

// ServicePool is domain service pool, a domain service type owns a pool
// instancePool key: the reflect.Type of domain service instancePool value: *sync.Pool
type ServicePool struct {
	instancePool map[reflect.Type]*sync.Pool
}

type serviceInstance struct {
	workers       []reflect.Value
	calls         []BeginRequest
	serviceObject interface{}
}

// get .
func (pool *ServicePool) get(rt *worker, service interface{}) {
	svalue := reflect.ValueOf(service)
	if svalue.Kind() != reflect.Ptr {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the service object,%v", svalue.Type())
	}
	ptr := svalue.Elem()

	var newService interface{}
	newService = pool.malloc(ptr.Type())
	if newService == nil {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the service object,%v", ptr.Type())
	}

	ptr.Set(reflect.ValueOf(newService.(serviceInstance).serviceObject))
	pool.beginRequest(rt, newService)
	rt.freeServices = append(rt.freeServices, newService)
}

// get .
func (pool *ServicePool) pop(rt Worker, service reflect.Type) interface{} {
	newService := pool.malloc(service)
	if newService == nil {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the service object,%v", service)
	}

	pool.beginRequest(rt, newService)
	return newService
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
func (pool *ServicePool) free(service interface{}) {
	t := reflect.TypeOf(service.(serviceInstance).serviceObject)
	syncpool, ok := pool.instancePool[t]
	if !ok {
		return
	}

	syncpool.Put(service)
}

func (pool *ServicePool) bind(t reflect.Type, f interface{}) {
	pool.instancePool[t] = &sync.Pool{
		New: func() interface{} {
			values := reflect.ValueOf(f).Call([]reflect.Value{})
			if len(values) == 0 {
				globalApp.Logger().Fatalf("[Freedom] BindService: func return to empty, %v", reflect.TypeOf(f))
			}

			newService := values[0].Interface()
			result := serviceInstance{serviceObject: newService, calls: []BeginRequest{}, workers: []reflect.Value{}}
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
func (pool *ServicePool) malloc(t reflect.Type) interface{} {
	// 判断此 领域服务类型是否存在 pool
	syncpool, ok := pool.instancePool[t]
	if !ok {
		return nil
	}
	// Get 其实是在 BindService 时注入的 生成 service 对象的函数
	newService := syncpool.Get()
	if newService == nil {
		globalApp.Logger().Fatalf("[Freedom] BindService: func return to empty, %v", t)
	}
	return newService
}

func (pool *ServicePool) allType() (list []reflect.Type) {
	for t := range pool.instancePool {
		list = append(list, t)
	}
	return
}

func (pool *ServicePool) factoryCall(rt Worker, wokrerValue reflect.Value, factoryFieldValue reflect.Value) {
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
func (pool *ServicePool) beginRequest(worker Worker, obj interface{}) {
	workerValue := reflect.ValueOf(worker)
	instance := obj.(serviceInstance)

	for i := 0; i < len(instance.workers); i++ {
		instance.workers[i].Set(workerValue)
	}
	for i := 0; i < len(instance.calls); i++ {
		instance.calls[i].BeginRequest(worker)
	}
}

func (pool *ServicePool) objBeginRequest2(rt Worker, obj interface{}) {
	rtValue := reflect.ValueOf(rt)
	allFields(obj, func(value reflect.Value) {
		kind := value.Kind()
		if kind == reflect.Interface && rtValue.Type().AssignableTo(value.Type()) && value.CanSet() {
			//如果是运行时对象
			value.Set(rtValue)
			return
		}
		if !value.CanInterface() {
			return
		}
		vi := value.Interface()
		//如果成员变量是工厂
		if globalApp.factoryPool.exist(value.Type()) {
			pool.factoryCall(rt, rtValue, value)
			return
		}
		allFieldsFromValue(value, func(fieldValue reflect.Value) {
			if !fieldValue.CanInterface() {
				return
			}
			infra := fieldValue.Interface()
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
