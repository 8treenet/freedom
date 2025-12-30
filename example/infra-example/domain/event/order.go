package event

// OrderPay 订单支付事件-普通消息示例
type OrderPay struct {
	OrderID int `json:"orderId"`
	UserID  int `json:"userId"`
}

// Topic .
func (op *OrderPay) Topic() string {
	return "OrderPay"
}
