package repository

import (
	"time"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/8treenet/freedom/example/fshop/domain/vo"
	"github.com/8treenet/freedom/example/fshop/infra/domainevent"
	"github.com/jinzhu/gorm"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建资源库函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindRepository(func() *User {
			return &User{} //创建User资源库
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.UserRepo = (*User)(nil)

// User .
type User struct {
	freedom.Repository
	EventRepository *domainevent.EventManager //领域事件组件
}

// Get .
func (repo *User) Get(ID int) (userEntity *entity.User, e error) {
	userEntity = &entity.User{}
	userEntity.ID = ID
	e = findUser(repo, userEntity)
	if e != nil {
		return
	}

	//注入基础Entity
	repo.InjectBaseEntity(userEntity)
	return
}

// FindByName .
func (repo *User) FindByName(userName string) (userEntity *entity.User, e error) {
	userEntity = &entity.User{User: po.User{Name: userName}}
	e = findUser(repo, userEntity)
	if e != nil {
		return
	}

	repo.InjectBaseEntity(userEntity)
	return
}

// Save .
func (repo *User) Save(entity *entity.User) error {
	_, e := saveUser(repo, &entity.User)
	if e != nil {
		return e
	}
	return repo.EventRepository.Save(&repo.Repository, entity)
}

// New .
func (repo *User) New(uservo vo.RegisterUserReq, money int) (entityUser *entity.User, e error) {
	user := po.User{Name: uservo.Name, Money: money, Password: uservo.Password, Created: time.Now(), Updated: time.Now()}

	_, e = createUser(repo, &user)
	if e != nil {
		return
	}
	entityUser = &entity.User{User: user}
	repo.InjectBaseEntity(entityUser)
	return
}

func (repo *User) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	db = db.New()
	db.SetLogger(repo.Worker().Logger())
	return db
}
