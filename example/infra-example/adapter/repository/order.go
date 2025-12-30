package repository

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/entity"
	"github.com/8treenet/freedom/example/infra-example/domain/event"
	"github.com/8treenet/freedom/example/infra-example/domain/po"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
	"github.com/8treenet/freedom/infra/kafka"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *OrderRepository {
			return &OrderRepository{}
		})
	})
}

// OrderRepository .
type OrderRepository struct {
	freedom.Repository
	EventManager *domainevent.EventManager
	Producer     kafka.Producer
}

// Get .
func (repo *OrderRepository) Get(ID, userID int) (result *entity.Order, e error) {
	result = &entity.Order{}
	result.ID = ID
	result.UserID = userID

	//注入基础Entity
	repo.InjectBaseEntity(result)

	e = findOrder(repo, &result.Order)
	return
}

// GetAll .
func (repo *OrderRepository) GetAll(userID int) (result []*entity.Order, e error) {
	list, e := findOrderList(repo, po.Order{UserID: userID})
	if e != nil {
		return
	}
	for _, v := range list {
		result = append(result, &entity.Order{Order: *v})
	}

	//注入基础Entity
	repo.InjectBaseEntitys(result)
	return
}

// Create .
func (repo *OrderRepository) Create(goodsID, num, userID int) error {
	_, e := createOrder(repo, &po.Order{
		UserID:  userID,
		GoodsID: goodsID,
		Num:     num,
		Status:  "NON_PAYMENT",
		Created: time.Now(),
		Updated: time.Now(),
	})
	return e
}

// Pay 订单支付.
func (repo *OrderRepository) Pay(orderID, userID int) error {
	orderPO, err := OrderToPoint(findOrderByWhere(repo, "id = ? and user_id = ?", []interface{}{orderID, userID}))
	if err != nil {
		return err
	}
	if orderPO.Status == "PAID" {
		return errors.New("订单已支付")
	}
	orderPO.SetStatus("PAID")
	if _, err := saveOrder(repo, orderPO); err != nil {
		return err
	}

	// 发送支付消息到 Kafka (普通 Kafka 消息，非领域事件)
	payEvent := event.OrderPay{
		OrderID: orderID,
		UserID:  userID,
	}
	data, _ := json.Marshal(payEvent)
	traceHead := repo.Worker().Bus().Header.Clone()
	return repo.Producer.NewMsg("OrderPay", data).SetHeader(map[string]interface{}{"x-request-id": traceHead.Get("x-request-id")}).Publish()
}

func (repo *OrderRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
