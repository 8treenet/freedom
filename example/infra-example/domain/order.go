package domain

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/adapter/repository"
	"github.com/8treenet/freedom/example/infra-example/domain/dto"
	"github.com/8treenet/freedom/infra/transaction"
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
	Worker    freedom.Worker
	GoodsRepo repository.GoodsInterface
	OrderRepo repository.OrderInterface
	Tx        transaction.Transaction
}

func (srv *OrderService) Get(id, userId int) (result dto.OrderRep, e error) {
	obj, e := srv.OrderRepo.Get(id, userId)
	if e != nil {
		return
	}
	goodsObj, e := srv.GoodsRepo.Get(obj.GoodsId)
	if e != nil {
		return
	}
	result.Id = obj.Id
	result.GoodsId = obj.GoodsId
	result.Num = obj.Num
	result.DateTime = obj.Created.Format("2006-01-02 15:04:05")
	result.GoodsName = goodsObj.Name
	return
}

func (srv *OrderService) GetAll(userId int) (result []dto.OrderRep, e error) {
	objs, e := srv.OrderRepo.GetAll(userId)
	if e != nil {
		return
	}

	for _, obj := range objs {
		goodsObj, err := srv.GoodsRepo.Get(obj.GoodsId)
		if err != nil {
			e = err
			return
		}

		result = append(result, dto.OrderRep{
			Id:        obj.Id,
			GoodsId:   obj.GoodsId,
			GoodsName: goodsObj.Name,
			Num:       obj.Num,
			DateTime:  obj.Created.Format("2006-01-02 15:04:05"),
		})
	}
	return
}

func (srv *OrderService) Add(goodsID, num, userId int) (resp string, e error) {
	goodsObj, e := srv.GoodsRepo.Get(goodsID)
	if e != nil {
		return
	}
	if goodsObj.Stock < num {
		resp = "库存不足"
		return
	}
	goodsObj.AddStock(-num)

	e = srv.Tx.Execute(func() error {
		if err := srv.GoodsRepo.Save(&goodsObj); err != nil {
			return err
		}

		return srv.OrderRepo.Create(goodsObj.Id, num, userId)
	})
	return
}
