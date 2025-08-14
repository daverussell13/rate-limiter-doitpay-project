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

type redisErrorExample struct{}

func (e redisErrorExample) Error() string {
	return "some redis error"
}

func setupRepoWithWindow(ctx context.Context, clientID string, windowCount int, windowDuration time.Duration) (*rdb.FixedWindowRepository, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	repo := rdb.NewFixedWindowRepository(db)

	window := ratelimit.Window{
		Count:   windowCount,
		EndTime: time.Now().Add(windowDuration),
	}
	data, _ := json.Marshal(window)
	ttl := time.Until(window.EndTime)
	if ttl <= 0 {
		ttl = time.Millisecond
	}
	mock.ExpectSet(clientID, data, ttl).SetVal("OK")
	_ = repo.SaveWindow(ctx, clientID, window)

	return repo, mock
}

func TestFixedWindowRepository_Get_NonExistingClient(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	repo := rdb.NewFixedWindowRepository(db)
	clientID := "nonexistent"

	mock.ExpectGet(clientID).RedisNil()

	got, err := repo.GetWindow(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 0 || !got.EndTime.IsZero() {
		t.Errorf("expected empty window for non-existing client, got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestFixedWindowRepository_Get_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client1"
	repo, mock := setupRepoWithWindow(ctx, clientID, 3, time.Minute)

	data, _ := json.Marshal(ratelimit.Window{
		Count:   3,
		EndTime: time.Now().Add(time.Minute),
	})
	mock.ExpectGet(clientID).SetVal(string(data))

	got, err := repo.GetWindow(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 3 {
		t.Errorf("expected Count=3, got %d", got.Count)
	}
	if !got.EndTime.After(time.Now()) {
		t.Errorf("expected EndTime in the future, got %v", got.EndTime)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestFixedWindowRepository_Get_RedisError(t *testing.T) {
	ctx := context.Background()
	clientID := "client-redis-error"

	db, mock := redismock.NewClientMock()
	repo := rdb.NewFixedWindowRepository(db)

	mock.ExpectGet(clientID).SetErr(redisErrorExample{})

	_, err := repo.GetWindow(ctx, clientID)
	if err == nil {
		t.Errorf("expected Redis error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestFixedWindowRepository_Get_UnmarshalError(t *testing.T) {
	ctx := context.Background()
	clientID := "client-unmarshal"

	db, mock := redismock.NewClientMock()
	repo := rdb.NewFixedWindowRepository(db)

	// kasih string yang bukan JSON valid
	mock.ExpectGet(clientID).SetVal("not-a-json")

	_, err := repo.GetWindow(ctx, clientID)
	if err == nil {
		t.Errorf("expected JSON unmarshal error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestFixedWindowRepository_Save_NewClient(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	repo := rdb.NewFixedWindowRepository(db)
	clientID := "client2"
	window := ratelimit.Window{
		Count:   1,
		EndTime: time.Now().Add(time.Minute),
	}

	data, _ := json.Marshal(window)
	ttl := time.Until(window.EndTime)
	if ttl <= 0 {
		ttl = time.Millisecond
	}
	mock.ExpectSet(clientID, data, ttl).SetVal("OK")

	if err := repo.SaveWindow(ctx, clientID, window); err != nil {
		t.Fatalf("unexpected error saving window: %v", err)
	}

	mock.ExpectGet(clientID).SetVal(string(data))
	got, err := repo.GetWindow(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 1 {
		t.Errorf("expected Count=1, got %d", got.Count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestFixedWindowRepository_Save_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client3"
	repo, mock := setupRepoWithWindow(ctx, clientID, 1, time.Minute)

	window2 := ratelimit.Window{
		Count:   5,
		EndTime: time.Now().Add(2 * time.Minute),
	}
	data2, _ := json.Marshal(window2)
	ttl2 := time.Until(window2.EndTime)
	if ttl2 <= 0 {
		ttl2 = time.Millisecond
	}
	mock.ExpectSet(clientID, data2, ttl2).SetVal("OK")

	if err := repo.SaveWindow(ctx, clientID, window2); err != nil {
		t.Fatalf("unexpected error saving window: %v", err)
	}

	mock.ExpectGet(clientID).SetVal(string(data2))
	got, err := repo.GetWindow(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 5 {
		t.Errorf("expected Count=5, got %d", got.Count)
	}
	if !got.EndTime.After(time.Now().Add(time.Minute)) {
		t.Errorf("expected EndTime updated, got %v", got.EndTime)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestSaveWindow_TTLEmpty(t *testing.T) {
	ctx := context.Background()
	db, mock := redismock.NewClientMock()
	repo := rdb.NewFixedWindowRepository(db)

	window := ratelimit.Window{
		Count:   1,
		EndTime: time.Now().Add(-time.Minute),
	}

	data, _ := json.Marshal(window)
	mock.ExpectSet("client1", data, time.Millisecond).SetVal("OK")

	err := repo.SaveWindow(ctx, "client1", window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
