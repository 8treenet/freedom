package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/event-example/domain/dto"
)

// Goods .
type Goods struct {
	freedom.Entity
	goodsObj dto.Goods
}

//Shopping .
func (g *Goods) Shopping() {
	/*
		DomainEvent(fun interface{}, object interface{}, header ...map[string]string)
		fun : Function for triggering the event, `k:v = EntityName:FuncName`
		object : Structure data, Could do json conversion
		header : k/v, Additional data
		Related shoping logic. . .
	*/

	// Trigger domain event, `Goods:Shopping`
	g.DomainEvent("Goods:Shopping", g.goodsObj)
}

// Identity .
func (g *Goods) Identity() string {
	return strconv.Itoa(g.goodsObj.ID)
}
