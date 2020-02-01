package freedom

import (
	"os"

	"github.com/kataras/iris/hero"

	"github.com/8treenet/freedom/general"
	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/golog"
	"github.com/kataras/iris"
)

var app *general.Application

func init() {
	app = general.NewApplication()
}

type (
	// Runtime .
	Runtime = general.Runtime

	// Initiator .
	Initiator = general.Initiator

	// Repository .
	Repository = general.Repository

	// GORMRepository .
	GORMRepository = general.GORMRepository

	// QueryBuilder .
	QueryBuilder = general.QueryBuilder

	// Infra .
	Infra = general.Infra

	// SingleBoot .
	SingleBoot = general.SingleBoot

	Result = hero.Result

	Context = iris.Context

	Entity = general.Entity

	DomainEventInfra = general.DomainEventInfra
)

// NewApplication .
func NewApplication() Application {
	return app
}

// Booting .
func Booting(f func(Initiator)) {
	general.Booting(f)
}

// Logger .
func Logger() *golog.Logger {
	return app.Logger()
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
		path = "./server/conf"
		_, err := os.Stat(path)
		if err != nil {
			path = ""
		}
	}

	if path == "" {
		panic("No profile directory found:" + "'$FREEDOM_PROJECT_CONFIG' or './conf' or './server/conf'")
	}
	_, err := toml.DecodeFile(path+"/"+fileName, obj)
	if err != nil && !def {
		panic(err)
	}
}

// Application .
type Application interface {
	InstallGorm(f func() (db *gorm.DB))
	InstallRedis(f func() (client redis.Cmdable))
	InstallMiddleware(handler iris.Handler)
	InstallParty(relativePath string)
	CreateH2CRunner(addr string) iris.Runner
	Iris() *iris.Application
	Logger() *golog.Logger
	Run(serve iris.Runner, c iris.Configuration)
	InstallDomainEventInfra(eventInfra DomainEventInfra)
}
