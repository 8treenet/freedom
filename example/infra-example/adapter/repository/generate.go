package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/po"
	"github.com/jinzhu/gorm"
)

// GORMRepository .
type GORMRepository interface {
	db() *gorm.DB
	GetWorker() freedom.Worker
}

// NewORMDescBuilder .
func NewORMDescBuilder(column string, columns ...string) *Reorder {
	return newReorder("desc", column, columns...)
}

// NewORMAscBuilder .
func NewORMAscBuilder(column string, columns ...string) *Reorder {
	return newReorder("asc", column, columns...)
}

// NewORMBuilder .
func NewORMBuilder() *Builder {
	return &Builder{}
}

// NewDescOrder .
func newReorder(sort, field string, args ...string) *Reorder {
	fields := []string{field}
	fields = append(fields, args...)
	orders := []string{}
	for index := 0; index < len(fields); index++ {
		orders = append(orders, sort)
	}
	return &Reorder{
		fields: fields,
		orders: orders,
	}
}

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

// Builder .
type Builder struct {
	reorder      *Reorder
	pageSize     int
	page         int
	totalPage    int
	selectColumn []string
}

// TotalPage .
func (b *Builder) TotalPage() int {
	return b.totalPage
}

// Order .
func (b *Builder) Order() interface{} {
	if b.reorder != nil {
		return b.Order()
	}
	return ""
}

// Execute .
func (b *Builder) Execute(db *gorm.DB, object interface{}) (e error) {
	pageFind := false
	if b.reorder != nil {
		db = db.Order(b.reorder.Order())
	} else {
		db = db.Set("gorm:order_by_primary_key", "DESC")
	}
	if b.page != 0 && b.pageSize != 0 {
		pageFind = true
		db = db.Offset((b.page - 1) * b.pageSize).Limit(b.pageSize)
	}

	if len(b.selectColumn) > 0 {
		db = db.Select(b.selectColumn)
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
		if count%b.pageSize == 0 {
			b.totalPage = count / b.pageSize
		} else {
			b.totalPage = count/b.pageSize + 1
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

func ormErrorLog(repo GORMRepository, model, method string, e error, expression ...interface{}) {
	if e == nil || e == gorm.ErrRecordNotFound {
		return
	}
	repo.GetWorker().Logger().Errorf("Orm error, model: %s, method: %s, expression :%v, reason for error:%v", model, method, expression, e)
}

// findAdmin .
func findAdmin(repo GORMRepository, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdmin", e, now)
		ormErrorLog(repo, "Admin", "findAdmin", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAdminListByPrimarys .
func findAdminListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByPrimarys", e, now)
	ormErrorLog(repo, "Admin", "findAdminsByPrimarys", e, primarys)
	return
}

// findAdminByWhere .
func findAdminByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminByWhere", e, now)
		ormErrorLog(repo, "Admin", "findAdminByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAdminByMap .
func findAdminByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminByMap", e, now)
		ormErrorLog(repo, "Admin", "findAdminByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAdminList .
func findAdminList(repo GORMRepository, query po.Admin, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminList", e, now)
		ormErrorLog(repo, "Admin", "findAdmins", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findAdminListByWhere .
func findAdminListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByWhere", e, now)
		ormErrorLog(repo, "Admin", "findAdminsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findAdminListByMap .
func findAdminListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByMap", e, now)
		ormErrorLog(repo, "Admin", "findAdminsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createAdmin .
func createAdmin(repo GORMRepository, object *po.Admin) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Admin", "createAdmin", e, now)
	ormErrorLog(repo, "Admin", "createAdmin", e, *object)
	return
}

// saveAdmin .
func saveAdmin(repo GORMRepository, object *po.Admin) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Admin", "saveAdmin", e, now)
	ormErrorLog(repo, "Admin", "saveAdmin", e, *object)
	return
}

// findCart .
func findCart(repo GORMRepository, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCart", e, now)
		ormErrorLog(repo, "Cart", "findCart", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findCartListByPrimarys .
func findCartListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByPrimarys", e, now)
	ormErrorLog(repo, "Cart", "findCartsByPrimarys", e, primarys)
	return
}

// findCartByWhere .
func findCartByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartByWhere", e, now)
		ormErrorLog(repo, "Cart", "findCartByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findCartByMap .
func findCartByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartByMap", e, now)
		ormErrorLog(repo, "Cart", "findCartByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findCartList .
func findCartList(repo GORMRepository, query po.Cart, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartList", e, now)
		ormErrorLog(repo, "Cart", "findCarts", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findCartListByWhere .
func findCartListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByWhere", e, now)
		ormErrorLog(repo, "Cart", "findCartsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findCartListByMap .
func findCartListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByMap", e, now)
		ormErrorLog(repo, "Cart", "findCartsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createCart .
func createCart(repo GORMRepository, object *po.Cart) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Cart", "createCart", e, now)
	ormErrorLog(repo, "Cart", "createCart", e, *object)
	return
}

// saveCart .
func saveCart(repo GORMRepository, object *po.Cart) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Cart", "saveCart", e, now)
	ormErrorLog(repo, "Cart", "saveCart", e, *object)
	return
}

// findDelivery .
func findDelivery(repo GORMRepository, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDelivery", e, now)
		ormErrorLog(repo, "Delivery", "findDelivery", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDeliveryListByPrimarys .
func findDeliveryListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByPrimarys", e, now)
	ormErrorLog(repo, "Delivery", "findDeliverysByPrimarys", e, primarys)
	return
}

// findDeliveryByWhere .
func findDeliveryByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryByWhere", e, now)
		ormErrorLog(repo, "Delivery", "findDeliveryByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDeliveryByMap .
func findDeliveryByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryByMap", e, now)
		ormErrorLog(repo, "Delivery", "findDeliveryByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDeliveryList .
func findDeliveryList(repo GORMRepository, query po.Delivery, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryList", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverys", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findDeliveryListByWhere .
func findDeliveryListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByWhere", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverysByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findDeliveryListByMap .
func findDeliveryListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByMap", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverysByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createDelivery .
func createDelivery(repo GORMRepository, object *po.Delivery) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Delivery", "createDelivery", e, now)
	ormErrorLog(repo, "Delivery", "createDelivery", e, *object)
	return
}

// saveDelivery .
func saveDelivery(repo GORMRepository, object *po.Delivery) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Delivery", "saveDelivery", e, now)
	ormErrorLog(repo, "Delivery", "saveDelivery", e, *object)
	return
}

// findGoods .
func findGoods(repo GORMRepository, result interface{}, builders ...*Builder) (e error) {
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
func findGoodsListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByPrimarys", e, now)
	ormErrorLog(repo, "Goods", "findGoodssByPrimarys", e, primarys)
	return
}

// findGoodsByWhere .
func findGoodsByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...*Builder) (e error) {
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
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findGoodsByMap .
func findGoodsByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsByMap", e, now)
		ormErrorLog(repo, "Goods", "findGoodsByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findGoodsList .
func findGoodsList(repo GORMRepository, query po.Goods, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsList", e, now)
		ormErrorLog(repo, "Goods", "findGoodss", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findGoodsListByWhere .
func findGoodsListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...*Builder) (e error) {
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
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findGoodsListByMap .
func findGoodsListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByMap", e, now)
		ormErrorLog(repo, "Goods", "findGoodssByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createGoods .
func createGoods(repo GORMRepository, object *po.Goods) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Goods", "createGoods", e, now)
	ormErrorLog(repo, "Goods", "createGoods", e, *object)
	return
}

// saveGoods .
func saveGoods(repo GORMRepository, object *po.Goods) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Goods", "saveGoods", e, now)
	ormErrorLog(repo, "Goods", "saveGoods", e, *object)
	return
}

// findOrder .
func findOrder(repo GORMRepository, result interface{}, builders ...*Builder) (e error) {
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
func findOrderListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByPrimarys", e, now)
	ormErrorLog(repo, "Order", "findOrdersByPrimarys", e, primarys)
	return
}

// findOrderByWhere .
func findOrderByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...*Builder) (e error) {
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
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderByMap .
func findOrderByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderByMap", e, now)
		ormErrorLog(repo, "Order", "findOrderByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderList .
func findOrderList(repo GORMRepository, query po.Order, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderList", e, now)
		ormErrorLog(repo, "Order", "findOrders", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderListByWhere .
func findOrderListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...*Builder) (e error) {
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
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderListByMap .
func findOrderListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...*Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByMap", e, now)
		ormErrorLog(repo, "Order", "findOrdersByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createOrder .
func createOrder(repo GORMRepository, object *po.Order) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Order", "createOrder", e, now)
	ormErrorLog(repo, "Order", "createOrder", e, *object)
	return
}

// saveOrder .
func saveOrder(repo GORMRepository, object *po.Order) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Order", "saveOrder", e, now)
	ormErrorLog(repo, "Order", "saveOrder", e, *object)
	return
}
