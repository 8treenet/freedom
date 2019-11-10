package components

import (
	"time"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindComponent(true, &SingleComponent{
			StartTime: time.Now().Unix(),
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
