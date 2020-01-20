package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/objects"
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
	result, e = objects.FindGoodsByPrimary(repo, id)
	return
}

func (repo *GoodsRepository) GetAll() (result []objects.Goods, e error) {
	result, e = objects.FindGoodss(repo, objects.Goods{})
	return
}

func (repo *GoodsRepository) ChangeStock(goods *objects.Goods, num int) (e error) {
	objects.UpdateGoods(repo, goods, objects.Goods{
		Stock: goods.Stock + num,
	})
	return
}
