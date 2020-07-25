package domain

import (
	"github.com/8treenet/freedom/example/fshop/domain/aggregate"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/example/fshop/domain/entity"

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
	Worker      freedom.Worker         //运行时，一个请求绑定一个运行时
	GoodsRepo   dependency.GoodsRepo   //依赖倒置商品资源库
	ShopFactory *aggregate.ShopFactory //依赖注入购买聚合根工厂
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
		g.Worker.Logger().Error("商品库存失败")
		return
	}

	g.Worker.Logger().Info("增加库存")
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
	//使用抽象工厂 创建商品类型
	shopType := g.ShopFactory.NewGoodsShopType(goodsId, goodsNum)
	//使用抽象工厂 创建抽象聚合根
	cmd, e := g.ShopFactory.NewShopCmd(userId, shopType)
	if e != nil {
		return
	}
	return cmd.Shop()
}
