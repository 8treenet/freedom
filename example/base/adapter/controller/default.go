package controller

import (
	"io"
	"os"

	"github.com/8treenet/freedom/example/base/domain"
	"github.com/8treenet/freedom/example/base/infra"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		/*
		   Common binding, default controller to path '/'.
		   initiator.BindController("/", &DefaultController{})
		*/

		// Middleware binding. Valid only for this controller.
		// If you need global middleware, please add in main.
		initiator.BindController("/", &Default{}, func(ctx freedom.Context) {
			worker := freedom.ToWorker(ctx)
			worker.Logger().Info("Hello middleware begin")
			ctx.Next()
			worker.Logger().Info("Hello middleware end")
		})
	})
}

// Default .
type Default struct {
	Sev     *domain.Default
	Worker  freedom.Worker
	Request *infra.Request
}

// Get handles the GET: / route.
func (c *Default) Get() freedom.Result {
	//curl http://127.0.0.1:8000
	c.Worker.Logger().Info("I'm Controller")
	remote := c.Sev.RemoteInfo()
	return &infra.JSONResponse{Object: remote}
}

// GetHello handles the GET: /hello route.
func (c *Default) GetHello() string {
	//curl http://127.0.0.1:8000/hello
	field := freedom.LogFields{
		"framework": "freedom",
		"like":      "DDD",
	}
	c.Worker.Logger().Info("hello", field)
	c.Worker.Logger().Infof("hello %s", "format", field)
	c.Worker.Logger().Debug("hello", field)
	c.Worker.Logger().Debugf("hello %s", "format", field)
	c.Worker.Logger().Error("hello", field)
	c.Worker.Logger().Errorf("hello %s", "format", field)
	c.Worker.Logger().Warn("hello", field)
	c.Worker.Logger().Warnf("hello %s", "format", field)
	c.Worker.Logger().Print("hello")
	c.Worker.Logger().Println("hello")
	return "hello"
}

// PutHello handles the PUT: /hello route.
func (c *Default) PutHello() freedom.Result {
	//curl -X PUT http://127.0.0.1:8000/hello
	return &infra.JSONResponse{Object: "putHello"}
}

// PostHello handles the POST: /hello route.
func (c *Default) PostHello() freedom.Result {
	//curl -X POST -d '{"userName":"freedom","userPassword":"freedom"}' http://127.0.0.1:8000/hello
	var postJSONData struct {
		UserName     string `json:"userName" validate:"required"`
		UserPassword string `json:"userPassword" validate:"required"`
	}
	if err := c.Request.ReadJSON(&postJSONData, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}

	return &infra.JSONResponse{Object: postJSONData}
}

// PutHello handles the DELETE: /hello route.
func (c *Default) DeleteHello() freedom.Result {
	//curl -X DELETE http://127.0.0.1:8000/hello
	return &infra.JSONResponse{Object: "deleteHello"}
}

/* Can use more than one, the factory will make sure
that the correct http methods are being registered for each route
for this controller, uncomment these if you want:
	func (c *Default) ConnectHello() {}
	func (c *Default) HeadHello() {}
	func (c *Default) PatchHello() {}
	func (c *Default) OptionsHello() {}
	func (c *Default) TraceHello() {}
*/

// BeforeActivation .
func (c *Default) BeforeActivation(b freedom.BeforeActivation) {
	b.Handle("GET", "/customPath/{id:int64}/{uid:int}/{username:string}", "CustomPath")
	b.Handle("ANY", "/custom", "Custom")
	//b.Handle("GET", "/custom", "Custom")
	//b.Handle("PUT", "/custom", "Custom")
	//b.Handle("POST", "/custom", "Custom")
	//b.Handle("DELETE", "/custom", "Custom")
}

// CustomPath handles the GET: /customPath/{id:int64}/{uid:int}/{username:string} route.
func (c *Default) CustomPath(id int64, uid int, username string) freedom.Result {
	//curl http://127.0.0.1:8000/customPath/1/2/freedom
	return &infra.JSONResponse{Object: map[string]interface{}{"id": id, "uid": uid, "username": username}}
}

// Custom handles the ANY: /custom route.
func (c *Default) Custom() freedom.Result {
	//curl http://127.0.0.1:8000/custom
	//curl -X PUT http://127.0.0.1:8000/custom
	//curl -X DELETE http://127.0.0.1:8000/custom
	//curl -X POST http://127.0.0.1:8000/custom
	method := c.Worker.IrisContext().Request().Method
	c.Worker.Logger().Info("CustomHello", freedom.LogFields{"method": method})
	return &infra.JSONResponse{Object: method + "/Custom"}
}

// GetUserBy handles the GET: /user/{username:string} route.
func (c *Default) GetUserBy(username string) freedom.Result {
	//curl 'http://127.0.0.1:8000/user/freedom?token=ftoken123&id=1&ip=192&ip=168&ip=1&ip=1'
	var query struct {
		Token string  `url:"token" validate:"required"`
		ID    int64   `url:"id" validate:"required"`
		IP    []int64 `url:"ip"`
	}
	if err := c.Request.ReadQuery(&query, true); err != nil {
		return &infra.JSONResponse{Error: err}
	}
	var data struct {
		Name  string
		Token string
		ID    int64
		IP    []int64
	}
	data.ID = query.ID
	data.Token = query.Token
	data.Name = username
	data.IP = query.IP
	return &infra.JSONResponse{Object: data}
}

// GetAgeByUserBy handles the GET: /age/{age:int}/user/{user:string} route.
func (c *Default) GetAgeByUserBy(age int, user string) freedom.Result {
	//curl http://127.0.0.1:8000/age/20/user/freedom
	var result struct {
		User string
		Age  int
	}
	result.Age = age
	result.User = user

	return &infra.JSONResponse{Object: result}
}

// PostForm handles the Post: /form route.
func (c *Default) PostForm() freedom.Result {
	//curl -X POST --data "userName=freedom&mail=freedom@freedom.com&myData=data1&myData=data2" http://127.0.0.1:8000/form
	var visitor struct {
		UserName string   `form:"userName" validate:"required"`
		Mail     string   `form:"mail" validate:"required"`
		Data     []string `form:"myData" validate:"required"`
	}
	err := c.Request.ReadForm(&visitor, true)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	c.Worker.Logger().Infof("%d, %s, %s, %v", c.Worker.IrisContext().Request().ContentLength, visitor.UserName, visitor.Mail, visitor.Data)
	return &infra.JSONResponse{Object: visitor}
}

// PostFile handles the Post: /file route.
func (c *Default) PostFile() freedom.Result {
	//curl -X POST -F "file=@example/base/adapter/controller/default.go" http://127.0.0.1:8000/file
	file, info, err := c.Worker.IrisContext().FormFile("file")
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}

	defer file.Close()
	fname := info.Filename
	out, err := os.OpenFile(os.TempDir()+fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return &infra.JSONResponse{Error: err}
	}
	defer out.Close()
	_, err = io.Copy(out, file)

	return &infra.JSONResponse{Error: err, Object: os.TempDir() + fname}
}
