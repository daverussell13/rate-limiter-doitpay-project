package rdb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/redis/go-redis/v9"
)

type TokenBucketRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewTokenBucketRepository(client *redis.Client, maxTokens float64, refillRate float64) *TokenBucketRepository {
	refillTime := time.Duration(maxTokens/refillRate) * time.Second
	ttl := (refillTime * 2) + (30 * time.Second)

	return &TokenBucketRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *TokenBucketRepository) GetBucket(ctx context.Context, clientID string) (ratelimit.TokenBucket, error) {
	val, err := r.client.Get(ctx, clientID).Result()
	if err != nil {
		if err == redis.Nil {
			return ratelimit.TokenBucket{}, nil
		}
		return ratelimit.TokenBucket{}, err
	}

	var bucket ratelimit.TokenBucket
	if err := json.Unmarshal([]byte(val), &bucket); err != nil {
		return ratelimit.TokenBucket{}, err
	}

	return bucket, nil
}

func (r *TokenBucketRepository) SaveBucket(ctx context.Context, clientID string, bucket ratelimit.TokenBucket) error {
	data, err := json.Marshal(bucket)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, clientID, data, r.ttl).Err()
}
