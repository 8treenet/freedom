package repository

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/common"
	"github.com/8treenet/freedom/example/infra-example/domain/entity"
	"github.com/8treenet/freedom/example/infra-example/domain/po"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *GoodsRepository {
			return &GoodsRepository{}
		})
	})
}

// GoodsRepository .
type GoodsRepository struct {
	freedom.Repository
	EventManager *domainevent.EventManager
}

// Get .
func (repo *GoodsRepository) Get(ID int) (result *entity.Goods, e error) {
	result = &entity.Goods{}
	result.ID = ID
	//注入基础Entity
	repo.InjectBaseEntity(result)

	e = findGoods(repo, &result.Goods)
	return
}

// GetAll .
func (repo *GoodsRepository) GetAll() (result []*entity.Goods, e error) {
	list, e := findGoodsList(repo, po.Goods{})
	if e != nil {
		return
	}
	for _, v := range list {
		result = append(result, &entity.Goods{Goods: v})
	}

	//注入基础Entity
	repo.InjectBaseEntitys(result)
	return
}

// Save .
func (repo *GoodsRepository) Save(goods *entity.Goods) (e error) {
	goods.Update("version", gorm.Expr("version + ?", 1)) //版本号加1

	affected, e := saveGoods(repo, goods)
	if e != nil {
		return e
	}
	if affected == 0 {
		return common.ErrVersionExpired
	}
	return repo.EventManager.Save(&repo.Repository, goods) //持久化实体上的领域事件
}

func (repo *GoodsRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
