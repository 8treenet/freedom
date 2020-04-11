package general

import (
	"reflect"
	"time"

	"github.com/8treenet/freedom/general/requests"
	"github.com/go-redis/redis"
	"github.com/kataras/iris"

	"github.com/jinzhu/gorm"
)

// Repository .
type Repository struct {
	Runtime Runtime
}

// BeginRequest .
func (repo *Repository) BeginRequest(rt Runtime) {
	repo.Runtime = rt
}

// DB .
func (repo *Repository) DB() (db *gorm.DB) {
	transactionData := repo.Runtime.Store().Get("freedom_local_transaction_db")
	if transactionData != nil {
		return transactionData.(*gorm.DB)
	}
	return globalApp.Database.db
}

// NewORMDescBuilder .
func (repo *Repository) NewORMDescBuilder(column string, columns ...string) *Reorder {
	return repo.newReorder("desc", column, columns...)
}

// NewORMAscBuilder .
func (repo *Repository) NewORMAscBuilder(column string, columns ...string) *Reorder {
	return repo.newReorder("asc", column, columns...)
}

// NewORMBuilder .
func (repo *Repository) NewORMBuilder() *Builder {
	return &Builder{}
}

// NewDescOrder .
func (repo *Repository) newReorder(sort, field string, args ...string) *Reorder {
	fields := []string{field}
	fields = append(fields, args...)
	orders := []string{}
	for index := 0; index < len(fields); index++ {
		orders = append(orders, sort)
	}
	return &Reorder{
		fields: fields,
		orders: orders,
	}
}

// Redis .
func (repo *Repository) Redis() redis.Cmdable {
	if globalApp.Cache.client == nil {
		panic("Redis not installed")
	}
	return globalApp.Cache.client
}

// Request .
type Request requests.Request

// NewHttpRequest, transferCtx : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (repo *Repository) NewHttpRequest(url string, transferCtx ...bool) Request {
	req := requests.NewHttpRequest(url)
	if len(transferCtx) > 0 && !transferCtx[0] {
		return req
	}
	HandleBusMiddleware(repo.Runtime)

	bus := repo.Runtime.Bus()
	req.SetHeader("x-freedom-bus", bus.ToJson())

	m := globalApp.Iris().ConfigurationReadOnly().GetOther()
	if serviceName, ok := m["service_name"]; ok {
		str, strok := serviceName.(string)
		if strok {
			req.SetHeader("user-agent", str)
		}
	}
	return req
}

// NewH2CRequest, transferCtx : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (repo *Repository) NewH2CRequest(url string, transferCtx ...bool) Request {
	req := requests.NewH2CRequest(url)
	if len(transferCtx) > 0 && !transferCtx[0] {
		return req
	}
	HandleBusMiddleware(repo.Runtime)

	bus := repo.Runtime.Bus()
	req.SetHeader("x-freedom-bus", bus.ToJson())

	m := globalApp.Iris().ConfigurationReadOnly().GetOther()
	if serviceName, ok := m["service_name"]; ok {
		str, strok := serviceName.(string)
		if strok {
			req.SetHeader("user-agent", str)
		}
	}
	return req
}

// // SingleFlight .
// func (repo *Repository) SingleFlight(key string, value, takeObject interface{}, fn func() (interface{}, error)) error {
// 	takeValue := reflect.ValueOf(takeObject)
// 	if takeValue.Kind() != reflect.Ptr {
// 		panic("'takeObject' must be a pointer")
// 	}
// 	takeValue = takeValue.Elem()
// 	if !takeValue.CanSet() {
// 		panic("'takeObject' cannot be set")
// 	}
// 	v, err, _ := globalApp.singleFlight.Do(key+"-"+fmt.Sprint(value), fn)
// 	if err != nil {
// 		return err
// 	}

// 	newValue := reflect.ValueOf(v)
// 	if takeValue.Type() != newValue.Type() {
// 		panic("'takeObject' type error")
// 	}
// 	takeValue.Set(reflect.ValueOf(v))
// 	return nil
// }

// InjectBaseEntity .
func (repo *Repository) InjectBaseEntity(entity Entity) {
	injectBaseEntity(repo.Runtime, entity)
	return
}

// InjectBaseEntity .
func (repo *Repository) InjectBaseEntitys(entitys interface{}) {
	entitysValue := reflect.ValueOf(entitys)
	if entitysValue.Kind() != reflect.Slice {
		panic("It's not a slice")
	}
	for i := 0; i < entitysValue.Len(); i++ {
		iface := entitysValue.Index(i).Interface()
		if _, ok := iface.(Entity); !ok {
			panic("This is not an entity")
		}
		injectBaseEntity(repo.Runtime, iface)
	}
	return
}

// GetOther .
func (repo *Repository) GetOther(obj interface{}) {
	globalApp.other.get(obj)
	return
}

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := int64(5)
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		sec = v.(int64)
	}
	requests.NewHttpClient(time.Duration(sec) * time.Second)
	requests.NewH2cClient(time.Duration(sec) * time.Second)
}

// GetRuntime .
func (repo *Repository) GetRuntime() Runtime {
	return repo.Runtime
}
