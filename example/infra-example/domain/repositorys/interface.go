package repositorys

import (
	"github.com/8treenet/freedom/example/infra-example/objects"
)

type GoodsInterface interface {
	Get(id int) (objects.Goods, error)
	GetAll() ([]objects.Goods, error)
	ChangeStock(*objects.Goods, int) error
}

type OrderInterface interface {
	Get(id int, userID int) (objects.Order, error)
	GetAll(userID int) ([]objects.Order, error)
	Create(goodsID, num, userID int) error
}
