package domain

import (
	"testing"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/domain/object"
	"github.com/8treenet/freedom/example/infra-example/infra/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func TestGoodsService_Get(t *testing.T) {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallGorm(func() (db *gorm.DB) {
		//这是连接数据库方式，mock方式参见TestGoodsService_MockGet
		conf := config.Get().DB
		var e error
		db, e = gorm.Open("mysql", conf.Addr)
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}

		db.DB().SetMaxIdleConns(conf.MaxIdleConns)
		db.DB().SetMaxOpenConns(conf.MaxOpenConns)
		db.DB().SetConnMaxLifetime(time.Duration(conf.ConnMaxLifeTime) * time.Second)
		return
	})

	//启动单元测试
	unitTest.Run()

	var srv *GoodsService
	//获取领域服务
	unitTest.GetService(&srv)
	//调用服务方法。 篇幅有限，没有具体逻辑 只是读取个数据
	rep, err := srv.Get(1)
	if err != nil {
		panic(rep)
	}
	t.Log(rep)
	//在这里做 case，如果不满足条件 触发panic
}

func TestGoodsService_MockGet(t *testing.T) {
	unitTest := freedom.NewUnitTest()
	unitTest.Run()

	var srv *GoodsService
	unitTest.GetService(&srv)
	srv.GoodsRepo = new(MockGoodsRepository)
	rep, err := srv.Get(1)
	if err != nil {
		panic(rep)
	}
	t.Log(rep)
}

type MockGoodsRepository struct {
}

func (repo *MockGoodsRepository) Get(id int) (result object.Goods, e error) {
	result.Id = 123
	result.Name = "mock商品名称"
	result.Price = 100
	result.Stock = 30
	result.Name = "mock商品名称"
	return
}

func (repo *MockGoodsRepository) GetAll() (goods []object.Goods, e error) { return }
func (repo *MockGoodsRepository) Save(*object.Goods) error                { return nil }
