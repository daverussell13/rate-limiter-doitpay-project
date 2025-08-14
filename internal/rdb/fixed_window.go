package rdb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/redis/go-redis/v9"
)

type FixedWindowRepository struct {
	client *redis.Client
}

func NewFixedWindowRepository(client *redis.Client) *FixedWindowRepository {
	return &FixedWindowRepository{
		client: client,
	}
}

func (r *FixedWindowRepository) GetWindow(ctx context.Context, clientID string) (ratelimit.Window, error) {
	val, err := r.client.Get(ctx, clientID).Result()
	if err != nil {
		if err == redis.Nil {
			return ratelimit.Window{}, nil
		}
		return ratelimit.Window{}, err
	}

	var w ratelimit.Window
	if err := json.Unmarshal([]byte(val), &w); err != nil {
		return ratelimit.Window{}, err
	}
	return w, nil
}

func (r *FixedWindowRepository) SaveWindow(ctx context.Context, clientID string, window ratelimit.Window) error {
	data, err := json.Marshal(window)
	if err != nil {
		return err
	}

	ttl := time.Until(window.EndTime)
	if ttl <= 0 {
		ttl = time.Millisecond
	}

	return r.client.Set(ctx, clientID, data, ttl).Err()
}
