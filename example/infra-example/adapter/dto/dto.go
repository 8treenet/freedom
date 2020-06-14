package dto

type GoodsRep struct {
	Id    int
	Name  string // 商品名称
	Price int    // 价格
	Stock int    // 库存
}

type OrderRep struct {
	Id        int    //订单ID
	GoodsId   int    // 商品ID
	GoodsName string // 商品名称
	Num       int    // 数量
	DateTime  string // 购买时间
}
