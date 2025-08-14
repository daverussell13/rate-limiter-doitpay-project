package service

import (
	"context"
	"time"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/config"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain/ratelimit"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/util"
)

type FixedWindowRepository interface {
	GetWindow(ctx context.Context, clientID string) (ratelimit.Window, error)
	SaveWindow(ctx context.Context, clientID string, window ratelimit.Window) error
}

type FixedWindowService struct {
	repo  FixedWindowRepository
	cfg   config.FixedWindow
	locks *util.StripedMutex
}

func NewFixedWindowService(repo FixedWindowRepository, cfg config.FixedWindow) *FixedWindowService {
	return &FixedWindowService{
		repo:  repo,
		cfg:   cfg,
		locks: util.NewStripedMutex(256),
	}
}

func (s *FixedWindowService) Allow(ctx context.Context, clientID string) (bool, error) {
	unlock := s.locks.Lock(clientID)
	defer unlock()

	window, err := s.repo.GetWindow(ctx, clientID)
	if err != nil {
		return false, err
	}

	now := time.Now()

	if window.EndTime.IsZero() || now.After(window.EndTime) {
		newWindow := ratelimit.Window{
			Count:   1,
			EndTime: now.Add(time.Duration(s.cfg.TimeFrameMs) * time.Millisecond),
		}
		if err := s.repo.SaveWindow(ctx, clientID, newWindow); err != nil {
			return false, err
		}
		return true, nil
	}

	if window.Count < s.cfg.MaxRequests {
		window.Count++
		if err := s.repo.SaveWindow(ctx, clientID, window); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
