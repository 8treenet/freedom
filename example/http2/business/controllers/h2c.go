package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/business/services"
	"github.com/kataras/iris"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		serFunc := func(ctx iris.Context) (m *services.H2cService) {
			initiator.GetService(ctx, &m)
			return
		}
		initiator.BindController("/", &H2cController{}, serFunc)
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
