package middleware

import (
	"strings"

	dt "DT-Go"
)

// NewBusFilter .
func NewBusFilter() func(dt.Worker) {
	return func(run dt.Worker) {
		bus := run.Bus()
		for key := range bus.Header {
			if strings.Index(key, "x-") == 0 || strings.Index(key, "X-") == 0 {
				continue
			}
			bus.Del(key)
		}
	}
}
