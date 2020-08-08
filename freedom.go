package freedom

import (
	"os"
	"runtime"
	"strings"

	"github.com/8treenet/freedom/internal"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/hero"
	"github.com/kataras/iris/v12/mvc"

	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis"
	"github.com/kataras/golog"
	iris "github.com/kataras/iris/v12"
)

var app *internal.Application

func init() {
	app = internal.NewApplication()
}

type (
	// Worker .
	Worker = internal.Worker

	// Initiator .
	Initiator = internal.Initiator

	// Repository .
	Repository = internal.Repository

	// Infra .
	Infra = internal.Infra

	// SingleBoot .
	SingleBoot = internal.SingleBoot

	Result = hero.Result

	Context = iris.Context

	Entity = internal.Entity

	DomainEventInfra = internal.DomainEventInfra

	UnitTest = internal.UnitTest

	Starter = internal.Starter

	Bus = internal.Bus

	BusHandler = internal.BusHandler

	Configuration    = iris.Configuration
	BeforeActivation = mvc.BeforeActivation
)

// NewApplication .
func NewApplication() Application {
	return app
}

func NewUnitTest() UnitTest {
	return new(internal.UnitTestImpl)
}

// Prepare .
func Prepare(f func(Initiator)) {
	internal.Prepare(f)
}

// Logger .
func Logger() *golog.Logger {
	return app.Logger()
}

// Prometheus .
func Prometheus() *internal.Prometheus {
	return app.Prometheus
}

var path string

type Configurer interface {
	Configure(obj interface{}, file string, must bool, metaData ...interface{})
}

var configurer Configurer

// SetConfigurer .
func SetConfigurer(confer Configurer) {
	configurer = confer
}

// Configure .
func Configure(obj interface{}, file string, must bool, metaData ...interface{}) {
	if configurer != nil {
		configurer.Configure(obj, file, must)
		return
	}
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
	_, err := toml.DecodeFile(path+"/"+file, obj)
	if err != nil && !must {
		panic(err)
	}
}

// Application .
type Application interface {
	InstallDB(f func() (db interface{}))
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
	InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
}

func ToWorker(ctx Context) Worker {
	if result, ok := ctx.Values().Get(internal.WorkerKey).(Worker); ok {
		return result
	}
	return nil
}

func DefaultConfiguration() Configuration {
	return iris.DefaultConfiguration()
}

func HandleBusMiddleware(worker Worker) {
	internal.HandleBusMiddleware(worker)
}
