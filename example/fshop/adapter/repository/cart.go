package repository

import (
	"time"

	"github.com/8treenet/freedom/example/fshop/application/entity"
	"github.com/8treenet/freedom/example/fshop/application/object"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *Cart {
			return &Cart{}
		})
	})
}

var _ CartRepo = new(Cart)

// Cart .
type Cart struct {
	freedom.Repository
}

// FindAll 获取用户购物车实体
func (repo *Cart) FindAll(userId int) (entitys []*entity.Cart, e error) {
	findCarts(repo, object.Cart{UserId: userId}, &entitys)
	e = findCarts(repo, object.Cart{UserId: userId}, &entitys)
	return
}

// FindByGoodsId 获取用户某商品的购物车
func (repo *Cart) FindByGoodsId(userId, goodsId int) (cartEntity *entity.Cart, e error) {
	cartEntity = &entity.Cart{}
	e = findCart(repo, object.Cart{UserId: userId, GoodsId: goodsId}, cartEntity)
	if e != nil {
		return
	}
	return
}

// Save 保存购物车
func (repo *Cart) Save(entity *entity.Cart) error {
	_, e := saveCart(repo, &entity.Cart)
	return e
}

// New 新增购物车
func (repo *Cart) New(userId, goodsId, num int) (cartEntity *entity.Cart, e error) {
	cart := object.Cart{UserId: userId, GoodsId: goodsId, Num: num, Created: time.Now(), Updated: time.Now()}

	_, e = createCart(repo, &cart)
	if e != nil {
		return
	}
	cartEntity = &entity.Cart{Cart: cart}
	repo.InjectBaseEntity(cartEntity)
	return
}

// DeleteAll 删除全部购物车
func (repo *Cart) DeleteAll(userId int) (e error) {
	e = repo.DB().Where(object.Cart{UserId: userId}).Delete(object.Cart{}).Error
	return
}
