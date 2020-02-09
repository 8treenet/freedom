package general

import (
	"net/http"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/context"
)

type UnitTest struct {
	rt      Runtime
	request *http.Request
}

func (u *UnitTest) GetService(service interface{}) *UnitTest {
	globalApp.GetService(u.rt.Ctx(), service)
	return u
}

func (u *UnitTest) GetRepository(repository interface{}) *UnitTest {
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
	return u
}

func (u *UnitTest) InstallGorm(f func() (db *gorm.DB)) *UnitTest {
	globalApp.InstallGorm(f)
	return u
}

func (u *UnitTest) InstallRedis(f func() (client redis.Cmdable)) *UnitTest {
	globalApp.InstallRedis(f)
	return u
}

func (u *UnitTest) Run() {
	for index := 0; index < len(boots); index++ {
		boots[index](globalApp)
	}
	u.rt = u.newRuntime()
	logLevel := "debug"
	globalApp.IrisApp.Logger().SetLevel(logLevel)
	globalApp.installDB()
	globalApp.comPool.singleBooting(globalApp)
}

func (u *UnitTest) newRuntime() Runtime {
	ctx := context.NewContext(globalApp.IrisApp)
	if u.request == nil {
		u.request = new(http.Request)
	}
	ctx.BeginRequest(nil, u.request)
	rt := newRuntime(ctx)
	ctx.Values().Set(runtimeKey, rt)
	return rt
}

func (u *UnitTest) SetRequest(request *http.Request) *UnitTest {
	u.request = request
	return u
}

func (u *UnitTest) InstallDomainEventInfra(eventInfra DomainEventInfra) *UnitTest {
	globalApp.InstallDomainEventInfra(eventInfra)
	return u
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

func (u *UnitTest) NewDomainEventInfra(call ...func(producer, topic string, data []byte, header map[string]string)) DomainEventInfra {
	result := new(logEvent)
	if len(call) > 0 {
		result.call = call[0]
	}
	return result
}

func (u *UnitTest) MadeEntity(entity interface{}) {
	newEntity(u.rt, entity)
}
