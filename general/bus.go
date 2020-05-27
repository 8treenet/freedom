package general

import "net/http"

func newBus(head http.Header) *Bus {
	result := &Bus{
		Header: head.Clone(),
	}
	return result
}

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

// Set .
func (b *Bus) Del(key string) {
	b.Header.Del(key)
}

type BusHandler func(Worker)

var busMiddlewares []BusHandler

func HandleBusMiddleware(worker Worker) {
	for i := 0; i < len(busMiddlewares); i++ {
		busMiddlewares[i](worker)
	}
}
