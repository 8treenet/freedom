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

// GetOther .
func (repo *Infra) GetOther(obj interface{}) {
	globalApp.other.get(obj)
	return
}

// NewHttpRequest, transferCtx : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (c *Infra) NewHttpRequest(url string, transferCtx ...bool) Request {
	req := requests.NewHttpRequest(url)
	if len(transferCtx) > 0 && !transferCtx[0] {
		return req
	}
	HandleBusMiddleware(c.Runtime)

	bus := c.Runtime.Bus()
	req.SetHeader("x-freedom-bus", bus.ToJson())
	return req
}

// NewH2CRequest, transferCtx : Whether to pass the context, turned on by default. Typically used for tracking internal services.
func (c *Infra) NewH2CRequest(url string, transferCtx ...bool) Request {
	req := requests.NewH2CRequest(url)
	if len(transferCtx) > 0 && !transferCtx[0] {
		return req
	}
	HandleBusMiddleware(c.Runtime)

	bus := c.Runtime.Bus()
	req.SetHeader("x-freedom-bus", bus.ToJson())
	return req
}
