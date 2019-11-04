package general

import (
	"net/http"
	"sync"

	"github.com/kataras/golog"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Initiator .
type Initiator interface {
	CreateParty(relativePath string, handlers ...context.Handler) iris.Party
	BindController(relativePath string, controller interface{}, service ...interface{})
	BindControllerByParty(party iris.Party, controller interface{}, service ...interface{})
	BindService(obj interface{}, f func() interface{})
	GetService(ctx iris.Context, service interface{})
	AsyncCachePreheat(f func(repoDB *RepositoryDB, repoCache *RepositoryCache))
	CachePreheat(f func(repoDB *RepositoryDB, repoCache *RepositoryCache))
}

// BeginRequest .
type BeginRequest interface {
	BeginRequest(runtime Runtime)
}

// Runtime .
type Runtime interface {
	Context() iris.Context
	Logger() *golog.Logger
	Store() *Store
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
