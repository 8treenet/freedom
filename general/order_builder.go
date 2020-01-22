package general

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

// Reorder .
type Reorder struct {
	fields []string
	orders []string
}

// NewPager .
func (o *Reorder) NewPager(page, pageSize int) *Pager {
	pager := new(Pager)
	pager.reorder = *o
	pager.page = page
	pager.pageSize = pageSize
	return pager
}

// Order .
func (o *Reorder) Order() interface{} {
	args := []string{}
	for index := 0; index < len(o.fields); index++ {
		args = append(args, fmt.Sprintf("`%s` %s", o.fields[index], o.orders[index]))
	}

	return strings.Join(args, ",")
}

// Pager 分页器
type Pager struct {
	reorder   Reorder
	pageSize  int
	page      int
	totalPage int
}

// TotalPage .
func (p *Pager) TotalPage() int {
	return p.totalPage
}

// Execute .
func (p *Pager) Execute(db *gorm.DB, object interface{}) (e error) {
	orderBy := p.reorder.Order()
	resultDB := db.Order(orderBy).Offset((p.page - 1) * p.pageSize).Limit(p.pageSize).Find(object)
	if resultDB.Error != nil {
		return resultDB.Error
	}
	var count int
	e = db.Model(object).Count(&count).Error
	if e == nil && count != 0 {
		//计算分页
		if count%p.pageSize == 0 {
			p.totalPage = count / p.pageSize
		} else {
			p.totalPage = count/p.pageSize + 1
		}
	}
	return
}

// Order .
func (p *Pager) Order() interface{} {
	return p.reorder.Order()
}