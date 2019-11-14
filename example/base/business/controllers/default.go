package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/business/services"
	"github.com/8treenet/freedom/example/base/models"
	"github.com/kataras/iris"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		serFunc := func(ctx iris.Context) (m *services.DefaultService) {
			initiator.GetService(ctx, &m)
			return
		}
		initiator.BindController("/", &DefaultController{}, serFunc)
	})
}

type DefaultController struct {
	Sev     *services.DefaultService
	ASev    *services.AlbumService
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
func (c *DefaultController) GetUserBy(id int) (album *models.Album, e error) {
	c.Runtime.Logger().Infof("我是 Album 控制器")
	return c.ASev.GetAlbum(id)
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
