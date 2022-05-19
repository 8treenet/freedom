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
var _ BootManager = (*Application)(nil)

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

	// Service locator .
	serviceLocator *ServiceLocatorImpl

	// A pool of reusable services .
	pool *servicePool

	// Repository pools that can be reused .
	rpool *repositoryPool

	// Reusable Fact pool .
	factoryPool *factoryPool

	// A pool of reusable components .
	comPool *infraPool

	// Subscribe to the message conversion HTTP API .
	subEventManager *eventPathManager

	// Custom data sources .
	other *custom

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

// NewApplication Creates an instance of Application exactly once while the program
// is running.
func NewApplication() *Application {
	globalAppOnce.Do(func() {
		globalApp = new(Application)
		globalApp.IrisApp = iris.New()
		globalApp.pool = newServicePool()
		globalApp.rpool = newRepoPool()
		globalApp.factoryPool = newFactoryPool()
		globalApp.comPool = newInfraPool()
		globalApp.subEventManager = newEventPathManager()
		globalApp.other = newCustom()
		globalApp.marshal = json.Marshal
		globalApp.unmarshal = json.Unmarshal
		globalApp.Prometheus = newPrometheus()
		globalApp.serviceLocator = newServiceLocator()
		globalApp.IrisApp.Logger().SetTimeFormat("2006-01-02 15:04:05.000")
	})
	return globalApp
}

// Iris Go back to the iris app.
func (app *Application) Iris() *IrisApplication {
	return app.IrisApp
}

// Logger Return to global logger.
func (app *Application) Logger() *golog.Logger {
	return app.Iris().Logger()
}

// GetServiceLocator .
// Because of ambiguous naming, I've been create ServiceLocator as an alternative.
// Considering remove this function in the future.
func (app *Application) GetServiceLocator() *ServiceLocatorImpl {
	return app.serviceLocator
}

// GetService accepts an IrisContext and a pointer to the typed service. GetService
// looks up a service from the ServicePool by the type of the pointer and fill
// the pointer with the service.
func (app *Application) GetService(ctx IrisContext, service interface{}) {
	app.FetchService(ctx, service)
}

// FetchService accepts an IrisContext and a pointer to the typed service. FetchService
// looks up a service from the ServicePool by the type of the pointer and fill
// the pointer with the service.
func (app *Application) FetchService(ctx IrisContext, service interface{}) {
	app.pool.get(ctx.Values().Get(WorkerKey).(*worker), service)
}

// GetInfra accepts an IrisContext and a pointer to the typed service. GetInfra
// looks up a service from the InfraPool by the type of the pointer and fill the
// pointer with the service.
func (app *Application) GetInfra(ctx IrisContext, com interface{}) {
	app.FetchInfra(ctx, com)
}

// FetchInfra accepts an IrisContext and a pointer to the typed service. GetInfra
// looks up a service from the InfraPool by the type of the pointer and fill the
// pointer with the service.
func (app *Application) FetchInfra(ctx IrisContext, com interface{}) {
	app.comPool.get(ctx.Values().Get(WorkerKey).(*worker), reflect.ValueOf(com).Elem())
}

// FetchSingleInfra .
func (app *Application) FetchSingleInfra(com interface{}) bool {
	return app.comPool.FetchSingleInfra(reflect.ValueOf(com).Elem())
}

// SetPrefixPath assigns specified prefix to the application-level router.
func (app *Application) SetPrefixPath(prefixPath string) {
	app.prefixPath = prefixPath
}

// InstallParty assigns specified prefix to the application-level router.
// Because of ambiguous naming, I've been create SetPrefixPath as an alternative.
// Considering remove this function in the future.
func (app *Application) InstallParty(prefixPath string) {
	app.SetPrefixPath(prefixPath)
}

// SubRouter accepts a string with the route path, and one or more IrisHandler.
// SubRouter makes a new path by concatenating the application-level router's
// prefix and the specified route path, creates an IrisRouter with the new path
// and the those specified IrisHandler, and returns the IrisRouter.
func (app *Application) SubRouter(relativePath string, handlers ...IrisHandler) IrisParty {
	return app.Iris().Party(app.withPrefix(relativePath), handlers...)
}

// CreateParty accepts a string with the route path, and one or more IrisHandler.
// CreateParty makes a new path by concatenating the application-level router's
// prefix and the specified route path, creates an IrisRouter with the new path
// and the those specified IrisHandler, and returns the IrisRouter.
func (app *Application) CreateParty(relativePath string, handlers ...IrisHandler) IrisParty {
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

	app.subEventManager.addController(controller)
}

// BindControllerWithParty accepts an IrisRouter and an IrisController.
// BindControllerWithParty resolves the application's dependencies, and creates
// an IrisRouter and an IrisMVCApplication. the IrisMVCApplication would be attached
// on IrisRouter and the EventBus after it has created.
func (app *Application) BindControllerWithParty(router IrisParty, controller IrisController) {
	mvcApp := mvc.New(router)
	mvcApp.Register(app.resolveDependencies()...)
	mvcApp.Handle(controller)
}

// BindControllerByParty accepts an IrisRouter and an IrisController.
// BindControllerByParty resolves the application's dependencies, and creates
// an IrisRouter and an IrisMVCApplication. the IrisMVCApplication would be attached
// on IrisRouter and the EventBus after it has created.
func (app *Application) BindControllerByParty(router IrisParty, controller interface{}) {
	app.BindControllerWithParty(router, controller)
}

// BindService Bind a function that creates the service.
func (app *Application) BindService(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		globalApp.Logger().Fatalf("[Freedom] BindService: The binding function is incorrect, %v : %s", f, err.Error())
	}
	app.pool.bind(outType, f)
}

// BindRepository Bind a function that creates the repository.
func (app *Application) BindRepository(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		globalApp.Logger().Fatalf("[Freedom] BindRepository: The binding function is incorrect, %v : %s", f, err.Error())
	}
	app.rpool.bind(outType, f)
}

// BindFactory Bind a function that creates the repository.
func (app *Application) BindFactory(f interface{}) {
	outType, err := parsePoolFunc(f)
	if err != nil {
		globalApp.Logger().Fatalf("[Freedom] BindFactory: The binding function is incorrect, %v : %s", f, err.Error())
	}
	app.factoryPool.bind(outType, f)
}

// BindInfra Bind a function that creates the repository infrastructure.
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

// ListenEvent You need to listen for message events using http agents.
// eventName is the name of the event.
// objectMethod is the controller that needs to be received.
// appointInfra can delegate infrastructure.
func (app *Application) ListenEvent(eventName string, objectMethod string, appointInfra ...interface{}) {
	app.subEventManager.addEvent(objectMethod, eventName, appointInfra...)
}

// EventsPath Gets the event name and HTTP routing address for Listen.
func (app *Application) EventsPath(infra interface{}) map[string]string {
	return app.subEventManager.EventsPath(infra)
}

// InjectIntoController Adds a Dependency for iris controller.
func (app *Application) InjectIntoController(f Dependency) {
	app.controllerDependencies = append(app.controllerDependencies, f)
}

// InjectController adds a Dependency for iris controller.
// Because of ambiguous naming, I've been create InjectIntoController as an
// alternative. Considering remove this function in the future.
func (app *Application) InjectController(f Dependency) {
	app.InjectIntoController(f)
}

// RegisterShutdown Register an inflatable callback function.
func (app *Application) RegisterShutdown(f func()) {
	app.comPool.registerShutdown(f)
}

// InstallDB Install the database.
// A function that returns the Interface type is passed in as a handle to the database.
func (app *Application) InstallDB(f func() interface{}) {
	app.Database.Install = f
}

// InstallRedis Install the redis.
// A function that returns the redis.Cmdable is passed in as a handle to the client.
func (app *Application) InstallRedis(f func() (client redis.Cmdable)) {
	app.Cache.Install = f
}

// InstallMiddleware Install the middleware for http routing globally.
func (app *Application) InstallMiddleware(handler IrisHandler) {
	app.Middleware = append(app.Middleware, handler)
}

// InstallCustom Install a custom data source.
// A function that returns the Interface type is passed in as a handle to the custom data sources.
func (app *Application) InstallCustom(f func() interface{}) {
	app.other.add(f)
}

// InstallBusMiddleware Install the middleware for the message bus.
func (app *Application) InstallBusMiddleware(handle ...BusHandler) {
	busMiddlewares = append(busMiddlewares, handle...)
}

// InstallSerializer .
func (app *Application) InstallSerializer(marshal func(v interface{}) ([]byte, error), unmarshal func(data []byte, v interface{}) error) {
	app.marshal = marshal
	app.unmarshal = unmarshal
}

// BindBooting Adds a builder function which builds a BootManager.
func (app *Application) BindBooting(f func(bootManager BootManager)) {
	bootManagers = append(bootManagers, f)
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

// NewH2CRunner create a Runner for H2C .
// intercepting any h2c
// traffic. If a request is an h2c connection, it's hijacked and redirected to
// s.ServeConn. Otherwise the returned Handler just forwards requests to h. This
// works because h2c is designed to be parseable as valid HTTP/1, but ignored by
// any HTTP server that does not handle h2c. Therefore we leverage the HTTP/1
// compatible parts of the Go http library to parse and recognize h2c requests.
// Once a request is recognized as h2c, we hijack the connection and convert it
// to an HTTP/2 connection which is understandable to s.ServeConn. (s.ServeConn
// understands HTTP/2 except for the h2c part of it.)
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
	logLevel := "debug"
	if level, ok := conf.Other["logger_level"]; ok {
		logLevel = level.(string)
	}
	globalApp.IrisApp.Logger().SetLevel(logLevel)

	app.addMiddlewares(conf)
	app.installDB()
	app.other.booting()
	for index := 0; index < len(prepares); index++ {
		prepares[index](app)
	}

	repositoryAPIRun(conf)
	app.subEventManager.building()

	app.comPool.singleBooting(app)
	for i := 0; i < len(bootManagers); i++ {
		bootManagers[i](app)
	}

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
