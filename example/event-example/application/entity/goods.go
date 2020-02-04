package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/application/objects"
)

type Goods struct {
	freedom.Entity
	goodsObj objects.Goods
}

/*
	DomainEvent(fun interface{}, object interface{}, header ...map[string]string)
	fun : 为触发事件的方法。 基于强类型的原则，框架已经做了方法和字符串Topic的映射,`实体名字:方法名`
	object : 结构体数据,会做json转换
	header : k/v 附加数据
*/
func (g *Goods) Shopping() {
	/*
		相关购买逻辑。。。
	*/

	//触发领域事件 `Goods:Shopping`
	g.DomainEvent(g.Shopping, g.goodsObj)
}

func (g *Goods) Identity() string {
	return strconv.Itoa(g.goodsObj.ID)
}
