package freedom

import (
	"github.com/8treenet/freedom/internal"
	"github.com/go-redis/redis"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
)

var (
	// app is a singleton of internal.Application. It is never nil.
	app *internal.Application
)

func initApp() {
	app = internal.NewApplication()
}

// Application represents an application of freedom
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
	Iris() *iris.Application
	Logger() *golog.Logger
	Run(serve iris.Runner, c iris.Configuration)
	Start(f func(starter Starter))
	InstallBusMiddleware(handle ...BusHandler)
	InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error)
}

// NewApplication .
func NewApplication() Application {
	return app
}

// Logger .
func Logger() *golog.Logger {
	return app.Logger()
}

// Prometheus .
func Prometheus() *internal.Prometheus {
	return app.Prometheus
}

// ServiceLocator .
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
//
//TODO(coco):
// Because of ambiguous naming, I've been create WorkerFromCtx as an alternative.
// Considering remove this function in the future.
func ToWorker(ctx Context) Worker {
	return WorkerFromCtx(ctx)
}
