package services

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/repositorys"
	"github.com/8treenet/freedom/example/infra-example/objects"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *GoodsService {
			return &GoodsService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *GoodsService) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// GoodsService .
type GoodsService struct {
	Runtime   freedom.Runtime
	GoodsRepo repositorys.GoodsInterface
}

func (srv *GoodsService) Get(id int) (rep objects.GoodsRep, e error) {
	obj, e := srv.GoodsRepo.Get(id)
	if e != nil {
		return
	}
	rep.ID = obj.ID
	rep.Name = obj.Name
	rep.Stock = obj.Stock
	rep.Price = obj.Price
	return
}

func (srv *GoodsService) GetAll() (result []objects.GoodsRep, e error) {
	objs, e := srv.GoodsRepo.GetAll()
	if e != nil {
		return
	}
	for _, goodsModel := range objs {
		result = append(result, objects.GoodsRep{
			ID:    goodsModel.ID,
			Name:  goodsModel.Name,
			Price: goodsModel.Price,
			Stock: goodsModel.Stock,
		})
	}
	return
}

func (srv *GoodsService) AddStock(goodsID, num int) error {
	obj, e := srv.GoodsRepo.Get(goodsID)
	if e != nil {
		return e
	}

	return srv.GoodsRepo.ChangeStock(&obj, num)
}
