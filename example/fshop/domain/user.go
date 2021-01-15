package domain

import (
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/vo"
	"github.com/8treenet/freedom/example/fshop/infra/domainevent"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建领域服务函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindService(func() *User {
			return &User{} //创建User领域服务
		})
		//控制器客户需要明确使用 InjectController
		initiator.InjectController(func(ctx freedom.Context) (service *User) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// User 用户领域服务.
type User struct {
	Worker      freedom.Worker                //运行时，一个请求绑定一个运行时
	UserRepo    dependency.UserRepo           //依赖倒置用户资源库
	Transaction *domainevent.EventTransaction //依赖注入事务组件
}

// ChangePassword 修改密码
func (user *User) ChangePassword(userID int, newPassword, oldPassword string) (e error) {
	//使用用户仓库读取用户实体
	userEntity, e := user.UserRepo.Get(userID)
	if e != nil {
		return
	}

	//修改密码
	if e = userEntity.ChangePassword(newPassword, oldPassword); e != nil {
		return
	}

	//使用事务组件保证一致性 1.修改密码属性, 2.事件表增加记录
	//Execute 如果返回错误 会触发回滚。成功会调用infra/domainevent/EventManager.push
	e = user.Transaction.Execute(func() error {
		return user.UserRepo.Save(userEntity)
	})
	user.Worker.Logger().Infof("ChangePassword newPassword:%s oldPassword:%s err:%v", newPassword, oldPassword, e)
	return
}

// Register .
func (user *User) Register(req vo.RegisterUserReq) (result vo.UserInfoRes, e error) {
	userEntity, e := user.UserRepo.New(req, 10000)
	if e != nil {
		return
	}
	result.ID = userEntity.ID
	result.Money = userEntity.Money
	result.Name = userEntity.Name
	return
}

// Get .
func (user *User) Get(userID int) (result vo.UserInfoRes, e error) {
	userEntity, e := user.UserRepo.Get(userID)
	if e != nil {
		return
	}
	result.ID = userEntity.ID
	result.Money = userEntity.Money
	result.Name = userEntity.Name
	return
}
