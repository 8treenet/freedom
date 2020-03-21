package repository

import (
	"time"

	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/example/fshop/application/entity"
	"github.com/8treenet/freedom/example/fshop/application/object"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
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

func (repo *User) Find(id int) (userEntity *entity.User, e error) {
	userEntity = &entity.User{}
	e = findUserByPrimary(repo, userEntity, id)
	return
}

func (repo *User) FindByName(userName string) (userEntity *entity.User, e error) {
	userEntity = &entity.User{}
	e = findUser(repo, object.User{Name: userName}, userEntity)
	return
}

func (repo *User) Save(entity *entity.User) error {
	_, e := saveUser(repo, &entity.User)
	return e
}

func (repo *User) New(userDto dto.RegisterUserReq, money int) (entityUser *entity.User, e error) {
	user := object.User{Name: userDto.Name, Money: money, Password: userDto.Password, Created: time.Now(), Updated: time.Now()}

	_, e = createUser(repo, &user)
	if e != nil {
		return
	}
	entityUser = &entity.User{User: user}
	repo.MadeEntity(entityUser)
	return
}
