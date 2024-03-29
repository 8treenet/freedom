package repository

import (
	"time"

	"github.com/8treenet/freedom/infra/store"
	"gorm.io/gorm"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/8treenet/freedom/example/fshop/infra/domainevent"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建资源库函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindRepository(func() *GoodsRepository {
			return &GoodsRepository{} //创建Godds资源库
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.GoodsRepo = (*GoodsRepository)(nil)

// GoodsRepository .
type GoodsRepository struct {
	freedom.Repository
	Cache        store.EntityCache         //实体缓存组件
	EventManager *domainevent.EventManager //领域事件组件
}

// BeginRequest .
func (repo *GoodsRepository) BeginRequest(worker freedom.Worker) {
	repo.Repository.BeginRequest(worker)

	//设置缓存的持久化数据源,旁路缓存模型，如果缓存未有数据，将回调该函数。
	repo.Cache.SetSource(func(result freedom.Entity) error {
		return findGoods(repo, &result.(*entity.Goods).Goods)
	})
	//缓存30秒, 不设置默认5分钟
	repo.Cache.SetExpiration(30 * time.Second)
	//设置缓存前缀
	repo.Cache.SetPrefix("freedom")
}

// Get 通过ID 获取商品实体.
func (repo *GoodsRepository) Get(ID int) (goodsEntity *entity.Goods, e error) {
	goodsEntity = &entity.Goods{}
	goodsEntity.ID = ID
	//注入基础Entity
	repo.InjectBaseEntity(goodsEntity)

	//读取缓存
	return goodsEntity, repo.Cache.GetEntity(goodsEntity)
}

// Save 持久化实体.
func (repo *GoodsRepository) Save(entity *entity.Goods) error {
	_, e := saveGoods(repo, entity)
	if e != nil {
		return e
	}

	//清空缓存
	defer repo.Cache.Delete(entity)
	return repo.EventManager.Save(&repo.Repository, entity) //持久化实体的事件
}

// Finds .
func (repo *GoodsRepository) Finds(IDs []int) (entitys []*entity.Goods, e error) {
	var primarys []interface{}
	for i := 0; i < len(IDs); i++ {
		primarys = append(primarys, IDs[i])
	}
	list, e := findGoodsListByPrimarys(repo, primarys...)
	if e != nil {
		return
	}
	for _, v := range list {
		entitys = append(entitys, &entity.Goods{Goods: v})
	}

	//注入基础Entity
	repo.InjectBaseEntitys(entitys)
	return
}

// FindsByPage .
func (repo *GoodsRepository) FindsByPage(page, pageSize int, tag string) (entitys []*entity.Goods, e error) {
	pager := NewDescPager("id").SetPage(page, pageSize)
	list, e := findGoodsList(repo, po.Goods{Tag: tag}, pager)
	if e != nil {
		return
	}
	for _, v := range list {
		entitys = append(entitys, &entity.Goods{Goods: v})
	}

	repo.Worker().Logger().Info("FindsByPage", freedom.LogFields{
		"page":      page,
		"pageSize":  pageSize,
		"totalPage": pager.TotalPage(),
	})
	//注入基础Entity
	repo.InjectBaseEntitys(entitys)
	return
}

// New .
func (repo *GoodsRepository) New(name, tag string, price, stock int) (entityGoods *entity.Goods, e error) {
	goods := po.Goods{Name: name, Price: price, Stock: stock, Tag: tag, Created: time.Now(), Updated: time.Now()}

	_, e = createGoods(repo, &goods)
	if e != nil {
		return
	}
	entityGoods = &entity.Goods{Goods: goods}
	repo.InjectBaseEntity(entityGoods)
	return
}

func (repo *GoodsRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
