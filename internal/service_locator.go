package internal

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/kataras/iris/v12/context"
)

func newServiceLocator() *ServiceLocatorImpl {
	return &ServiceLocatorImpl{}
}

// ServiceLocatorImpl .
type ServiceLocatorImpl struct {
	beginCallBack []func(Worker)
	endCallBack   []func(Worker)
}

// InstallBeginCallBack .
func (locator *ServiceLocatorImpl) InstallBeginCallBack(f func(Worker)) {
	locator.beginCallBack = append(locator.beginCallBack, f)
}

// InstallEndCallBack .
func (locator *ServiceLocatorImpl) InstallEndCallBack(f func(Worker)) {
	locator.endCallBack = append(locator.endCallBack, f)
}

// Call .
func (locator *ServiceLocatorImpl) Call(fun interface{}) {
	ctx := context.NewContext(globalApp.IrisApp)
	ctx.BeginRequest(nil, new(http.Request))
	worker := newWorker(ctx)
	ctx.Values().Set(WorkerKey, worker)
	worker.bus = newBus(make(http.Header))

	serviceObj, err := parseCallServiceFunc(fun)
	if err != nil {
		panic(fmt.Sprintf("[Freedom] ServiceLocatorImpl.Call, %v : %s", fun, err.Error()))
	}
	newService := globalApp.pool.create(worker, serviceObj)
	for _, beginCallBack := range locator.beginCallBack {
		beginCallBack(worker)
	}
	reflect.ValueOf(fun).Call([]reflect.Value{reflect.ValueOf(newService.(serviceElement).serviceObject)})
	for _, endCallBack := range locator.endCallBack {
		endCallBack(worker)
	}

	if worker.IsDeferRecycle() {
		return
	}
	globalApp.pool.free(newService)
}
