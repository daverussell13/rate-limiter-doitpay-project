package rdb_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/rdb"
	"github.com/go-redis/redismock/v9"
)

func setupTokenBucketRepoWithBucket(ctx context.Context, clientID string, tokens float64, lastRefill time.Time, maxTokens, refillRate float64) (*rdb.TokenBucketRepository, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	repo := rdb.NewTokenBucketRepository(db, maxTokens, refillRate)

	bucket := ratelimit.TokenBucket{
		Tokens:     tokens,
		LastRefill: lastRefill,
	}
	data, _ := json.Marshal(bucket)
	refillTime := time.Duration(maxTokens/refillRate) * time.Second
	expectedTTL := (refillTime * 2) + (30 * time.Second)

	mock.ExpectSet(clientID, data, expectedTTL).SetVal("OK")
	_ = repo.SaveBucket(ctx, clientID, bucket)

	return repo, mock
}

func TestTokenBucketRepository_GetBucket_NonExistingClient(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	repo := rdb.NewTokenBucketRepository(db, 100.0, 10.0)
	clientID := "nonexistent"

	mock.ExpectGet(clientID).RedisNil()

	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != 0 || !got.LastRefill.IsZero() {
		t.Errorf("expected empty bucket for non-existing client, got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTokenBucketRepository_GetBucket_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client1"
	expectedTokens := 50.0
	expectedLastRefill := time.Now().Add(-time.Minute)

	repo, mock := setupTokenBucketRepoWithBucket(ctx, clientID, expectedTokens, expectedLastRefill, 100.0, 10.0)

	data, _ := json.Marshal(ratelimit.TokenBucket{
		Tokens:     expectedTokens,
		LastRefill: expectedLastRefill,
	})
	mock.ExpectGet(clientID).SetVal(string(data))

	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != expectedTokens {
		t.Errorf("expected Tokens=%.2f, got %.2f", expectedTokens, got.Tokens)
	}
	if !got.LastRefill.Equal(expectedLastRefill) {
		t.Errorf("expected LastRefill=%v, got %v", expectedLastRefill, got.LastRefill)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTokenBucketRepository_GetBucket_RedisError(t *testing.T) {
	ctx := context.Background()
	clientID := "client-redis-error"

	db, mock := redismock.NewClientMock()
	repo := rdb.NewTokenBucketRepository(db, 100.0, 10.0)

	mock.ExpectGet(clientID).SetErr(redisErrorExample{})

	_, err := repo.GetBucket(ctx, clientID)
	if err == nil {
		t.Errorf("expected Redis error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTokenBucketRepository_GetBucket_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	clientID := "client-unmarshal"

	db, mock := redismock.NewClientMock()
	repo := rdb.NewTokenBucketRepository(db, 100.0, 10.0)

	mock.ExpectGet(clientID).SetVal("not-a-json")

	_, err := repo.GetBucket(ctx, clientID)
	if err == nil {
		t.Errorf("expected JSON unmarshal error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTokenBucketRepository_SaveBucket_NewClient(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	maxTokens, refillRate := 100.0, 10.0
	repo := rdb.NewTokenBucketRepository(db, maxTokens, refillRate)
	clientID := "client2"

	bucket := ratelimit.TokenBucket{
		Tokens:     75.5,
		LastRefill: time.Now(),
	}

	data, _ := json.Marshal(bucket)

	refillTime := time.Duration(maxTokens/refillRate) * time.Second
	expectedTTL := (refillTime * 2) + (30 * time.Second)

	mock.ExpectSet(clientID, data, expectedTTL).SetVal("OK")

	if err := repo.SaveBucket(ctx, clientID, bucket); err != nil {
		t.Fatalf("unexpected error saving bucket: %v", err)
	}

	mock.ExpectGet(clientID).SetVal(string(data))
	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != 75.5 {
		t.Errorf("expected Tokens=75.5, got %.2f", got.Tokens)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTokenBucketRepository_SaveBucket_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client3"
	maxTokens, refillRate := 100.0, 10.0

	repo, mock := setupTokenBucketRepoWithBucket(ctx, clientID, 50.0, time.Now().Add(-time.Minute), maxTokens, refillRate)

	updatedBucket := ratelimit.TokenBucket{
		Tokens:     25.0,
		LastRefill: time.Now(),
	}

	data2, _ := json.Marshal(updatedBucket)

	refillTime := time.Duration(maxTokens/refillRate) * time.Second
	expectedTTL := (refillTime * 2) + (30 * time.Second)

	mock.ExpectSet(clientID, data2, expectedTTL).SetVal("OK")

	if err := repo.SaveBucket(ctx, clientID, updatedBucket); err != nil {
		t.Fatalf("unexpected error saving bucket: %v", err)
	}

	mock.ExpectGet(clientID).SetVal(string(data2))
	got, err := repo.GetBucket(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Tokens != 25.0 {
		t.Errorf("expected Tokens=25.0, got %.2f", got.Tokens)
	}
	if got.LastRefill.Before(time.Now().Add(-time.Second)) {
		t.Errorf("expected LastRefill updated, got %v", got.LastRefill)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTokenBucketRepository_SaveBucket_RedisSetError(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	repo := rdb.NewTokenBucketRepository(db, 100.0, 10.0)
	clientID := "client-set-error"

	bucket := ratelimit.TokenBucket{
		Tokens:     50.0,
		LastRefill: time.Now(),
	}

	data, _ := json.Marshal(bucket)

	refillTime := time.Duration(100.0/10.0) * time.Second
	expectedTTL := (refillTime * 2) + (30 * time.Second)

	mock.ExpectSet(clientID, data, expectedTTL).SetErr(redisErrorExample{})

	err := repo.SaveBucket(ctx, clientID, bucket)
	if err == nil {
		t.Errorf("expected Redis SET error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
