package internal

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/8treenet/freedom/infra/requests"
	iris "github.com/8treenet/iris/v12"
	"github.com/go-redis/redis"
)

const (
	//TransactionKey for transaction DB handles in the first-level cache
	TransactionKey = "freedom_local_transaction_db"
)

// Repository .
// The parent class of the repository, which can be inherited using the parent class's methods.
type Repository struct {
	worker Worker
}

// BeginRequest Polymorphic method, subclasses can override overrides overrides.
// The request is triggered after entry.
func (repo *Repository) BeginRequest(rt Worker) {
	repo.worker = rt
}

// FetchDB Gets the installed database handle.
// DB can be changed through AOP, such as transaction processing.
func (repo *Repository) FetchDB(db interface{}) error {
	resultDB := globalApp.Database.db

	transactionData := repo.worker.Store().Get(TransactionKey)
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

// FetchOnlyDB Gets the installed database handle.
func (repo *Repository) FetchOnlyDB(db interface{}) error {
	resultDB := globalApp.Database.db
	if resultDB == nil {
		return errors.New("DB not found, please install")
	}
	if !fetchValue(db, resultDB) {
		return errors.New("DB not found, please install")
	}
	return nil
}

// Redis Gets the installed redis client.
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

// InjectBaseEntity The base class object that is injected into the entity.
func (repo *Repository) InjectBaseEntity(entity Entity) {
	injectBaseEntity(repo.worker, entity)
	return
}

// InjectBaseEntitys The base class object that is injected into the entity.
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

// FetchCustom Gets the installed custom data sources.
func (repo *Repository) FetchCustom(obj interface{}) {
	globalApp.other.get(obj)
	return
}

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := 5
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		i, err := strconv.Atoi(fmt.Sprint(v))
		if err != nil {
			panic(err)
		}
		sec = i
	}
	requests.InitHTTPClient(time.Duration(sec) * time.Second)
	requests.InitH2CClient(time.Duration(sec) * time.Second)
}

// Worker Returns the requested Worer.
func (repo *Repository) Worker() Worker {
	return repo.worker
}
