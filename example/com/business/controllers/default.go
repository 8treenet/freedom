package controllers

import (
	"fmt"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/com/business/services"
	"github.com/8treenet/freedom/example/com/components"
	"github.com/kataras/iris"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		serFunc := func(ctx iris.Context) (m *services.DefaultService) {
			initiator.GetService(ctx, &m)
			return
		}
		comFunc := func(ctx iris.Context) (c *components.DefaultComponent) {
			initiator.GetComponent(ctx, &c)
			return
		}
		singleFunc := func(ctx iris.Context) (s *components.SingleComponent) {
			initiator.GetComponent(ctx, &s)
			return
		}
		initiator.BindController("/", &DefaultController{}, serFunc, comFunc, singleFunc)
	})
}

type DefaultController struct {
	Sev       *services.DefaultService
	SingleCom *components.SingleComponent
	Defcom    *components.DefaultComponent
	Runtime   freedom.Runtime
}

// Get handles the GET: / route.
func (c *DefaultController) Get() (string, error) {
	c.Runtime.Logger().Infof("我是控制器Get")

	value := fmt.Sprintf("Good freedom, 多层共用组件 time:%d", time.Now().Unix())
	c.Defcom.SetValue(value)

	c.SingleCom.Print()
	return c.Sev.Get(), nil
}

// GetCom handles the GET: /com route.
func (c *DefaultController) GetCom() (string, error) {
	c.Runtime.Logger().Info("我是控制器GetCom", "get设置的组件数据：", c.Defcom.Value)
	return "", nil
}
