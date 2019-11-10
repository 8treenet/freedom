package services

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/com/business/repositorys"
	"github.com/8treenet/freedom/example/com/components"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *DefaultService {
			return &DefaultService{}
		})
	})
}

// DefaultService .
type DefaultService struct {
	Runtime   freedom.Runtime
	DefRepo   *repositorys.DefaultRepository
	Defcom    *components.DefaultComponent
	SingleCom *components.SingleComponent
}

// Get .
func (s *DefaultService) Get() string {
	s.Defcom.Print()
	s.SingleCom.Print()
	return s.DefRepo.Get()
}
