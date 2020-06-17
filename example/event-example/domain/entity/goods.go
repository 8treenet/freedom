package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/adapter/dto"
)

type Goods struct {
	freedom.Entity
	goodsObj dto.Goods
}

/*
	DomainEvent(fun interface{}, object interface{}, header ...map[string]string)
	fun : 为触发事件的方法, `实体名字:方法名`
	object : 结构体数据,会做json转换
	header : k/v 附加数据
*/
func (g *Goods) Shopping() {
	/*
		相关购买逻辑。。。
	*/

	//触发领域事件 `Goods:Shopping`
	g.DomainEvent("Goods:Shopping", g.goodsObj)
}

func (g *Goods) Identity() string {
	return strconv.Itoa(g.goodsObj.ID)
}
