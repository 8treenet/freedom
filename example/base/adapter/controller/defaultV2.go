package controller

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/domain"
	"github.com/8treenet/freedom/example/base/infra"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/v2", &DefaultV2{}, func(ctx freedom.Context) {
			worker := freedom.ToWorker(ctx)
			worker.Logger().Info("Hello middleware v2 begin")
			ctx.Next()
			worker.Logger().Info("Hello middleware v2 end")
		})
	})
}

// DefaultV2 .
type DefaultV2 struct {
	Default      //super
	DefServiceV2 *domain.DefaultV2
}

// Get handles the GET: / route.
func (c *DefaultV2) Get() freedom.Result {
	c.Worker.Logger().Info("I'm v2 Controller")
	remote := c.DefServiceV2.RemoteInfo()
	return &infra.JSONResponse{Object: remote}
}

// GetHello handles the GET: /hello route.
func (c *DefaultV2) GetHello() string {
	field := freedom.LogFields{
		"framework": "freedom",
		"like":      "DDD",
	}
	c.Worker.Logger().Info("hello v2", field)
	c.Worker.Logger().Infof("hello v2 %s", "format", field)
	c.Worker.Logger().Debug("hello v2", field)
	c.Worker.Logger().Debugf("hello v2 %s", "format", field)
	c.Worker.Logger().Error("hello v2", field)
	c.Worker.Logger().Errorf("hello v2 %s", "format", field)
	c.Worker.Logger().Warn("hello v2", field)
	c.Worker.Logger().Warnf("hello v2 %s", "format", field)
	c.Worker.Logger().Print("hello v2")
	c.Worker.Logger().Println("hello v2")
	return "hello v2"
}
