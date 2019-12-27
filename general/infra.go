package general

import (
	"github.com/8treenet/freedom/general/requests"
	"github.com/8treenet/gcache"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

// Infra .
type Infra struct {
	Runtime Runtime `json:"-"`
}

// BeginRequest .子实现多态
func (c *Infra) BeginRequest(rt Runtime) {
	c.Runtime = rt
}

// DB .
func (c *Infra) DB(name ...string) (db *gorm.DB) {
	if len(name) != 0 {
		if _, ok := globalApp.Database.Multi[name[0]]; !ok {
			return nil
		}
		return globalApp.Database.Multi[name[0]].db
	}
	return globalApp.Database.db
}

// DBCache .
func (c *Infra) DBCache(name ...string) gcache.Plugin {
	if len(name) != 0 {
		if _, ok := globalApp.Database.Multi[name[0]]; !ok {
			return nil
		}
		return globalApp.Database.Multi[name[0]].cache
	}
	return globalApp.Database.cache
}

// Redis .
func (c *Infra) Redis(name ...string) redis.Cmdable {
	if len(name) != 0 {
		if _, ok := globalApp.Redis.Multi[name[0]]; !ok {
			return nil
		}
		return globalApp.Redis.Multi[name[0]]
	}
	return globalApp.Redis.client
}

// NewFastRequest .
func (c *Infra) NewFastRequest(url string) Request {
	bus := GetBus(c.Runtime.Ctx())
	req := requests.NewFastRequest(url, bus.ToJson())
	return req
}

// NewH2CRequest .
func (c *Infra) NewH2CRequest(url string) Request {
	bus := GetBus(c.Runtime.Ctx())
	req := requests.NewH2CRequest(url, bus.ToJson())
	return req
}
