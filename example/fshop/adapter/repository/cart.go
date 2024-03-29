package repository

import (
	"time"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"gorm.io/gorm"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建资源库函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindRepository(func() *CartRepository {
			return &CartRepository{} //创建Cart资源库
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.CartRepo = (*CartRepository)(nil)

// CartRepository .
type CartRepository struct {
	freedom.Repository
}

// FindAll 获取用户购物车实体
func (repo *CartRepository) FindAll(userID int) (entitys []*entity.Cart, e error) {
	list, e := findCartList(repo, po.Cart{UserID: userID})
	if e != nil {
		return
	}

	for _, obj := range list {
		entitys = append(entitys, &entity.Cart{Cart: obj})
	}
	//注入基础Entity
	repo.InjectBaseEntitys(entitys)
	return
}

// FindByGoodsID 获取用户某商品的购物车
func (repo *CartRepository) FindByGoodsID(userID, goodsID int) (cartEntity *entity.Cart, e error) {
	cartEntity = &entity.Cart{Cart: po.Cart{UserID: userID, GoodsID: goodsID}}
	e = findCart(repo, &cartEntity.Cart)
	if e != nil {
		return
	}
	repo.InjectBaseEntity(cartEntity)
	return
}

// Save 保存购物车
func (repo *CartRepository) Save(entity *entity.Cart) error {
	_, e := saveCart(repo, entity)
	return e
}

// New 新增购物车
func (repo *CartRepository) New(userID, goodsID, num int) (cartEntity *entity.Cart, e error) {
	cart := po.Cart{UserID: userID, GoodsID: goodsID, Num: num, Created: time.Now(), Updated: time.Now()}

	_, e = createCart(repo, &cart)
	if e != nil {
		return
	}
	cartEntity = &entity.Cart{Cart: cart}
	repo.InjectBaseEntity(cartEntity)
	return
}

// DeleteAll 删除全部购物车
func (repo *CartRepository) DeleteAll(userID int) (e error) {
	e = repo.db().Where(po.Cart{UserID: userID}).Delete(po.Cart{}).Error
	return
}

func (repo *CartRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	return db
}
