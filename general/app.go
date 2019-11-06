package general

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"

	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/mvc"

	"github.com/8treenet/freedom/general/middleware"
	"github.com/8treenet/gcache"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// NewApplication .
func NewApplication() *Application {
	globalAppOnce.Do(func() {
		globalApp = new(Application)
		globalApp.IrisApp = iris.New()
		globalApp.pool = newServicePool()
		globalApp.rpool = newRepoPool()
		boots = make([]func(Initiator), 0)
	})
	return globalApp
}

// Application .
type Application struct {
	IrisApp  *iris.Application
	pool     *ServicePool
	rpool    *RepositoryPool
	Database struct {
		db      *gorm.DB
		cache   gcache.Plugin
		Install func() (db *gorm.DB, cache gcache.Plugin)
	}

	Redis struct {
		client  *redis.Client
		Install func() (client *redis.Client)
	}
	traceIDKey string
	Middleware []context.Handler
}

// CreateParty .
func (app *Application) CreateParty(relativePath string, handlers ...context.Handler) iris.Party {
	return app.IrisApp.Party(relativePath, handlers...)
}

// BindController .
func (app *Application) BindController(relativePath string, controller interface{}, service ...interface{}) {
	mvcApp := mvc.New(app.IrisApp.Party(relativePath))
	deps := append(app.generalDep(), service...)
	mvcApp.Register(deps...)
	mvcApp.Handle(controller)
	return
}

// BindControllerByParty .
func (app *Application) BindControllerByParty(party iris.Party, controller interface{}, service ...interface{}) {
	mvcApp := mvc.New(party)
	deps := append(app.generalDep(), service...)
	mvcApp.Register(deps...)
	mvcApp.Handle(controller)
	return
}

// GetService .
func (app *Application) GetService(ctx iris.Context, service interface{}) {
	app.pool.get(ctx.Values().Get(runtimeKey).(*appRuntime), service)
	return
}

// BindService .
func (app *Application) BindService(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		panic(fmt.Sprintf("%v : %s", f, fmt.Sprint(err)))
	}
	app.pool.bind(outType, f)
}

// BindRepository .
func (app *Application) BindRepository(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		panic(fmt.Sprintf("%v : %s", f, fmt.Sprint(err)))
	}
	app.rpool.bind(outType, f)
}

// AsyncCachePreheat .
func (app *Application) AsyncCachePreheat(f func(repoDB *RepositoryDB, repoCache *RepositoryCache)) {
	rb := new(RepositoryDB)
	rc := new(RepositoryCache)
	go f(rb, rc)
}

// CachePreheat .
func (app *Application) CachePreheat(f func(repoDB *RepositoryDB, repoCache *RepositoryCache)) {
	rb := new(RepositoryDB)
	rc := new(RepositoryCache)
	f(rb, rc)
}

func (app *Application) generalDep() (result []interface{}) {
	result = append(result, func(ctx iris.Context) (rt Runtime) {
		rt = ctx.Values().Get(runtimeKey).(Runtime)
		return
	})
	return
}

// Run .
func (app *Application) Run(serve iris.Runner, irisConf iris.Configuration) {
	app.traceIDKey = irisConf.Other["trace_key"].(string)
	app.addMiddlewares(irisConf)
	if app.Database.Install != nil {
		app.Database.db, app.Database.cache = app.Database.Install()
	}

	if app.Redis.Install != nil {
		app.Redis.client = app.Redis.Install()
	}

	for index := 0; index < len(boots); index++ {
		boots[index](app)
	}

	repositoryAPIRun(irisConf)
	app.IrisApp.Run(serve, iris.WithConfiguration(irisConf))
}

func (app *Application) addMiddlewares(irisConf iris.Configuration) {
	app.IrisApp.Use(newRuntimeHandle())
	app.IrisApp.Use(globalApp.pool.freeHandle())

	loggerConf := logger.DefaultConfig()
	loggerConf.IP = false
	loggerConf.MessageContextKeys = []string{"logger_trace", "logger_message"}
	app.IrisApp.Use(logger.New(loggerConf))

	if pladdr, ok := irisConf.Other["prometheus_listen_addr"]; ok {
		m := middleware.NewPrometheus(irisConf.Other["service_name"].(string), pladdr.(string))
		globalApp.IrisApp.Use(m.ServeHTTP)
	}
	globalApp.IrisApp.Use(app.Middleware...)
	globalApp.IrisApp.Use(newRecover())
}
