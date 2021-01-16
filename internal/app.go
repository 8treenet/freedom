package internal

import (
	stdcontext "context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/go-redis/redis"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

var _ Initiator = (*Application)(nil)
var _ SingleBoot = (*Application)(nil)
var _ Starter = (*Application)(nil)

type (
	// Dependency is a function which represents a dependency of the application.
	// A valid definition of dependency should like:
	//  func foo(ctx iris.Context) Bar {
	//    return Bar{}
	//  }
	//
	// Example:
	// If you want to inject a `Foo` into somewhere, you could written:
	//  func (ctx iris.Context) Foo {
	//    return NewFoo() // you'd replace NewFoo() with your own logic for creating Foo.
	//  }
	Dependency = interface{}
)

// Application represents a manager of freedom.
type Application struct {
	// prefixPath is the prefix of the application-level router
	prefixPath string

	// serviceLocator .
	serviceLocator *ServiceLocatorImpl

	// pool .
	pool *ServicePool

	// rpool .
	rpool *RepositoryPool

	// factoryPool .
	factoryPool *FactoryPool

	// comPool .
	comPool *InfraPool

	// msgsBus .
	msgsBus *EventBus

	// other .
	other *other

	// unmarshal is a global deserializer for deserialize every []byte into object.
	unmarshal func(data []byte, v interface{}) error

	// marshal is a global serializer for serialize every object into []byte.
	marshal func(v interface{}) ([]byte, error)

	// controllerDependencies contains the dependencies that provided for each
	// iris controller.
	controllerDependencies []interface{}

	// TODO(coco):
	//  Here is bad smell I felt, so we shouldn't export this member. Considering
	//  sets this member unexported in the future.
	//
	// IrisApp is an iris application
	IrisApp *IrisApplication

	// Database represents a database connection
	Database struct {
		db      interface{}
		Install func() (db interface{})
	}

	// Cache represents a redis client
	Cache struct {
		client  redis.Cmdable
		Install func() (client redis.Cmdable)
	}

	// Middleware is a set that satisfies the iris handler's definition
	Middleware []IrisHandler

	// Prometheus .
	Prometheus *Prometheus
}

// NewApplication creates an instance of Application exactly once while the program
// is running.
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
		globalApp.serviceLocator = newServiceLocator()
		globalApp.IrisApp.Logger().SetTimeFormat("2006-01-02 15:04:05.000")
	})
	return globalApp
}

// Iris .
func (app *Application) Iris() *IrisApplication {
	return app.IrisApp
}

// Logger .
func (app *Application) Logger() *golog.Logger {
	return app.Iris().Logger()
}

// ServiceLocator .
func (app *Application) ServiceLocator() *ServiceLocatorImpl {
	return app.serviceLocator
}

// TODO(coco):
//  Because of ambiguous naming, I've been create ServiceLocator as an alternative.
//  Considering remove this function in the future.
//
// GetServiceLocator .
func (app *Application) GetServiceLocator() *ServiceLocatorImpl {
	return app.ServiceLocator()
}

// GetService accepts an IrisContext and a pointer to the typed service. GetService
// looks up a service from the ServicePool by the type of the pointer and fill
// the pointer with the service.
func (app *Application) GetService(ctx IrisContext, service interface{}) {
	app.pool.get(ctx.Values().Get(WorkerKey).(*worker), service)
}

// GetInfra accepts an IrisContext and a pointer to the typed service. GetInfra
// looks up a service from the InfraPool by the type of the pointer and fill the
// pointer with the service.
func (app *Application) GetInfra(ctx IrisContext, com interface{}) {
	app.comPool.get(ctx.Values().Get(WorkerKey).(*worker), reflect.ValueOf(com).Elem())
}

// GetSingleInfra .
func (app *Application) GetSingleInfra(com interface{}) bool {
	return app.comPool.GetSingleInfra(reflect.ValueOf(com).Elem())
}

// SetPrefixPath assigns specified prefix to the application-level router.
func (app *Application) SetPrefixPath(prefixPath string) {
	app.prefixPath = prefixPath
}

// TODO(coco):
//  Because of ambiguous naming, I've been create SetPrefixPath as an alternative.
//  Considering remove this function in the future.
//
// InstallParty assigns specified prefix to the application-level router.
func (app *Application) InstallParty(prefixPath string) {
	app.SetPrefixPath(prefixPath)
}

// SubRouter accepts a string with the route path, and one or more IrisHandler.
// SubRouter makes a new path by concatenating the application-level router's
// prefix and the specified route path, creates an IrisRouter with the new path
// and the those specified IrisHandler, and returns the IrisRouter.
func (app *Application) SubRouter(relativePath string, handlers ...IrisHandler) IrisRouter {
	return app.Iris().Party(app.withPrefix(relativePath), handlers...)
}

// TODO(coco):
//  Because of ambiguous naming, I've been create SubRouter as an alternative.
//  Considering remove this function in the future.
//
// CreateParty accepts a string with the route path, and one or more IrisHandler.
// CreateParty makes a new path by concatenating the application-level router's
// prefix and the specified route path, creates an IrisRouter with the new path
// and the those specified IrisHandler, and returns the IrisRouter.
func (app *Application) CreateParty(relativePath string, handlers ...IrisHandler) IrisRouter {
	return app.SubRouter(relativePath, handlers...)
}

// BindController accepts a string with the route path, an IrisController, and
// one or more IrisHandler. BindController resolves the application's dependencies,
// and creates an IrisRouter and an IrisMVCApplication. the IrisMVCApplication
// would be attached on IrisRouter and the EventBus after it has created.
func (app *Application) BindController(relativePath string, controller IrisController, handlers ...IrisHandler) {
	router := app.SubRouter(relativePath, handlers...)

	mvcApp := mvc.New(router)
	mvcApp.Register(app.resolveDependencies()...)
	mvcApp.Handle(controller)

	app.msgsBus.addController(controller)
}

// BindControllerWithRouter accepts an IrisRouter and an IrisController.
// BindControllerWithRouter resolves the application's dependencies, and creates
// an IrisRouter and an IrisMVCApplication. the IrisMVCApplication would be attached
// on IrisRouter and the EventBus after it has created.
func (app *Application) BindControllerWithRouter(router IrisRouter, controller IrisController) {
	mvcApp := mvc.New(router)
	mvcApp.Register(app.resolveDependencies()...)
	mvcApp.Handle(controller)
}

// TODO(coco):
//  Because of naming ambiguous, I've been create BindControllerWithRouter as
//  an alternative. Considering remove this function in the future.
//
// BindControllerByParty accepts an IrisRouter and an IrisController.
// BindControllerByParty resolves the application's dependencies, and creates
// an IrisRouter and an IrisMVCApplication. the IrisMVCApplication would be attached
// on IrisRouter and the EventBus after it has created.
func (app *Application) BindControllerByParty(router IrisRouter, controller interface{}) {
	app.BindControllerWithRouter(router, controller)
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

// ListenEvent .
func (app *Application) ListenEvent(eventName string, objectMethod string, appointInfra ...interface{}) {
	app.msgsBus.addEvent(objectMethod, eventName, appointInfra...)
}

// EventsPath .
func (app *Application) EventsPath(infra interface{}) map[string]string {
	return app.msgsBus.EventsPath(infra)
}

// CacheWarmUp .
func (app *Application) CacheWarmUp(f func(repo *Repository)) {
	rb := new(Repository)
	f(rb)
}

// AsyncCacheWarmUp .
func (app *Application) AsyncCacheWarmUp(f func(repo *Repository)) {
	rb := new(Repository)
	go f(rb)
}

// InjectIntoController adds a Dependency for iris controller.
func (app *Application) InjectIntoController(f Dependency) {
	app.controllerDependencies = append(app.controllerDependencies, f)
}

// TODO(coco):
//  Because of ambiguous naming, I've been create InjectIntoController as an
//  alternative. Considering remove this function in the future.
//
// InjectController adds a Dependency for iris controller.
func (app *Application) InjectController(f Dependency) {
	app.InjectIntoController(f)
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

// InstallMiddleware .
func (app *Application) InstallMiddleware(handler IrisHandler) {
	app.Middleware = append(app.Middleware, handler)
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

// AddStarter adds a builder function which builds a Starter.
func (app *Application) AddStarter(f func(starter Starter)) {
	starters = append(starters, f)
}

// Start adds a builder function which builds a Starter.
func (app *Application) Start(f func(starter Starter)) {
	starters = append(starters, f)
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
func (app *Application) NewRunner(addr string, configurators ...IrisHostConfigurator) IrisRunner {
	return iris.Addr(addr, configurators...)
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
func (app *Application) NewAutoTLSRunner(addr string, domain string, email string, configurators ...IrisHostConfigurator) IrisRunner {
	return func(irisApp *IrisApplication) error {
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
func (app *Application) NewTLSRunner(addr string, certFile, keyFile string, configurators ...IrisHostConfigurator) IrisRunner {
	return func(irisApp *IrisApplication) error {
		return irisApp.NewHost(&http.Server{Addr: addr}).
			Configure(configurators...).
			ListenAndServeTLS(certFile, keyFile)
	}
}

//NewH2CRunner .
func (app *Application) NewH2CRunner(addr string, configurators ...IrisHostConfigurator) IrisRunner {
	h2cSer := &http2.Server{}
	ser := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(app.Iris(), h2cSer),
	}
	return func(irisApp *IrisApplication) error {
		return irisApp.NewHost(ser).Configure(configurators...).ListenAndServe()
	}
}

// Run accepts an IrisRunner and an IrisConfiguration. Run initializes the
// middleware and the infrastructure of the application likes database connection,
// cache, repository and etc. , and serving an iris application as the HTTP server
// to handle the incoming requests.
func (app *Application) Run(runner IrisRunner, conf IrisConfiguration) {
	app.addMiddlewares(conf)
	app.installDB()
	app.other.booting()
	for index := 0; index < len(prepares); index++ {
		prepares[index](app)
	}

	logLevel := "debug"
	if level, ok := conf.Other["logger_level"]; ok {
		logLevel = level.(string)
	}
	globalApp.IrisApp.Logger().SetLevel(logLevel)

	repositoryAPIRun(conf)
	for i := 0; i < len(starters); i++ {
		starters[i](app)
	}
	app.msgsBus.building()
	app.comPool.singleBooting(app)
	shutdownSecond := int64(2)
	if level, ok := conf.Other["shutdown_second"]; ok {
		shutdownSecond = level.(int64)
	}
	app.shutdown(shutdownSecond)
	app.Iris().Run(runner, iris.WithConfiguration(conf))
}

// withPrefix returns a string with a path prefixed by prefixPath.
func (app *Application) withPrefix(path string) string {
	return app.prefixPath + path
}

func (app *Application) shutdown(timeout int64) {
	iris.RegisterOnInterrupt(func() {
		//读取配置的关闭最长时间
		ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		close := func() {
			if err := recover(); err != nil {
				app.Logger().Errorf("[Freedom] An error was encountered during the program shutdown, %v", err)
			}
			app.comPool.shutdown()
		}
		close()
		//通知组件服务即将关闭
		app.Iris().Shutdown(ctx)
	})
}

func (app *Application) resolveDependencies() []Dependency {
	var result []Dependency

	result = append(result, func(ctx IrisContext) (rt Worker) {
		rt = ctx.Values().Get(WorkerKey).(Worker)
		return
	})
	result = append(result, app.controllerDependencies...)

	return result
}

func (app *Application) installDB() {
	if app.Database.Install != nil {
		app.Database.db = app.Database.Install()
	}

	if app.Cache.Install != nil {
		app.Cache.client = app.Cache.Install()
	}
}

func (app *Application) addMiddlewares(irisConf IrisConfiguration) {
	app.Iris().Use(newWorkerHandle())
	app.Iris().Use(globalApp.pool.freeHandle())
	app.Iris().Use(globalApp.comPool.freeHandle())
	if pladdr, ok := irisConf.Other["prometheus_listen_addr"]; ok {
		registerPrometheus(app.Prometheus, irisConf.Other["service_name"].(string), pladdr.(string))
		globalApp.IrisApp.Use(newPrometheusHandle(app.Prometheus))
	}
	globalApp.IrisApp.Use(app.Middleware...)
}
