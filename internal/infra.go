package internal

import (
	"fmt"
	"reflect"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/go-redis/redis"
)

// Infra .
type Infra struct {
	worker Worker `json:"-"`
	single bool
}

// BeginRequest .子实现多态
func (infra *Infra) BeginRequest(rt Worker) {
	if infra.single {
		return
	}
	infra.worker = rt
}

// SourceDB .
func (infra *Infra) SourceDB() (db interface{}) {
	return globalApp.Database.db
}

// Redis .
func (infra *Infra) Redis() redis.Cmdable {
	return globalApp.Cache.client
}

// Other .
func (infra *Infra) Other(obj interface{}) {
	globalApp.oneShotPool.get(obj)
	return
}

// NewHTTPRequest transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (infra *Infra) NewHTTPRequest(url string, transferBus ...bool) requests.Request {
	req := requests.NewHTTPRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	if infra.worker == nil {
		//The singleton object does not have a Worker component
		return req
	}
	req.SetHeader(infra.worker.Bus().Header)
	return req
}

// NewH2CRequest transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (infra *Infra) NewH2CRequest(url string, transferBus ...bool) requests.Request {
	req := requests.NewH2CRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	if infra.worker == nil {
		//The singleton object does not have a Worker component
		return req
	}
	req.SetHeader(infra.worker.Bus().Header)
	return req
}

// InjectBaseEntity .
func (infra *Infra) InjectBaseEntity(entity Entity) {
	if infra.worker == nil {
		panic("The singleton object does not have a Worker component")
	}
	injectBaseEntity(infra.worker, entity)
	return
}

// InjectBaseEntitys .
func (infra *Infra) InjectBaseEntitys(entitys interface{}) {
	if infra.worker == nil {
		panic("The singleton object does not have a Worker component")
	}
	entitysValue := reflect.ValueOf(entitys)
	if entitysValue.Kind() != reflect.Slice {
		panic(fmt.Sprintf("[Freedom] InjectBaseEntitys: It's not a slice, %v", entitysValue.Type()))
	}
	for i := 0; i < entitysValue.Len(); i++ {
		iface := entitysValue.Index(i).Interface()
		if _, ok := iface.(Entity); !ok {
			panic(fmt.Sprintf("[Freedom] InjectBaseEntitys: This is not an entity, %v", entitysValue.Type()))
		}
		injectBaseEntity(infra.worker, iface)
	}
	return
}

// Worker .
func (infra *Infra) Worker() Worker {
	if infra.worker == nil {
		panic("The singleton object does not have a Worker component")
	}
	return infra.worker
}

// GetSingleInfra .
func (infra *Infra) GetSingleInfra(com interface{}) bool {
	return globalApp.GetSingleInfra(com)
}

func (infra *Infra) setSingle() {
	infra.single = true
}
