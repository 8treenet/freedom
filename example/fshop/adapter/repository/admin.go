package repository

import (
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/jinzhu/gorm"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建资源库函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindRepository(func() *Admin {
			return &Admin{} //创建Admin资源库
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.AdminRepo = (*Admin)(nil)

// Admin .
type Admin struct {
	freedom.Repository
}

// Get .
func (repo *Admin) Get(id int) (adminEntity *entity.Admin, e error) {
	adminEntity = &entity.Admin{}
	adminEntity.ID = id
	e = findUser(repo, adminEntity)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(adminEntity)
	return
}

func (repo *Admin) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	db = db.New()
	db.SetLogger(repo.Worker.Logger())
	return db
}
