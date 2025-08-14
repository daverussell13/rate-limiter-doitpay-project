package ratelimit

import "time"

type TokenBucket struct {
	Tokens     float64
	LastRefill time.Time
}
