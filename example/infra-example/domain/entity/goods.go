package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/po"
)

// Goods 商品实体
type Goods struct {
	freedom.Entity
	po.Goods
}

// Identity 唯一
func (g *Goods) Identity() string {
	return strconv.Itoa(g.ID)
}

// Location .
func (g *Goods) Location() map[string]interface{} {
	return map[string]interface{}{"id": g.ID, "version": g.Version}
}
