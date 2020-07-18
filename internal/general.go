package internal

import (
	"sync"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

// Initiator .
type Initiator interface {
	CreateParty(relativePath string, handlers ...context.Handler) iris.Party
	BindController(relativePath string, controller interface{}, handlers ...context.Handler)
	BindControllerByParty(party iris.Party, controller interface{})
	BindService(f interface{})
	InjectController(f interface{})
	BindRepository(f interface{})
	BindFactory(f interface{})
	GetService(ctx iris.Context, service interface{})
	//BindInfra 如果是单例 com是对象， 如果是多例，com是函数
	BindInfra(single bool, com interface{})
	GetInfra(ctx iris.Context, com interface{})
	//监听事件
	ListenEvent(eventName string, objectMethod string, appointInfra ...interface{})
	Start(f func(starter Starter))
	Iris() *iris.Application
}

// Starter .
type Starter interface {
	Iris() *iris.Application
	//异步缓存预热
	AsyncCachePreheat(f func(repo *Repository))
	//同步缓存预热
	CachePreheat(f func(repo *Repository))
	GetSingleInfra(com interface{})
}

// SingleBoot .
type SingleBoot interface {
	Iris() *iris.Application
	EventsPath(infra interface{}) map[string]string
	Closeing(func())
}

// BeginRequest .
type BeginRequest interface {
	BeginRequest(Worker Worker)
}

var (
	globalApp     *Application
	globalAppOnce sync.Once
	prepares      []func(Initiator)
	starters      []func(starter Starter)
)

// Prepare app.BindController or app.BindControllerByParty.
func Prepare(f func(Initiator)) {
	prepares = append(prepares, f)
}
