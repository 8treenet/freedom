package domain

import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *DefaultV2 {
			return &DefaultV2{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *DefaultV2) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// DefaultV2 .
type DefaultV2 struct {
	Default //super
}

// RemoteInfo .
func (s *DefaultV2) RemoteInfo() (result struct {
	IP string
	Ua string
}) {
	s.Worker.Logger().Info("I'm v2 service")
	result.IP = "mock v2"
	result.Ua = "mock v2"
	return
}
