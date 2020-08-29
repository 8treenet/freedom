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
	//Worker .
	Worker = internal.Worker

	//Initiator .
	Initiator = internal.Initiator

	//Repository .
	Repository = internal.Repository

	//Infra .
	Infra = internal.Infra

	//SingleBoot .
	SingleBoot = internal.SingleBoot

	//Result is the controller return type.
	Result = hero.Result

	//Context is the context type.
	Context = iris.Context

	//Entity is the entity's father interface.
	Entity = internal.Entity

	//DomainEventInfra is a dependency inverted interface for domain events.
	DomainEventInfra = internal.DomainEventInfra

	//UnitTest is a unit test tool.
	UnitTest = internal.UnitTest

	//Starter is the startup interface.
	Starter = internal.Starter

	//Bus is the bus message type.
	Bus = internal.Bus

	//BusHandler is the bus message middleware type.
	BusHandler = internal.BusHandler

	//Configuration is the configuration type of the app.
	Configuration = iris.Configuration

	//BeforeActivation is Is the start-up pre-processing of the action..
	BeforeActivation = mvc.BeforeActivation

	//LogFields is the column type of the log.
	LogFields = golog.Fields

	//LogRow is the log per line callback.
	LogRow = golog.Log
)

// NewApplication .
func NewApplication() Application {
	return app
}

// NewUnitTest .
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

// Configurer .
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
	CallService(fun interface{}, worker ...Worker)
}

//ToWorker the context is converted to a worker.
func ToWorker(ctx Context) Worker {
	if result, ok := ctx.Values().Get(internal.WorkerKey).(Worker); ok {
		return result
	}
	return nil
}

//DefaultConfiguration the default profile.
func DefaultConfiguration() Configuration {
	return iris.DefaultConfiguration()
}

//HandleBusMiddleware middleware processing.
func HandleBusMiddleware(worker Worker) {
	internal.HandleBusMiddleware(worker)
}
