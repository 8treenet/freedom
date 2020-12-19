package internal

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/kataras/iris/v12/context"
)

func callService(fun interface{}, worker ...Worker) {
	if len(worker) == 0 {
		worker = make([]Worker, 1)
		ctx := context.NewContext(globalApp.IrisApp)
		ctx.BeginRequest(nil, new(http.Request))
		rt := newWorker(ctx)
		ctx.Values().Set(WorkerKey, rt)
		worker[0] = rt
	}
	serviceObj, err := parseCallServiceFunc(fun)
	if err != nil {
		panic(fmt.Sprintf("[Freedom] CallService, %v : %s", fun, err.Error()))
	}
	newService := globalApp.pool.create(worker[0], serviceObj)
	reflect.ValueOf(fun).Call([]reflect.Value{reflect.ValueOf(newService.(serviceElement).serviceObject)})

	if worker[0].IsDeferRecycle() {
		return
	}
	globalApp.pool.free(newService)
}
