package general

import (
	"github.com/kataras/golog"
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
	rt.logger = ctx.Application().Logger().Clone()
	rt.logger.Handle(rt.traceLog)

	rt.traceID = ctx.GetHeader(globalApp.traceIDKey)
	if rt.traceID == "" {
		rt.traceID = UUID()
	}
	rt.store.Set(globalApp.traceIDKey, rt.traceID)
	ctx.Values().Set("logger_trace", rt.traceID)
	return rt
}

// appRuntime .
type appRuntime struct {
	ctx      iris.Context
	services []interface{}
	store    *Store
	logger   *golog.Logger
	traceID  string
}

// Context .
func (rt *appRuntime) Context() iris.Context {
	return rt.ctx
}

// Logger .
func (rt *appRuntime) Logger() *golog.Logger {
	return rt.logger
}

// Store .
func (rt *appRuntime) Store() *Store {
	return rt.store
}

// traceLog .
func (rt *appRuntime) traceLog(value *golog.Log) bool {
	value.Message = rt.traceID + " " + value.Message
	return false
}
