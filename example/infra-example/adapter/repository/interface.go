package repository

import (
	"github.com/8treenet/freedom/example/infra-example/domain/object"
)

type GoodsInterface interface {
	Get(id int) (object.Goods, error)
	GetAll() ([]object.Goods, error)
	Save(*object.Goods) error
}

type OrderInterface interface {
	Get(id int, userID int) (object.Order, error)
	GetAll(userID int) ([]object.Order, error)
	Create(goodsID, num, userID int) error
}
