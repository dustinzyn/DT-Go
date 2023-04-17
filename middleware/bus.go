package middleware

import (
	"strings"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"
)

// NewBusFilter .
func NewBusFilter() func(hive.Worker) {
	return func(run hive.Worker) {
		bus := run.Bus()
		for key := range bus.Header {
			if strings.Index(key, "x-") == 0 || strings.Index(key, "X-") == 0 {
				continue
			}
			bus.Del(key)
		}
	}
}
