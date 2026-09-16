package service

import (
	"log/slog"
	"os"
	"testing"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestFollowSelfRejected(t *testing.T) {
	svc := NewFollowService(repository.NewUserFollowRepository(nil), newTestLogger())
	if _, err := svc.Follow(1, 1); err == nil {
		t.Error("expected error when following self")
	}
}
