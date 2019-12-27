package kafka

import (
	"math/rand"
	"sync/atomic"
	"time"
)

func newLimiter(maximum int32) *Limiter {
	return &Limiter{processing: 0, maximum: maximum}
}

type Limiter struct {
	processing int32
	maximum    int32
}

func (s *Limiter) Open(deltas ...int32) {
	var delta int32 = 1
	if len(deltas) > 0 {
		delta = deltas[0]
	}
	atomic.AddInt32(&s.processing, delta)
	for index := 0; index < 5; index++ {
		num := atomic.LoadInt32(&s.processing)
		if num > s.maximum {
			Sleep(index, 500*time.Millisecond, 1500*time.Millisecond)
			continue
		}
		break
	}
}

func (s *Limiter) Close(deltas ...int32) {
	var delta int32 = 1
	if len(deltas) > 0 {
		delta = deltas[0]
	}
	atomic.AddInt32(&s.processing, -delta)
}

func Sleep(retry int, minBackoff time.Duration, maxBackoff time.Duration) {
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
