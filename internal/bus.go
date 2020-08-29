package internal

import "net/http"

func newBus(head http.Header) *Bus {
	result := &Bus{
		Header: head.Clone(),
	}
	return result
}

// Bus message bus, using http header to pass through data.
type Bus struct {
	http.Header
}

// Add .
func (b *Bus) Add(key, obj string) {
	b.Header.Add(key, obj)
}

// Get .
func (b *Bus) Get(key string) string {
	return b.Header.Get(key)
}

// Set .
func (b *Bus) Set(key, obj string) {
	b.Header.Set(key, obj)
}

// Del .
func (b *Bus) Del(key string) {
	b.Header.Del(key)
}

//BusHandler The middleware type of the message bus. .
type BusHandler func(Worker)

var busMiddlewares []BusHandler

// HandleBusMiddleware .
func HandleBusMiddleware(worker Worker) {
	for i := 0; i < len(busMiddlewares); i++ {
		busMiddlewares[i](worker)
	}
}
