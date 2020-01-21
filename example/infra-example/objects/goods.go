package objects

import (
	"github.com/8treenet/freedom"
	"time"
)

// Goods .
type Goods struct {
	changes map[string]interface{}
	ID      int       `gorm:"primary_key" column:"id"`
	Name    string    `gorm:"column:name"`  // 商品名称
	Price   int       `gorm:"column:price"` // 价格
	Stock   int       `gorm:"column:stock"` // 库存
	Created time.Time `gorm:"column:created"`
	Updated time.Time `gorm:"column:updated"`
}

func (obj *Goods) TableName() string {
	return "goods"
}

// Updates .
func (obj *Goods) Updates(rep freedom.GORMRepository) (affected int64, e error) {
	if obj.changes == nil {
		return
	}
	db := rep.DB().Model(obj).Updates(obj.changes)
	e = db.Error
	affected = db.RowsAffected
	return
}

// SetID .
func (obj *Goods) SetID(iD int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.ID = iD
	obj.changes["id"] = iD
}

// SetName .
func (obj *Goods) SetName(name string) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Name = name
	obj.changes["name"] = name
}

// SetPrice .
func (obj *Goods) SetPrice(price int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Price = price
	obj.changes["price"] = price
}

// SetStock .
func (obj *Goods) SetStock(stock int) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Stock = stock
	obj.changes["stock"] = stock
}

// SetCreated .
func (obj *Goods) SetCreated(created time.Time) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Created = created
	obj.changes["created"] = created
}

// SetUpdated .
func (obj *Goods) SetUpdated(updated time.Time) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.Updated = updated
	obj.changes["updated"] = updated
}

// FindGoodsByPrimary .
func FindGoodsByPrimary(rep freedom.GORMRepository, primary interface{}) (result Goods, e error) {
	e = rep.DB().Find(&result, primary).Error
	return
}

// FindGoodssByPrimarys .
func FindGoodssByPrimarys(rep freedom.GORMRepository, primarys ...interface{}) (results []Goods, e error) {
	e = rep.DB().Find(&results, primarys).Error
	return
}

// FindGoods .
func FindGoods(rep freedom.GORMRepository, query Goods, builders ...freedom.QueryBuilder) (result Goods, e error) {
	db := rep.DB().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = db.Limit(1).Order(builders[0].Order()).Find(&result).Error
	return
}

// FindGoodsByWhere .
func FindGoodsByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (result Goods, e error) {
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

// FindGoodss .
func FindGoodss(rep freedom.GORMRepository, query Goods, builders ...freedom.QueryBuilder) (results []Goods, e error) {
	db := rep.DB().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// FindGoodssByWhere .
func FindGoodssByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (results []Goods, e error) {
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

// CreateGoods .
func CreateGoods(rep freedom.GORMRepository, entity *Goods) (rowsAffected int64, e error) {
	db := rep.DB().Create(entity)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// FindToUpdateGoodss .
func FindToUpdateGoodss(rep freedom.GORMRepository, query Goods, values map[string]interface{}, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Goods{}).Where(query).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(values)
	} else {
		db = db.Model(&Goods{}).Where(query).Updates(values)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}

// FindByWhereToUpdateGoodss .
func FindByWhereToUpdateGoodss(rep freedom.GORMRepository, query string, args []interface{}, values map[string]interface{}, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Goods{}).Where(query, args...).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(values)
	} else {
		db = db.Model(&Goods{}).Where(query, args...).Updates(values)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}
