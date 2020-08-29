package repository

import (
	"github.com/8treenet/freedom/example/infra-example/domain/po"
)

// GoodsInterface .
type GoodsInterface interface {
	Get(id int) (po.Goods, error)
	GetAll() ([]po.Goods, error)
	Save(*po.Goods) error
}

// OrderInterface .
type OrderInterface interface {
	Get(id int, userID int) (po.Order, error)
	GetAll(userID int) ([]po.Order, error)
	Create(goodsID, num, userID int) error
}
