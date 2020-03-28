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

// NewPageBuilder .
func (o *Reorder) NewPageBuilder(page, pageSize int) *Builder {
	pager := new(Builder)
	pager.reorder = o
	pager.page = page
	pager.pageSize = pageSize
	return pager
}

// NewBuilder .
func (o *Reorder) NewBuilder() *Builder {
	pager := new(Builder)
	pager.reorder = o
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

// Builder
type Builder struct {
	reorder       *Reorder
	pageSize      int
	page          int
	totalPage     int
	selectColumn  []string
	selectPrimary bool
}

// TotalPage .
func (p *Builder) TotalPage() int {
	return p.totalPage
}

func (b *Builder) Order() interface{} {
	if b.reorder != nil {
		return b.Order()
	}
	return ""
}

// Execute .
func (p *Builder) Execute(db *gorm.DB, object interface{}) (e error) {
	pageFind := false
	if p.reorder != nil {
		db = db.Order(p.reorder.Order())
	} else {
		db = db.Set("gorm:order_by_primary_key", "DESC")
	}
	if p.page != 0 && p.pageSize != 0 {
		pageFind = true
		db = db.Offset((p.page - 1) * p.pageSize).Limit(p.pageSize)
	}

	if len(p.selectColumn) > 0 && !p.selectPrimary {
		db = db.Select(p.selectColumn)
	}
	if p.selectPrimary {
		db = db.Select(getEntityPrimary(object))
	}

	resultDB := db.Find(object)
	if resultDB.Error != nil {
		return resultDB.Error
	}

	if !pageFind {
		return
	}

	var count int
	e = resultDB.Offset(0).Limit(1).Count(&count).Error
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

// SetPage .
func (b *Builder) SetPage(page, pageSize int) *Builder {
	b.page = page
	b.pageSize = pageSize
	return b
}

// SelectColumn .
func (b *Builder) SelectColumn(column ...string) *Builder {
	b.selectColumn = append(b.selectColumn, column...)
	return b
}

// SelectPrimary .
func (b *Builder) SelectPrimary() *Builder {
	b.selectPrimary = true
	return b
}
