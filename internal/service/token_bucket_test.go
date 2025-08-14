package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/config"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/service"
)

// mock repository simple
type mockRepo struct {
	data map[string]ratelimit.TokenBucket
}

func newTokenBucketMockRepo() *mockRepo {
	return &mockRepo{data: make(map[string]ratelimit.TokenBucket)}
}

func (m *mockRepo) GetBucket(ctx context.Context, clientID string) (ratelimit.TokenBucket, error) {
	if b, ok := m.data[clientID]; ok {
		return b, nil
	}
	return ratelimit.TokenBucket{}, nil
}

func (m *mockRepo) SaveBucket(ctx context.Context, clientID string, bucket ratelimit.TokenBucket) error {
	m.data[clientID] = bucket
	return nil
}

func TestTokenBucketService_Allow(t *testing.T) {
	cfg := config.TokenBucket{
		MaxTokens:  5,
		RefillRate: 1,
	}

	repo := newTokenBucketMockRepo()
	svc := service.NewTokenBucketService(repo, cfg)
	ctx := context.Background()
	clientID := "test-client"

	allowed, err := svc.Allow(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected allowed=true on first request")
	}

	for i := 0; i < 4; i++ {
		allowed, _ := svc.Allow(ctx, clientID)
		if !allowed {
			t.Fatalf("expected allowed=true for token %d", i+2)
		}
	}

	allowed, _ = svc.Allow(ctx, clientID)
	if allowed {
		t.Fatalf("expected allowed=false when tokens exhausted")
	}

	bucket := repo.data[clientID]
	bucket.LastRefill = bucket.LastRefill.Add(-2 * time.Second)
	repo.data[clientID] = bucket

	allowed, _ = svc.Allow(ctx, clientID)
	if !allowed {
		t.Fatalf("expected allowed=true after refill")
	}
}
