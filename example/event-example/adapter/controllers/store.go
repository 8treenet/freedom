package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/domain/dto"
)

const (
	// EventSell .
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

// StoreController .
type StoreController struct {
	Worker freedom.Worker
}

// PostSellGoodsBy The event method starts with Post, and the parameter is the event id.
func (s *StoreController) PostSellGoodsBy(eventID string) error {
	//rawData, err := ioutil.ReadAll(s.Worker.Ctx().Request().Body)
	var goods dto.Goods
	s.Worker.IrisContext().ReadJSON(&goods)

	action := s.Worker.IrisContext().GetHeader("x-action")
	s.Worker.Logger().Infof("PostSellGoodsBy goods ID:%d, amount: %d, action: %s, eventId: %s", goods.ID, goods.Amount, action, eventID)
	s.Worker.Logger().Info("PostSellGoodsBy", freedom.LogFields{"goods": goods.ID, "amount": goods.Amount, "action": action, "eventId": eventID})
	return nil
}
