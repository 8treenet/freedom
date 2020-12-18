package repository

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/entity"
	"github.com/8treenet/freedom/example/infra-example/domain/po"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
	"github.com/jinzhu/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *OrderRepository {
			return &OrderRepository{}
		})
	})
}

// OrderRepository .
type OrderRepository struct {
	freedom.Repository
	EventManager *domainevent.EventManager
}

// Get .
func (repo *OrderRepository) Get(ID, userID int) (result *entity.Order, e error) {
	result = &entity.Order{}
	result.ID = ID
	result.UserID = userID

	//注入基础Entity
	repo.InjectBaseEntity(result)

	e = findOrder(repo, &result)
	return
}

// GetAll .
func (repo *OrderRepository) GetAll(userID int) (result []*entity.Order, e error) {
	e = findOrderList(repo, po.Order{UserID: userID}, &result)
	if e != nil {
		return
	}

	//注入基础Entity
	repo.InjectBaseEntitys(result)
	return
}

// Create .
func (repo *OrderRepository) Create(goodsID, num, userID int) error {
	_, e := createOrder(repo, &po.Order{
		UserID:  userID,
		GoodsID: goodsID,
		Num:     num,
		Created: time.Now(),
		Updated: time.Now(),
	})
	return e
}

func (repo *OrderRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	db.SetLogger(repo.Worker().Logger())
	return db
}
