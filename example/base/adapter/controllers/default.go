package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/application"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//initiator.BindController("/", &DefaultController{})
		//加入中间件， 只对本控制器生效，全局中间件请在main加入。
		initiator.BindController("/", &DefaultController{}, func(ctx freedom.Context) {
			runtime := freedom.PickRuntime(ctx)
			runtime.Logger().Info("Hello middleware begin")
			ctx.Next()
			runtime.Logger().Info("Hello middleware end")
		})
	})
}

type DefaultController struct {
	Sev     *application.DefaultService
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
