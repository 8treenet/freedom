package entity

import (
	"strconv"

	"github.com/8treenet/freedom/example/fshop/domain/object"

	"github.com/8treenet/freedom"
)

type Delivery struct {
	freedom.Entity
	object.Delivery
}

// Identity 唯一
func (d *Delivery) Identity() string {
	return strconv.Itoa(d.Id)
}
