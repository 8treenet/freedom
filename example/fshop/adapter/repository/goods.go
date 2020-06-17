package repository

import (
	"time"

	"github.com/8treenet/freedom/infra/store"

	"github.com/8treenet/freedom/example/fshop/adapter/po"
	"github.com/8treenet/freedom/example/fshop/domain/entity"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *Goods {
			return &Goods{}
		})
	})
}

var _ GoodsRepo = new(Goods)

// Goods .
type Goods struct {
	freedom.Repository
	Cache store.EntityCache //实体缓存组件
}

// BeginRequest
func (repo *Goods) BeginRequest(worker freedom.Worker) {
	repo.Repository.BeginRequest(worker)

	//设置缓存的持久化数据源,旁路缓存模型，如果缓存未有数据，将回调该函数。
	repo.Cache.SetSource(func(result freedom.Entity) error {
		return findGoods(repo, result)
	})
	//缓存30秒, 不设置默认5分钟
	repo.Cache.SetExpiration(30 * time.Second)
	//设置缓存前缀
	repo.Cache.SetPrefix("freedom")
}

// Get 通过id 获取商品实体.
func (repo *Goods) Get(id int) (goodsEntity *entity.Goods, e error) {
	goodsEntity = &entity.Goods{}
	goodsEntity.Id = id
	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(goodsEntity)

	//读取缓存
	return goodsEntity, repo.Cache.GetEntity(goodsEntity)
}

// Save 持久化实体.
func (repo *Goods) Save(entity *entity.Goods) error {
	_, e := saveGoods(repo, &entity.Goods)
	//清空缓存
	repo.Cache.Delete(entity)
	return e
}

func (repo *Goods) Finds(ids []int) (entitys []*entity.Goods, e error) {
	var primarys []interface{}
	for i := 0; i < len(ids); i++ {
		primarys = append(primarys, ids[i])
	}
	e = findGoodsListByPrimarys(repo, &entitys, primarys...)
	if e != nil {
		return
	}

	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntitys(entitys)
	return
}

func (repo *Goods) FindsByPage(page, pageSize int, tag string) (entitys []*entity.Goods, e error) {
	build := repo.NewORMDescBuilder("id").NewPageBuilder(page, pageSize)
	e = findGoodsList(repo, po.Goods{Tag: tag}, &entitys, build)
	if e != nil {
		return
	}
	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntitys(entitys)
	return
}

func (repo *Goods) New(name, tag string, price, stock int) (entityGoods *entity.Goods, e error) {
	goods := po.Goods{Name: name, Price: price, Stock: stock, Tag: tag, Created: time.Now(), Updated: time.Now()}

	_, e = createGoods(repo, &goods)
	if e != nil {
		return
	}
	entityGoods = &entity.Goods{Goods: goods}
	repo.InjectBaseEntity(entityGoods)
	return
}
