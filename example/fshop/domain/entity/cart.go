package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/adapter/po"
)

// 购物车项实体
type Cart struct {
	freedom.Entity
	po.Cart
}

// Identity 唯一
func (c *Cart) Identity() string {
	return strconv.Itoa(c.Id)
}
