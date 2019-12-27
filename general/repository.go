package general

import (
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
	Runtime         Runtime
	selectDBName    string
	selectRedisName string
}

// BeginRequest .
func (repo *Repository) BeginRequest(rt Runtime) {
	repo.Runtime = rt
	repo.selectDBName = ""
}

// DB .
func (repo *Repository) DB() (db *gorm.DB) {
	transactionData := repo.Runtime.Store().Get("freedom_local_transaction_db")
	for {
		if transactionData != nil {
			m := transactionData.(map[string]interface{})
			name := m["name"].(string)
			transactionDB := m["db"].(*gorm.DB)
			if name == repo.selectDBName {
				db = transactionDB
				return
			}
		}
		if repo.selectDBName != "" {
			return globalApp.Database.Multi[repo.selectDBName].db
		}
		break
	}
	return globalApp.Database.db
}

// SelectDB .
func (repo *Repository) SelectDB(name string) *Repository {
	if _, ok := globalApp.Database.Multi[name]; !ok {
		panic(fmt.Sprintf("db '%s' does not exist", name))
	}

	newRepository := new(Repository)
	newRepository.Runtime = repo.Runtime
	newRepository.selectDBName = name
	newRepository.selectRedisName = repo.selectRedisName
	return newRepository
}

// SelectDB .
func (repo *Repository) SelectRedis(name string) *Repository {
	if _, ok := globalApp.Redis.Multi[name]; !ok {
		panic(fmt.Sprintf("redis '%s' does not exist", name))
	}

	newRepository := new(Repository)
	newRepository.Runtime = repo.Runtime
	newRepository.selectDBName = repo.selectDBName
	newRepository.selectRedisName = name
	return newRepository
}

// DBCache .
func (repo *Repository) DBCache() gcache.Plugin {
	if repo.selectDBName != "" {
		return globalApp.Database.Multi[repo.selectDBName].cache
	}
	return globalApp.Database.cache
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

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := int64(5)
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		sec = v.(int64)
	}
	requests.NewFastClient(time.Duration(sec) * time.Second)
	requests.Newh2cClient(time.Duration(sec) * time.Second)
}
