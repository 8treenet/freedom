package repository

import (
	"errors"
	"fmt"
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"gorm.io/gorm"
	"strings"
	"time"
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
	pageFind := false
	orderValue := p.Order()
	if orderValue != nil {
		db = db.Order(orderValue)
	}
	if p.page != 0 && p.pageSize != 0 {
		pageFind = true
		db = db.Offset((p.page - 1) * p.pageSize).Limit(p.pageSize)
	}

	resultDB := db.Find(object)
	if resultDB.Error != nil {
		return resultDB.Error
	}

	if !pageFind {
		return
	}

	var count64 int64
	e = resultDB.Offset(0).Limit(1).Count(&count64).Error
	count := int(count64)
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

func ormErrorLog(repo GORMRepository, model, method string, e error, expression ...interface{}) {
	if e == nil || e == gorm.ErrRecordNotFound {
		return
	}
	repo.Worker().Logger().Errorf("error: %v, model: %s, method: %s", e, model, method)
}

// findUser .
func findUser(repo GORMRepository, result *po.User, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUser", e, now)
		ormErrorLog(repo, "User", "findUser", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findUserListByPrimarys .
func findUserListByPrimarys(repo GORMRepository, primarys ...interface{}) (results []po.User, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserListByPrimarys", e, now)
		ormErrorLog(repo, "User", "findUsersByPrimarys", e, primarys)
	}()

	e = repo.db().Find(&results, primarys).Error
	return
}

// findUserByWhere .
func findUserByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (result po.User, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserByWhere", e, now)
		ormErrorLog(repo, "User", "findUserByWhere", e, query, args)
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

// findUserByMap .
func findUserByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (result po.User, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserByMap", e, now)
		ormErrorLog(repo, "User", "findUserByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findUserList .
func findUserList(repo GORMRepository, query po.User, builders ...Builder) (results []po.User, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserList", e, now)
		ormErrorLog(repo, "User", "findUsers", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findUserListByWhere .
func findUserListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (results []po.User, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserListByWhere", e, now)
		ormErrorLog(repo, "User", "findUsersByWhere", e, query, args)
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

// findUserListByMap .
func findUserListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (results []po.User, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserListByMap", e, now)
		ormErrorLog(repo, "User", "findUsersByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// createUser .
func createUser(repo GORMRepository, object *po.User) (rowsAffected int64, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "createUser", e, now)
		ormErrorLog(repo, "User", "createUser", e, *object)
	}()

	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// saveUser .
func saveUser(repo GORMRepository, object saveObject) (rowsAffected int64, e error) {
	if len(object.Location()) == 0 {
		return 0, errors.New("location cannot be empty")
	}

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "saveUser", e, now)
		ormErrorLog(repo, "User", "saveUser", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(object.GetChanges())
	e = db.Error
	rowsAffected = db.RowsAffected
	return
}

// findOrderDetail .
func findOrderDetail(repo GORMRepository, result *po.OrderDetail, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetail", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetail", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderDetailListByPrimarys .
func findOrderDetailListByPrimarys(repo GORMRepository, primarys ...interface{}) (results []po.OrderDetail, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailListByPrimarys", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailsByPrimarys", e, primarys)
	}()

	e = repo.db().Find(&results, primarys).Error
	return
}

// findOrderDetailByWhere .
func findOrderDetailByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (result po.OrderDetail, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailByWhere", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailByWhere", e, query, args)
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

// findOrderDetailByMap .
func findOrderDetailByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (result po.OrderDetail, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailByMap", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findOrderDetailList .
func findOrderDetailList(repo GORMRepository, query po.OrderDetail, builders ...Builder) (results []po.OrderDetail, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailList", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetails", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findOrderDetailListByWhere .
func findOrderDetailListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (results []po.OrderDetail, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailListByWhere", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailsByWhere", e, query, args)
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

// findOrderDetailListByMap .
func findOrderDetailListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (results []po.OrderDetail, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailListByMap", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// createOrderDetail .
func createOrderDetail(repo GORMRepository, object *po.OrderDetail) (rowsAffected int64, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "createOrderDetail", e, now)
		ormErrorLog(repo, "OrderDetail", "createOrderDetail", e, *object)
	}()

	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// saveOrderDetail .
func saveOrderDetail(repo GORMRepository, object saveObject) (rowsAffected int64, e error) {
	if len(object.Location()) == 0 {
		return 0, errors.New("location cannot be empty")
	}

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "saveOrderDetail", e, now)
		ormErrorLog(repo, "OrderDetail", "saveOrderDetail", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(object.GetChanges())
	e = db.Error
	rowsAffected = db.RowsAffected
	return
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

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "saveOrder", e, now)
		ormErrorLog(repo, "Order", "saveOrder", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(object.GetChanges())
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

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "saveGoods", e, now)
		ormErrorLog(repo, "Goods", "saveGoods", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(object.GetChanges())
	e = db.Error
	rowsAffected = db.RowsAffected
	return
}

// findDelivery .
func findDelivery(repo GORMRepository, result *po.Delivery, builders ...Builder) (e error) {
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
func findDeliveryListByPrimarys(repo GORMRepository, primarys ...interface{}) (results []po.Delivery, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByPrimarys", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverysByPrimarys", e, primarys)
	}()

	e = repo.db().Find(&results, primarys).Error
	return
}

// findDeliveryByWhere .
func findDeliveryByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (result po.Delivery, e error) {
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
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findDeliveryByMap .
func findDeliveryByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (result po.Delivery, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryByMap", e, now)
		ormErrorLog(repo, "Delivery", "findDeliveryByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findDeliveryList .
func findDeliveryList(repo GORMRepository, query po.Delivery, builders ...Builder) (results []po.Delivery, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryList", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverys", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findDeliveryListByWhere .
func findDeliveryListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (results []po.Delivery, e error) {
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
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findDeliveryListByMap .
func findDeliveryListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (results []po.Delivery, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByMap", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverysByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// createDelivery .
func createDelivery(repo GORMRepository, object *po.Delivery) (rowsAffected int64, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "createDelivery", e, now)
		ormErrorLog(repo, "Delivery", "createDelivery", e, *object)
	}()

	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// saveDelivery .
func saveDelivery(repo GORMRepository, object saveObject) (rowsAffected int64, e error) {
	if len(object.Location()) == 0 {
		return 0, errors.New("location cannot be empty")
	}

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "saveDelivery", e, now)
		ormErrorLog(repo, "Delivery", "saveDelivery", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(object.GetChanges())
	e = db.Error
	rowsAffected = db.RowsAffected
	return
}

// findCart .
func findCart(repo GORMRepository, result *po.Cart, builders ...Builder) (e error) {
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
func findCartListByPrimarys(repo GORMRepository, primarys ...interface{}) (results []po.Cart, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByPrimarys", e, now)
		ormErrorLog(repo, "Cart", "findCartsByPrimarys", e, primarys)
	}()

	e = repo.db().Find(&results, primarys).Error
	return
}

// findCartByWhere .
func findCartByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (result po.Cart, e error) {
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
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findCartByMap .
func findCartByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (result po.Cart, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartByMap", e, now)
		ormErrorLog(repo, "Cart", "findCartByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findCartList .
func findCartList(repo GORMRepository, query po.Cart, builders ...Builder) (results []po.Cart, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartList", e, now)
		ormErrorLog(repo, "Cart", "findCarts", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findCartListByWhere .
func findCartListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (results []po.Cart, e error) {
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
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findCartListByMap .
func findCartListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (results []po.Cart, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByMap", e, now)
		ormErrorLog(repo, "Cart", "findCartsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// createCart .
func createCart(repo GORMRepository, object *po.Cart) (rowsAffected int64, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "createCart", e, now)
		ormErrorLog(repo, "Cart", "createCart", e, *object)
	}()

	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// saveCart .
func saveCart(repo GORMRepository, object saveObject) (rowsAffected int64, e error) {
	if len(object.Location()) == 0 {
		return 0, errors.New("location cannot be empty")
	}

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "saveCart", e, now)
		ormErrorLog(repo, "Cart", "saveCart", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(object.GetChanges())
	e = db.Error
	rowsAffected = db.RowsAffected
	return
}

// findAdmin .
func findAdmin(repo GORMRepository, result *po.Admin, builders ...Builder) (e error) {
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
func findAdminListByPrimarys(repo GORMRepository, primarys ...interface{}) (results []po.Admin, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByPrimarys", e, now)
		ormErrorLog(repo, "Admin", "findAdminsByPrimarys", e, primarys)
	}()

	e = repo.db().Find(&results, primarys).Error
	return
}

// findAdminByWhere .
func findAdminByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (result po.Admin, e error) {
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
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findAdminByMap .
func findAdminByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (result po.Admin, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminByMap", e, now)
		ormErrorLog(repo, "Admin", "findAdminByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), &result)
	return
}

// findAdminList .
func findAdminList(repo GORMRepository, query po.Admin, builders ...Builder) (results []po.Admin, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminList", e, now)
		ormErrorLog(repo, "Admin", "findAdmins", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findAdminListByWhere .
func findAdminListByWhere(repo GORMRepository, query string, args []interface{}, builders ...Builder) (results []po.Admin, e error) {
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
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// findAdminListByMap .
func findAdminListByMap(repo GORMRepository, query map[string]interface{}, builders ...Builder) (results []po.Admin, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByMap", e, now)
		ormErrorLog(repo, "Admin", "findAdminsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// createAdmin .
func createAdmin(repo GORMRepository, object *po.Admin) (rowsAffected int64, e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "createAdmin", e, now)
		ormErrorLog(repo, "Admin", "createAdmin", e, *object)
	}()

	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// saveAdmin .
func saveAdmin(repo GORMRepository, object saveObject) (rowsAffected int64, e error) {
	if len(object.Location()) == 0 {
		return 0, errors.New("location cannot be empty")
	}

	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "saveAdmin", e, now)
		ormErrorLog(repo, "Admin", "saveAdmin", e, object)
	}()

	db := repo.db().Table(object.TableName()).Where(object.Location()).Updates(object.GetChanges())
	e = db.Error
	rowsAffected = db.RowsAffected
	return
}
