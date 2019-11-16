package freedom

import (
	"fmt"
	"os"
	"strings"

	"github.com/8treenet/freedom/general"
	"github.com/8treenet/gcache"
	"github.com/BurntSushi/toml"
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

	// Repository .
	Repository = general.Repository

	// GORMRepository .
	GORMRepository = general.GORMRepository

	// QueryBuilder .
	QueryBuilder = general.QueryBuilder

	// Component .
	Component = general.Component
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
	if strings.Index(addr, ":") == 0 {
		fmt.Printf("Now h2c listening on: http://0.0.0.0%s\n", addr)
	} else {
		fmt.Printf("Now h2c listening on: http://%s\n", addr)
	}
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

// Configure .
func Configure(obj interface{}, fileName string, def bool) {
	path := os.Getenv("FREEDOM_PROJECT_CONFIG")
	if path != "" {
		_, err := os.Stat(path)
		if err != nil {
			path = ""
		}
	}
	if path == "" {
		path = "./conf"
		_, err := os.Stat(path)
		if err != nil {
			path = ""
		}
	}

	if path == "" {
		path = "./cmd/conf"
		_, err := os.Stat(path)
		if err != nil {
			path = ""
		}
	}

	if path == "" {
		panic("No profile directory found:" + "'$FREEDOM_PROJECT_CONFIG' or './conf' or './cmd/conf'")
	}
	_, err := toml.DecodeFile(path+"/"+fileName, obj)
	if err != nil && !def {
		panic(err)
	}
}
