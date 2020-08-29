package dto

// UserInfoRes .
type UserInfoRes struct {
	ID    int    //用户id
	Name  string //用户名
	Money int    //用户金钱
}

// GoodsItemRes .
type GoodsItemRes struct {
	ID    int    //商品id
	Name  string //商品名称
	Price int    //商品价格
	Stock int    //商品库存
	Tag   string //商品tag
}

// CartItemRes .
type CartItemRes struct {
	TotalPrice int //总价
	Items      []struct {
		ID         int    //购物车项ID
		GoodsID    int    //商品id
		GoodsName  string //商品名称
		GoodsNum   int    //商品数量
		TotalPrice int    //商品价格
	}
}

// OrderItemRes .
type OrderItemRes struct {
	OrderNo    string
	TotalPrice int
	Status     string
	GoodsItems []struct {
		GoodsID   int    // 商品id
		Num       int    // 数量
		GoodsName string // 商品名称
	}
}

// OrderPayMsg .
type OrderPayMsg struct {
	OrderNo    string
	TotalPrice int
}
