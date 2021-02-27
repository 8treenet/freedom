package entity

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/po"
)

const (
	//GoodsHotTag 热销
	GoodsHotTag = "HOT"
	//GoodsNewTag 新品
	GoodsNewTag = "NEW"
	//GoodsNoneTag 默认
	GoodsNoneTag = "NONE"
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

// MarkedTag 为商品打tag
func (g *Goods) MarkedTag(tag string) error {
	if tag != GoodsHotTag && tag != GoodsNewTag && tag != GoodsNoneTag {
		return errors.New("Tag doesn't exist")
	}
	g.SetTag(tag)
	return nil
}

// MarshalJSON 序列化json.
func (g *Goods) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Goods)
}
