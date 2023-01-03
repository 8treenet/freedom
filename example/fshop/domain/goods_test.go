package domain_test

import (
	"testing"

	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/fshop/adapter/repository" //引入输出适配器 repository资源库。不引入会报错！！！！！！！！！！！！！！
	"github.com/8treenet/freedom/example/fshop/domain"
	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func getUnitTest() freedom.UnitTest {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} {
		db, e := gorm.Open(mysql.Open("root:123123@tcp(127.0.0.1:3306)/fshop?charset=utf8&parseTime=True&loc=Local"))
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}
		db = db.Debug()
		return db
	})

	opt := &redis.Options{
		Addr: "127.0.0.1:6379",
	}
	redisClient := redis.NewClient(opt)
	if e := redisClient.Ping().Err(); e != nil {
		freedom.Logger().Fatal(e.Error())
	}
	unitTest.InstallRedis(func() (client redis.Cmdable) {
		return redisClient
	})

	return unitTest
}

// TestGoodsServiceNew 创建商品
func TestGoodsServiceNew(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()

	var srv *domain.GoodsService
	//获取领域服务
	unitTest.FetchService(&srv)
	t.Log(srv.New("freedom-test", 50))
}

// TestGoodsServiceAddStock 增加库存
func TestGoodsServiceAddStock(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()

	var srv *domain.GoodsService
	//获取领域服务
	unitTest.FetchService(&srv)
	t.Log(srv.AddStock(1, 100))
}

// TestGoodsServiceShop 购买
func TestGoodsServiceShop(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()

	var srv *domain.GoodsService
	//获取领域服务
	unitTest.FetchService(&srv)
	t.Log(srv.Shop(1, 1, 1))
}
