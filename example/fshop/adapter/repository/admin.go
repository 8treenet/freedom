package repository

import (
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *Admin {
			return &Admin{}
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.AdminRepo = new(Admin)

// Admin .
type Admin struct {
	freedom.Repository
}

func (repo *Admin) Get(id int) (adminEntity *entity.Admin, e error) {
	adminEntity = &entity.Admin{}
	adminEntity.Id = id
	e = findUser(repo, adminEntity)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(adminEntity)
	return
}
