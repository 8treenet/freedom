package freedom

import (
	"github.com/8treenet/freedom/general"
	"github.com/8treenet/gcache"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/golog"
	"github.com/kataras/iris"
)

func init() {
	app = general.NewApplication()
}

type (
	// Runtime .
	Runtime = general.Runtime

	// Application .
	Application = general.Application

	// Initiator .
	Initiator = general.Initiator
	// RepositoryCache .
	RepositoryCache = general.RepositoryCache

	// RepositoryDB .
	RepositoryDB = general.RepositoryDB

	// GORMRepository .
	GORMRepository = general.GORMRepository

	// QueryBuilder .
	QueryBuilder = general.QueryBuilder
)

var app *Application

// Run .
func Run(serve iris.Runner, c *iris.Configuration) {
	app.Run(serve, *c)
}

// Booting app.BindController or app.BindControllerByParty.
func Booting(f func(Initiator)) {
	general.Booting(f)
}

// CreateH2CRunner -
func CreateH2CRunner(addr string) iris.Runner {
	ser := general.CreateH2Server(app, addr)
	return iris.Raw(ser.ListenAndServe)
}

// InstallGorm .
func InstallGorm(f func() (db *gorm.DB, cache gcache.Plugin)) {
	app.Database.Install = f
}

// InstallRedis .
func InstallRedis(f func() (client *redis.Client)) {
	app.Redis.Install = f
}

// Logger .
func Logger() *golog.Logger {
	return app.IrisApp.Logger()
}

// UseMiddleware .
func UseMiddleware(handler iris.Handler) {
	app.Middleware = append(app.Middleware, handler)
}
