package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/kataras/golog"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/mvc"

	stdContext "context"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

var _ Initiator = new(Application)
var _ SingleBoot = new(Application)
var _ Starter = new(Application)

// NewApplication .
func NewApplication() *Application {
	globalAppOnce.Do(func() {
		globalApp = new(Application)
		globalApp.IrisApp = iris.New()
		globalApp.pool = newServicePool()
		globalApp.rpool = newRepoPool()
		globalApp.comPool = newInfraPool()
		globalApp.msgsBus = newMessageBus()
		globalApp.other = newOther()
		globalApp.marshal = json.Marshal
		globalApp.unmarshal = json.Unmarshal
		globalApp.Prometheus = newPrometheus()
		globalApp.IrisApp.Logger().SetTimeFormat("2006-01-02 15:04:05.000")
	})
	return globalApp
}

// Application .
type Application struct {
	IrisApp     *iris.Application
	pool        *ServicePool
	rpool       *RepositoryPool
	comPool     *InfraPool
	msgsBus     *EventBus
	prefixParty string
	Database    struct {
		db      *gorm.DB
		Install func() (db *gorm.DB)
	}

	Cache struct {
		client  redis.Cmdable
		Install func() (client redis.Cmdable)
	}
	other         *other
	Middleware    []context.Handler
	Prometheus    *Prometheus
	ControllerDep []interface{}
	eventInfra    DomainEventInfra
	unmarshal     func(data []byte, v interface{}) error
	marshal       func(v interface{}) ([]byte, error)
}

// InstallParty .
func (app *Application) InstallParty(relativePath string) {
	app.prefixParty = relativePath
}

// CreateParty .
func (app *Application) CreateParty(relativePath string, handlers ...context.Handler) iris.Party {
	return app.IrisApp.Party(app.prefixParty+relativePath, handlers...)
}

// BindController .
func (app *Application) BindController(relativePath string, controller interface{}, handlers ...context.Handler) {
	mvcApp := mvc.New(app.IrisApp.Party(app.prefixParty+relativePath, handlers...))
	mvcApp.Register(app.generalDep()...)
	mvcApp.Handle(controller)
	app.msgsBus.addController(controller)
	return
}

// BindControllerByParty .
func (app *Application) BindControllerByParty(party iris.Party, controller interface{}) {
	mvcApp := mvc.New(party)
	mvcApp.Register(app.generalDep()...)
	mvcApp.Handle(controller)
	return
}

// GetService .
func (app *Application) GetService(ctx iris.Context, service interface{}) {
	app.pool.get(ctx.Values().Get(WorkerKey).(*worker), service)
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

// ListenMessage .
func (app *Application) ListenEvent(eventName string, objectMethod string, appointInfra ...interface{}) {
	app.msgsBus.addEvent(objectMethod, eventName, appointInfra...)
}

// EventsPath .
func (app *Application) EventsPath(infra interface{}) map[string]string {
	return app.msgsBus.EventsPath(infra)
}

// BindInfra .
func (app *Application) BindInfra(single bool, com interface{}) {
	if !single {
		outType, err := parsePoolFunc(com)
		if err != nil {
			panic(fmt.Sprintf("%v : %s", com, fmt.Sprint(err)))
		}
		app.comPool.bind(single, outType, com)
		return
	}
	if reflect.TypeOf(com).Kind() != reflect.Ptr {
		panic("single:true, The infra must be an object")
	}
	app.comPool.bind(single, reflect.TypeOf(com), com)
}

// GetInfra .
func (app *Application) GetInfra(ctx iris.Context, com interface{}) {
	app.comPool.get(ctx.Values().Get(WorkerKey).(*worker), reflect.ValueOf(com).Elem())
}

// AsyncCachePreheat .
func (app *Application) AsyncCachePreheat(f func(repo *Repository)) {
	rb := new(Repository)
	go f(rb)
}

// CachePreheat .
func (app *Application) CachePreheat(f func(repo *Repository)) {
	rb := new(Repository)
	f(rb)
}

func (app *Application) generalDep() (result []interface{}) {
	result = append(result, func(ctx iris.Context) (rt Worker) {
		rt = ctx.Values().Get(WorkerKey).(Worker)
		return
	})
	result = append(result, app.ControllerDep...)
	return
}

// InjectController .
func (app *Application) InjectController(f interface{}) {
	app.ControllerDep = append(app.ControllerDep, f)
}

// Run .
func (app *Application) Run(serve iris.Runner, irisConf iris.Configuration) {
	app.addMiddlewares(irisConf)
	app.installDB()
	app.other.booting()
	for index := 0; index < len(prepares); index++ {
		prepares[index](app)
	}

	logLevel := "debug"
	if level, ok := irisConf.Other["logger_level"]; ok {
		logLevel = level.(string)
	}
	globalApp.IrisApp.Logger().SetLevel(logLevel)

	repositoryAPIRun(irisConf)
	for i := 0; i < len(starters); i++ {
		starters[i](app)
	}
	app.msgsBus.building()
	app.comPool.singleBooting(app)
	shutdownSecond := int64(2)
	if level, ok := irisConf.Other["shutdown_second"]; ok {
		shutdownSecond = level.(int64)
	}
	app.closeing(shutdownSecond)
	app.IrisApp.Run(serve, iris.WithConfiguration(irisConf))
}

func (app *Application) CreateH2CRunner(addr string, configurators ...host.Configurator) iris.Runner {
	h2cSer := &http2.Server{}
	ser := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(app.IrisApp, h2cSer),
	}
	return func(irisApp *iris.Application) error {
		return irisApp.NewHost(ser).Configure(configurators...).ListenAndServe()
	}
}

func (app *Application) CreateRunner(addr string, configurators ...host.Configurator) iris.Runner {
	return iris.Addr(addr, configurators...)
}

func (app *Application) closeing(timeout int64) {
	iris.RegisterOnInterrupt(func() {
		//读取配置的关闭最长时间
		ctx, cancel := stdContext.WithTimeout(stdContext.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		close := func() {
			if err := recover(); err != nil {
				app.IrisApp.Logger().Error(err)
			}
			app.comPool.closeing()
		}
		close()
		//通知组件服务即将关闭
		app.IrisApp.Shutdown(ctx)
	})
}

// Register .
func (app *Application) Closeing(f func()) {
	app.comPool.Closeing(f)
}

// InstallGorm .
func (app *Application) InstallGorm(f func() (db *gorm.DB)) {
	app.Database.Install = f
}

// InstallRedis .
func (app *Application) InstallRedis(f func() (client redis.Cmdable)) {
	app.Cache.Install = f
}

func (app *Application) installDB() {
	if app.Database.Install != nil {
		app.Database.db = app.Database.Install()
	}

	if app.Cache.Install != nil {
		app.Cache.client = app.Cache.Install()
	}
}

// InstallDomainEventInfra .
func (app *Application) InstallDomainEventInfra(eventInfra DomainEventInfra) {
	app.eventInfra = eventInfra
}

// Logger .
func (app *Application) Logger() *golog.Logger {
	return app.IrisApp.Logger()
}

// InstallMiddleware .
func (app *Application) InstallMiddleware(handler iris.Handler) {
	app.Middleware = append(app.Middleware, handler)
}

// Iris .
func (app *Application) Iris() *iris.Application {
	return app.IrisApp
}

// Start .
func (app *Application) Start(f func(starter Starter)) {
	starters = append(starters, f)
}

// GetSingleInfra .
func (app *Application) GetSingleInfra(com interface{}) {
	if !app.comPool.GetSingleInfra(reflect.ValueOf(com).Elem()) {
		panic("Component not found")
	}
}

func (app *Application) addMiddlewares(irisConf iris.Configuration) {
	app.IrisApp.Use(newWorkerHandle())
	app.IrisApp.Use(globalApp.pool.freeHandle())
	app.IrisApp.Use(globalApp.comPool.freeHandle())
	if pladdr, ok := irisConf.Other["prometheus_listen_addr"]; ok {
		registerPrometheus(app.Prometheus, irisConf.Other["service_name"].(string), pladdr.(string))
		globalApp.IrisApp.Use(newPrometheusHandle(app.Prometheus))
	}
	globalApp.IrisApp.Use(app.Middleware...)
}

// InstallOther .
func (app *Application) InstallOther(f func() interface{}) {
	app.other.add(f)
}

// InstallBusMiddleware .
func (app *Application) InstallBusMiddleware(handle ...BusHandler) {
	busMiddlewares = append(busMiddlewares, handle...)
}

// InstallSerializer .
func (app *Application) InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error) {
	app.marshal = marshal
	app.unmarshal = unmarshal
}
