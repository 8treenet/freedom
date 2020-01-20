package objects

type GoodsRep struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`  // 商品名称
	Price int    `json:"price"` // 价格
	Stock int    `json:"stock"` // 库存
}

type OrderRep struct {
	ID        int    `json:"id"`        //订单ID
	GoodsID   int    `json:"goodsId"`   // 商品ID
	GoodsName string `json:"goodsName"` // 商品名称
	Num       int    `json:"num"`       // 数量
	DateTime  string `json:"datetime"`  // 购买时间
}
