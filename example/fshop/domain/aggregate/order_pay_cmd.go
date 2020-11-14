package aggregate

import (
	"errors"

	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/example/fshop/domain/entity"
	"github.com/8treenet/freedom/example/fshop/infra/domainevent"
)

//OrderPayCmd 支付订单聚合根
type OrderPayCmd struct {
	entity.Order
	userEntity *entity.User

	userRepo  dependency.UserRepo
	orderRepo dependency.OrderRepo
	tx        *domainevent.EventTransaction
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
	return e
}
