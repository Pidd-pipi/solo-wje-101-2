package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/dto"
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

func TestDecorateFavoriteStateConsistencyAcrossRequests(t *testing.T) {
	db := newFavoriteTestDB(t)
	beanRepo := repository.NewCoffeeBeanRepository(db)
	noteRepo := repository.NewTastingNoteRepository(db)
	favRepo := repository.NewBeanFavoriteRepository(db)
	svc := NewFavoriteService(favRepo, beanRepo, noteRepo, newTestLogger())

	const uid uint = 1
	b1 := &model.CoffeeBean{Name: "豆一", ProcessMethod: "washed", FlavorTags: `["柑橘"]`}
	b2 := &model.CoffeeBean{Name: "豆二", ProcessMethod: "natural", FlavorTags: `["莓果"]`}
	_ = beanRepo.Create(b1)
	_ = beanRepo.Create(b2)
	all := []model.CoffeeBean{*b1, *b2}

	findCard := func(cards []dto.BeanCard, id uint) *dto.BeanCard {
		for i := range cards {
			if cards[i].ID == id {
				return &cards[i]
			}
		}
		return nil
	}

	// Request 1: not favored yet.
	cards, err := svc.Decorate(all, uid)
	if err != nil {
		t.Fatalf("decorate: %v", err)
	}
	if findCard(cards, b1.ID).IsFavored || findCard(cards, b2.ID).IsFavored {
		t.Fatal("both should be un-favored initially")
	}

	// Favorite b1; repeated readbacks must show b1 favored, b2 not.
	if _, err := svc.Favorite(uid, b1.ID); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	for i := 0; i < 3; i++ { // consecutive requests
		cards, err = svc.Decorate(all, uid)
		if err != nil {
			t.Fatalf("decorate repeat %d: %v", i, err)
		}
		if !findCard(cards, b1.ID).IsFavored || findCard(cards, b2.ID).IsFavored {
			t.Fatalf("request %d: expected b1=true b2=false", i)
		}
	}

	// Anonymous viewer must always see false even though uid favors b1.
	anon, err := svc.Decorate(all, 0)
	if err != nil {
		t.Fatalf("decorate anon: %v", err)
	}
	if findCard(anon, b1.ID).IsFavored {
		t.Fatal("anonymous viewer must not see favored state")
	}

	// Cancel -> next readback reflects false; another viewer's state is untouched.
	if _, err := svc.Unfavorite(uid, b1.ID); err != nil {
		t.Fatalf("unfavorite: %v", err)
	}
	cards, _ = svc.Decorate(all, uid)
	if findCard(cards, b1.ID).IsFavored {
		t.Fatal("b1 should read back un-favored after cancel")
	}

	// Different user favorites b1; uid still sees it as un-favored (per-user state).
	const other uint = 2
	if _, err := svc.Favorite(other, b1.ID); err != nil {
		t.Fatalf("other favorite: %v", err)
	}
	cards, _ = svc.Decorate(all, uid)
	if findCard(cards, b1.ID).IsFavored {
		t.Fatal("uid must not inherit another user's favorite")
	}
	cardsOther, _ := svc.Decorate(all, other)
	if !findCard(cardsOther, b1.ID).IsFavored {
		t.Fatal("other user should see their own favored state")
	}
}
