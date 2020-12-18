package internal

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/8treenet/freedom/infra/requests"
	"github.com/go-redis/redis"
	iris "github.com/kataras/iris/v12"
)

// Repository .
type Repository struct {
	worker Worker
}

// BeginRequest .
func (repo *Repository) BeginRequest(rt Worker) {
	repo.worker = rt
}

// FetchDB .
func (repo *Repository) FetchDB(db interface{}) error {
	resultDB := globalApp.Database.db

	transactionData := repo.worker.Store().Get("freedom_local_transaction_db")
	if transactionData != nil {
		resultDB = transactionData
	}
	if resultDB == nil {
		return errors.New("DB not found, please install")
	}
	if !fetchValue(db, resultDB) {
		return errors.New("DB not found, please install")
	}
	// db = globalApp.Database.db.New()
	// db.SetLogger(repo.Worker.Logger())
	return nil
}

// FetchSourceDB .
func (repo *Repository) FetchSourceDB(db interface{}) error {
	resultDB := globalApp.Database.db
	if resultDB == nil {
		return errors.New("DB not found, please install")
	}
	if !fetchValue(db, resultDB) {
		return errors.New("DB not found, please install")
	}
	return nil
}

// Redis .
func (repo *Repository) Redis() redis.Cmdable {
	return globalApp.Cache.client
}

// NewHTTPRequest transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (repo *Repository) NewHTTPRequest(url string, transferBus ...bool) requests.Request {
	req := requests.NewHTTPRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	req.SetHeader(repo.worker.Bus().Header)
	return req
}

// NewH2CRequest transferBus : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (repo *Repository) NewH2CRequest(url string, transferBus ...bool) requests.Request {
	req := requests.NewH2CRequest(url)
	if len(transferBus) > 0 && !transferBus[0] {
		return req
	}
	req.SetHeader(repo.worker.Bus().Header)
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
	injectBaseEntity(repo.worker, entity)
	return
}

// InjectBaseEntitys .
func (repo *Repository) InjectBaseEntitys(entitys interface{}) {
	entitysValue := reflect.ValueOf(entitys)
	if entitysValue.Kind() != reflect.Slice {
		panic(fmt.Sprintf("[Freedom] InjectBaseEntitys: It's not a slice, %v", entitysValue.Type()))
	}
	for i := 0; i < entitysValue.Len(); i++ {
		iface := entitysValue.Index(i).Interface()
		if _, ok := iface.(Entity); !ok {
			panic(fmt.Sprintf("[Freedom] InjectBaseEntitys: This is not an entity, %v", entitysValue.Type()))
		}
		injectBaseEntity(repo.worker, iface)
	}
	return
}

// Other .
func (repo *Repository) Other(obj interface{}) {
	globalApp.other.get(obj)
	return
}

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := int64(5)
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		sec = v.(int64)
	}
	requests.InitHTTPClient(time.Duration(sec) * time.Second)
	requests.InitH2cClient(time.Duration(sec) * time.Second)
}

// Worker .
func (repo *Repository) Worker() Worker {
	return repo.worker
}
