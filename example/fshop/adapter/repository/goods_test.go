package repository

import (
	"context"
	"testing"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/redis/go-redis/v9"
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if e := redisClient.Ping(ctx).Err(); e != nil {
		freedom.Logger().Fatal(e.Error())
	}
	unitTest.InstallRedis(func() (client redis.Cmdable) {
		return redisClient
	})
	return unitTest
}

// TestGoodsEntity 商品实体单测
func TestGoodsEntity(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()

	var repo *GoodsRepository
	//获取资源库
	unitTest.FetchRepository(&repo)
	goodsEnity, err := repo.Get(1)
	if err != nil {
		t.Error(err)
		return
	}
	err = goodsEnity.MarkedTag(entity.GoodsNewTag)
	if err != nil {
		t.Error(err)
		return
	}
	err = repo.Save(goodsEnity)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("ok", goodsEnity)
}

// TestGoodsEntity 商品列表测试
func TestGoodssEntity(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()

	var repo *GoodsRepository
	//获取资源库
	unitTest.FetchRepository(&repo)
	goodsEnitys, err := repo.FindsByPage(1, 3, "")

	t.Log("ok", goodsEnitys, err)
}

// TestPager 商品列表分页测试
func TestPager(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()

	var repo *GoodsRepository
	//获取资源库
	unitTest.FetchRepository(&repo)
	pager := NewDescPager("id").SetPage(1, 10)
	list, e := findGoodsList(repo, po.Goods{}, pager)
	if e != nil {
		panic(e)
	}
	t.Log(pager.TotalPage())
	t.Log(list)

	tagPager := NewDescPager("id").SetPage(1, 10)
	list, e = findGoodsList(repo, po.Goods{Tag: "NEW"}, tagPager)
	if e != nil {
		panic(e)
	}
	t.Log(tagPager.TotalPage())
	t.Log(list)
}

// TestFind
func TestFind(t *testing.T) {
	//获取单测工具
	unitTest := getUnitTest()
	unitTest.Run()
	var repo *GoodsRepository
	//获取资源库
	unitTest.FetchRepository(&repo)
	var pObject po.Goods
	err := findGoods(repo, &pObject)
	if err != nil {
		panic(err)
	}
	t.Log("findGoods", pObject)
	list, err := findGoodsListByPrimarys(repo, 1, 2, 3)
	if err != nil || len(list) == 0 {
		panic(err)
	}
	t.Log("findGoodsListByPrimarys", list)

	pObject, err = findGoodsByWhere(repo, "price = ? and name = ?", []interface{}{1000, "iMac"})
	if err != nil {
		panic(err)
	}
	t.Log("findGoodsByWhere", pObject)

	list, err = findGoodsListByWhere(repo, "price > ? and tag != ?", []interface{}{99, "NEW"})
	if err != nil || len(list) == 0 {
		panic(err)
	}
	t.Log("findGoodsListByWhere", list)

	pObject, err = findGoodsByMap(repo, map[string]interface{}{"price": 1000, "name": "iMac"})
	if err != nil {
		panic(err)
	}
	t.Log("findGoodsByMap", pObject)

	list, err = findGoodsListByMap(repo, map[string]interface{}{"tag": "HOT"})
	if err != nil || len(list) == 0 {
		panic(err)
	}
	t.Log("findGoodsListByMap", list)
}
