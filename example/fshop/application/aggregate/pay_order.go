package aggregate

import (
	"errors"

	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/application/dto"
	"github.com/8treenet/freedom/example/fshop/application/entity"
	"github.com/8treenet/freedom/infra/transaction"
)

// NewOrderPayCmd 订单支付聚合根，传入相关仓库的接口
func NewOrderPayCmd(userRepo repository.UserRepo, orderRepo repository.OrderRepo, tx transaction.Transaction) *OrderPayCmd {
	return &OrderPayCmd{
		userRepo:  userRepo,
		orderRepo: orderRepo,
		tx:        tx,
	}
}

// 支付订单聚合根
type OrderPayCmd struct {
	entity.Order
	userRepo   repository.UserRepo
	orderRepo  repository.OrderRepo
	tx         transaction.Transaction
	userEntity *entity.User
}

// LoadEntity 加载依赖实体
func (cmd *OrderPayCmd) LoadEntity(orderNo string, userId int) error {
	orderEntity, err := cmd.orderRepo.Find(orderNo, userId)
	if err != nil {
		return err
	}
	cmd.Order = *orderEntity

	cmd.userEntity, err = cmd.userRepo.Get(cmd.UserId)
	if err != nil {
		return err
	}
	return nil
}

// Pay 支付.
func (cmd *OrderPayCmd) Pay() error {
	if cmd.Status != entity.OrderStatusNonPayment {
		return errors.New("未知错误")
	}
	if cmd.userEntity.Money < cmd.TotalPrice {
		return errors.New("余额不足")
	}
	cmd.userEntity.AddMoney(-cmd.TotalPrice)

	cmd.Order.Pay()

	//事务执行 修改订单状态、扣用户金币
	e := cmd.tx.Execute(func() error {
		if e := cmd.orderRepo.Save(&cmd.Order); e != nil {
			return e
		}

		return cmd.userRepo.Save(cmd.userEntity)
	})

	if e == nil {
		msg := dto.OrderPayMsg{
			OrderNo:    cmd.OrderNo,
			TotalPrice: cmd.TotalPrice,
		}
		//发布领域事件 订单支付, 需要配置 server/conf/infra/kafka.toml 生产者相关配置
		cmd.DomainEvent("order-pay", msg)
	}
	return e
}
