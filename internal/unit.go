package internal

import (
	"net/http"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/kataras/iris/v12/context"
)

var _ UnitTest = new(UnitTestImpl)

type UnitTest interface {
	GetService(service interface{})
	GetRepository(repository interface{})
	GetFactory(factory interface{})
	InstallDB(f func() (db interface{}))
	InstallRedis(f func() (client redis.Cmdable))
	Run()
	SetRequest(request *http.Request)
	InjectBaseEntity(entity interface{})
	InstallDomainEventInfra(eventInfra DomainEventInfra)
	NewDomainEventInfra(call ...func(producer, topic string, data []byte, header map[string]string)) DomainEventInfra
}

type UnitTestImpl struct {
	rt      *worker
	request *http.Request
}

func (u *UnitTestImpl) GetService(service interface{}) {
	globalApp.GetService(u.rt.IrisContext(), service)
}

func (u *UnitTestImpl) GetRepository(repository interface{}) {
	value := reflect.ValueOf(repository).Elem()
	ok, newfield := globalApp.rpool.get(value.Type())
	if !ok {
		globalApp.IrisApp.Logger().Fatalf("[freedom]No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		globalApp.IrisApp.Logger().Fatalf("[freedom]This use repository object must be a capital variable, %v" + value.Type().String())
	}

	repo := newfield.Interface()
	globalApp.comPool.diInfra(repo)
	globalApp.pool.objBeginRequest(u.rt, repo)
	value.Set(newfield)
}

func (u *UnitTestImpl) GetFactory(factory interface{}) {
	value := reflect.ValueOf(factory).Elem()
	ok, newfield := globalApp.factoryPool.get(value.Type())
	if !ok {
		globalApp.IrisApp.Logger().Fatalf("[freedom]No dependency injection was found for the object,%v", value.Type().String())
	}
	if !value.CanSet() {
		globalApp.IrisApp.Logger().Fatalf("[freedom]This use repository object must be a capital variable, %v" + value.Type().String())
	}

	factoryObj := newfield.Interface()
	globalApp.comPool.diInfra(factoryObj)
	globalApp.rpool.diRepo(factoryObj)
	globalApp.pool.factoryCall(u.rt, reflect.ValueOf(u.rt), newfield)
	value.Set(newfield)
}

func (u *UnitTestImpl) InstallDB(f func() (db interface{})) {
	globalApp.InstallDB(f)
}

func (u *UnitTestImpl) InstallRedis(f func() (client redis.Cmdable)) {
	globalApp.InstallRedis(f)
}

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

func (u *UnitTestImpl) SetRequest(request *http.Request) {
	u.request = request
}

func (u *UnitTestImpl) InstallDomainEventInfra(eventInfra DomainEventInfra) {
	globalApp.InstallDomainEventInfra(eventInfra)
}

type logEvent struct {
	call func(producer, topic string, data []byte, header map[string]string)
}

// DomainEvent .
func (log *logEvent) DomainEvent(producer, topic string, data []byte, Worker Worker, header ...map[string]string) {
	h := map[string]string{}
	if len(header) > 0 {
		h = header[0]
	}

	if log.call != nil {
		log.call(producer, topic, data, h)
		return
	}
	Worker.Logger().Infof("Domain event:  %s   %s   %v", topic, string(data), header, h)
}

func (u *UnitTestImpl) NewDomainEventInfra(call ...func(producer, topic string, data []byte, header map[string]string)) DomainEventInfra {
	result := new(logEvent)
	if len(call) > 0 {
		result.call = call[0]
	}
	return result
}

func (u *UnitTestImpl) InjectBaseEntity(entity interface{}) {
	injectBaseEntity(u.rt, entity)
	return
}
