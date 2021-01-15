package domainevent

import (
	"github.com/8treenet/freedom"
	"github.com/jinzhu/gorm"
)

var eventManager *EventManager

func init() {
	eventManager = &EventManager{publisherChan: make(chan freedom.DomainEvent, 1000)}
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, eventManager) //单例绑定
		initiator.InjectController(func(ctx freedom.Context) (com *EventManager) {
			initiator.GetInfra(ctx, &com)
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
	publisherChan chan freedom.DomainEvent
}

// GetPublisherChan .
func (manager *EventManager) GetPublisherChan() <-chan freedom.DomainEvent {
	return manager.publisherChan
}

// Booting .
func (manager *EventManager) Booting(sb freedom.SingleBoot) {
	if err := manager.db().AutoMigrate(&domainEventPublish{}).Error; err != nil {
		panic(err)
	}
	if err := manager.db().AutoMigrate(&domainEventSubscribe{}).Error; err != nil {
		panic(err)
	}
}

// push .
func (manager *EventManager) push(event freedom.DomainEvent) {
	freedom.Logger().Infof("领域事件发布 Topic:%s, %+v", event.Topic(), event)
	//eventID := event.Identity()
	go func() {
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
		//publish := &domainEventPublish{ID: eventID}

		// Push成功后删除事件
		// if err := manager.db().Delete(&publish).Error; err != nil {
		// 	freedom.Logger().Error(err)
		// }
	}()
}

func (manager *EventManager) db() *gorm.DB {
	return manager.SourceDB().(*gorm.DB)
}

// Save .
func (manager *EventManager) Save(repo *freedom.Repository, entity freedom.Entity) (e error) {

	return
}

// InsertSubEvent .
func (manager *EventManager) InsertSubEvent(event freedom.DomainEvent) error {
	return nil
}

// Retry .
func (manager *EventManager) Retry() {
	//定时器扫描表中失败的Pub/Sub事件
	freedom.Logger().Info("EventManager Retry")
}

// addPubToWorker 增加发布事件到worker的Store.
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
