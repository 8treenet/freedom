package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/models"
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

func (repo *GoodsRepository) Get(id int) (result models.Goods, e error) {
	result, e = models.FindGoodsByPrimary(repo, id)
	return
}

func (repo *GoodsRepository) GetAll() (result []models.Goods, e error) {
	result, e = models.FindGoodss(repo, models.Goods{})
	return
}

func (repo *GoodsRepository) ChangeStock(goods *models.Goods, num int) (e error) {
	models.UpdateGoods(repo, goods, models.Goods{
		Stock: goods.Stock + num,
	})
	return
}
