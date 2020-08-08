package repository

import (
	"time"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/jinzhu/gorm"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *User {
			return &User{}
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.UserRepo = new(User)

// User .
type User struct {
	freedom.Repository
}

func (repo *User) Get(id int) (userEntity *entity.User, e error) {
	userEntity = &entity.User{}
	userEntity.Id = id
	e = findUser(repo, userEntity)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(userEntity)
	return
}

func (repo *User) FindByName(userName string) (userEntity *entity.User, e error) {
	userEntity = &entity.User{User: po.User{Name: userName}}
	e = findUser(repo, userEntity)
	if e != nil {
		return
	}

	repo.InjectBaseEntity(userEntity)
	return
}

func (repo *User) Save(entity *entity.User) error {
	_, e := saveUser(repo, &entity.User)
	return e
}

func (repo *User) New(userDto dto.RegisterUserReq, money int) (entityUser *entity.User, e error) {
	user := po.User{Name: userDto.Name, Money: money, Password: userDto.Password, Created: time.Now(), Updated: time.Now()}

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
	db.SetLogger(repo.Worker.Logger())
	return db
}
