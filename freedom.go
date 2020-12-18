package freedom

import (
	"os"

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

	//UnitTest is a unit test tool.
	UnitTest = internal.UnitTest

	//Starter is the startup interface.
	Starter = internal.Starter

	// Bus is the bus message type.
	Bus = internal.Bus

	// BusHandler is the bus message middleware type.
	BusHandler = internal.BusHandler

	// Configuration is the configuration type of the app.
	Configuration = iris.Configuration

	// BeforeActivation is Is the start-up pre-processing of the action.
	BeforeActivation = mvc.BeforeActivation

	// LogFields is the column type of the log.
	LogFields = golog.Fields

	// LogRow is the log per line callback.
	LogRow = golog.Log

	// DomainEvent .
	DomainEvent = internal.DomainEvent
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
	Configure(obj interface{}, file string, metaData ...interface{}) error
}

var configurer Configurer

// SetConfigurer .
func SetConfigurer(confer Configurer) {
	configurer = confer
}

// ProfileENV The name of the environment variable for the profile directory
var ProfileENV = "FREEDOM_PROJECT_CONFIG"

// Configure .
func Configure(obj interface{}, file string, metaData ...interface{}) error {
	if configurer != nil {
		return configurer.Configure(obj, file)
	}
	path = os.Getenv(ProfileENV)
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
	_, err := toml.DecodeFile(path+"/"+file, obj)
	if err != nil {
		Logger().Errorf("[Freedom] configure decode error: %s", err.Error())
	} else {
		Logger().Infof("[Freedom] configure decode: %s", path+"/"+file)
	}
	return err
}

// Application .
type Application interface {
	InstallDB(f func() (db interface{}))
	InstallRedis(f func() (client redis.Cmdable))
	InstallOther(f func() interface{})
	InstallMiddleware(handler iris.Handler)
	InstallParty(relativePath string)
	NewRunner(addr string, configurators ...host.Configurator) iris.Runner
	NewH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
	NewAutoTLSRunner(addr string, domain string, email string, configurators ...host.Configurator) iris.Runner
	NewTLSRunner(addr string, certFile, keyFile string, configurators ...host.Configurator) iris.Runner
	CreateRunner(addr string, configurators ...host.Configurator) iris.Runner //become invalid after a specified date.
	Iris() *iris.Application
	Logger() *golog.Logger
	Run(serve iris.Runner, c iris.Configuration)
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
