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

// DBByName .
func (repo *Repository) DBByName(name string) (db *gorm.DB) {
	if _, ok := globalApp.Database.Multi[name]; !ok {
		return nil
	}
	return globalApp.Database.Multi[name].db
}

// DBCache .
func (repo *Repository) DBCache() gcache.Plugin {
	return globalApp.Database.cache
}

// DBCacheByName .
func (repo *Repository) DBCacheByName(name string) gcache.Plugin {
	if _, ok := globalApp.Database.Multi[name]; !ok {
		return nil
	}
	return globalApp.Database.Multi[name].cache
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
func (repo *Repository) NewDescOrder(field string, fields ...string) *Reorder {
	return repo.newReorder("desc", field, fields...)
}

// NewAscOrder .
func (repo *Repository) NewAscOrder(field string, fields ...string) *Reorder {
	return repo.newReorder("asc", field, fields...)
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
	if globalApp.Redis.client == nil {
		panic("Redis not installed")
	}
	return globalApp.Redis.client
}

// RedisByName .
func (repo *Repository) RedisByName(name string) redis.Cmdable {
	if _, ok := globalApp.Redis.Multi[name]; !ok {
		return nil
	}
	return globalApp.Redis.Multi[name]
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
