package vo

// CartAddReq .
type CartAddReq struct {
	UserID   int `validate:"required"` //用户id
	GoodsID  int `validate:"required"` //商品id
	GoodsNum int `validate:"required"` //商品数
}

// CartShopReq .
type CartShopReq struct {
	UserID int `validate:"required"` //用户id
}

// RegisterUserReq .
type RegisterUserReq struct {
	Name     string `validate:"required"` //用户名称
	Password string `validate:"required"` //用户密码
}

// ChangePasswordReq .
type ChangePasswordReq struct {
	ID          int
	NewPassword string `validate:"required"`
	OldPassword string `validate:"required"`
}

// GoodsAddReq .
type GoodsAddReq struct {
	Name  string `validate:"required"`
	Price int    `validate:"min=10,max=100000"` //最小价格10，最大价格100000
}

// GoodsTagReq .
type GoodsTagReq struct {
	ID  int    `validate:"required"`
	Tag string `validate:"oneof=HOT NEW NONE"` //要设置的标签必须是 热门，新品，默认
}

// GoodsShopReq .
type GoodsShopReq struct {
	UserID int `validate:"required"` //用户id
	ID     int `validate:"required"` //商品id
	Num    int `validate:"required"` //商品数量
}

// OrderPayReq .
type OrderPayReq struct {
	UserID  int    `validate:"required"` //用户id
	OrderNo string `validate:"required"` //订单id
}

// DeliveryReq .
type DeliveryReq struct {
	OrderNo        string `validate:"required"` //订单id
	TrackingNumber string `validate:"required"` //快递号
	AdminID        int    `validate:"required"` //管理员id
}
