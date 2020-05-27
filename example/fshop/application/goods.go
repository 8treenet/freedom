package application

import (
	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/application/aggregate"
	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/example/fshop/application/entity"
	"github.com/8treenet/freedom/infra/transaction"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *Goods {
			return &Goods{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *Goods) {
			initiator.GetService(ctx, &service)
			return
		})
	})
}

// Goods 商品领域服务.
type Goods struct {
	Worker    freedom.Worker       //运行时，一个请求绑定一个运行时
	GoodsRepo repository.GoodsRepo //商品仓库
	OrderRepo repository.OrderRepo //订单仓库
	UserRepo  repository.UserRepo  //用户仓库

	Transaction transaction.Transaction //事务组件
}

// New 创建商品
func (g *Goods) New(name string, price int) (e error) {
	_, e = g.GoodsRepo.New(name, entity.GoodsNoneTag, price, 100)
	return
}

// Items 分页商品列表
func (g *Goods) Items(page, pagesize int, tag string) (items []dto.GoodsItemRes, e error) {
	entitys, e := g.GoodsRepo.FindsByPage(page, pagesize, tag)
	if e != nil {
		return
	}

	for i := 0; i < len(entitys); i++ {
		items = append(items, dto.GoodsItemRes{
			Id:    entitys[i].Id,
			Name:  entitys[i].Name,
			Price: entitys[i].Price,
			Stock: entitys[i].Stock,
			Tag:   entitys[i].Tag,
		})
	}
	return
}

// AddStock 增加商品库存
func (g *Goods) AddStock(goodsId, num int) (e error) {
	entity, e := g.GoodsRepo.Get(goodsId)
	if e != nil {
		return
	}

	entity.AddStock(num)
	return g.GoodsRepo.Save(entity)
}

// MarkedTag 商品打tag
func (g *Goods) MarkedTag(goodsId int, tag string) (e error) {
	goodsEntity, e := g.GoodsRepo.Get(goodsId)
	if e != nil {
		return
	}
	e = goodsEntity.MarkedTag(tag)
	if e != nil {
		return
	}

	return g.GoodsRepo.Save(goodsEntity)
}

// Shop 购买商品
func (g *Goods) Shop(goodsId, goodsNum, userId int) (e error) {
	defer func() {
		if e != nil {
			g.Worker.Logger().Error("shop失败", e)
		} else {
			g.Worker.Logger().Info("shop 成功", goodsId, goodsNum, userId)
		}
	}()

	// cqrs 创建购买商品聚合根命令
	cmd := aggregate.NewShopGoodsCmd(g.UserRepo, g.OrderRepo, g.GoodsRepo, g.Transaction)
	if e = cmd.LoadEntity(userId, goodsId); e != nil {
		return e
	}

	e = cmd.Shop(goodsNum)
	return
}
