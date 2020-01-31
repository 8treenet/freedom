package repositorys

import "github.com/8treenet/freedom/example/http2/application/objects"

type GoodsInterface interface {
	GetGoods(goodsID int) objects.GoodsModel
}
