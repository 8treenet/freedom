package freedom

import (
	"os"
	"runtime"
	"strings"

	"github.com/kataras/iris/core/host"
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

	UnitTest = general.UnitTest

	Starter = general.Starter

	Bus = general.Bus

	BusHandler = general.BusHandler
)

// NewApplication .
func NewApplication() Application {
	return app
}

func NewUnitTest() UnitTest {
	return new(general.UnitTestImpl)
}

// Prepare .
func Prepare(f func(Initiator)) {
	general.Prepare(f)
}

// Logger .
func Logger() *golog.Logger {
	return app.Logger()
}

// Prometheus .
func Prometheus() *general.Prometheus {
	return app.Prometheus
}

var path string

// Configure .
func Configure(obj interface{}, fileName string, def bool) {
	if path == "" {
		path = os.Getenv("FREEDOM_PROJECT_CONFIG")
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
	callerProjectDir := ""
	if path == "" {
		_, filego, _, _ := runtime.Caller(1)
		list := strings.Split(filego, "/infra/config")
		if len(list) == 2 {
			callerProjectDir = list[0] + "/server/conf"
			path = callerProjectDir
			_, err := os.Stat(path)
			if err != nil {
				path = ""
			}
		}
	}
	if path == "" {
		panic("No profile directory found:" + "'$FREEDOM_PROJECT_CONFIG' or './conf' or './server/conf' or '" + callerProjectDir + "'")
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
	InstallOther(f func() interface{})
	InstallMiddleware(handler iris.Handler)
	InstallParty(relativePath string)
	CreateH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
	CreateRunner(addr string, configurators ...host.Configurator) iris.Runner
	Iris() *iris.Application
	Logger() *golog.Logger
	Run(serve iris.Runner, c iris.Configuration)
	InstallDomainEventInfra(eventInfra DomainEventInfra)
	Start(f func(starter Starter))
	InstallBusMiddleware(handle ...BusHandler)
}

func PickRuntime(ctx Context) Runtime {
	if result, ok := ctx.Values().Get(general.RuntimeKey).(Runtime); ok {
		return result
	}
	return nil
}
