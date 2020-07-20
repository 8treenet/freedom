package aggregate

import (
	"errors"

	"github.com/8treenet/freedom/example/fshop/adapter/repository"
	"github.com/8treenet/freedom/example/fshop/domain/dto"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/infra/transaction"
)

// 支付订单聚合根
type OrderPayCmd struct {
	entity.Order
	userEntity *entity.User

	userRepo  repository.UserRepo
	orderRepo repository.OrderRepo
	tx        transaction.Transaction
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
