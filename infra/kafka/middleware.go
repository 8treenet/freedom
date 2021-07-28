package kafka

var middlewares []ProducerHandler

// ProducerHandler The function declaration of the Kafka Producer middleware..
type ProducerHandler func(*Msg)

// InstallMiddleware Install the middleware..
// You can control the publishing of messages by installing middleware.
func InstallMiddleware(handle ...ProducerHandler) {
	middlewares = append(middlewares, handle...)
}
