package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/com/components"
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
	Defcom    *components.DefaultComponent
	SingleCom *components.SingleComponent
}

// Get .
func (repo *DefaultRepository) Get() string {
	repo.Defcom.Print()
	repo.SingleCom.Print()
	req := repo.NewFastRequest("http://127.0.0.1:8000/com").Get()
	go func() {
		req.ToString()
	}()
	return repo.Defcom.GetValue()
}
