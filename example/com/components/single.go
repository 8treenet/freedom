package components

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/kataras/iris"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindComponent(true, &SingleComponent{
			StartTime: time.Now().Unix(),
		})

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

// Print .
func (s *SingleComponent) Print() {
	freedom.Logger().Info("SingleComponent 服务器启动时间", s.StartTime)
}
