package repository

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/domain/po"
	"github.com/jinzhu/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		//绑定创建资源库函数到框架，框架会根据客户的使用做依赖倒置和依赖注入的处理。
		initiator.BindRepository(func() *Delivery {
			return &Delivery{} //创建Delivery资源库
		})
	})
}

//实现领域模型内的依赖倒置
var _ dependency.DeliveryRepo = (*Delivery)(nil)

// Delivery .
type Delivery struct {
	freedom.Repository
}

// New 创建实体
func (repo *Delivery) New() (deliveryEntity *entity.Delivery, err error) {
	deliveryEntity = &entity.Delivery{Delivery: po.Delivery{Created: time.Now(), Updated: time.Now()}}
	//注入基础Entity 包含运行时和领域事件的producer
	repo.InjectBaseEntity(deliveryEntity)
	return
}

// Save 保存实体
func (repo *Delivery) Save(deliveryEntity *entity.Delivery) error {
	if deliveryEntity.ID == 0 {
		_, err := createDelivery(repo, &deliveryEntity.Delivery)
		return err
	}

	_, err := saveDelivery(repo, &deliveryEntity.Delivery)
	return err
}

func (repo *Delivery) db() *gorm.DB {
	var db *gorm.DB
	if err := repo.FetchDB(&db); err != nil {
		panic(err)
	}
	db = db.New()
	db.SetLogger(repo.Worker.Logger())
	return db
}
