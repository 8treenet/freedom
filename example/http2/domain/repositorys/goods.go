package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/objects"
	"strconv"
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

// GetGoods implment Goods interface
func (repo *GoodsRepository) GetGoods(goodsID int) (result objects.GoodsModel) {
	repo.Runtime.Logger().Info("我是GoodsRepository")
	//通过h2c request 访问本服务 /goods/:id
	addr := "http://127.0.0.1:8000/goods/" + strconv.Itoa(goodsID)
	repo.NewH2CRequest(addr).Get().ToJSON(&result)
	return result
}
