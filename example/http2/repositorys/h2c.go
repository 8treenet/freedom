package repositorys

import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *H2cRepository {
			return &H2cRepository{}
		})
	})
}

// H2cRepository .
type H2cRepository struct {
	freedom.Repository
}

// GetHello .
func (repo *H2cRepository) GetHello() string {
	repo.Runtime.Logger().Infof("我是H2cRepository")
	//通过h2c request 访问本服务 /hello
	result, _ := repo.NewH2CRequest("http://127.0.0.1:8000/hello").Get().ToString()
	return result
}
