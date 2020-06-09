package repositorys

import "github.com/8treenet/freedom/example/http2/domain/objects"

type GoodsInterface interface {
	GetGoods(goodsID int) objects.GoodsModel
}
