package freedom

import (
	"github.com/8treenet/freedom/internal"
	"github.com/8treenet/iris/v12"
	"github.com/8treenet/iris/v12/core/host"
	"github.com/go-redis/redis"
	"github.com/kataras/golog"
)

var (
	// app is a singleton of internal.Application. It is never nil.
	app *internal.Application
)

func initApp() {
	app = internal.NewApplication()
}

// Application Represents an application of freedom
type Application interface {
	// Install the database. A function that returns the Interface type is passed in as a handle to the database.
	InstallDB(f func() (db interface{}))
	// Install the redis. A function that returns the redis.Cmdable is passed in as a handle to the client.
	InstallRedis(f func() (client redis.Cmdable))
	// Install a custom data source. A function that returns the Interface type is passed in as a handle to the custom data sources.
	InstallCustom(f func() interface{})
	// Install the middleware for http routing globally.
	InstallMiddleware(handler iris.Handler)
	// Assigns specified prefix to the application-level router.
	// Because of ambiguous naming, I've been create SetPrefixPath as an alternative.
	// Considering remove this function in the future.
	InstallParty(relativePath string)
	// NewRunner Can be used as an argument for the `Run` method.
	NewRunner(addr string, configurators ...host.Configurator) iris.Runner
	// NewH2CRunner Create a Runner for H2C .
	NewH2CRunner(addr string, configurators ...host.Configurator) iris.Runner
	// NewAutoTLSRunner Can be used as an argument for the `Run` method.
	NewAutoTLSRunner(addr string, domain string, email string, configurators ...host.Configurator) iris.Runner
	// NewTLSRunner Can be used as an argument for the `Run` method.
	NewTLSRunner(addr string, certFile, keyFile string, configurators ...host.Configurator) iris.Runner
	// Go back to the iris app.
	Iris() *iris.Application
	// Return to global logger.
	Logger() *golog.Logger
	// Run serving an iris application as the HTTP server to handle the incoming requests
	Run(serve iris.Runner, c iris.Configuration)
	// Adds a builder function which builds a BootManager.
	BindBooting(f func(bootManager BootManager))
	// Install the middleware for the message bus.
	InstallBusMiddleware(handle ...BusHandler)
	InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
}

// NewApplication Create an application.
func NewApplication() Application {
	return app
}

// Logger Return to global logger.
func Logger() *golog.Logger {
	return app.Logger()
}

// Prometheus Return prometheus.
func Prometheus() *internal.Prometheus {
	return app.Prometheus
}

// ServiceLocator Return serviceLocator.
// Use the service locator to get the service.
func ServiceLocator() *internal.ServiceLocatorImpl {
	return app.GetServiceLocator()
}

// WorkerFromCtx extracts Worker from a Context.
func WorkerFromCtx(ctx Context) Worker {
	if result, ok := ctx.Values().Get(internal.WorkerKey).(Worker); ok {
		return result
	}
	return nil
}

// ToWorker proxy a call to WorkerFromCtx.
func ToWorker(ctx Context) Worker {
	return WorkerFromCtx(ctx)
}
