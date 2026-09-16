package service

import (
	"testing"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
)

func newFavoriteServiceForTest() *FavoriteService {
	return NewFavoriteService(
		repository.NewBeanFavoriteRepository(nil),
		repository.NewCoffeeBeanRepository(nil),
		repository.NewTastingNoteRepository(nil),
		newTestLogger(),
	)
}

func TestDecorateAnonymousAllFalse(t *testing.T) {
	svc := newFavoriteServiceForTest()
	beans := []model.CoffeeBean{{ID: 1}, {ID: 2}}
	cards, err := svc.Decorate(beans, 0)
	if err != nil {
		t.Fatalf("decorate anonymous: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	for i, c := range cards {
		if c.IsFavored {
			t.Errorf("card %d should not be favored for anonymous viewer", i)
		}
	}
}

func TestSortPreferencesByCountThenLabel(t *testing.T) {
	items := []dto.PreferenceItem{
		{Label: "b", Count: 1},
		{Label: "a", Count: 2},
		{Label: "c", Count: 2},
	}
	sortPreferences(items)
	if items[0].Label != "a" || items[1].Label != "c" || items[2].Label != "b" {
		t.Fatalf("unexpected order: %v", items)
	}
}
