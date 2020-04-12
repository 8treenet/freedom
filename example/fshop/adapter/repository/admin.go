package repository

import (
	"github.com/8treenet/freedom/example/fshop/application/entity"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *Admin {
			return &Admin{}
		})
	})
}

var _ AdminRepo = new(Admin)

// Admin .
type Admin struct {
	freedom.Repository
}

func (repo *Admin) Get(id int) (adminEntity *entity.Admin, e error) {
	adminEntity = &entity.Admin{}
	e = findUserByPrimary(repo, adminEntity, id)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(adminEntity)
	return
}
