package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
	"github.com/wjecoffeetaste/wjecoffeetaste/internal/repository"
)

func newFavoriteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&model.User{}, &model.CoffeeBean{}, &model.BeanFavorite{}, &model.TastingNote{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestFavoriteIdempotentAndProfileReadback(t *testing.T) {
	db := newFavoriteTestDB(t)
	beanRepo := repository.NewCoffeeBeanRepository(db)
	noteRepo := repository.NewTastingNoteRepository(db)
	favRepo := repository.NewBeanFavoriteRepository(db)
	svc := NewFavoriteService(favRepo, beanRepo, noteRepo, newTestLogger())

	const uid uint = 1
	b1 := &model.CoffeeBean{Name: "耶加雪菲", Origin: "埃塞俄比亚", ProcessMethod: "washed", FlavorTags: `["柑橘","茉莉"]`}
	b2 := &model.CoffeeBean{Name: "曼特宁", Origin: "印度尼西亚", ProcessMethod: "natural", FlavorTags: `["黑巧克力"]`}
	if err := beanRepo.Create(b1); err != nil {
		t.Fatalf("create b1: %v", err)
	}
	if err := beanRepo.Create(b2); err != nil {
		t.Fatalf("create b2: %v", err)
	}
	if err := noteRepo.Create(&model.TastingNote{
		UserID: uid, CoffeeName: "耶加雪菲", RoastLevel: "light", FlavorTags: `["柑橘","花香"]`, OverallScore: 8.0,
	}); err != nil {
		t.Fatalf("create note: %v", err)
	}

	// First favorite creates a row.
	if _, err := svc.Favorite(uid, b1.ID); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	// Repeat favorite must keep exactly one relation.
	if _, err := svc.Favorite(uid, b1.ID); err != nil {
		t.Fatalf("repeat favorite: %v", err)
	}
	if total, _ := svc.CountByUser(uid); total != 1 {
		t.Fatalf("expected single favorite row, got %d", total)
	}

	pref, err := svc.Preference(uid)
	if err != nil {
		t.Fatalf("preference: %v", err)
	}
	if len(pref.Process) != 1 || pref.Process[0].Key != "washed" || pref.Process[0].Count != 1 {
		t.Fatalf("unexpected process preference: %+v", pref.Process)
	}
	if len(pref.Roast) != 1 || pref.Roast[0].Key != "light" {
		t.Fatalf("unexpected roast preference: %+v", pref.Roast)
	}
	// 柑橘 appears on both the favored bean and the note -> merged count 2.
	var citrus int
	for _, f := range pref.Flavor {
		if f.Key == "柑橘" {
			citrus = f.Count
		}
	}
	if citrus != 2 {
		t.Fatalf("expected merged 柑橘 count 2, got %d (%+v)", citrus, pref.Flavor)
	}

	// Cancel: list/count/preference must read back the new state.
	if _, err := svc.Unfavorite(uid, b1.ID); err != nil {
		t.Fatalf("unfavorite: %v", err)
	}
	if total, _ := svc.CountByUser(uid); total != 0 {
		t.Fatalf("expected 0 favorites after cancel, got %d", total)
	}
	pref2, _ := svc.Preference(uid)
	if len(pref2.Process) != 0 {
		t.Fatalf("expected no process preference after cancel, got %+v", pref2.Process)
	}
	// Canceling twice is idempotent and must not error.
	if _, err := svc.Unfavorite(uid, b1.ID); err != nil {
		t.Fatalf("second unfavorite should be idempotent, got %v", err)
	}
}

func TestFavoriteCleanupOnBeanDelete(t *testing.T) {
	db := newFavoriteTestDB(t)
	beanRepo := repository.NewCoffeeBeanRepository(db)
	noteRepo := repository.NewTastingNoteRepository(db)
	favRepo := repository.NewBeanFavoriteRepository(db)
	favSvc := NewFavoriteService(favRepo, beanRepo, noteRepo, newTestLogger())
	beanSvc := NewBeanService(beanRepo, favRepo, newTestLogger())

	const uid uint = 1
	b1 := &model.CoffeeBean{Name: "慧兰", ProcessMethod: "honey", FlavorTags: `["焦糖"]`}
	b2 := &model.CoffeeBean{Name: "蜜处理", ProcessMethod: "honey", FlavorTags: `["莓果"]`}
	_ = beanRepo.Create(b1)
	_ = beanRepo.Create(b2)
	if _, err := favSvc.Favorite(uid, b1.ID); err != nil {
		t.Fatalf("favorite b1: %v", err)
	}
	if _, err := favSvc.Favorite(uid, b2.ID); err != nil {
		t.Fatalf("favorite b2: %v", err)
	}

	// Admin de-listing removes the bean and its favorites in one transaction.
	if err := beanSvc.Delete(b1.ID); err != nil {
		t.Fatalf("delete bean: %v", err)
	}
	if total, _ := favSvc.CountByUser(uid); total != 1 {
		t.Fatalf("expected 1 remaining favorite after bean delete, got %d", total)
	}
	items, _ := favSvc.ListByUser(uid, 10)
	if len(items) != 1 || items[0].BeanID != b2.ID {
		t.Fatalf("expected only b2 to remain, got %+v", items)
	}
	if _, err := beanRepo.FindByID(b1.ID); err == nil {
		t.Fatal("expected de-listed bean to be gone")
	}
}
