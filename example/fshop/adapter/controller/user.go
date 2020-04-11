package controller

import (
	"github.com/8treenet/freedom/example/fshop/application"
	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/example/fshop/infra"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/user", &User{})
	})
}

// User 控制器
type User struct {
	Sev         *application.User  //用户领域服务
	Runtime     freedom.Runtime    //运行时，一个请求绑定一个运行时
	JSONRequest *infra.JSONRequest //基础设施 用于处理客户端请求io的json数据和验证
}

// Put 修改密码, PUT: /user route.
func (u *User) Put() freedom.Result {
	var req dto.ChangePasswordReq
	if e := u.JSONRequest.ReadBodyJSON(&req); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	//调用领域服务
	e := u.Sev.ChangePassword(req.Id, req.NewPassword, req.OldPassword)
	if e != nil {
		//返回错误
		return &infra.JSONResponse{Error: e}
	}
	return &infra.JSONResponse{}
}

// GetBy 获取用户信息, GET: /user/:id route.
func (c *User) GetBy(id int) freedom.Result {
	dto, e := c.Sev.Get(id)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}
	return &infra.JSONResponse{Object: dto}
}

// Post 注册用户, POST: /user route.
func (u *User) Post() freedom.Result {
	var req dto.RegisterUserReq
	if e := u.JSONRequest.ReadBodyJSON(&req); e != nil {
		return &infra.JSONResponse{Error: e}
	}

	dto, e := u.Sev.Register(req)
	if e != nil {
		return &infra.JSONResponse{Error: e}
	}
	return &infra.JSONResponse{Object: dto}
}
