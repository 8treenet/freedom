package services

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/com/business/repositorys"
	"github.com/8treenet/freedom/example/com/components"
	"github.com/kataras/iris"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *DefaultService {
			return &DefaultService{}
		})
		initiator.InjectController(func(ctx iris.Context) (service *DefaultService) {
			initiator.GetService(ctx, &service)
			return
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
