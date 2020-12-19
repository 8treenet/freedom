package repositorys

import "github.com/8treenet/freedom/example/http2/domain/vo"

// GoodsInterface .
type GoodsInterface interface {
	GetGoods(goodsID int) vo.Goods
}
