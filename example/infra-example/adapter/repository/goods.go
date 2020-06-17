package repository

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/adapter/po"
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
}

func (repo *GoodsRepository) Get(id int) (result po.Goods, e error) {
	result.Id = id
	e = findGoods(repo, &result)
	return
}

func (repo *GoodsRepository) GetAll() (result []po.Goods, e error) {
	e = findGoodsList(repo, po.Goods{}, &result)
	return
}

func (repo *GoodsRepository) Save(goods *po.Goods) (e error) {
	_, e = saveGoods(repo, goods)
	return
}
