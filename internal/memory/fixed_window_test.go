package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/memory"
)

func setupRepoWithWindow(ctx context.Context, clientID string, windowCount int, windowDuration time.Duration) *memory.FixedWindowRepository {
	repo := memory.NewFixedWindowRepository()
	window := ratelimit.Window{
		Count:   windowCount,
		EndTime: time.Now().Add(windowDuration),
	}
	repo.SaveWindow(ctx, clientID, window)
	return repo
}

func TestFixedWindowRepository_Get_NonExistingClient(t *testing.T) {
	repo := memory.NewFixedWindowRepository()
	clientID := "nonexistent"
	ctx := context.Background()

	got, err := repo.GetWindow(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 0 || !got.EndTime.IsZero() {
		t.Errorf("expected empty window for non-existing client, got %+v", got)
	}
}

func TestFixedWindowRepository_Get_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client1"
	repo := setupRepoWithWindow(ctx, clientID, 3, time.Minute)

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
}

func TestFixedWindowRepository_Save_NewClient(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewFixedWindowRepository()
	clientID := "client2"
	window := ratelimit.Window{
		Count:   1,
		EndTime: time.Now().Add(time.Minute),
	}

	if err := repo.SaveWindow(ctx, clientID, window); err != nil {
		t.Fatalf("unexpected error saving window: %v", err)
	}

	got, err := repo.GetWindow(ctx, clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 1 {
		t.Errorf("expected Count=1, got %d", got.Count)
	}
}

func TestFixedWindowRepository_Save_ExistingClient(t *testing.T) {
	ctx := context.Background()
	clientID := "client3"
	repo := setupRepoWithWindow(ctx, clientID, 1, time.Minute)

	window2 := ratelimit.Window{
		Count:   5,
		EndTime: time.Now().Add(2 * time.Minute),
	}
	if err := repo.SaveWindow(ctx, clientID, window2); err != nil {
		t.Fatalf("unexpected error saving window: %v", err)
	}

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
}
