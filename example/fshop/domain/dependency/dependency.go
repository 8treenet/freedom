package dependency

import (
	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
)

//依赖倒置的接口 外部adapter负责实现

// UserRepo .
type UserRepo interface {
	Get(ID int) (obj *entity.User, e error)
	FindByName(userName string) (userEntity *entity.User, e error)
	Save(entity *entity.User) error
	New(userDto dto.RegisterUserReq, money int) (entityUser *entity.User, e error)
}

// CartRepo .
type CartRepo interface {
	FindAll(userID int) (entitys []*entity.Cart, e error)
	FindByGoodsID(userID, goodsID int) (cartEntity *entity.Cart, e error)
	Save(entity *entity.Cart) error
	New(userID, goodsID, num int) (cartEntity *entity.Cart, e error)
	DeleteAll(userID int) (e error)
}

// GoodsRepo .
type GoodsRepo interface {
	Get(ID int) (goodsEntity *entity.Goods, e error)
	Finds(IDs []int) (entitys []*entity.Goods, e error)
	FindsByPage(page, pageSize int, tag string) (entitys []*entity.Goods, e error)
	Save(entity *entity.Goods) error
	New(name, tag string, price, stock int) (entityGoods *entity.Goods, e error)
}

// OrderRepo .
type OrderRepo interface {
	New() (orderEntity *entity.Order, e error)
	Save(orderEntity *entity.Order) (e error)
	Find(orderNO string, userID int) (orderEntity *entity.Order, e error)
	Get(orderNO string) (orderEntity *entity.Order, e error)
	Finds(userID int, page, pageSize int) (entitys []*entity.Order, totalPage int, e error)
}

// DeliveryRepo .
type DeliveryRepo interface {
	New() (*entity.Delivery, error)
	Save(*entity.Delivery) error
}

// AdminRepo .
type AdminRepo interface {
	Get(int) (*entity.Admin, error)
}
