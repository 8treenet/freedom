package general

import (
	"encoding/json"
)

func newBus(headerStr string) *Bus {
	// headerStr
	result := &Bus{
		m: make(map[string]interface{}),
	}
	json.Unmarshal([]byte(headerStr), &result.m)
	return result
}

type Bus struct {
	m map[string]interface{}
}

// Add .
func (b *Bus) Add(key string, obj interface{}) {
	b.m[key] = obj
}

// Get .
func (b *Bus) Get(key string) (obj interface{}, ok bool) {
	obj, ok = b.m[key]
	return
}

// ToJson .
func (b *Bus) ToJson() string {
	bys, _ := json.Marshal(b.m)
	return string(bys)
}

type BusHandler func(Runtime)

var busMiddlewares []BusHandler

func HandleBusMiddleware(rt Runtime) {
	for i := 0; i < len(busMiddlewares); i++ {
		busMiddlewares[i](rt)
	}
}
