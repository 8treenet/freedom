package project

func init() {
	content["/business/services/default.go"] = servicesTemplate()
}

func servicesTemplate() string {
	return `package services

	import (
		"github.com/8treenet/freedom"
		"{{.PackagePath}}/business/repositorys"
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

	`
}
