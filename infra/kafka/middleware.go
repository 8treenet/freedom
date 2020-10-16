package kafka

var middlewares []ProducerHandler

// ProducerHandler .
type ProducerHandler func(*Msg)

// InstallMiddleware .
func InstallMiddleware(handle ...ProducerHandler) {
	middlewares = append(middlewares, handle...)
}
