package internal

import (
	"sync"

	iris "github.com/8treenet/iris/v12"
	"github.com/8treenet/iris/v12/context"
)

var (
	globalApp     *Application
	globalAppOnce sync.Once
	prepares      []func(Initiator)
	bootManagers  []func(bootManager BootManager)
)

// Initiator Initialize the binding relationship between the user's code and the associated code.
type Initiator interface {
	// Accepts a string with the route path, and one or more IrisHandler.
	// CreateParty Makes a new path by concatenating the application-level router's
	// prefix and the specified route path, creates an IrisRouter with the new path
	// and the those specified IrisHandler, and returns the IrisRouter.
	CreateParty(relativePath string, handlers ...context.Handler) iris.Party
	// Accepts a string with the route path, an IrisController, and
	// one or more IrisHandler. BindController resolves the application's dependencies,
	// and creates an IrisRouter and an IrisMVCApplication. the IrisMVCApplication
	// would be attached on IrisRouter and the EventBus after it has created.
	BindController(relativePath string, controller interface{}, handlers ...context.Handler)
	// Accepts an IrisRouter and an IrisController.
	// BindControllerWithParty Resolves the application's dependencies, and creates
	// an IrisRouter and an IrisMVCApplication. the IrisMVCApplication would be attached
	// on IrisRouter and the EventBus after it has created.
	BindControllerWithParty(party iris.Party, controller interface{})
	// Bind a function that creates the service.
	BindService(f interface{})
	// Adds a Dependency for iris controller.
	// Because of ambiguous naming, I've been create InjectIntoController as an
	// alternative. Considering remove this function in the future.
	InjectController(f interface{})
	// Bind a function that creates the repository.
	BindRepository(f interface{})
	// Bind a function that creates the repository.
	BindFactory(f interface{})

	// Accepts an IrisContext and a pointer to the typed service. GetService
	// looks up a service from the ServicePool by the type of the pointer and fill
	// the pointer with the service.
	GetService(ctx iris.Context, service interface{})
	// Accepts an IrisContext and a pointer to the typed service. FetchService
	// looks up a service from the ServicePool by the type of the pointer and fill
	// the pointer with the service.
	FetchService(ctx iris.Context, service interface{})
	// Bind a function that creates the repository infrastructure.
	BindInfra(single bool, com interface{})
	// Accepts an IrisContext and a pointer to the typed service. GetInfra
	// looks up a service from the InfraPool by the type of the pointer and fill the
	// pointer with the service.
	GetInfra(ctx iris.Context, com interface{})
	// Accepts an IrisContext and a pointer to the typed service. GetInfra
	// looks up a service from the InfraPool by the type of the pointer and fill the
	// pointer with the service.
	FetchInfra(ctx iris.Context, com interface{})
	// You need to listen for message events using http agents.
	// eventName is the name of the event.
	// objectMethod is the controller that needs to be received.
	// sequential specifies whether to process messages sequentially (true) or concurrently (false).
	ListenEvent(eventName string, objectMethod string, sequential bool)
	// Adds a builder function which builds a BootManager.
	BindBooting(f func(bootManager BootManager))
	// Go back to the iris app.
	Iris() *iris.Application
}

// BootManager can be used to launch in the application.
type BootManager interface {
	// Go back to the iris app.
	Iris() *iris.Application
	// Gets a single-case component.
	FetchSingleInfra(infra interface{}) bool
	// Gets the event name and HTTP routing address for Listen.
	EventsPath(infra interface{}) map[string]string
	// Gets the sequential/concurrent configuration for each topic.
	EventsSequential(infra interface{}) map[string]bool
	// Register an inflatable callback function.
	RegisterShutdown(func())
}

// BeginRequest Requests to start the interface, and the call is triggered when the instance is implemented.
type BeginRequest interface {
	BeginRequest(Worker Worker)
}

// Prepare A prepared function is passed in for initialization.
func Prepare(f func(Initiator)) {
	prepares = append(prepares, f)
}
