package domain

import (
	"fmt"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/adapter/repository"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *ShopService {
			return &ShopService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *ShopService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// ShopService .
type ShopService struct {
	Worker freedom.Worker
	Goods  repository.GoodsInterface
}

// Shopping implment Shopping interface
func (s *ShopService) Shopping(goodsID int) string {
	s.Worker.Logger().Info("我是ShopService.Shopping")
	entity := s.Goods.GetGoods(goodsID)
	return fmt.Sprintf("您花费了%d元, 购买了一件 %s。", entity.Price, entity.Name)
}
