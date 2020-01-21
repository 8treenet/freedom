package objects

import (
	"github.com/8treenet/freedom"
	"time"
)

// Order .
type Order struct {
	changes map[string]interface{}
	ID      int       `gorm:"primary_key" column:"id"`
	UserID  int       `gorm:"column:user_id"`  // 用户id
	GoodsID int       `gorm:"column:goods_id"` // 商品id
	Num     int       `gorm:"column:num"`      // 数量
	Created time.Time `gorm:"column:created"`
	Updated time.Time `gorm:"column:updated"`
}

func (obj *Order) TableName() string {
	return "order"
}

// Updates .
func (obj *Order) Updates(rep freedom.GORMRepository) (affected int64, e error) {
	if obj.changes == nil {
		return
	}
	db := rep.DB().Model(obj).Updates(obj.changes)
	e = db.Error
	affected = db.RowsAffected
	return
}

// SetID .
func (obj *Order) SetID(iD int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.ID = iD
	obj.changes["id"] = iD
}

// SetUserID .
func (obj *Order) SetUserID(userID int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.UserID = userID
	obj.changes["user_id"] = userID
}

// SetGoodsID .
func (obj *Order) SetGoodsID(goodsID int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.GoodsID = goodsID
	obj.changes["goods_id"] = goodsID
}

// SetNum .
func (obj *Order) SetNum(num int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Num = num
	obj.changes["num"] = num
}

// SetCreated .
func (obj *Order) SetCreated(created time.Time) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Created = created
	obj.changes["created"] = created
}

// SetUpdated .
func (obj *Order) SetUpdated(updated time.Time) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Updated = updated
	obj.changes["updated"] = updated
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

// FindToUpdateOrders .
func FindToUpdateOrders(rep freedom.GORMRepository, query Order, values map[string]interface{}, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Order{}).Where(query).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(values)
	} else {
		db = db.Model(&Order{}).Where(query).Updates(values)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}

// FindByWhereToUpdateOrders .
func FindByWhereToUpdateOrders(rep freedom.GORMRepository, query string, args []interface{}, values map[string]interface{}, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Order{}).Where(query, args...).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(values)
	} else {
		db = db.Model(&Order{}).Where(query, args...).Updates(values)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}
