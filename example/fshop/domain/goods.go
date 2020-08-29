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
		//绑定创建领域服务函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindService(func() *Goods {
			return &Goods{} //创建Goods领域服务
		})
		//控制器客户需要明确使用 InjectController
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
			ID:    entitys[i].ID,
			Name:  entitys[i].Name,
			Price: entitys[i].Price,
			Stock: entitys[i].Stock,
			Tag:   entitys[i].Tag,
		})
	}
	return
}

// AddStock 增加商品库存
func (g *Goods) AddStock(goodsID, num int) (e error) {
	entity, e := g.GoodsRepo.Get(goodsID)
	if e != nil {
		g.Worker.Logger().Error("商品库存失败", freedom.LogFields{"goodsId": goodsID, "num": num})
		return
	}

	g.Worker.Logger().Info("增加库存", freedom.LogFields{"goodsId": goodsID, "num": num})
	entity.AddStock(num)
	return g.GoodsRepo.Save(entity)
}

// MarkedTag 商品打tag
func (g *Goods) MarkedTag(goodsID int, tag string) (e error) {
	goodsEntity, e := g.GoodsRepo.Get(goodsID)
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
func (g *Goods) Shop(goodsID, goodsNum, userID int) (e error) {
	//使用抽象工厂 创建商品类型
	shopType := g.ShopFactory.NewGoodsShopType(goodsID, goodsNum)
	//使用抽象工厂 创建抽象聚合根
	cmd, e := g.ShopFactory.NewShopCmd(userID, shopType)
	if e != nil {
		return
	}
	return cmd.Shop()
}
