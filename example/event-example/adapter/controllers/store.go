package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/domain/dto"
)

const (
	EventSell = "event-sell"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		store := &StoreController{}
		initiator.BindController("/store", store)

		initiator.ListenEvent(EventSell, "StoreController.PostSellGoodsBy")
		/*
			Binding Event, ListenEvent(eventName string, fun interface{}, appointInfra ...interface{})
			eventName : Event Name
			fun : Binding Function
			appointInfra: Specified infrastructure components, all infrastructure components are visible by default

			// Can specify kafka, rocket, timer, etc.
			initiator.ListenEvent(EventSell, "StoreController.PostSellGoodsBy", &kafka.Consumer{}) // Binding Kafka message components
		*/
	})
}

type StoreController struct {
	Worker freedom.Worker
}

// PostSellGoodsBy. The event method starts with Post, and the parameter is the event id.
func (s *StoreController) PostSellGoodsBy(eventID string) error {
	//rawData, err := ioutil.ReadAll(s.Worker.Ctx().Request().Body)
	var goods dto.Goods
	s.Worker.IrisContext().ReadJSON(&goods)

	action := s.Worker.IrisContext().GetHeader("x-action")
	s.Worker.Logger().Infof("Consumed goods ID:%d, Amount: %d, Action: %s, Event Key: %s", goods.ID, goods.Amount, action, eventID)
	return nil
}
