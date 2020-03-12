package repositorys

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/application/object"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *OrderRepository {
			return &OrderRepository{}
		})
	})
}

// OrderRepository .
type OrderRepository struct {
	freedom.Repository
}

func (repo *OrderRepository) Get(id, userID int) (result object.Order, e error) {
	e = findOrder(repo, object.Order{Id: id, UserId: userID}, &result)
	return
}

func (repo *OrderRepository) GetAll(userID int) (result []object.Order, e error) {
	e = findOrders(repo, object.Order{UserId: userID}, &result)
	return
}

func (repo *OrderRepository) Create(goodsID, num, userID int) error {
	_, e := createOrder(repo, &object.Order{
		UserId:  userID,
		GoodsId: goodsID,
		Num:     num,
		Created: time.Now(),
		Updated: time.Now(),
	})
	return e
}
