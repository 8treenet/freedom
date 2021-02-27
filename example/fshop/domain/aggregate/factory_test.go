package aggregate_test

import (
	"testing"

	"github.com/8treenet/freedom"
	_ "github.com/8treenet/freedom/example/fshop/adapter/repository" //引入输出适配器 repository资源库。不引入会报错！！！！！！！！！！！！！！
	"github.com/8treenet/freedom/example/fshop/domain/aggregate"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func getUnitTest() freedom.UnitTest {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} {
		db, e := gorm.Open("mysql", "root:123123@tcp(127.0.0.1:3306)/fshop?charset=utf8&parseTime=True&loc=Local")
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

// TestCartItemAggregate 测试购物车聚合根
func TestCartItemAggregate(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()

	var aggregate *aggregate.CartFactory
	//获取工厂
	unitTest.FetchFactory(&aggregate)
	cartItemQuery, err := aggregate.NewCartItemQuery(1)
	if err != nil {
		t.Error(err)
		return
	}
	//获取全部购物车价格
	totalPrice := cartItemQuery.AllItemTotalPrice()
	t.Log(totalPrice)
}
