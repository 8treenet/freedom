package repositorys

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/application/objects"
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

	correctReq := repo.NewH2CRequest(addr)
	go func() {
		var model objects.GoodsModel
		/*
			panic 错误方式 : repo.NewH2CRequest方法使用了请求的运行时,用来设置传递给下游的trace相关数据，并发中请求运行时结束会导致panic
			repo.NewH2CRequest(addr).Get().ToJSON(&model)
		*/
		//正确方式1, 在请求内创建
		correctReq.Get().ToJSON(&model)

		//正确方式2, 不向下游传递trace相关数据,不会访问请求运行时。通常访问外部服务设置false。
		repo.NewH2CRequest(addr, false).Get().ToJSON(&model)
	}()
	return result
}
