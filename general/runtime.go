package general

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

func newRuntimeHandle() context.Handler {
	return func(ctx context.Context) {
		rt := newRuntime(ctx)
		ctx.Values().Set(runtimeKey, rt)
		ctx.Next()
	}
}

func newRuntime(ctx iris.Context) *appRuntime {
	rt := new(appRuntime)
	rt.services = make([]interface{}, 0)
	rt.ctx = ctx
	rt.store = newStore()
	rt.traceID = ctx.GetHeader(globalApp.traceIDKey)
	if rt.traceID == "" {
		rt.traceID = UUID()
	}
	rt.store.Set(globalApp.traceIDKey, rt.traceID)
	ctx.Values().Set("logger_trace", rt.traceID)
	rt.logger = newFreedomLogger(rt.traceID)
	return rt
}

// appRuntime .
type appRuntime struct {
	ctx      iris.Context
	services []interface{}
	store    *Store
	logger   *freedomLogger
	traceID  string
}

// Ctx .
func (rt *appRuntime) Ctx() iris.Context {
	return rt.ctx
}

// Logger .
func (rt *appRuntime) Logger() Logger {
	return rt.logger
}

// Store .
func (rt *appRuntime) Store() *Store {
	return rt.store
}
