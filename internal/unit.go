package internal

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/8treenet/iris/v12/context"
	"github.com/go-redis/redis"
)

var _ UnitTest = (*UnitTestImpl)(nil)

// UnitTest Unit testing tools.
type UnitTest interface {
	FetchService(service interface{})
	FetchRepository(repository interface{})
	FetchFactory(factory interface{})
	InstallDB(f func() (db interface{}))
	InstallRedis(f func() (client redis.Cmdable))
	InstallCustom(f func() interface{})
	Run()
	SetRequest(request *http.Request)
	InjectBaseEntity(entity interface{})
}

// UnitTestImpl Unit testing tools.
type UnitTestImpl struct {
	rt *worker
}

// NewUnitTest .
func NewUnitTest() UnitTest {
	return new(UnitTestImpl)
}

// FetchService .
func (u *UnitTestImpl) FetchService(service interface{}) {
	globalApp.FetchService(u.rt.IrisContext(), service)
}

// FetchRepository .
func (u *UnitTestImpl) FetchRepository(repository interface{}) {
	instance := serviceElement{calls: []BeginRequest{}, workers: []reflect.Value{}}
	value := reflect.ValueOf(repository).Elem()
	ok := globalApp.rpool.diRepoFromValue(value, &instance)
	if !ok {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable, %v" + value.Type().String())
	}

	if br, ok := value.Interface().(BeginRequest); ok {
		instance.calls = append(instance.calls, br)
	}
	globalApp.pool.beginRequest(u.rt, instance)
}

// FetchFactory .
func (u *UnitTestImpl) FetchFactory(factory interface{}) {
	instance := serviceElement{calls: []BeginRequest{}, workers: []reflect.Value{}}
	value := reflect.ValueOf(factory).Elem()
	ok := globalApp.factoryPool.diFactoryFromValue(value, &instance)
	if !ok {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable, %v" + value.Type().String())
	}

	globalApp.pool.beginRequest(u.rt, instance)
}

// InstallDB .
func (u *UnitTestImpl) InstallDB(f func() (db interface{})) {
	globalApp.InstallDB(f)
}

// InstallRedis .
func (u *UnitTestImpl) InstallRedis(f func() (client redis.Cmdable)) {
	globalApp.InstallRedis(f)
}

// InstallCustom .
func (u *UnitTestImpl) InstallCustom(f func() interface{}) {
	globalApp.InstallCustom(f)
}

// Run .
func (u *UnitTestImpl) Run() {
	for index := 0; index < len(prepares); index++ {
		prepares[index](globalApp)
	}
	u.rt = u.newRuntime()
	logLevel := "debug"

	globalApp.IrisApp.Logger().SetLevel(logLevel)
	globalApp.installDB()
	globalApp.other.booting()
	globalApp.comPool.singleBooting(globalApp)
}

func (u *UnitTestImpl) newRuntime() *worker {
	ctx := context.NewContext(globalApp.IrisApp)
	request := new(http.Request)
	request.URL = &url.URL{}
	ctx.ResetRequest(request)

	rt := newWorker(ctx)
	ctx.Values().Set(WorkerKey, rt)
	rt.bus = newBus(make(http.Header))
	ctx.Values().Set(WorkerKey, rt)
	return rt
}

// SetRequest .
func (u *UnitTestImpl) SetRequest(request *http.Request) {
	u.rt.IrisContext().ResetRequest(request)
}

// InjectBaseEntity .
func (u *UnitTestImpl) InjectBaseEntity(entity interface{}) {
	injectBaseEntity(u.rt, entity)
	return
}
