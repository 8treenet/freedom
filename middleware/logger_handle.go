package middleware

import (
	"fmt"
	"path"
	"runtime"
	"sort"
	"strings"

	"github.com/8treenet/freedom"
	"github.com/kataras/golog"
)

// DefaultLogRowHandle The middleware output of the log line .
func DefaultLogRowHandle(value *freedom.LogRow) bool {
	//logRow中间件，每一行日志都会触发回调。如果返回true，将停止中间件遍历回调。
	callerField(value) //打印代码行号

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

func init() {
	levelMap = map[uint32]struct{}{}
	levelMap[uint32(golog.WarnLevel)] = struct{}{}
	levelMap[uint32(golog.DebugLevel)] = struct{}{}
	levelMap[uint32(golog.ErrorLevel)] = struct{}{}
	levelMap[uint32(golog.FatalLevel)] = struct{}{}
}

var levelMap map[uint32]struct{}

// callerField
func callerField(value *freedom.LogRow) {
	if len(value.Fields) != 0 {
		return
	}
	if _, ok := levelMap[uint32(value.Level)]; !ok {
		return
	}

	value.Fields = make(golog.Fields)
	_, file, line, _ := runtime.Caller(6)
	if strings.Contains(file, "pkg/mod/github.com") {
		return
	}
	value.Fields["caller"] = fmt.Sprintf("%s:%d", path.Base(file), line)
	return
}
