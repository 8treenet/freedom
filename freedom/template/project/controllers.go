package project

func init() {
	content["/adapter/controllers/default.go"] = controllerTemplate()
}

func controllerTemplate() string {
	return `package controllers

	import (
		"github.com/8treenet/freedom"
		"{{.PackagePath}}/application"
		"{{.PackagePath}}/infra"
	)
	
	func init() {
		freedom.Booting(func(initiator freedom.Initiator) {
			initiator.BindController("/", &DefaultController{})
		})
	}
	
	type DefaultController struct {
		Sev     *application.DefaultService
		Runtime freedom.Runtime
	}
	
	// Get handles the GET: / route.
	func (c *DefaultController) Get() freedom.Result {
		c.Runtime.Logger().Infof("我是控制器")
		remote := c.Sev.RemoteInfo()
		return &infra.JSONResponse{Object: remote}
	}
	
	// GetHello handles the GET: /hello route.
	func (c *DefaultController) GetHello() string {
		return "hello"
	}
	
	// GetUserBy handles the GET: /user/{username:string} route.
	func (c *DefaultController) GetUserBy(username string) string {
		return username
	}
	
	// GetAgeByUserBy handles the GET: /age/{age:int}/user/{user:string} route.
	func (c *DefaultController) GetAgeByUserBy(age int, user string) freedom.Result {
		var result struct {
			User string
			Age  int
		}
		result.Age = age
		result.User = user

		return &infra.JSONResponse{Object: result}
	}
	`
}
