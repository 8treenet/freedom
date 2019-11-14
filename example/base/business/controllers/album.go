package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/business/services"
	"github.com/8treenet/freedom/example/base/models"
	"github.com/kataras/iris"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		serFunc := func(ctx iris.Context) (m *services.AlbumService) {
			initiator.GetService(ctx, &m)
			return
		}
		// initiator.BindControllerByParty(iris.Party.Party("/albums"), &AlbumController{}, serFunc)
		initiator.BindController("/", &DefaultController{}, serFunc)
	})
}

// AlbumController .
type AlbumController struct {
	Sev     *services.AlbumService
	Runtime freedom.Runtime
}

// Get handles the GET: /{id:int} route.
func (c *AlbumController) GetAlbumsBy(id int) (album *models.Album, e error) {
	c.Runtime.Logger().Infof("我是 Album 控制器")
	return c.Sev.GetAlbum(id)
}
