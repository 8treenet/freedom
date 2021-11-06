package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/po"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GORMRepository .
type GORMRepository interface {
	db() *gorm.DB
	Worker() freedom.Worker
}

type saveObject interface {
	TableName() string
	Location() map[string]interface{}
	GetChanges() map[string]interface{}
}

// Builder .
type Builder interface {
	Execute(db *gorm.DB, object interface{}) error
}

// Pager .
type Pager struct {
	pageSize  int
	page      int
	totalPage int
	fields    []string
	orders    []string
}

// NewDescPager .
func NewDescPager(column string, columns ...string) *Pager {
	return newDefaultPager("desc", column, columns...)
}

// NewAscPager .
func NewAscPager(column string, columns ...string) *Pager {
	return newDefaultPager("asc", column, columns...)
}

// NewDescOrder .
func newDefaultPager(sort, field string, args ...string) *Pager {
	fields := []string{field}
	fields = append(fields, args...)
	orders := []string{}
	for index := 0; index < len(fields); index++ {
		orders = append(orders, sort)
	}
	return &Pager{
		fields: fields,
		orders: orders,
	}
}

// Order .
func (p *Pager) Order() interface{} {
	if len(p.fields) == 0 {
		return nil
	}
	args := []string{}
	for index := 0; index < len(p.fields); index++ {
		args = append(args, fmt.Sprintf("`%s` %s", p.fields[index], p.orders[index]))
	}

	return strings.Join(args, ",")
}

// TotalPage .
func (p *Pager) TotalPage() int {
	return p.totalPage
}

// SetPage .
func (p *Pager) SetPage(page, pageSize int) *Pager {
	p.page = page
	p.pageSize = pageSize
	return p
}

// Execute .
func (p *Pager) Execute(db *gorm.DB, object interface{}) (e error) {
	if p.page != 0 && p.pageSize != 0 {
		var count64 int64
		e = db.Model(object).Count(&count64).Error
		count := int(count64)
		if e != nil {
			return
		}
		if count != 0 {
			//Calculate the length of the pagination
			if count%p.pageSize == 0 {
				p.totalPage = count / p.pageSize
			} else {
				p.totalPage = count/p.pageSize + 1
			}
		}
		db = db.Offset((p.page - 1) * p.pageSize).Limit(p.pageSize)
	}

	orderValue := p.Order()
	if orderValue != nil {
		db = db.Order(orderValue)
	}

	resultDB := db.Find(object)
	if resultDB.Error != nil {
		return resultDB.Error
	}
	return
}

// Limiter .
type Limiter struct {
	size   int
	column string
	desc   bool
}

// NewDescLimiter .
func NewDescLimiter(column string, size int) *Limiter {
	return &Limiter{column: column, size: size, desc: true}
}

// NewAscLimiter .
func NewAscLimiter(column string, size int) *Limiter {
	return &Limiter{column: column, size: size, desc: false}
}

// Execute .
func (limiter *Limiter) Execute(db *gorm.DB, object interface{}) (e error) {
	db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: limiter.column}, Desc: limiter.desc}).Limit(limiter.size)
	resultDB := db.Find(object)
	if resultDB.Error != nil {
		return resultDB.Error
	}
	return
}

func ormErrorLog(repo GORMRepository, model, method string, e error, expression ...interface{}) {
	if e == nil || e == gorm.ErrRecordNotFound {
		return
	}
	repo.Worker().Logger().Errorf("error: %v, model: %s, method: %s", e, model, method)
}

// findOrder .
func findOrder(repo GORMRepository, result *po.Order, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrder", e, now)
		ormErrorLog(repo, "Order", "findOrder", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderListByPrimarys .
func findOrderListByPrimarys(repo GORMRepository, primarys ...interface{}) (results []po.Order, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByPrimarys", e, now)
		ormErrorLog(repo, "Order", "findOrdersByPrimarys", e, primarys)
	}()

	e = repo.db().Find(&results, primarys).Error
	return
}

// findOrderByWhere .
func findOrderByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (result po.Order, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderByWhere", e, now)
		ormErrorLog(repo, "Order", "findOrderByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findOrderByMap .
func findOrderByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (result po.Order, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderByMap", e, now)
		ormErrorLog(repo, "Order", "findOrderByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findOrderList .
func findOrderList(repo GORMRepository, query po.Order, builders ...Builder) (results []po.Order, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderList", e, now)
		ormErrorLog(repo, "Order", "findOrders", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findOrderListByWhere .
func findOrderListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (results []po.Order, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByWhere", e, now)
		ormErrorLog(repo, "Order", "findOrdersByWhere", e, query, args)
	}()
	db := repo.db()
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

// findOrderListByMap .
func findOrderListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (results []po.Order, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByMap", e, now)
		ormErrorLog(repo, "Order", "findOrdersByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// createOrder .
func createOrder(repo GORMRepository, object *po.Order) (rowsAffected int64, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "createOrder", e, now)
		ormErrorLog(repo, "Order", "createOrder", e, *object)
	}()

	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// saveOrder .
func saveOrder(repo GORMRepository, object saveObject) (rowsAffected int64, e error) {
	if len(object.Location()) == 0 {
		return 0, errors.New("location cannot be empty")
	}
	updateValues := object.GetChanges()
	if len(updateValues) == 0 {
		return 0, nil
	}

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "saveOrder", e, now)
		ormErrorLog(repo, "Order", "saveOrder", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(updateValues)
	e = db.Error
	rowsAffected = db.RowsAffected
	return
}

// findGoods .
func findGoods(repo GORMRepository, result *po.Goods, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoods", e, now)
		ormErrorLog(repo, "Goods", "findGoods", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findGoodsListByPrimarys .
func findGoodsListByPrimarys(repo GORMRepository, primarys ...interface{}) (results []po.Goods, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByPrimarys", e, now)
		ormErrorLog(repo, "Goods", "findGoodssByPrimarys", e, primarys)
	}()

	e = repo.db().Find(&results, primarys).Error
	return
}

// findGoodsByWhere .
func findGoodsByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (result po.Goods, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsByWhere", e, now)
		ormErrorLog(repo, "Goods", "findGoodsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findGoodsByMap .
func findGoodsByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (result po.Goods, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsByMap", e, now)
		ormErrorLog(repo, "Goods", "findGoodsByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findGoodsList .
func findGoodsList(repo GORMRepository, query po.Goods, builders ...Builder) (results []po.Goods, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsList", e, now)
		ormErrorLog(repo, "Goods", "findGoodss", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findGoodsListByWhere .
func findGoodsListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (results []po.Goods, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByWhere", e, now)
		ormErrorLog(repo, "Goods", "findGoodssByWhere", e, query, args)
	}()
	db := repo.db()
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

// findGoodsListByMap .
func findGoodsListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (results []po.Goods, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByMap", e, now)
		ormErrorLog(repo, "Goods", "findGoodssByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// createGoods .
func createGoods(repo GORMRepository, object *po.Goods) (rowsAffected int64, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "createGoods", e, now)
		ormErrorLog(repo, "Goods", "createGoods", e, *object)
	}()

	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// saveGoods .
func saveGoods(repo GORMRepository, object saveObject) (rowsAffected int64, e error) {
	if len(object.Location()) == 0 {
		return 0, errors.New("location cannot be empty")
	}
	updateValues := object.GetChanges()
	if len(updateValues) == 0 {
		return 0, nil
	}

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "saveGoods", e, now)
		ormErrorLog(repo, "Goods", "saveGoods", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(updateValues)
	e = db.Error
	rowsAffected = db.RowsAffected
	return
}
