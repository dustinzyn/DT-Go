package internal

import "net/http"

// Bus message bus, using http header to pass through data
type Bus struct {
	Header http.Header
}

func newBus(head http.Header) *Bus {
	return &Bus{
		Header: head.Clone(),
	}
}

func (b *Bus) Add(key, value string) {
	b.Header.Add(key, value)
}

func (b *Bus) Get(key string) (string) {
	return b.Header.Get(key)
}

func (b *Bus) Set(key, value string) {
	b.Header.Set(key, value)
}

func (b *Bus) Del(key string) {
	b.Header.Del(key)
}

// BusHandler the middleware type of message bus.
type BusHandler func(Worker)

var busMiddlewares []BusHandler

// HandlerBusMiddleware call middleware
func HandlerBusMiddleware(worker Worker) {
	for i := 0; i < len(busMiddlewares); i++ {
		busMiddlewares[i](worker)
	}
}