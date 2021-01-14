package domainevent

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/8treenet/freedom"
)

func newEventRetry() *eventRetry {
	return &eventRetry{
		pubPool:  make(map[string]reflect.Type),
		subPool:  make(map[string]subRetry),
		delay:    60 * time.Second, //新插入数据，延迟60秒被扫描
		retries:  3,                //重试次数
		interval: time.Second * 30, //定时间隔
	}
}

type eventRetry struct {
	pubPool  map[string]reflect.Type
	subPool  map[string]subRetry
	delay    time.Duration
	retries  int
	interval time.Duration
}

type subRetry struct {
	event    reflect.Type
	function interface{}
}

func (retry *eventRetry) RetryPubEvent(event freedom.DomainEvent) {
	topic := event.Topic()
	if topic == "" {
		panic("Topic Cannot be empty")
	}
	if _, ok := retry.pubPool[topic]; ok {
		panic(fmt.Sprintf("Topic:%s already exists", topic))
	}
	retry.pubPool[topic] = reflect.TypeOf(event)
}

func (retry *eventRetry) RetrySubEvent(event freedom.DomainEvent, fun interface{}) {
	topic := event.Topic()
	if topic == "" {
		panic("Topic Cannot be empty")
	}

	if _, ok := retry.subPool[topic]; ok {
		panic(fmt.Sprintf("Topic:%s already exists", topic))
	}
	eventType := reflect.TypeOf(event)
	eventInType, err := parseRetryCallFunc(fun)
	if err != nil {
		panic(err)
	}
	if eventType != eventInType {
		panic("The function's argument is wrong")
	}

	retry.subPool[topic] = subRetry{
		event:    eventType,
		function: fun,
	}
}

func (retry *eventRetry) PubExist(topic string) bool {
	_, ok := retry.pubPool[topic]
	return ok
}

func (retry *eventRetry) SubExist(topic string) bool {
	_, ok := retry.subPool[topic]
	return ok
}

func (retry *eventRetry) SetRetryPolicy(delay, interval time.Duration, retries int) {
	retry.delay = delay
	retry.retries = retries
	retry.interval = interval
}

func (retry *eventRetry) Booting() {
	go func() {
		time.Sleep(5 * time.Second)
		timer := time.NewTimer(retry.interval)
		for range timer.C {
			retry.scanSub()
			retry.scanPub()
			timer.Reset(retry.interval)
		}
	}()
}

func (retry *eventRetry) scanSub() {
	rows := 200
	var list []*subEventPO
	err := GetEventManager().db().Where("retries < ? and created < ?", retry.retries, time.Now().Add(-retry.delay)).Order("sequence ASC").Limit(rows).Find(&list).Error
	if err != nil {
		freedom.Logger().Info("RetrySubEvent:", err)
		return
	}

	var filterList []*subEventPO
	for _, po := range list {
		if !retry.SubExist(po.Topic) {
			GetEventManager().db().Delete(&subEventPO{}, "identity = ?", po.Identity) //未注册重试，直接删除
			continue
		}

		po.AddRetries(1)
		err := GetEventManager().db().Table(po.TableName()).Where(po.Location()).Updates(po.GetChanges()).Error //修改重试次数
		if err != nil {
			freedom.Logger().Error("RetrySubEvent:", err)
			continue
		}
		filterList = append(filterList, po)
	}

	for _, po := range filterList {
		retry.callSub(po)
	}
}

func (retry *eventRetry) callSub(po *subEventPO) {
	defer func() {
		if perr := recover(); perr != nil {
			freedom.Logger().Error("RetrySubEvent:", perr)
		}
	}()

	subRetryObj := retry.subPool[po.Topic]
	newEvent := reflect.New(subRetryObj.event.Elem()).Interface()

	if err := newEvent.(freedom.DomainEvent).Unmarshal([]byte(po.Content)); err != nil {
		freedom.Logger().Error("RetrySubEvent:", err)
		return
	}

	reflect.ValueOf(subRetryObj.function).Call([]reflect.Value{reflect.ValueOf(newEvent)})
}

func (retry *eventRetry) scanPub() {
	rows := 200
	var list []*pubEventPO
	err := GetEventManager().db().Where("retries < ? and created < ?", retry.retries, time.Now().Add(-retry.delay)).Order("sequence ASC").Limit(rows).Find(&list).Error
	if err != nil {
		freedom.Logger().Info("RetryPubEvent:", err)
		return
	}

	var filterList []*pubEventPO
	for _, po := range list {
		if !retry.PubExist(po.Topic) {
			GetEventManager().db().Delete(&pubEventPO{}, "identity = ?", po.Identity) //未注册重试，直接删除
			continue
		}

		po.AddRetries(1)
		err := GetEventManager().db().Table(po.TableName()).Where(po.Location()).Updates(po.GetChanges()).Error //修改重试次数
		if err != nil {
			freedom.Logger().Error("RetryPubEvent:", err)
			continue
		}
		filterList = append(filterList, po)
	}

	for _, po := range filterList {
		retry.callPub(po)
	}
}

func (retry *eventRetry) callPub(po *pubEventPO) {
	defer func() {
		if perr := recover(); perr != nil {
			freedom.Logger().Error("RetrySubEvent:", perr)
		}
	}()

	pubRetryObj := retry.pubPool[po.Topic]
	newEvent := reflect.New(pubRetryObj.Elem()).Interface()
	GetEventManager().push(newEvent.(freedom.DomainEvent))
}

func parseRetryCallFunc(f interface{}) (inType reflect.Type, e error) {
	ftype := reflect.TypeOf(f)
	if ftype.Kind() != reflect.Func {
		e = errors.New("It's not a func")
		return
	}
	if ftype.NumIn() != 1 {
		e = errors.New("The function's argument is wrong")
		return
	}
	if ftype.NumOut() != 0 {
		e = errors.New("The function's argument is wrong")
		return
	}
	inType = ftype.In(0)
	if inType.Kind() != reflect.Ptr {
		e = errors.New("The function's argument is wrong")
		return
	}
	return
}
