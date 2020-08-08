package repository

import (
	"testing"
	"time"

	"github.com/8treenet/freedom"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func TestGoodsRepository_Get(t *testing.T) {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} {
		//这是连接数据库方式，mock方式参见TestGoodsService_MockGet
		db, e := gorm.Open("mysql", "root:123123@tcp(127.0.0.1:3306)/freedom?charset=utf8&parseTime=True&loc=Local")
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}
		db = db.Debug()
		return db
	})
	unitTest.Run()

	var repo *GoodsRepository
	unitTest.GetRepository(&repo)
	for i := 0; i < 30; i++ {
		if i < 5 {
			go func() {
				t.Log(repo.Get(2))
			}()
		}
		go func() {
			t.Log(repo.Get(1))
		}()
	}
	time.Sleep(1 * time.Second)
}
