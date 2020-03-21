package application

import (
	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/application/dto"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *User {
			return &User{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *User) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// User 用户领域服务.
type User struct {
	Runtime  freedom.Runtime     //运行时，一个请求绑定一个运行时
	UserRepo repository.UserRepo //用户仓库
}

// ChangePassword 修改密码
func (s *User) ChangePassword(userId int, newPassword, oldPassword string) (e error) {
	//使用用户仓库读取用户实体
	userEntity, e := s.UserRepo.Find(userId)
	if e != nil {
		return
	}

	//修改密码
	if e = userEntity.ChangePassword(newPassword, oldPassword); e != nil {
		return
	}

	//使用用户仓库持久化实体
	e = s.UserRepo.Save(userEntity)
	return
}

// Register .
func (s *User) Register(user dto.RegisterUserReq) (result dto.UserInfoRes, e error) {
	userEntity, e := s.UserRepo.New(user, 10000)
	if e != nil {
		return
	}
	result.Id = userEntity.Id
	result.Money = userEntity.Money
	result.Name = userEntity.Name
	return
}

// Get .
func (s *User) Get(userId int) (result dto.UserInfoRes, e error) {
	userEntity, e := s.UserRepo.Find(userId)
	if e != nil {
		return
	}
	result.Id = userEntity.Id
	result.Money = userEntity.Money
	result.Name = userEntity.Name
	return
}
