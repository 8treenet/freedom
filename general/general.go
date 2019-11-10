package general

import (
	"net/http"
	"sync"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/core/memstore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Initiator .
type Initiator interface {
	CreateParty(relativePath string, handlers ...context.Handler) iris.Party
	BindController(relativePath string, controller interface{}, service ...interface{})
	BindControllerByParty(party iris.Party, controller interface{}, service ...interface{})
	BindService(f interface{})
	BindRepository(f interface{})
	GetService(ctx iris.Context, service interface{})
	AsyncCachePreheat(f func(repo *Repository))
	CachePreheat(f func(repo *Repository))
	//BindComponent 如果是单例 com是对象， 如果是多例，com是函数
	BindComponent(single bool, com interface{})
	GetComponent(ctx iris.Context, com interface{})
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

// CreateH2Server .
func CreateH2Server(app *Application, addr string) *http.Server {
	h2cSer := &http2.Server{}
	ser := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(app.IrisApp, h2cSer),
	}
	return ser
}
