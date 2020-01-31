package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/application/objects"
	"time"
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

func (repo *OrderRepository) Get(id, userID int) (result objects.Order, e error) {
	result, e = findOrder(repo, objects.Order{ID: id, UserID: userID})
	return
}

func (repo *OrderRepository) GetAll(userID int) (result []objects.Order, e error) {
	result, e = findOrders(repo, objects.Order{UserID: userID})
	return
}

func (repo *OrderRepository) Create(goodsID, num, userID int) error {
	_, e := createOrder(repo, &objects.Order{
		UserID:  userID,
		GoodsID: goodsID,
		Num:     num,
		Created: time.Now(),
		Updated: time.Now(),
	})
	return e
}
