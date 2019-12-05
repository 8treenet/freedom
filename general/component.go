package general

import (
	"github.com/8treenet/freedom/general/requests"
	"github.com/8treenet/gcache"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

// Component .
type Component struct {
	Runtime Runtime `json:"-"`
}

// BeginRequest .子实现多态
func (c *Component) BeginRequest(rt Runtime) {
	c.Runtime = rt
}

// AddBus .
func (c *Component) AddBus(key string, component interface{}) {
	bus := GetBus(c.Runtime.Ctx())
	bus.Add(key, component)
}

// GetBus .
func (c *Component) GetBus() *Bus {
	bus := GetBus(c.Runtime.Ctx())
	return bus
}

// DB .
func (c *Component) DB() (db *gorm.DB) {
	return globalApp.Database.db
}

// DBCache .
func (c *Component) DBCache() gcache.Plugin {
	return globalApp.Database.cache
}

// Redis .
func (c *Component) Redis() redis.Cmdable {
	if globalApp.Redis.client == nil {
		panic("Redis not installed")
	}
	return globalApp.Redis.client
}

// NewFastRequest .
func (c *Component) NewFastRequest(url string) Request {
	bus := GetBus(c.Runtime.Ctx())
	req := requests.NewFastRequest(url, bus.toJson())

	return req
}

// NewH2CRequest .
func (c *Component) NewH2CRequest(url string) Request {
	bus := GetBus(c.Runtime.Ctx())
	req := requests.NewH2CRequest(url, bus.toJson())
	return req
}
