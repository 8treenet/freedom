package general

import (
	"errors"
	"fmt"
	"time"

	"github.com/8treenet/freedom/general/requests"
	"github.com/8treenet/gcache"
	"github.com/go-redis/redis"
	"github.com/kataras/iris"

	"github.com/jinzhu/gorm"
)

// Repository .
type Repository struct {
	Runtime       Runtime
	transactionDB *gorm.DB
}

// BeginRequest .
func (repo *Repository) BeginRequest(rt Runtime) {
	repo.Runtime = rt
}

// DB .
func (repo *Repository) DB() (db *gorm.DB) {
	if repo.transactionDB != nil {
		return repo.transactionDB
	}

	return globalApp.Database.db
}

// DBCache .
func (repo *Repository) DBCache() gcache.Plugin {
	return globalApp.Database.cache
}

// Transaction .
func (repo *Repository) Transaction(fun func() error) (e error) {
	repo.transactionDB = globalApp.Database.db.Begin()

	defer func() {
		if perr := recover(); perr != nil {
			repo.transactionDB.Rollback()
			repo.transactionDB = nil
			e = errors.New(fmt.Sprint(perr))
			return
		}

		deferdb := repo.transactionDB
		repo.transactionDB = nil
		if e != nil {
			e = deferdb.Rollback().Error
			return
		}
		e = deferdb.Commit().Error
	}()

	e = fun()
	return
}

// NewDescOrder .
func (repo *Repository) NewDescOrder(orderField string, orderFields ...string) *Order {
	fields := []string{orderField}
	fields = append(fields, orderFields...)
	return &Order{
		orderFields: fields,
		order:       "desc",
	}
}

// NewAscOrder .
func (repo *Repository) NewAscOrder(orderField string, orderFields ...string) *Order {
	fields := []string{orderField}
	fields = append(fields, orderFields...)
	return &Order{
		orderFields: fields,
		order:       "asc",
	}
}

// Redis .
func (repo *Repository) Redis() *redis.Client {
	if globalApp.Redis.client == nil {
		panic("Redis not installed")
	}
	return globalApp.Redis.client
}

// Request .
type Request requests.Request

// NewFastRequest .
func (repo *Repository) NewFastRequest(url string) Request {
	bus := GetBus(repo.Runtime.Ctx())
	req := requests.NewFastRequest(url, bus.toJson())

	return req
}

// NewH2CRequest .
func (repo *Repository) NewH2CRequest(url string) Request {
	bus := GetBus(repo.Runtime.Ctx())
	req := requests.NewH2CRequest(url, bus.toJson())
	return req
}

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := int64(5)
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		sec = v.(int64)
	}
	requests.NewFastClient(time.Duration(sec) * time.Second)
	requests.Newh2cClient(time.Duration(sec) * time.Second)
}
