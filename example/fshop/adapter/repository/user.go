package repository

import (
	"time"

	"github.com/8treenet/freedom/example/fshop/adapter/dto"
	"github.com/8treenet/freedom/example/fshop/adapter/po"
	"github.com/8treenet/freedom/example/fshop/domain/entity"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *User {
			return &User{}
		})
	})
}

var _ UserRepo = new(User)

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
