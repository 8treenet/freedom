package domainevent

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/infra/kafka"
	"github.com/jinzhu/gorm"
)

var eventManager *EventManager

func init() {
	eventManager = &EventManager{eventRetry: newEventRetry(), publisherChan: make(chan freedom.DomainEvent)}
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, eventManager) //单例绑定
		initiator.InjectController(func(ctx freedom.Context) (com *EventManager) {
			initiator.FetchInfra(ctx, &com)
			return
		})
	})
}

// GetEventManager .
func GetEventManager() *EventManager {
	return eventManager
}

// EventManager .
type EventManager struct {
	freedom.Infra
	eventRetry    *eventRetry         //重试处理
	kafkaProducer *kafka.ProducerImpl //Kafka Producer组件
	publisherChan chan freedom.DomainEvent
}

// Booting .
func (manager *EventManager) Booting(bootManager freedom.BootManager) {
	if err := manager.db().AutoMigrate(&pubEventObject{}).Error; err != nil {
		panic(err)
	}
	if err := manager.db().AutoMigrate(&subEventObject{}).Error; err != nil {
		panic(err)
	}

	//获取Kafka Producer组件, 只支持单例组件的获取.
	if !manager.FetchSingleInfra(&manager.kafkaProducer) {
		panic("No KafkaProducer found")
	}

	//重试参考 example/infra-example
}

// GetPublisherChan .
func (manager *EventManager) GetPublisherChan() <-chan freedom.DomainEvent {
	return manager.publisherChan
}

// SetRetryPolicy Set the rules for retry.
func (manager *EventManager) SetRetryPolicy(delay time.Duration, retries int) {
	manager.eventRetry.setRetryPolicy(delay, retries)
}

// push EventTransaction在事务成功后触发 .
func (manager *EventManager) push(event freedom.DomainEvent) {
	freedom.Logger().Debugf("Publish Event Topic:%s, %+v", event.Topic(), event)
	identity := event.Identity()
	go func() {
		defer func() {
			if perr := recover(); perr != nil {
				freedom.Logger().Error("Publish Event:", perr)
			}
		}()

		/*
			Kafka 消息模式参考 example/infra-example
			if err := manager.kafkaProducer.NewMsg(event.Topic(), event.Marshal()).SetHeader(event.GetPrototypes()).Publish(); err != nil {
				freedom.Logger().Error(err)
				return
			}

			REST 模式参考 example/http2
			manager.NewHTTPRequest(URL)
			manager.NewH2CRequest(URL)
		*/
		manager.publisherChan <- event

		if !manager.eventRetry.pubExist(event.Topic()) {
			return //未注册重试,结束
		}

		// Push成功后删除
		if err := manager.db().Delete(&pubEventObject{}, "identity = ?", identity).Error; err != nil {
			freedom.Logger().Error(err)
		}
	}()
}

func (manager *EventManager) db() *gorm.DB {
	var db *gorm.DB
	if err := manager.FetchOnlyDB(&db); err != nil {
		panic(err)
	}
	return db
}

// Save .
func (manager *EventManager) Save(repo *freedom.Repository, entity freedom.Entity) (e error) {
	txDB := getTxDB(repo) //获取事务db

	//删除实体里的全部事件
	defer entity.RemoveAllPubEvent()
	defer entity.RemoveAllSubEvent()

	//Insert PubEvent
	for _, domainEvent := range entity.GetPubEvent() {
		if !manager.eventRetry.pubExist(domainEvent.Topic()) {
			continue //未注册重试,无需存储
		}

		content, err := domainEvent.Marshal()
		if err != nil {
			return err
		}

		model := pubEventObject{
			Identity: domainEvent.Identity(),
			Topic:    domainEvent.Topic(),
			Content:  string(content),
			Created:  time.Now(),
			Updated:  time.Now(),
		}
		if e = txDB.Create(&model).Error; e != nil { //插入发布事件表。
			return
		}
	}
	manager.addPubToWorker(repo.Worker(), entity.GetPubEvent())

	//Delete SubEvent
	for _, subEvent := range entity.GetSubEvent() {
		if !manager.eventRetry.subExist(subEvent.Topic()) {
			continue //未注册重试,无需修改
		}

		eventID := subEvent.Identity()
		if e = getTxDB(repo).Delete(&subEventObject{}, "identity = ?", eventID).Error; e != nil { //删除消费事件表
			return
		}
	}
	return
}

// InsertSubEvent .
func (manager *EventManager) InsertSubEvent(event freedom.DomainEvent) error {
	if !manager.eventRetry.subExist(event.Topic()) {
		return nil //未注册重试,无需存储
	}

	content, err := event.Marshal()
	if err != nil {
		return err
	}

	model := subEventObject{
		Identity: event.Identity(),
		Topic:    event.Topic(),
		Content:  string(content),
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	return manager.db().Create(&model).Error //插入消费事件表。
}

// Retry .
func (manager *EventManager) Retry() {
	manager.eventRetry.retry()
}

// BindRetryPubEvent .
func (manager *EventManager) BindRetryPubEvent(event freedom.DomainEvent) {
	manager.eventRetry.bindRetryPubEvent(event)
}

// BindRetrySubEvent .
func (manager *EventManager) BindRetrySubEvent(event freedom.DomainEvent, function interface{}) {
	manager.eventRetry.bindRetrySubEvent(event, function)
}

// addPubToWorker 增加发布事件到Worker.Store.
func (manager *EventManager) addPubToWorker(worker freedom.Worker, pubs []freedom.DomainEvent) {
	if len(pubs) == 0 {
		return
	}
	for _, pubEvent := range pubs {
		//把worker的请求信息 写到事件的属性里
		m := map[string]interface{}{}
		for key, item := range worker.Bus().Header.Clone() {
			if len(item) <= 0 {
				continue
			}
			m[key] = item[0]
		}

		pubEvent.SetPrototypes(m)
	}

	//把发布事件加到store内, EventTransaction在事务结束后会触发push
	var storePubEvents []freedom.DomainEvent
	store := worker.Store().Get(workerStorePubEventKey)
	if store != nil {
		list, ok := store.([]freedom.DomainEvent)
		if ok {
			storePubEvents = list
		}
	}
	storePubEvents = append(storePubEvents, pubs...)
	worker.Store().Set(workerStorePubEventKey, storePubEvents)
}

// getTxDB 获取请求内事务的DB
func getTxDB(repo *freedom.Repository) *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	db = db.New()
	db.SetLogger(repo.Worker().Logger())
	return db
}
