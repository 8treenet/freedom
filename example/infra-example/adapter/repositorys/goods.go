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
	e = findGoodsByPrimary(repo, &result, id)
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
