package repositorys

import "github.com/8treenet/freedom/example/http2/domain/dto"

// GoodsInterface .
type GoodsInterface interface {
	GetGoods(goodsID int) dto.Goods
}
