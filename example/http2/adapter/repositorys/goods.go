package repositorys

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/http2/domain/dto"
	"github.com/jinzhu/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *GoodsRepository {
			return &GoodsRepository{}
		})
	})
}

// GoodsRepository .
type GoodsRepository struct {
	freedom.Repository
}

// GetGoods implment Goods interface
func (repo *GoodsRepository) GetGoods(goodsID int) (result dto.Goods) {
	repo.Worker.Logger().Info("我是GoodsRepository")
	repo.Worker.Bus().Add("x-sender-name", "GoodsRepository")
	//通过h2c request 访问本服务 /goods/:id
	addr := "http://127.0.0.1:8000/goods/" + strconv.Itoa(goodsID)
	repo.NewH2CRequest(addr).Get().ToJSON(&result)

	//开启go 并发,并且没有group wait。请求结束触发相关对象回收，会快于当前并发go的读取数据，所以使用DeferRecycle
	repo.Worker.DeferRecycle()
	go func() {
		var model dto.Goods
		repo.NewH2CRequest(addr).Get().ToJSON(&model)
		repo.NewHTTPRequest(addr, false).Get().ToJSON(&model)
	}()
	return result
}

// db .
func (repo *GoodsRepository) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	db = db.New()
	db.SetLogger(repo.Worker.Logger())
	return db
}

/*
	// xorm
	func (repo *Default) db() *xorm.Engine {
		var db *xorm.Engine
		if err := repo.FetchDB(&db); err != nil {
			panic(err)
		}
		return db
	}
	func main {
		app.InstallDB(func() interface{} {
			db, _ := xorm.NewEngine("mysql", "root:root@tcp(127.0.0.1:3306)/xorm?charset=utf8")
			return db
		})
	}
*/
