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
		"{{.PackagePath}}/adapter/repository"
	)
	
	func init() {
		freedom.Prepare(func(initiator freedom.Initiator) {
			initiator.BindService(func() *Default {
				return &Default{}
			})
			initiator.InjectController(func(ctx freedom.Context) (service *Default) {
				initiator.GetService(ctx, &service)
				return
			})
		})
	}
	
	// Default .
	type Default struct {
		Runtime   freedom.Runtime
		DefRepo   *repository.Default
		DefRepoIF repository.DefaultRepoInterface
	}
	
	// RemoteInfo .
	func (s *Default) RemoteInfo() (result struct {
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
