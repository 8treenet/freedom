package repositorys

import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *DefaultRepository {
			return &DefaultRepository{}
		})
	})
}

// DefaultRepository .
type DefaultRepository struct {
	freedom.Repository
}

// GetIP .
func (repo *DefaultRepository) GetIP() string {
	//repo.DB().Find()
	repo.Runtime.Logger().Infof("我是Repository GetIP")
	return repo.Runtime.Ctx().RemoteAddr()
}

// GetUA - implment DefaultRepoInterface interface
func (repo *DefaultRepository) GetUA() string {
	repo.Runtime.Logger().Infof("我是Repository GetUA")
	return repo.Runtime.Ctx().Request().UserAgent()
}
