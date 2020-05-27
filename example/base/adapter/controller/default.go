package controller

import (
	"github.com/8treenet/freedom/example/base/application"
	"github.com/8treenet/freedom/example/base/infra"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		/*
		   普通方式绑定 Default控制器到路径 /
		   initiator.BindController("/", &DefaultController{})
		*/

		//中间件方式绑定， 只对本控制器生效，全局中间件请在main加入。
		initiator.BindController("/", &Default{}, func(ctx freedom.Context) {
			worker := freedom.ToWorker(ctx)
			worker.Logger().Info("Hello middleware begin")
			ctx.Next()
			worker.Logger().Info("Hello middleware end")
		})
	})
}

type Default struct {
	Sev    *application.Default
	Worker freedom.Worker
}

// Get handles the GET: / route.
func (c *Default) Get() freedom.Result {
	c.Worker.Logger().Infof("我是控制器")
	remote := c.Sev.RemoteInfo()
	return &infra.JSONResponse{Object: remote}
}

// GetHello handles the GET: /hello route.
func (c *Default) GetHello() string {
	return "hello"
}

// PutHello handles the PUT: /hello route.
func (c *Default) PutHello() freedom.Result {
	return &infra.JSONResponse{Object: "putHello"}
}

// PostHello handles the POST: /hello route.
func (c *Default) PostHello() freedom.Result {
	/*
		var postJsonData struct {
			UserName     string validate:"required"
			UserPassword string validate:"required"
		}
		if err := c.JSONRequest.ReadJSON(&postJsonData); err != nil {
			return &infra.JSONResponse{Error: err}
		}
	*/
	return &infra.JSONResponse{Object: "postHello"}
}

func (m *Default) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("ANY", "/custom", "CustomHello")
	//b.Handle("GET", "/custom", "CustomHello")
	//b.Handle("PUT", "/custom", "CustomHello")
	//b.Handle("POST", "/custom", "CustomHello")
}

// PostHello handles the POST: /hello route.
func (c *Default) CustomHello() freedom.Result {
	method := c.Worker.IrisContext().Request().Method
	c.Worker.Logger().Info(method, "CustomHello")
	return &infra.JSONResponse{Object: method + "CustomHello"}
}

// GetUserBy handles the GET: /user/{username:string} route.
func (c *Default) GetUserBy(username string) string {
	return username
}

// GetAgeByUserBy handles the GET: /age/{age:int}/user/{user:string} route.
func (c *Default) GetAgeByUserBy(age int, user string) freedom.Result {
	var result struct {
		User string
		Age  int
	}
	result.Age = age
	result.User = user

	return &infra.JSONResponse{Object: result}
}
