package matcher

import (
	"net/http"
	"time"
)

type throttle struct {
	lastCall time.Time
	timeout  float64
}

func newThrottle(timeout float64) *throttle {
	return &throttle{
		// lastCall: time.Now(),
		timeout: timeout,
	}
}

func (t *throttle) Match(r *http.Request) bool {
	now := time.Now()
	if now.Sub(t.lastCall).Seconds()*1000 > t.timeout {
		t.lastCall = now
		return true
	}
	return false
}
