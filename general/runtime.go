package general

import (
	"reflect"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/memstore"
)

func newRuntimeHandle() context.Handler {
	return func(ctx context.Context) {
		rt := newRuntime(ctx)
		ctx.Values().Set(RuntimeKey, rt)
		ctx.Next()
		ctx.Values().Reset()
		rt.ctx = nil
		rt.store = nil
	}
}

func newRuntime(ctx iris.Context) *appRuntime {
	rt := new(appRuntime)
	rt.freeServices = make([]interface{}, 0)
	rt.coms = make(map[reflect.Type]interface{})
	rt.ctx = ctx
	rt.store = ctx.Values()
	rt.bus = newBus(ctx.Request().Header)
	return rt
}

// appRuntime .
type appRuntime struct {
	ctx          iris.Context
	freeServices []interface{}
	coms         map[reflect.Type]interface{}
	store        *memstore.Store
	logger       Logger
	bus          *Bus
}

// Ctx .
func (rt *appRuntime) Ctx() iris.Context {
	return rt.ctx
}

// Logger .
func (rt *appRuntime) Logger() Logger {
	if rt.logger == nil {
		l := rt.ctx.Values().Get("logger_impl")
		if l == nil {
			rt.logger = rt.ctx.Application().Logger()
		} else {
			rt.logger = l.(Logger)
		}
	}
	return rt.logger
}

// Store .
func (rt *appRuntime) Store() *memstore.Store {
	return rt.ctx.Values()
}

// Bus .
func (rt *appRuntime) Bus() *Bus {
	return rt.bus
}
