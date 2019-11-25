package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/business/services"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindController("/", &H2cController{})
	})
}

type H2cController struct {
	Sev     *services.H2cService
	Runtime freedom.Runtime
}

// Get handles the GET: / route.
func (c *H2cController) Get() string {
	c.Runtime.Logger().Info("我是控制器", "Get")
	return c.Sev.Get()
}

// GetHello handles the GET: /hello route.
func (c *H2cController) GetHello() string {
	c.Runtime.Logger().Info("我是控制器", "GetHello")
	return "hello h2c"
}
