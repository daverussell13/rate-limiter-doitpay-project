package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/config"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/service"
)

type mockFixedWindowRepo struct {
	storage map[string]ratelimit.Window
}

func newMockRepo() *mockFixedWindowRepo {
	return &mockFixedWindowRepo{
		storage: make(map[string]ratelimit.Window),
	}
}

func (m *mockFixedWindowRepo) GetWindow(ctx context.Context, clientID string) (ratelimit.Window, error) {
	w, ok := m.storage[clientID]
	if !ok {
		return ratelimit.Window{}, nil
	}
	return w, nil
}

func (m *mockFixedWindowRepo) SaveWindow(ctx context.Context, clientID string, window ratelimit.Window) error {
	m.storage[clientID] = window
	return nil
}

func TestFixedWindowService_Allow(t *testing.T) {
	repo := newMockRepo()
	cfg := config.FixedWindow{MaxRequests: 2, TimeFrameMs: 1000}
	svc := service.NewFixedWindowService(repo, cfg)
	clientID := "client1"

	ctx := context.Background()

	allowed, err := svc.Allow(ctx, clientID)
	if err != nil || !allowed {
		t.Error("expected first request to be allowed")
	}

	allowed, err = svc.Allow(ctx, clientID)
	if err != nil || !allowed {
		t.Error("expected second request to be allowed")
	}
}

func TestFixedWindowService_Allow_HitMaxRequests(t *testing.T) {
	repo := newMockRepo()
	cfg := config.FixedWindow{MaxRequests: 1, TimeFrameMs: 1000}
	svc := service.NewFixedWindowService(repo, cfg)
	clientID := "client2"

	ctx := context.Background()

	allowed, err := svc.Allow(ctx, clientID)
	if err != nil || !allowed {
		t.Error("expected first request to be allowed")
	}

	allowed, err = svc.Allow(ctx, clientID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected second request to be denied")
	}
}

func TestFixedWindowService_Allow_WindowExpired(t *testing.T) {
	repo := newMockRepo()
	cfg := config.FixedWindow{MaxRequests: 1, TimeFrameMs: 1000}
	svc := service.NewFixedWindowService(repo, cfg)
	clientID := "client3"

	ctx := context.Background()

	allowed, err := svc.Allow(ctx, clientID)
	if err != nil || !allowed {
		t.Error("expected first request to be allowed")
	}

	win := repo.storage[clientID]
	win.EndTime = time.Now().Add(-time.Millisecond)
	repo.storage[clientID] = win

	allowed, err = svc.Allow(ctx, clientID)
	if err != nil || !allowed {
		t.Error("expected request to be allowed after window expired")
	}
}
