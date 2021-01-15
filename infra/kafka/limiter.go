package kafka

import (
	"math/rand"
	"runtime"
	"time"
)

func newLimiter() *LimiterImpl {
	return &LimiterImpl{
		progress: make(chan int8, runtime.NumCPU()*512),
	}
}

// LimiterImpl .
type LimiterImpl struct {
	maximum  int32
	progress chan int8
}

// SetChanSize .
func (l *LimiterImpl) SetChanSize(maximum int) {
	l.progress = make(chan int8, maximum)
}

func (l *LimiterImpl) add() {
	l.progress <- 1
}

func (l *LimiterImpl) sub() {
	_ = <-l.progress
}

func sleep(retry int, minBackoff time.Duration, maxBackoff time.Duration) {
	if retry < 0 {
		retry = 0
	}

	backoff := minBackoff << uint(retry)
	if backoff > maxBackoff || backoff < minBackoff {
		backoff = maxBackoff
	}

	if backoff == 0 {
		return
	}

	d := time.Duration(rand.Int63n(int64(backoff)))
	time.Sleep(d)
	return
}
