package repositorys

import (
	"github.com/8treenet/freedom/example/infra-example/models"
)

type GoodsInterface interface {
	Get(id int) (models.Goods, error)
	GetAll() ([]models.Goods, error)
	ChangeStock(*models.Goods, int) error
}

type OrderInterface interface {
	Get(id int, userID int) (models.Order, error)
	GetAll(userID int) ([]models.Order, error)
	Create(goodsID, num, userID int) error
}
