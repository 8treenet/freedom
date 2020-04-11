package controllers

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/application/objects"
	"github.com/8treenet/freedom/infra/kafka"
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
			绑定事件 ListenEvent(eventName string, fun interface{}, appointInfra ...interface{})
			eventName : 事件名称
			fun : 绑定方法
			appointInfra: 指定的基础设施组件, 默认全部基础设施组件可见

			//可以指定 kafka, rocket, 定时器, 等
			initiator.ListenEvent(EventSell, "StoreController.PostSellGoodsBy", &kafka.Consumer{}) //绑定kafka消息组件
		*/
	})
}

type StoreController struct {
	Runtime freedom.Runtime
}

// PostSellGoodsBy 事件方法为 Post开头, 参数是事件id
func (s *StoreController) PostSellGoodsBy(eventID string) error {
	//rawData, err := ioutil.ReadAll(s.Runtime.Ctx().Request().Body)
	var goods objects.Goods
	s.Runtime.Ctx().ReadJSON(&goods)

	action := s.Runtime.Ctx().GetHeader("x-action")
	retryCount := s.Runtime.Ctx().GetHeader(kafka.XRetryCount)
	s.Runtime.Logger().Infof("消耗商品ID:%d, %d件, 行为:%s, 消息key:%s, 重试次数:%v", goods.ID, goods.Amount, action, eventID, retryCount)
	return nil
}
