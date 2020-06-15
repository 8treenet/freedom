package internal

import (
	"reflect"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

// Infra .
type Infra struct {
	Worker Worker `json:"-"`
}

// BeginRequest .子实现多态
func (c *Infra) BeginRequest(rt Worker) {
	c.Worker = rt
}

// DB .
func (c *Infra) DB() (db *gorm.DB) {
	return globalApp.Database.db
}

// Redis .
func (c *Infra) Redis() redis.Cmdable {
	return globalApp.Cache.client
}

// GetOther .
func (infra *Infra) Other(obj interface{}) {
	globalApp.other.get(obj)
	return
}

// NewHttpRequest, transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (infra *Infra) NewHttpRequest(url string, transferBus ...bool) requests.Request {
	req := requests.NewHttpRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	req.SetHeader(infra.Worker.Bus().Header)
	return req
}

// NewH2CRequest, transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (infra *Infra) NewH2CRequest(url string, transferBus ...bool) requests.Request {
	req := requests.NewH2CRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	req.SetHeader(infra.Worker.Bus().Header)
	return req
}

// InjectBaseEntity .
func (infra *Infra) InjectBaseEntity(entity Entity) {
	injectBaseEntity(infra.Worker, entity)
	return
}

// InjectBaseEntity .
func (infra *Infra) InjectBaseEntitys(entitys interface{}) {
	entitysValue := reflect.ValueOf(entitys)
	if entitysValue.Kind() != reflect.Slice {
		panic("It's not a slice")
	}
	for i := 0; i < entitysValue.Len(); i++ {
		iface := entitysValue.Index(i).Interface()
		if _, ok := iface.(Entity); !ok {
			panic("This is not an entity")
		}
		injectBaseEntity(infra.Worker, iface)
	}
	return
}

// GetWorker .
func (infra *Infra) GetWorker() Worker {
	return infra.Worker
}
