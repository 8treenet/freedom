package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/po"
)

// Order 订单实体
type Order struct {
	freedom.Entity
	po.Order
}

// Identity 唯一
func (o *Order) Identity() string {
	return strconv.Itoa(o.ID)
}
