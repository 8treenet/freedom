package internal

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/go-redis/redis"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/kataras/golog"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/mvc"

	stdContext "context"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

var _ Initiator = (*Application)(nil)
var _ SingleBoot = (*Application)(nil)
var _ Starter = (*Application)(nil)

// NewApplication create an instance of freedom.
func NewApplication() *Application {
	globalAppOnce.Do(func() {
		globalApp = new(Application)
		globalApp.IrisApp = iris.New()
		globalApp.pool = newServicePool()
		globalApp.rpool = newRepoPool()
		globalApp.factoryPool = newFactoryPool()
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

// Application is the framework's application.
// Create an application of freedom, by using NewApplication().
type Application struct {
	// IrisApp is an iris application
	IrisApp     *iris.Application
	pool        *ServicePool
	rpool       *RepositoryPool
	factoryPool *FactoryPool

	comPool *InfraPool
	msgsBus *EventBus
	// prefixParty means the prefix of url path
	prefixParty string
	// Database contains a database connection object and an installation function
	Database struct {
		db      interface{}
		Install func() (db interface{})
	}
	// Cache contains a redis client and an installation function
	Cache struct {
		client  redis.Cmdable
		Install func() (client redis.Cmdable)
	}
	other *other
	// Middleware is a set that satisfies the iris handler's definition
	Middleware []context.Handler
	Prometheus *Prometheus
	// ControllerDep TODO
	ControllerDep []interface{}
	// unmarshal deserializes [] byte into object
	unmarshal func(data []byte, v interface{}) error
	// marshal serializes objects into []byte
	marshal func(v interface{}) ([]byte, error)
}

// InstallParty installs prefixParty of Application
func (app *Application) InstallParty(relativePath string) {
	app.prefixParty = relativePath
}

// CreateParty returns a sub-router of iris which may have the same prefix and share same handlers
func (app *Application) CreateParty(relativePath string, handlers ...context.Handler) iris.Party {
	return app.IrisApp.Party(app.prefixParty+relativePath, handlers...)
}

// BindController binds the controller that satisfies the iris's definition to the Application
// and adds the controller into msgsBus
func (app *Application) BindController(relativePath string, controller interface{}, handlers ...context.Handler) {
	mvcApp := mvc.New(app.IrisApp.Party(app.prefixParty+relativePath, handlers...))
	mvcApp.Register(app.generalDep()...)
	mvcApp.Handle(controller)
	app.msgsBus.addController(controller)
	return
}

// BindControllerByParty binds the controller by iris's party
func (app *Application) BindControllerByParty(party iris.Party, controller interface{}) {
	mvcApp := mvc.New(party)
	mvcApp.Register(app.generalDep()...)
	mvcApp.Handle(controller)
	return
}

// GetService TODO
func (app *Application) GetService(ctx iris.Context, service interface{}) {
	app.pool.get(ctx.Values().Get(WorkerKey).(*worker), service)
	return
}

// BindService .
func (app *Application) BindService(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		globalApp.Logger().Fatalf("[Freedom] BindService: The binding function is incorrect, %v : %s", f, err.Error())
	}
	app.pool.bind(outType, f)
}

// BindRepository .
func (app *Application) BindRepository(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		globalApp.Logger().Fatalf("[Freedom] BindRepository: The binding function is incorrect, %v : %s", f, err.Error())
	}
	app.rpool.bind(outType, f)
}

// BindFactory .
func (app *Application) BindFactory(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		globalApp.Logger().Fatalf("[Freedom] BindFactory: The binding function is incorrect, %v : %s", f, err.Error())
	}
	app.factoryPool.bind(outType, f)
}

// ListenEvent .
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
			globalApp.Logger().Fatalf("[Freedom] BindInfra: The binding function is incorrect, %v : %s", reflect.TypeOf(com), err.Error())
		}
		app.comPool.bind(single, outType, com)
		return
	}
	if reflect.TypeOf(com).Kind() != reflect.Ptr {
		globalApp.Logger().Fatalf("[Freedom] BindInfra: This is not a single-case object, %v", reflect.TypeOf(com))
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
	app.shutdown(shutdownSecond)
	app.IrisApp.Run(serve, iris.WithConfiguration(irisConf))
}

// NewRunner can be used as an argument for the `Run` method.
// It accepts a host address which is used to build a server
// and a listener which listens on that host and port.
//
// Addr should have the form of [host]:port, i.e localhost:8080 or :8080.
//
// Second argument is optional, it accepts one or more
// `func(*host.Configurator)` that are being executed
// on that specific host that this function will create to start the server.
// Via host configurators you can configure the back-end host supervisor,
// i.e to add events for shutdown, serve or error.
func (app *Application) NewRunner(addr string, configurators ...host.Configurator) iris.Runner {
	return iris.Addr(addr, configurators...)
}

// CreateRunner become invalid after a specified date.
func (app *Application) CreateRunner(addr string, configurators ...host.Configurator) iris.Runner {
	return app.NewRunner(addr, configurators...)
}

// NewAutoTLSRunner can be used as an argument for the `Run` method.
// It will start the Application's secure server using
// certifications created on the fly by the "autocert" golang/x package,
// so localhost may not be working, use it at "production" machine.
//
// Addr should have the form of [host]:port, i.e mydomain.com:443.
//
// The whitelisted domains are separated by whitespace in "domain" argument,
// i.e "8tree.net", can be different than "addr".
// If empty, all hosts are currently allowed. This is not recommended,
// as it opens a potential attack where clients connect to a server
// by IP address and pretend to be asking for an incorrect host name.
// Manager will attempt to obtain a certificate for that host, incorrectly,
// eventually reaching the CA's rate limit for certificate requests
// and making it impossible to obtain actual certificates.
//
// For an "e-mail" use a non-public one, letsencrypt needs that for your own security.
//
// Note: `AutoTLS` will start a new server for you
// which will redirect all http versions to their https, including subdomains as well.
//
// Last argument is optional, it accepts one or more
// `func(*host.Configurator)` that are being executed
// on that specific host that this function will create to start the server.
// Via host configurators you can configure the back-end host supervisor,
// i.e to add events for shutdown, serve or error.
// Look at the `ConfigureHost` too.
func (app *Application) NewAutoTLSRunner(addr string, domain string, email string, configurators ...host.Configurator) iris.Runner {
	return func(irisApp *iris.Application) error {
		return irisApp.NewHost(&http.Server{Addr: addr}).
			Configure(configurators...).
			ListenAndServeAutoTLS(domain, email, "letscache")
	}
}

// NewTLSRunner can be used as an argument for the `Run` method.
// It will start the Application's secure server.
//
// Use it like you used to use the http.ListenAndServeTLS function.
//
// Addr should have the form of [host]:port, i.e localhost:443 or :443.
// CertFile & KeyFile should be filenames with their extensions.
//
// Second argument is optional, it accepts one or more
// `func(*host.Configurator)` that are being executed
// on that specific host that this function will create to start the server.
// Via host configurators you can configure the back-end host supervisor,
// i.e to add events for shutdown, serve or error.
// An example of this use case can be found at:
func (app *Application) NewTLSRunner(addr string, certFile, keyFile string, configurators ...host.Configurator) iris.Runner {
	return func(irisApp *iris.Application) error {
		return irisApp.NewHost(&http.Server{Addr: addr}).
			Configure(configurators...).
			ListenAndServeTLS(certFile, keyFile)
	}
}

//NewH2CRunner .
func (app *Application) NewH2CRunner(addr string, configurators ...host.Configurator) iris.Runner {
	h2cSer := &http2.Server{}
	ser := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(app.IrisApp, h2cSer),
	}
	return func(irisApp *iris.Application) error {
		return irisApp.NewHost(ser).Configure(configurators...).ListenAndServe()
	}
}

func (app *Application) shutdown(timeout int64) {
	iris.RegisterOnInterrupt(func() {
		//读取配置的关闭最长时间
		ctx, cancel := stdContext.WithTimeout(stdContext.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		close := func() {
			if err := recover(); err != nil {
				app.IrisApp.Logger().Errorf("[Freedom] An error was encountered during the program shutdown, %v", err)
			}
			app.comPool.shutdown()
		}
		close()
		//通知组件服务即将关闭
		app.IrisApp.Shutdown(ctx)
	})
}

// RegisterShutdown .
func (app *Application) RegisterShutdown(f func()) {
	app.comPool.registerShutdown(f)
}

// InstallDB .
func (app *Application) InstallDB(f func() interface{}) {
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
func (app *Application) GetSingleInfra(com interface{}) bool {
	return app.comPool.GetSingleInfra(reflect.ValueOf(com).Elem())
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

// CallService .
func (app *Application) CallService(fun interface{}, worker ...Worker) {
	callService(fun, worker...)
}
