package project

func init() {
	content["/adapter/repository/default.go"] = repositoryTemplate()
	content["/adapter/repository/interface.go"] = repositoryInterfaceTemplate()
}

func repositoryInterfaceTemplate() string {
	return `package repository
// DefaultRepoInterface .
type DefaultRepoInterface interface {
	GetUA() string
}
`
}

func repositoryTemplate() string {
	return `package repository

	import (
		"github.com/8treenet/freedom"
	)
	
	func init() {
		freedom.Prepare(func(initiator freedom.Initiator) {
			initiator.BindRepository(func() *Default {
				return &Default{}
			})
		})
	}
	
	// Default .
	type Default struct {
		freedom.Repository
	}
	
	// GetIP .
	func (repo *Default) GetIP() string {
		//repo.DB().Find()
		repo.Runtime.Logger().Infof("我是Repository GetIP")
		return repo.Runtime.Ctx().RemoteAddr()
	}
	
	// GetUA - implment DefaultRepoInterface interface
	func (repo *Default) GetUA() string {
		repo.Runtime.Logger().Infof("我是Repository GetUA")
		return repo.Runtime.Ctx().Request().UserAgent()
	}
	
	`
}
