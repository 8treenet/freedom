package components

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/kataras/iris"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindComponent(true, &SingleComponent{})

		initiator.InjectController(func(ctx iris.Context) (com *SingleComponent) {
			initiator.GetComponent(ctx, &com)
			return
		})
	})
}

// SingleComponent . 单例组件
type SingleComponent struct {
	StartTime int64
}

// Booting 单例组件 http listen 前会调用.
func (s *SingleComponent) Booting(sb freedom.SingleBoot) {
	s.StartTime = time.Now().Unix()
	freedom.Logger().Info("SingleComponent Booting", s.StartTime)
}

// Print .
func (s *SingleComponent) Print() {
	freedom.Logger().Info("SingleComponent 服务器启动时间", s.StartTime)
}
