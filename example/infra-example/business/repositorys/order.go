package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/models"
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

func (repo *OrderRepository) Get(id, userID int) (result models.Order, e error) {
	result, e = models.FindOrder(repo, models.Order{ID: id, UserID: userID})
	return
}

func (repo *OrderRepository) GetAll(userID int) (result []models.Order, e error) {
	limiter := repo.NewDescOrder("id").NewLimiter(50)
	result, e = models.FindOrders(repo, models.Order{UserID: userID}, limiter)
	return
}

func (repo *OrderRepository) Create(goodsID, num, userID int) error {
	_, e := models.CreateOrder(repo, &models.Order{
		UserID:  userID,
		GoodsID: goodsID,
		Num:     num,
		Created: time.Now(),
		Updated: time.Now(),
	})
	return e
}
