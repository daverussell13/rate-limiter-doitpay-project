package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/memory"
)

func setupRepoWithBucket(ctx context.Context, clientID string, tokens float64, lastRefill time.Time) *memory.TokenBucketRepository {
	repo := memory.NewTokenBucketRepository()
	bucket := ratelimit.TokenBucket{
		Tokens:     tokens,
		LastRefill: lastRefill,
	}
	repo.SaveBucket(ctx, clientID, bucket)
	return repo
}

func TestTokenBucketRepository_Get_NonExistingClient(t *testing.T) {
	repo := memory.NewTokenBucketRepository()
	clientID := "nonexistent"
	ctx := context.Background()

	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != 0 || !got.LastRefill.IsZero() {
		t.Errorf("expected empty bucket for non-existing client, got %+v", got)
	}
}

func TestTokenBucketRepository_Get_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client1"
	repo := setupRepoWithBucket(ctx, clientID, 3.5, time.Now())

	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != 3.5 {
		t.Errorf("expected Tokens=3.5, got %f", got.Tokens)
	}
	if got.LastRefill.IsZero() {
		t.Errorf("expected LastRefill to be set, got zero value")
	}
}

func TestTokenBucketRepository_Save_NewClient(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewTokenBucketRepository()
	clientID := "client2"
	bucket := ratelimit.TokenBucket{
		Tokens:     2.0,
		LastRefill: time.Now(),
	}

	if err := repo.SaveBucket(ctx, clientID, bucket); err != nil {
		t.Fatalf("unexpected error saving bucket: %v", err)
	}

	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != 2.0 {
		t.Errorf("expected Tokens=2.0, got %f", got.Tokens)
	}
}

func TestTokenBucketRepository_Save_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client3"
	repo := setupRepoWithBucket(ctx, clientID, 1.0, time.Now())

	bucket2 := ratelimit.TokenBucket{
		Tokens:     5.5,
		LastRefill: time.Now().Add(time.Minute),
	}
	if err := repo.SaveBucket(ctx, clientID, bucket2); err != nil {
		t.Fatalf("unexpected error saving bucket: %v", err)
	}

	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != 5.5 {
		t.Errorf("expected Tokens=5.5, got %f", got.Tokens)
	}
	if !got.LastRefill.After(time.Now()) {
		t.Errorf("expected LastRefill updated, got %v", got.LastRefill)
	}
}
