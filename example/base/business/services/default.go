package services

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/business/repositorys"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *DefaultService {
			return &DefaultService{}
		})
	})
}

// DefaultRepoInterface .
type DefaultRepoInterface interface {
	GetUA() string
}

// DefaultService .
type DefaultService struct {
	Runtime   freedom.Runtime
	DefRepo   *repositorys.DefaultRepository
	DefRepoIF DefaultRepoInterface
}

// RemoteInfo .
func (s *DefaultService) RemoteInfo() (result struct {
	IP string
	UA string
}) {
	s.Runtime.Logger().Infof("我是service")
	result.IP = s.DefRepo.GetIP()
	result.UA = s.DefRepoIF.GetUA()
	return
}
