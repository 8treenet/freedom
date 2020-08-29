package middleware

import (
	"fmt"
	"sort"

	"github.com/8treenet/freedom"
)

// DefaultLogRowHandle .
func DefaultLogRowHandle(value *freedom.LogRow) bool {
	//logRow中间件，每一行日志都会触发回调。如果返回true，将停止中间件遍历回调。
	fieldKeys := []string{}
	for k := range value.Fields {
		fieldKeys = append(fieldKeys, k)
	}
	sort.Strings(fieldKeys)
	for i := 0; i < len(fieldKeys); i++ {
		fieldMsg := value.Fields[fieldKeys[i]]
		if value.Message != "" {
			value.Message += "  "
		}
		value.Message += fmt.Sprintf("%s:%v", fieldKeys[i], fieldMsg)
	}
	return false

	/*
		logrus.WithFields(value.Fields).Info(value.Message)
		return true
	*/
	/*
		zapLogger, _ := zap.NewProduction()
		zapLogger.Info(value.Message)
		return true
	*/
}
