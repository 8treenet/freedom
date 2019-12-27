package repositorys

import "github.com/8treenet/freedom/example/http2/models"

type GoodsInterface interface {
	GetGoods(goodsID int) models.GoodsModel
}
