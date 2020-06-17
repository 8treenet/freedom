package repository

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/adapter/po"
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
}

func (repo *OrderRepository) Get(id, userID int) (result po.Order, e error) {
	result.Id = id
	result.UserId = userID
	e = findOrder(repo, &result)
	return
}

func (repo *OrderRepository) GetAll(userID int) (result []po.Order, e error) {
	e = findOrderList(repo, po.Order{UserId: userID}, &result)
	return
}

func (repo *OrderRepository) Create(goodsID, num, userID int) error {
	_, e := createOrder(repo, &po.Order{
		UserId:  userID,
		GoodsId: goodsID,
		Num:     num,
		Created: time.Now(),
		Updated: time.Now(),
	})
	return e
}
