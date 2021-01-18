package internal

import (
	"sync"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

var (
	globalApp     *Application
	globalAppOnce sync.Once
	prepares      []func(Initiator)
	starters      []func(starter Starter)
)

// Initiator .
type Initiator interface {
	CreateParty(relativePath string, handlers ...context.Handler) iris.Party
	BindController(relativePath string, controller interface{}, handlers ...context.Handler)
	BindControllerWithParty(party iris.Party, controller interface{})
	BindService(f interface{})
	InjectController(f interface{})
	BindRepository(f interface{})
	BindFactory(f interface{})
	GetService(ctx iris.Context, service interface{})
	// BindInfra if is a singleton, com is an object. if is multiton, com is a function
	BindInfra(single bool, com interface{})
	GetInfra(ctx iris.Context, com interface{})
	// Listen Event
	ListenEvent(eventName string, objectMethod string, appointInfra ...interface{})
	Start(f func(starter Starter))
	Iris() *iris.Application
}

// Starter .
type Starter interface {
	Iris() *iris.Application
	// Asynchronous cache warm-up
	AsyncCacheWarmUp(f func(repo *Repository))
	// Sync cache warm-up
	CacheWarmUp(f func(repo *Repository))
	GetSingleInfra(com interface{}) bool
}

// SingleBoot represents singleton startup component.
type SingleBoot interface {
	Iris() *iris.Application
	//Pass in the current component to get the event path, which can bind the specified component.
	EventsPath(infra interface{}) map[string]string
	//Register for a shutdown event.
	RegisterShutdown(func())
}

// BeginRequest .
type BeginRequest interface {
	BeginRequest(Worker Worker)
}

// Prepare app.BindController or app.BindControllerByParty.
func Prepare(f func(Initiator)) {
	prepares = append(prepares, f)
}
