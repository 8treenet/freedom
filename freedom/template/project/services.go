package project

func init() {
	content["/application/default.go"] = servicesTemplate()
	content["/application/interface.go"] = servicesInterfaceTemplate()
}

func servicesInterfaceTemplate() string {
	return `package application`
}

func servicesTemplate() string {
	return `package application

	import (
		"github.com/8treenet/freedom"
		"{{.PackagePath}}/adapter/repositorys"
	)
	
	func init() {
		freedom.Booting(func(initiator freedom.Initiator) {
			initiator.BindService(func() *DefaultService {
				return &DefaultService{}
			})
			initiator.InjectController(func(ctx freedom.Context) (service *DefaultService) {
				initiator.GetService(ctx, &service)
				return
			})
		})
	}
	
	// DefaultService .
	type DefaultService struct {
		Runtime   freedom.Runtime
		DefRepo   *repositorys.DefaultRepository
		DefRepoIF repositorys.DefaultRepoInterface
	}
	
	// RemoteInfo .
	func (s *DefaultService) RemoteInfo() (result struct {
		Ip string
		Ua string
	}) {
		s.Runtime.Logger().Infof("我是service")
		result.Ip = s.DefRepo.GetIP()
		result.Ua = s.DefRepoIF.GetUA()
		return
	}

	`
}
