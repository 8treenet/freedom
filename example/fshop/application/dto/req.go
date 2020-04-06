package dto

type CartAddReq struct {
	UserId   int `validate:"required"` //用户id
	GoodsId  int `validate:"required"` //商品id
	GoodsNum int `validate:"required"` //商品数
}

type CartShopReq struct {
	UserId int `validate:"required"` //用户id
}

type RegisterUserReq struct {
	Name     string `validate:"required"` //用户名称
	Password string `validate:"required"` //用户密码
}

type ChangePasswordReq struct {
	Id          int
	NewPassword string `validate:"required"`
	OldPassword string `validate:"required"`
}

type GoodsAddReq struct {
	Name  string `validate:"required"`
	Price int    `validate:"min=10,max=100000"` //最小价格10，最大价格100000
}

type GoodsTagReq struct {
	Id  int    `validate:"required"`
	Tag string `validate:"oneof=HOT NEW NONE"` //要设置的标签必须是 热门，新品，默认
}

type GoodsShopReq struct {
	UserId int `validate:"required"` //用户id
	Id     int `validate:"required"` //商品id
	Num    int `validate:"required"` //商品数量
}

type OrderPayReq struct {
	UserId  int    `validate:"required"` //用户id
	OrderNo string `validate:"required"` //订单id
}
