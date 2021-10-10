package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/kataras/iris/v12/context"
)

func newServiceLocator() *ServiceLocatorImpl {
	return &ServiceLocatorImpl{}
}

// ServiceLocatorImpl The implementation of service positioning.
type ServiceLocatorImpl struct {
	beginCallBack []func(Worker)
	endCallBack   []func(Worker)
}

// InstallBeginCallBack Install the callback function that started.
// Triggered before the callback function.
func (locator *ServiceLocatorImpl) InstallBeginCallBack(f func(Worker)) {
	locator.beginCallBack = append(locator.beginCallBack, f)
}

// InstallEndCallBack The callback function at the end of the installation.
// Triggered after the callback function.
func (locator *ServiceLocatorImpl) InstallEndCallBack(f func(Worker)) {
	locator.endCallBack = append(locator.endCallBack, f)
}

// Call Called with a service locator.
func (locator *ServiceLocatorImpl) Call(fun interface{}) error {
	ctx := context.NewContext(globalApp.IrisApp)
	request := new(http.Request)
	request.URL = &url.URL{}
	ctx.ResetRequest(request)

	worker := newWorker(ctx)
	ctx.Values().Set(WorkerKey, worker)
	worker.bus = newBus(make(http.Header))

	serviceObj, err := parseCallServiceFunc(fun)
	if err != nil {
		return fmt.Errorf("[Freedom] ServiceLocatorImpl.Call, %v : %w", fun, err)
	}
	newService, err := globalApp.pool.create(worker, serviceObj)
	if err != nil {
		return fmt.Errorf("[Freedom] ServiceLocatorImpl.Call, %v : %w", fun, err)
	}
	for _, beginCallBack := range locator.beginCallBack {
		beginCallBack(worker)
	}
	returnValue := reflect.ValueOf(fun).Call([]reflect.Value{reflect.ValueOf(newService.(serviceElement).serviceObject)})
	for _, endCallBack := range locator.endCallBack {
		endCallBack(worker)
	}

	for {
		if worker.IsDeferRecycle() {
			break
		}
		globalApp.pool.free(newService)
		break
	}
	if len(returnValue) > 0 && !returnValue[0].IsNil() {
		returnInterface := returnValue[0].Interface()
		err, _ := returnInterface.(error)
		return err
	}
	return nil
}
