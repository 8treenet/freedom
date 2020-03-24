package general

import (
	"net/http"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/context"
)

var _ UnitTest = new(UnitTestImpl)

type UnitTest interface {
	GetService(service interface{})
	GetRepository(repository interface{})
	InstallGorm(f func() (db *gorm.DB))
	InstallRedis(f func() (client redis.Cmdable))
	Run()
	SetRequest(request *http.Request)
	InjectBaseEntity(entity interface{})
	InstallDomainEventInfra(eventInfra DomainEventInfra)
	NewDomainEventInfra(call ...func(producer, topic string, data []byte, header map[string]string)) DomainEventInfra
}

type UnitTestImpl struct {
	rt      Runtime
	request *http.Request
}

func (u *UnitTestImpl) GetService(service interface{}) {
	globalApp.GetService(u.rt.Ctx(), service)
}

func (u *UnitTestImpl) GetRepository(repository interface{}) {
	value := reflect.ValueOf(repository).Elem()
	ok, newfield := globalApp.rpool.get(value.Type())
	if !ok {
		panic("not found")
	}
	if !value.CanSet() {
		globalApp.IrisApp.Logger().Fatal("The member variable must be publicly visible, Its type is " + value.Type().String())
	}

	br, ok := newfield.Interface().(BeginRequest)
	if ok {
		br.BeginRequest(u.rt)
	}
	value.Set(newfield)
}

func (u *UnitTestImpl) InstallGorm(f func() (db *gorm.DB)) {
	globalApp.InstallGorm(f)
}

func (u *UnitTestImpl) InstallRedis(f func() (client redis.Cmdable)) {
	globalApp.InstallRedis(f)
}

func (u *UnitTestImpl) Run() {
	for index := 0; index < len(boots); index++ {
		boots[index](globalApp)
	}
	u.rt = u.newRuntime()
	logLevel := "debug"
	globalApp.IrisApp.Logger().SetLevel(logLevel)
	globalApp.installDB()
	globalApp.comPool.singleBooting(globalApp)
}

func (u *UnitTestImpl) newRuntime() Runtime {
	ctx := context.NewContext(globalApp.IrisApp)
	if u.request == nil {
		u.request = new(http.Request)
	}
	ctx.BeginRequest(nil, u.request)
	rt := newRuntime(ctx)
	ctx.Values().Set(RuntimeKey, rt)
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
func (log *logEvent) DomainEvent(producer, topic string, data []byte, runtime Runtime, header ...map[string]string) {
	h := map[string]string{}
	if len(header) > 0 {
		h = header[0]
	}

	if log.call != nil {
		log.call(producer, topic, data, h)
		return
	}
	runtime.Logger().Infof("Domain event:  %s   %s   %v", topic, string(data), header, h)
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
