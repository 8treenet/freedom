package project

func init() {
	content["/adapter/controllers/default.go"] = controllerTemplate()
}

func controllerTemplate() string {
	return `package controllers

	import (
		"github.com/8treenet/freedom"
		"{{.PackagePath}}/application"
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
	func (c *DefaultController) Get() (result struct {
		IP string
		UA string
	}, e error) {
		c.Runtime.Logger().Infof("我是控制器")
		remote := c.Sev.RemoteInfo()
		result.IP = remote.IP
		result.UA = remote.UA
		return
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
	func (c *DefaultController) GetAgeByUserBy(age int, user string) (result struct {
		User string
		Age  int
	}, e error) {
		result.Age = age
		result.User = user
		return
	}
	`
}
