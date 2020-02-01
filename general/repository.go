package general

import (
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
	if globalApp.Cache.client == nil {
		panic("Redis not installed")
	}
	return globalApp.Cache.client
}

// Request .
type Request requests.Request

// NewFastRequest .
func (repo *Repository) NewFastRequest(url string) Request {
	bus := GetBus(repo.Runtime.Ctx())
	req := requests.NewFastRequest(url, bus.ToJson())

	return req
}

// NewH2CRequest .
func (repo *Repository) NewH2CRequest(url string) Request {
	bus := GetBus(repo.Runtime.Ctx())
	req := requests.NewH2CRequest(url, bus.ToJson())
	return req
}

// MadeEntity .
func (repo *Repository) MadeEntity(entity interface{}) {
	newEntity(repo.Runtime, entity)
	return
}

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := int64(5)
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		sec = v.(int64)
	}
	requests.NewFastClient(time.Duration(sec) * time.Second)
	requests.Newh2cClient(time.Duration(sec) * time.Second)
}
