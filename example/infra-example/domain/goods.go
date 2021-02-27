package domain

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/adapter/repository"
	"github.com/8treenet/freedom/example/infra-example/domain/event"
	"github.com/8treenet/freedom/example/infra-example/domain/vo"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *GoodsService {
			return &GoodsService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *GoodsService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// GoodsService .
type GoodsService struct {
	Worker           freedom.Worker
	GoodsRepo        *repository.GoodsRepository
	EventTransaction *domainevent.EventTransaction //事务组件
}

// Get .
func (srv *GoodsService) Get(ID int) (rep vo.GoodsRep, e error) {
	entity, e := srv.GoodsRepo.Get(ID)
	if e != nil {
		return
	}
	rep.ID = entity.ID
	rep.Name = entity.Name
	rep.Stock = entity.Stock
	rep.Price = entity.Price
	return
}

// GetAll .
func (srv *GoodsService) GetAll() (result []vo.GoodsRep, e error) {
	entitys, e := srv.GoodsRepo.GetAll()
	if e != nil {
		return
	}
	for _, goodsModel := range entitys {
		result = append(result, vo.GoodsRep{
			ID:    goodsModel.ID,
			Name:  goodsModel.Name,
			Price: goodsModel.Price,
			Stock: goodsModel.Stock,
		})
	}
	return
}

// ShopEvent .
func (srv *GoodsService) ShopEvent(shopEvent *event.ShopGoods) error {
	entity, e := srv.GoodsRepo.Get(shopEvent.GoodsID) //通过id取商品实体
	if e != nil {
		return e
	}

	entity.AddStock(shopEvent.GoodsNum) //增加库存
	entity.AddSubEvent(shopEvent)       //为实体加入消费事件

	//使用事务组件保证一致性 1.修改商品库存, 2.事件表修改状态为已处理
	return srv.EventTransaction.Execute(func() error {
		return srv.GoodsRepo.Save(entity)
	})
}
