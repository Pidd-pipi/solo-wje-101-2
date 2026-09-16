package service

import (
	"testing"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
)

func TestNoteCreateInvalidRoast(t *testing.T) {
	svc := NewNoteService(repository.NewTastingNoteRepository(nil), newTestLogger())
	n := &model.TastingNote{CoffeeName: "测试", RoastLevel: "blue"}
	if _, err := svc.Create(1, n); err == nil {
		t.Error("expected error for invalid roast level")
	}
}
