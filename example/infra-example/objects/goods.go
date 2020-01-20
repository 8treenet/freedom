package objects

import (
	"github.com/8treenet/freedom"
	"time"
)

// Goods .
type Goods struct {
	ID      int       `gorm:"primary_key" column:"id"`
	Name    string    `gorm:"column:name"`  // 商品名称
	Price   int       `gorm:"column:price"` // 价格
	Stock   int       `gorm:"column:stock"` // 库存
	Created time.Time `gorm:"column:created"`
	Updated time.Time `gorm:"column:updated"`
}

func (m *Goods) TableName() string {
	return "goods"
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

// UpdateGoods .
func UpdateGoods(rep freedom.GORMRepository, query *Goods, value Goods) (affected int64, e error) {
	db := rep.DB().Model(query).Updates(value)
	e = db.Error
	affected = db.RowsAffected
	return
}

// FindToUpdateGoodss .
func FindToUpdateGoodss(rep freedom.GORMRepository, query Goods, value Goods, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Goods{}).Where(query).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&Goods{}).Where(query).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}

// FindByWhereToUpdateGoodss .
func FindByWhereToUpdateGoodss(rep freedom.GORMRepository, query string, args []interface{}, value Goods, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Goods{}).Where(query, args...).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&Goods{}).Where(query, args...).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}
