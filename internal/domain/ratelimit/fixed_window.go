package ratelimit

import "time"

type Window struct {
	Count   int
	EndTime time.Time
}
