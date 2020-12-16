package internal

import (
	"net/http"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/kataras/iris/v12/context"
)

var _ UnitTest = (*UnitTestImpl)(nil)

// UnitTest .
type UnitTest interface {
	GetService(service interface{})
	GetRepository(repository interface{})
	GetFactory(factory interface{})
	InstallDB(f func() (db interface{}))
	InstallRedis(f func() (client redis.Cmdable))
	Run()
	SetRequest(request *http.Request)
	InjectBaseEntity(entity interface{})
}

// UnitTestImpl .
type UnitTestImpl struct {
	rt      *worker
	request *http.Request
}

// GetService .
func (u *UnitTestImpl) GetService(service interface{}) {
	globalApp.GetService(u.rt.IrisContext(), service)
}

// GetRepository .
func (u *UnitTestImpl) GetRepository(repository interface{}) {
	value := reflect.ValueOf(repository).Elem()
	ok, newfield := globalApp.rpool.get(value.Type())
	if !ok {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable, %v" + value.Type().String())
	}

	repo := newfield.Interface()
	globalApp.comPool.diInfra(repo)
	globalApp.pool.beginRequest(u.rt, repo)
	value.Set(newfield)
}

// GetFactory .
func (u *UnitTestImpl) GetFactory(factory interface{}) {
	value := reflect.ValueOf(factory).Elem()
	ok, newfield := globalApp.factoryPool.get(value.Type())
	if !ok {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		globalApp.IrisApp.Logger().Fatalf("[Freedom] This use repository object must be a capital variable, %v" + value.Type().String())
	}

	factoryObj := newfield.Interface()
	globalApp.comPool.diInfra(factoryObj)
	globalApp.rpool.diRepo(factoryObj, nil)
	globalApp.pool.factoryCall(u.rt, reflect.ValueOf(u.rt), newfield)
	value.Set(newfield)
}

// InstallDB .
func (u *UnitTestImpl) InstallDB(f func() (db interface{})) {
	globalApp.InstallDB(f)
}

// InstallRedis .
func (u *UnitTestImpl) InstallRedis(f func() (client redis.Cmdable)) {
	globalApp.InstallRedis(f)
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
	globalApp.comPool.singleBooting(globalApp)
}

func (u *UnitTestImpl) newRuntime() *worker {
	ctx := context.NewContext(globalApp.IrisApp)
	if u.request == nil {
		u.request = new(http.Request)
	}
	ctx.BeginRequest(nil, u.request)
	rt := newWorker(ctx)
	ctx.Values().Set(WorkerKey, rt)
	return rt
}

// SetRequest .
func (u *UnitTestImpl) SetRequest(request *http.Request) {
	u.request = request
}

// InjectBaseEntity .
func (u *UnitTestImpl) InjectBaseEntity(entity interface{}) {
	injectBaseEntity(u.rt, entity)
	return
}
