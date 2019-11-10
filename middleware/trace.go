package middleware

import (
	"hash/crc32"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/8treenet/freedom/general"

	"github.com/kataras/iris/context"
	"github.com/zheng-ji/goSnowFlake"
)

// NewTrace .
func NewTrace(traceIDName string) func(context.Context) {
	return func(ctx context.Context) {
		bus := general.GetBus(ctx)
		traceID, ok := bus.Get(traceIDName)
		if !ok || traceID == "" {
			traceID = uuid()
		}
		ctx.Values().Set(traceIDName, traceID)
		bus.Add(traceIDName, traceID)
		ctx.Next()
	}
}

// uuid .
func uuid() string {
	if snowFlakeWorker == nil {
		return ""
	}
	ts, _ := snowFlakeWorker.NextId()
	return strings.ToUpper(hostID + strconv.FormatInt(ts, 36))
}

var snowFlakeWorker *goSnowFlake.IdWorker
var hostID string

func init() {
	hostName, err := os.Hostname()
	if err != nil {
		return
	}
	hostID = strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(hostName))), 36)
	rand.Seed(time.Now().UnixNano())
	machineID := int64(1 + rand.Intn(950))
	snowFlakeWorker, _ = goSnowFlake.NewIdWorker(machineID)
}
