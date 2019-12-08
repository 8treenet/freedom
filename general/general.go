package general

import (
	"sync"

	"github.com/kataras/golog"
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
	//BindComponent 如果是单例 com是对象， 如果是多例，com是函数
	BindComponent(single bool, com interface{})
	GetComponent(ctx iris.Context, com interface{})
	//监听消息
	ListenMessage(topic string, controller interface{}, funName string)
}

// SingleBoot .
type SingleBoot interface {
	Iris() *iris.Application
	Logger() *golog.Logger
	MessagesPath() map[string]string
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
