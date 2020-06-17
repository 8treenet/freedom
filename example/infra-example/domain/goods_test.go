package domain

import (
	"testing"

	"github.com/8treenet/freedom/example/infra-example/adapter/po"

	"github.com/8treenet/freedom"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func TestGoodsService_Get(t *testing.T) {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallGorm(func() (db *gorm.DB) {
		//这是连接数据库方式，mock方式参见TestGoodsService_MockGet
		var e error
		db, e = gorm.Open("mysql", "root:123123@tcp(127.0.0.1:3306)/freedom?charset=utf8&parseTime=True&loc=Local")
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}
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

func (repo *MockGoodsRepository) Get(id int) (result po.Goods, e error) {
	result.Id = 123
	result.Name = "mock商品名称"
	result.Price = 100
	result.Stock = 30
	result.Name = "mock商品名称"
	return
}

func (repo *MockGoodsRepository) GetAll() (goods []po.Goods, e error) { return }
func (repo *MockGoodsRepository) Save(*po.Goods) error                { return nil }
