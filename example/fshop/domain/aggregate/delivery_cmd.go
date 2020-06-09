package aggregate

import (
	"errors"

	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/infra/transaction"
)

// NewDeliveryCmd 订单发货聚合根，传入相关仓库的接口
func NewDeliveryCmd(adminRepo repository.AdminRepo, orderRepo repository.OrderRepo, deliveryRepo repository.DeliveryRepo, tx transaction.Transaction) *DeliveryCmd {
	return &DeliveryCmd{
		orderRepo:    orderRepo,
		adminRepo:    adminRepo,
		deliveryRepo: deliveryRepo,
		tx:           tx,
	}
}

// 订单发货聚合根
type DeliveryCmd struct {
	entity.Order
	adminEntity  *entity.Admin
	orderRepo    repository.OrderRepo
	adminRepo    repository.AdminRepo
	deliveryRepo repository.DeliveryRepo
	tx           transaction.Transaction
}

// LoadEntity 加载依赖实体
func (cmd *DeliveryCmd) LoadEntity(orderNo string, adminId int) error {
	orderEntity, err := cmd.orderRepo.Get(orderNo)
	if err != nil {
		return err
	}
	cmd.Order = *orderEntity

	cmd.adminEntity, err = cmd.adminRepo.Get(adminId)
	if err != nil {
		return err
	}
	return nil
}

// Run .
func (cmd *DeliveryCmd) Run(trackingNumber string) error {
	//调用订单父类 判断是否支付
	if !cmd.IsPay() {
		return errors.New("该订单未支付")
	}

	deliveryEntity, err := cmd.deliveryRepo.New()
	if err != nil {
		return err
	}

	//设置发货数据
	deliveryEntity.SetOrderNo(cmd.OrderNo)
	deliveryEntity.SetAdminId(cmd.adminEntity.Id)
	deliveryEntity.SetTrackingNumber(trackingNumber)

	//调用订单父类发货
	cmd.Shipment()

	return cmd.tx.Execute(func() error {
		if e := cmd.orderRepo.Save(&cmd.Order); e != nil {
			return e
		}

		return cmd.deliveryRepo.Save(deliveryEntity)
	})
}
