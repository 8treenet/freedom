package internal

import (
	"math/rand"
	"reflect"
	"time"

	stdContext "context"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/memstore"
)

const (
	// WorkerKey specify where is the Worker should be located in Context
	WorkerKey = "STORE-WORKER-KEY"
)

// Worker describes a global context which use to share the internal component
// (i.e infrastructure, transaction, logger and so on) with middleware,
// controller, domain service and etc.
type Worker interface {
	// IrisContext point to current iris.Context.
	IrisContext() iris.Context

	// Logger returns current Logger.
	Logger() Logger

	// SetLogger set current Logger instead Logger.
	SetLogger(Logger)

	// Store returns an address to current memstore.Store.
	//
	// memstore.Store is a collection of key/value entries. usually use to store metadata produced by freedom runtime.
	Store() *memstore.Store

	// Bus returns an address to current Bus.
	Bus() *Bus

	// Context returns current context.
	Context() stdContext.Context

	// WithContext set current context instead Context.
	WithContext(stdContext.Context)

	// StartTime returns a time since this Worker be created.
	StartTime() time.Time

	// DeferRecycle marks the resource won't be recycled immediately after
	// the request has ended.
	//
	// It is a bit hard to understand what is, here is a simply explain about
	// this.
	//
	// When an HTTP request is incoming, the program will probably serve a bunch
	// of business logic services, DB connections, transactions, Redis caches,
	// and so on. Once those procedures has done, the system should write
	// response and release those resource immediately. In other words, the
	// system should do some clean up procedures for this request. You might
	// thought it is a matter of course. But in special cases, such as goroutine
	// without synchronizing-signal. When all business procedures has completed
	// on business goroutine, and prepare to respond. GC system may be run before
	// the http handler goroutine to respond to the client. Once this opportunity
	// was met, the client will got an "Internal Server Error" or other wrong
	// result, because resource has been recycled by GC before to respond to client.
	DeferRecycle()

	// IsDeferRecycle indicates system need to wait a while for recycle resource.
	IsDeferRecycle() bool

	// Rand returns a rand.Rand act a random number seeder.
	Rand() *rand.Rand
}

func newWorkerHandle() context.Handler {
	return func(ctx context.Context) {
		work := newWorker(ctx)
		ctx.Values().Set(WorkerKey, work)
		ctx.Next()

		if work.IsDeferRecycle() {
			return
		}
		work.logger = nil
		work.ctx = nil
		ctx.Values().Reset()
	}
}

// newWorker creates a Worker from iris.Context. This procedure will read the
// request header, and also initialize bus with that.
func newWorker(ctx iris.Context) *worker {
	work := new(worker)
	work.freeServices = make([]interface{}, 0)
	work.freeComs = make([]interface{}, 0)
	work.ctx = ctx
	work.bus = newBus(ctx.Request().Header)
	work.stdCtx = ctx.Request().Context()
	work.time = time.Now()
	work.deferRecycle = false
	HandleBusMiddleware(work)
	return work
}

// worker act as a default implementation to Worker.
type worker struct {
	ctx          iris.Context
	freeServices []interface{}
	freeComs     []interface{}
	logger       Logger
	bus          *Bus
	stdCtx       stdContext.Context
	time         time.Time
	values       memstore.Store
	deferRecycle bool
	randInstance *rand.Rand
}

// Ctx .
func (rt *worker) IrisContext() iris.Context {
	return rt.ctx
}

// Context .
func (rt *worker) Context() stdContext.Context {
	return rt.stdCtx
}

// WithContext .
func (rt *worker) WithContext(ctx stdContext.Context) {
	rt.stdCtx = ctx
}

// StartTime .
func (rt *worker) StartTime() time.Time {
	return rt.time
}

// Logger .
func (rt *worker) Logger() Logger {
	if rt.logger == nil {
		l := rt.values.Get("logger_impl")
		if l == nil {
			rt.logger = globalApp.Logger()
		} else {
			rt.logger = l.(Logger)
		}
	}
	return rt.logger
}

// Store .
func (rt *worker) Store() *memstore.Store {
	return &rt.values
}

// Bus .
func (rt *worker) Bus() *Bus {
	return rt.bus
}

// DeferRecycle .
func (rt *worker) DeferRecycle() {
	rt.deferRecycle = true
}

// IsDeferRecycle .
func (rt *worker) IsDeferRecycle() bool {
	return rt.deferRecycle
}

// Rand .
func (rt *worker) Rand() *rand.Rand {
	if rt.randInstance == nil {
		rt.randInstance = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return rt.randInstance
}

// SetLogger .
func (rt *worker) SetLogger(l Logger) {
	rt.logger = l
}

var workerType = reflect.TypeOf(&worker{})
