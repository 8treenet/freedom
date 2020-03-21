package dto

type UserInfoRes struct {
	Id    int    //用户id
	Name  string //用户名
	Money int    //用户金钱
}

type GoodsItemRes struct {
	Id    int    //商品id
	Name  string //商品名称
	Price int    //商品价格
	Stock int    //商品库存
	Tag   string //商品tag
}

type CartItemRes struct {
	TotalPrice int //总价
	Items      []struct {
		Id         int    //购物车项ID
		GoodsId    int    //商品id
		GoodsName  string //商品名称
		GoodsNum   int    //商品数量
		TotalPrice int    //商品价格
	}
}
