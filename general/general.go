package general

import (
	"sync"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/core/memstore"
)

// Initiator .
type Initiator interface {
	CreateParty(relativePath string, handlers ...context.Handler) iris.Party
	BindController(relativePath string, controller interface{}, service ...interface{})
	BindControllerByParty(party iris.Party, controller interface{}, service ...interface{})
	BindService(f interface{})
	InjectController(f interface{})
	BindRepository(f interface{})
	GetService(ctx iris.Context, service interface{})
	AsyncCachePreheat(f func(repo *Repository))
	CachePreheat(f func(repo *Repository))
	//BindInfra 如果是单例 com是对象， 如果是多例，com是函数
	BindInfra(single bool, com interface{})
	GetInfra(ctx iris.Context, com interface{})
	//监听事件
	ListenEvent(eventName string, fun interface{}, appointInfra ...interface{})
}

// SingleBoot .
type SingleBoot interface {
	Iris() *iris.Application
	EventsPath(infra interface{}) map[string][]string
}

// BeginRequest .
type BeginRequest interface {
	BeginRequest(runtime Runtime)
}

// Runtime .
type Runtime interface {
	Ctx() iris.Context
	Logger() Logger
	Store() *memstore.Store
	Prometheus() *Prometheus
}

var (
	globalApp     *Application
	globalAppOnce sync.Once
	boots         []func(Initiator)
)

// Booting app.BindController or app.BindControllerByParty.
func Booting(f func(Initiator)) {
	boots = append(boots, f)
}
