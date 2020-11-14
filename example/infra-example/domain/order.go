package domain

import (
	"errors"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/adapter/repository"
	"github.com/8treenet/freedom/example/infra-example/domain/dto"
	"github.com/8treenet/freedom/example/infra-example/domain/event"
	"github.com/8treenet/freedom/example/infra-example/infra/domainevent"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *OrderService {
			return &OrderService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *OrderService) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// OrderService .
type OrderService struct {
	Worker           freedom.Worker
	GoodsRepo        *repository.GoodsRepository
	OrderRepo        *repository.OrderRepository
	EventTransaction *domainevent.EventTransaction //事务组件
}

// Get .
func (srv *OrderService) Get(ID, userID int) (result dto.OrderRep, e error) {
	entity, e := srv.OrderRepo.Get(ID, userID)
	if e != nil {
		return
	}
	goodsEntity, e := srv.GoodsRepo.Get(entity.GoodsID)
	if e != nil {
		return
	}
	result.ID = entity.ID
	result.GoodsID = entity.GoodsID
	result.Num = entity.Num
	result.DateTime = entity.Created.Format("2006-01-02 15:04:05")
	result.GoodsName = goodsEntity.Name
	return
}

// GetAll .
func (srv *OrderService) GetAll(userID int) (result []dto.OrderRep, e error) {
	entitys, e := srv.OrderRepo.GetAll(userID)
	if e != nil {
		return
	}

	for _, obj := range entitys {
		goodsEntity, err := srv.GoodsRepo.Get(obj.GoodsID)
		if err != nil {
			e = err
			return
		}

		result = append(result, dto.OrderRep{
			ID:        obj.ID,
			GoodsID:   obj.GoodsID,
			GoodsName: goodsEntity.Name,
			Num:       obj.Num,
			DateTime:  obj.Created.Format("2006-01-02 15:04:05"),
		})
	}
	return
}

// Shop 这不是一个正确的示例，只是为展示领域事件和Kafka的结合, 请参考fshop的聚合根.
func (srv *OrderService) Shop(goodsID, num, userID int) (e error) {
	goodsEntity, e := srv.GoodsRepo.Get(goodsID)
	if e != nil {
		return
	}
	if goodsEntity.Stock < num {
		e = errors.New("库存不足")
		return
	}
	goodsEntity.AddStock(-num) //扣库存

	//为商品实体增加购买事件
	goodsEntity.AddPubEvent(&event.ShopGoods{
		UserID:    userID,
		GoodsID:   goodsID,
		GoodsNum:  num,
		GoodsName: goodsEntity.Name,
	})

	//使用事务组件保证一致性 1.修改商品库存, 2.创建订单, 3.事件表增加记录
	//Execute 如果返回错误 会触发回滚。成功会调用infra/domainevent/EventManager.push
	e = srv.EventTransaction.Execute(func() error {
		if err := srv.GoodsRepo.Save(goodsEntity); err != nil {
			return err
		}

		return srv.OrderRepo.Create(goodsEntity.ID, num, userID)
	})
	return
}
