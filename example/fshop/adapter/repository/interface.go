package repository

import (
	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/example/fshop/application/entity"
)

type UserRepo interface {
	Get(id int) (obj *entity.User, e error)
	FindByName(userName string) (userEntity *entity.User, e error)
	Save(entity *entity.User) error
	New(userDto dto.RegisterUserReq, money int) (entityUser *entity.User, e error)
}

type CartRepo interface {
	FindAll(userId int) (entitys []*entity.Cart, e error)
	FindByGoodsId(userId, goodsId int) (cartEntity *entity.Cart, e error)
	Save(entity *entity.Cart) error
	New(userId, goodsId, num int) (cartEntity *entity.Cart, e error)
	DeleteAll(userId int) (e error)
}

type GoodsRepo interface {
	Get(id int) (goodsEntity *entity.Goods, e error)
	Finds(ids []int) (entitys []*entity.Goods, e error)
	FindsByPage(page, pageSize int, tag string) (entitys []*entity.Goods, e error)
	Save(entity *entity.Goods) error
	New(name, tag string, price, stock int) (entityGoods *entity.Goods, e error)
}

type OrderRepo interface {
	New() (orderEntity *entity.Order, e error)
	Save(orderEntity *entity.Order) (e error)
	Find(orderNO string, userId int) (orderEntity *entity.Order, e error)
	Get(orderNO string) (orderEntity *entity.Order, e error)
	Finds(userId int, page, pageSize int) (entitys []*entity.Order, totalPage int, e error)
}

type DeliveryRepo interface {
	New() (*entity.Delivery, error)
	Save(*entity.Delivery) error
}

type AdminRepo interface {
	Get(int) (*entity.Admin, error)
}
