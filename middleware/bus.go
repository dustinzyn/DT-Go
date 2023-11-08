package middleware

import (
	"strings"
)

// NewBusFilter .
func NewBusFilter() func(dhive.Worker) {
	return func(run dhive.Worker) {
		bus := run.Bus()
		for key := range bus.Header {
			if strings.Index(key, "x-") == 0 || strings.Index(key, "X-") == 0 {
				continue
			}
			bus.Del(key)
		}
	}
}
