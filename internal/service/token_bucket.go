package service

import (
	"context"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/config"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/util"
)

type TokenBucketRepository interface {
	GetBucket(ctx context.Context, clientID string) (ratelimit.TokenBucket, error)
	SaveBucket(ctx context.Context, clientID string, bucket ratelimit.TokenBucket) error
}

type TokenBucketService struct {
	repo  TokenBucketRepository
	cfg   config.TokenBucket
	locks *util.StripedMutex
}

func NewTokenBucketService(repo TokenBucketRepository, cfg config.TokenBucket) *TokenBucketService {
	return &TokenBucketService{
		repo:  repo,
		cfg:   cfg,
		locks: util.NewStripedMutex(256),
	}
}

func (s *TokenBucketService) Allow(ctx context.Context, clientID string) (bool, error) {
	unlock := s.locks.Lock(clientID)
	defer unlock()

	bucket, err := s.repo.GetBucket(ctx, clientID)
	if err != nil {
		return false, err
	}

	now := time.Now()

	if bucket.LastRefill.IsZero() {
		bucket.LastRefill = now
		bucket.Tokens = s.cfg.MaxTokens
	}

	elapsed := now.Sub(bucket.LastRefill).Seconds()
	if elapsed > 0 {
		bucket.Tokens += elapsed * s.cfg.RefillRate
		if bucket.Tokens > s.cfg.MaxTokens {
			bucket.Tokens = s.cfg.MaxTokens
		}
		bucket.LastRefill = now
	}

	allowed := false
	if bucket.Tokens >= 1.0 {
		bucket.Tokens -= 1.0
		allowed = true
	}

	if err := s.repo.SaveBucket(ctx, clientID, bucket); err != nil {
		return false, err
	}
	return allowed, nil
}
