package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/config"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain"
)

var validYAML = `
server:
  host: "localhost"
  port: 8080
rate-limiter:
  fixed-window:
    max-requests: 5
    time-frame-ms: 1000
`

var invalidYAML = `
wrongyaml
`

var invalidConfigTypeYAML = `
server:
  host: "localhost"
  port: "not-an-int"
rate-limiter:
  fixed-window:
    max-requests: 5
    time-frame-ms: 1000
`

func withTempConfig(t *testing.T, content string, fn func()) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(strings.TrimSpace(content)), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	fn()
}

func TestLoad_ValidConfig(t *testing.T) {
	withTempConfig(t, validYAML, func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("expected Load to succeed, got error: %v", err)
		}

		if cfg.Server.Host != "localhost" {
			t.Errorf("expected host=localhost, got %s", cfg.Server.Host)
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("expected port=8080, got %d", cfg.Server.Port)
		}
		if cfg.RateLimiter.FixedWindow.MaxRequests != 5 {
			t.Errorf("expected max-requests=5, got %d", cfg.RateLimiter.FixedWindow.MaxRequests)
		}
		if cfg.RateLimiter.FixedWindow.TimeFrameMs != 1000 {
			t.Errorf("expected time-frame-ms=1000, got %d", cfg.RateLimiter.FixedWindow.TimeFrameMs)
		}
	})
}

func TestLoad_InvalidConfig(t *testing.T) {
	withTempConfig(t, invalidYAML, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected Load to fail due to invalid YAML")
		}

		e, ok := err.(*domain.Error)
		if !ok {
			t.Fatalf("expected *domain.Error, got %T", err)
		}

		if e.Code() != domain.ErrUnknown {
			t.Errorf("expected error code ErrUnknown, got %v", e.Code())
		}
	})
}

func TestLoad_InvalidConfigType(t *testing.T) {
	withTempConfig(t, invalidConfigTypeYAML, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected Load to fail due to invalid YAML")
		}

		e, ok := err.(*domain.Error)
		if !ok {
			t.Fatalf("expected *domain.Error, got %T", err)
		}

		if e.Code() != domain.ErrUnknown {
			t.Errorf("expected error code ErrUnknown, got %v", e.Code())
		}
	})
}

func TestLoad_ConfigFileNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected Load to fail due to missing config file")
	}

	e, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected *domain.Error, got %T", err)
	}

	if e.Code() != domain.ErrNotFound {
		t.Errorf("expected error code ErrNotFound, got %v", e.Code())
	}
}
