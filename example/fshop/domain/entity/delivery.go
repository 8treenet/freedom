package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/adapter/po"
)

type Delivery struct {
	freedom.Entity
	po.Delivery
}

// Identity 唯一
func (d *Delivery) Identity() string {
	return strconv.Itoa(d.Id)
}
