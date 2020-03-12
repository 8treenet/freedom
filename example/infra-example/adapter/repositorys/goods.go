package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/application/object"
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

func (repo *GoodsRepository) Get(id int) (result object.Goods, e error) {
	// SingleFlight 防击穿
	e = repo.SingleFlight(result.TableName(), id, &result, func() (interface{}, error) {
		var obj object.Goods
		e := findGoodsByPrimary(repo, &obj, id)
		return obj, e
	})
	return
}

func (repo *GoodsRepository) GetAll() (result []object.Goods, e error) {
	e = findGoodss(repo, object.Goods{}, &result)
	return
}

func (repo *GoodsRepository) Save(goods *object.Goods) (e error) {
	_, e = saveGoods(repo, goods)
	return
}
