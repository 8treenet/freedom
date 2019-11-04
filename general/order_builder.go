package general

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// Order .
type Order struct {
	orderFields []string
	order       string
	pager       *struct {
		pageSize  int
		page      int
		TotalPage int
	}
	limit *int
}

// SetPager .
func (o *Order) SetPager(page, pageSize int) *Order {
	defer func() {
		o.limit = nil
	}()

	o.pager.pageSize = pageSize
	o.pager.page = page
	return o
}

// SetLimiter .
func (o *Order) SetLimiter(limit int) *Order {
	defer func() {
		o.pager = nil
	}()

	*o.limit = limit
	return o
}

// Order .
func (o *Order) Order() string {
	return strings.Join(o.orderFields, ",") + " " + o.order
}

// Limit .
func (o *Order) Limit() int {
	return *o.limit
}

// Execute .
func (o *Order) Execute(db *gorm.DB, object interface{}) (next *gorm.DB, e error) {
	orderBy := o.Order()
	if o.pager != nil {
		resultDB := db.Order(orderBy).Offset((o.pager.page - 1) * o.pager.pageSize).Find(object)
		if resultDB.Error != nil {
			return nil, resultDB.Error
		}
		var count int
		e := resultDB.Count(&count).Error
		if e == nil && count != 0 {
			//计算分页
			if count%o.pager.pageSize == 0 {
				o.pager.TotalPage = count / o.pager.pageSize
			} else {
				o.pager.TotalPage = count/o.pager.pageSize + 1
			}

		}
		return nil, nil
	}

	if o.limit != nil {
		e = db.Order(orderBy).Limit(*o.limit).Find(object).Error
		return nil, e
	}

	return nil, db.Order(orderBy).Find(&object).Error
}
