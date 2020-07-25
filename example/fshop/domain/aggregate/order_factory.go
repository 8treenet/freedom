package aggregate

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain/dependency"
	"github.com/8treenet/freedom/infra/transaction"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindFactory(func() *OrderFactory {
			return &OrderFactory{}
		})
	})
}

// OrderFactory 订单聚合根工厂
type OrderFactory struct {
	UserRepo     dependency.UserRepo     //依赖倒置用户资源库
	OrderRepo    dependency.OrderRepo    //依赖倒置订单资源库
	AdminRepo    dependency.AdminRepo    //依赖倒置管理员资源库
	DeliveryRepo dependency.DeliveryRepo //依赖倒置物流资源库
	TX           transaction.Transaction //依赖倒置事务组件
	Worker       freedom.Worker          //运行时，一个请求绑定一个运行时
}

// NewOrderPayCmd 创建订单支付聚合根
func (factory *OrderFactory) NewOrderPayCmd(orderNo string, userId int) (*OrderPayCmd, error) {
	factory.Worker.Logger().Info("创建订单支付聚合根")
	orderEntity, err := factory.OrderRepo.Find(orderNo, userId)
	if err != nil {
		return nil, err
	}

	userEntity, err := factory.UserRepo.Get(userId)
	if err != nil {
		return nil, err
	}
	cmd := &OrderPayCmd{
		Order:      *orderEntity,
		userEntity: userEntity,
		userRepo:   factory.UserRepo,
		orderRepo:  factory.OrderRepo,
		tx:         factory.TX,
	}
	return cmd, nil
}

// NewOrderDeliveryCmd 创建订单发货聚合根
func (factory *OrderFactory) NewOrderDeliveryCmd(orderNo string, adminId int) (*DeliveryCmd, error) {
	//factory.Worker.Logger().Info("创建订单发货聚合根")
	orderEntity, err := factory.OrderRepo.Get(orderNo)
	if err != nil {
		return nil, err
	}
	adminEntity, err := factory.AdminRepo.Get(adminId)
	if err != nil {
		return nil, err
	}
	cmd := &DeliveryCmd{
		Order:       *orderEntity,
		adminEntity: adminEntity,

		orderRepo:    factory.OrderRepo,
		deliveryRepo: factory.DeliveryRepo,
		tx:           factory.TX,
	}
	return cmd, nil
}
