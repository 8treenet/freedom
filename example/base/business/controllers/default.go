package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/business/services"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindController("/", &DefaultController{})
	})
}

type DefaultController struct {
	Sev     *services.DefaultService
	Runtime freedom.Runtime
}

// Get handles the GET: / route.
func (c *DefaultController) Get() (result struct {
	IP string
	UA string
}, e error) {
	c.Runtime.Logger().Infof("我是控制器")
	remote := c.Sev.RemoteInfo()
	result.IP = remote.IP
	result.UA = remote.UA
	return
}
