package memory

import (
	"context"
	"sync"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
)

type TokenBucketRepository struct {
	mu   sync.RWMutex
	data map[string]ratelimit.TokenBucket
}

func NewTokenBucketRepository() *TokenBucketRepository {
	return &TokenBucketRepository{
		data: make(map[string]ratelimit.TokenBucket),
	}
}

func (r *TokenBucketRepository) GetBucket(ctx context.Context, clientID string) (ratelimit.TokenBucket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if b, ok := r.data[clientID]; ok {
		return b, nil
	}
	return ratelimit.TokenBucket{}, nil
}

func (r *TokenBucketRepository) SaveBucket(ctx context.Context, clientID string, bucket ratelimit.TokenBucket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[clientID] = bucket
	return nil
}
