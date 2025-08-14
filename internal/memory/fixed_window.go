package memory

import (
	"context"
	"sync"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
)

type FixedWindowRepository struct {
	mu    sync.RWMutex
	store map[string]ratelimit.Window
}

func NewFixedWindowRepository() *FixedWindowRepository {
	return &FixedWindowRepository{
		store: make(map[string]ratelimit.Window),
	}
}

func (r *FixedWindowRepository) GetWindow(ctx context.Context, clientID string) (ratelimit.Window, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ws, ok := r.store[clientID]
	if !ok {
		return ratelimit.Window{}, nil
	}
	return ws, nil
}

func (r *FixedWindowRepository) SaveWindow(ctx context.Context, clientID string, state ratelimit.Window) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[clientID] = state
	return nil
}
