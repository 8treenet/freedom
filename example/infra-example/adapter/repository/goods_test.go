package repository

import (
	"testing"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/infra-example/server/conf"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func TestGoodsRepository_Get(t *testing.T) {
	//创建单元测试工具
	unitTest := freedom.NewUnitTest()
	unitTest.InstallGorm(func() (db *gorm.DB) {
		//这是连接数据库方式，mock方式参见TestGoodsService_MockGet
		conf := conf.Get().DB
		var e error
		db, e = gorm.Open("mysql", conf.Addr)
		if e != nil {
			freedom.Logger().Fatal(e.Error())
		}

		db.DB().SetMaxIdleConns(conf.MaxIdleConns)
		db.DB().SetMaxOpenConns(conf.MaxOpenConns)
		db.DB().SetConnMaxLifetime(time.Duration(conf.ConnMaxLifeTime) * time.Second)
		db = db.Debug()
		return
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
