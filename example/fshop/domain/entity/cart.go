package entity

import (
	"strconv"

	"github.com/8treenet/freedom/example/fshop/domain/object"

	"github.com/8treenet/freedom"
)

// 购物车项实体
type Cart struct {
	freedom.Entity
	object.Cart
}

// Identity 唯一
func (c *Cart) Identity() string {
	return strconv.Itoa(c.Id)
}
