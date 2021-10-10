package consumer

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain"
	"github.com/8treenet/freedom/example/fshop/domain/event"
	"github.com/8treenet/freedom/example/fshop/infra/domainevent"
	"github.com/8treenet/freedom/middleware"
)

func init() {
	//注册服务定位器的回调函数, freedom.ServiceLocator().Call 之前触发
	freedom.ServiceLocator().InstallBeginCallBack(func(worker freedom.Worker) {
		traceID, _ := middleware.GenerateTraceID()
		logger := middleware.NewLogger("x-request-id", traceID)
		worker.Bus().Set("x-request-id", traceID)
		worker.SetLogger(logger)
	})

	//注册启动回调
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindBooting(warmUp)
	})
}

// Start .
func Start() {
	fixedTime()
	eventLoop()
}

// fixedTime .
func fixedTime() {
	go func() {
		time.Sleep(10 * time.Second) //延迟，等待程序Application.Run
		t := time.NewTimer(10 * time.Second)
		for range t.C {
			freedom.ServiceLocator().Call(func(goodsService *domain.Goods) error {
				return goodsService.MarkedTag(4, "HOT")
			})
			t.Reset(10 * time.Second)
		}
	}()
}

// eventLoop .
func eventLoop() {
	eventRepository := domainevent.GetEventManager()
	go func() {
		for pubEvent := range eventRepository.GetPublisherChan() {
			time.Sleep(1 * time.Second)
			freedom.Logger().Infof("领域事件消费 Topic:%s, %+v", pubEvent.Topic(), pubEvent)
			shopEvent, ok := pubEvent.(*event.ShopGoods)
			if ok {
				if err := eventRepository.InsertSubEvent(shopEvent); err != nil {
					freedom.Logger().Error("领域事件消费错误", err)
					continue
				}
				freedom.ServiceLocator().Call(func(goodsService *domain.Goods) error {
					//购买事件
					return goodsService.ShopEvent(shopEvent)
				})
			}
		}
	}()
}

// warmUp 缓存预热
func warmUp(bootManager freedom.BootManager) {
	freedom.ServiceLocator().Call(func(goodsService *domain.Goods) error {
		for i := 1; i < 5; i++ {
			goodsService.Get(i)
		}
		return nil
	})
}
