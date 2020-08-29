package aggregate

// ShopType .
type ShopType interface {
	//返回购买的类型 单独商品 或购物车
	GetType() int
	//如果是直接购买类型 返回商品id和数量
	GetDirectGoods() (int, int)
}

// ShopCmd .
type ShopCmd interface {
	Shop() error
}
