package general

import (
	"github.com/8treenet/freedom/general/requests"
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
func (c *Infra) DB() (db *gorm.DB) {
	return globalApp.Database.db
}

// Redis .
func (c *Infra) Redis() redis.Cmdable {
	return globalApp.Cache.client
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
