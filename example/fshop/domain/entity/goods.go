package entity

import (
	"errors"
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/adapter/po"
)

const (
	//热销
	GoodsHotTag = "HOT"
	//新品
	GoodsNewTag  = "NEW"
	GoodsNoneTag = "NONE"
)

// 商品实体
type Goods struct {
	freedom.Entity
	po.Goods
}

// Identity 唯一
func (g *Goods) Identity() string {
	return strconv.Itoa(g.Id)
}

// MarkedTag 为商品打tag
func (g *Goods) MarkedTag(tag string) error {
	if tag != GoodsHotTag && tag != GoodsNewTag && tag != GoodsNoneTag {
		return errors.New("Tag doesn't exist")
	}
	g.SetTag(tag)
	return nil
}
