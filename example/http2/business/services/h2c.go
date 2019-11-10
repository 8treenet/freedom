package services

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/business/repositorys"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *H2cService {
			return &H2cService{}
		})
	})
}

// H2cService .
type H2cService struct {
	Runtime freedom.Runtime
	H2cRepo *repositorys.H2cRepository
}

// Get .
func (s *H2cService) Get() string {
	s.Runtime.Logger().Infof("我是H2cService")
	return s.H2cRepo.GetHello()
}
