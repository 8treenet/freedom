package models

import (
	"github.com/8treenet/freedom"
	"time"
)

// Order .
type Order struct {
	ID      int       `gorm:"primary_key" column:"id"`
	UserID  int       `gorm:"column:user_id"`  // 用户id
	GoodsID int       `gorm:"column:goods_id"` // 商品id
	Num     int       `gorm:"column:num"`      // 数量
	Created time.Time `gorm:"column:created"`
	Updated time.Time `gorm:"column:updated"`
}

func (m *Order) TableName() string {
	return "order"
}

// FindOrderByPrimary .
func FindOrderByPrimary(rep freedom.GORMRepository, primary interface{}) (result Order, e error) {
	e = rep.DB().Find(&result, primary).Error
	return
}

// FindOrdersByPrimarys .
func FindOrdersByPrimarys(rep freedom.GORMRepository, primarys ...interface{}) (results []Order, e error) {
	e = rep.DB().Find(&results, primarys).Error
	return
}

// FindOrder .
func FindOrder(rep freedom.GORMRepository, query Order, builders ...freedom.QueryBuilder) (result Order, e error) {
	db := rep.DB().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = db.Limit(1).Order(builders[0].Order()).Find(&result).Error
	return
}

// FindOrderByWhere .
func FindOrderByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (result Order, e error) {
	db := rep.DB()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = db.Limit(1).Order(builders[0].Order()).Find(&result).Error
	return
}

// FindOrders .
func FindOrders(rep freedom.GORMRepository, query Order, builders ...freedom.QueryBuilder) (results []Order, e error) {
	db := rep.DB().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// FindOrdersByWhere .
func FindOrdersByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (results []Order, e error) {
	db := rep.DB()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// CreateOrder .
func CreateOrder(rep freedom.GORMRepository, entity *Order) (rowsAffected int64, e error) {
	db := rep.DB().Create(entity)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// UpdateOrder .
func UpdateOrder(rep freedom.GORMRepository, query *Order, value Order) (affected int64, e error) {
	db := rep.DB().Model(query).Updates(value)
	e = db.Error
	affected = db.RowsAffected
	return
}

// FindToUpdateOrders .
func FindToUpdateOrders(rep freedom.GORMRepository, query Order, value Order, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Order{}).Where(query).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&Order{}).Where(query).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}

// FindByWhereToUpdateOrders .
func FindByWhereToUpdateOrders(rep freedom.GORMRepository, query string, args []interface{}, value Order, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Order{}).Where(query, args...).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&Order{}).Where(query, args...).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}
