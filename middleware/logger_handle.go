package middleware

import (
	"fmt"
	// "sort"

	"Hive"
)

// DefaultLogRowHandle .
func DefaultLogRowHandle(value *hive.LogRow) bool {
	//logRow中间件，每一行日志都会触发回调。如果返回true，将停止中间件遍历回调。
	fieldMap := map[string]interface{}{
		"status":  nil,
		"method":  nil,
		"path":    nil,
		"ip":      nil,
		"latency": nil,
	}
	fieldKeys := []string{"status", "method", "path", "ip", "latency"}
	for k := range value.Fields {
		if _, ok := fieldMap[k]; !ok {
			fieldKeys = append(fieldKeys, k)
		}
	}
	// sort.Strings(fieldKeys)
	for i := 0; i < len(fieldKeys); i++ {
		fieldMsg := value.Fields[fieldKeys[i]]
		if fieldMsg == nil {
			continue
		}
		if fieldKeys[i] == "x-request-id" {
			value.Message = fmt.Sprintf("[%v] %v", fieldMsg, value.Message)
			continue
		}
		if value.Message != "" {
			value.Message += " "
		}
		if _, ok := fieldMap[fieldKeys[i]]; ok {
			if fieldKeys[i] == "ip" {
				value.Message += fmt.Sprintf("(%v)", fieldMsg)
			} else {
				value.Message += fmt.Sprintf("%v", fieldMsg)
			}
		} else {
			value.Message += fmt.Sprintf("%s:%v", fieldKeys[i], fieldMsg)
		}
	}
	return false

	/*
		logrus.WithFields(value.Fields).Info(value.Message)
		return true
	*/
}
