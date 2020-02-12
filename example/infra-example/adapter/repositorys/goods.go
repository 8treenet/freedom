package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/application/objects"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *GoodsRepository {
			return &GoodsRepository{}
		})
	})
}

// GoodsRepository .
type GoodsRepository struct {
	freedom.Repository
}

func (repo *GoodsRepository) Get(id int) (result objects.Goods, e error) {
	// SingleFlight 防击穿
	e = repo.SingleFlight(result.TableName(), id, &result, func() (interface{}, error) {
		return findGoodsByPrimary(repo, id)
	})
	return
}

func (repo *GoodsRepository) GetAll() (result []objects.Goods, e error) {
	result, e = findGoodss(repo, objects.Goods{})
	return
}

func (repo *GoodsRepository) Save(goods *objects.Goods) (e error) {
	_, e = updateGoods(repo, goods)
	return
}
