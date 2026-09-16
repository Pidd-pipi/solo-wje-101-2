package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

func newBeanRepoTestDB(t *testing.T) (*gorm.DB, *CoffeeBeanRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.CoffeeBean{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db, NewCoffeeBeanRepository(db)
}

func seedBeans(t *testing.T, repo *CoffeeBeanRepository) []model.CoffeeBean {
	t.Helper()
	beans := []model.CoffeeBean{
		{Name: "耶加雪菲", Origin: "埃塞俄比亚", ProcessMethod: "washed", FlavorTags: `["柑橘","茉莉"]`, Description: "明亮柑橘酸质"},
		{Name: "哥伦比亚慧兰", Origin: "哥伦比亚", ProcessMethod: "washed", FlavorTags: `["坚果","焦糖"]`, Description: "甜感平衡"},
		{Name: "哥斯达黎加蜜处理", Origin: "哥斯达黎加", ProcessMethod: "honey", FlavorTags: `["莓果"]`, Description: "醇厚甜感"},
		{Name: "100%阿拉比卡", Origin: "测试产地", ProcessMethod: "natural", FlavorTags: `["黑巧克力"]`, Description: "含下划线_symbol"},
	}
	if err := repo.Create(&beans[0]); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := repo.Create(&beans[1]); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := repo.Create(&beans[2]); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := repo.Create(&beans[3]); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return beans
}

func ids(items []model.CoffeeBean) []uint {
	out := make([]uint, len(items))
	for i, b := range items {
		out[i] = b.ID
	}
	return out
}

func TestBeanListKeywordSearch(t *testing.T) {
	_, repo := newBeanRepoTestDB(t)
	beans := seedBeans(t, repo)
	b1, b2, b3, b4 := beans[0].ID, beans[1].ID, beans[2].ID, beans[3].ID

	cases := []struct {
		name    string
		origin  string
		process string
		keyword string
		wantIDs []uint
	}{
		{"Chinese name", "", "", "耶加", []uint{b1}},
		{"description match", "", "", "甜感", []uint{b2, b3}},
		{"flavor tag match", "", "", "茉莉", []uint{b1}},
		{"flavor tag match 2", "", "", "坚果", []uint{b2}},
		// '%' must be treated as an ordinary character, not a wildcard.
		{"literal percent", "", "", "100%", []uint{b4}},
		// '_' must be treated as an ordinary character, not a single-char wildcard.
		{"literal underscore", "", "", "_symbol", []uint{b4}},
		{"trim surrounding spaces", "", "", "  耶加  ", []uint{b1}},
		{"origin + keyword hit", "埃塞俄比亚", "", "柑橘", []uint{b1}},
		{"origin + keyword miss", "埃塞俄比亚", "", "坚果", []uint{}},
		{"process + keyword", "", "honey", "莓", []uint{b3}},
		{"origin only", "哥伦比亚", "", "", []uint{b2}},
		{"process only", "", "washed", "", []uint{b1, b2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			items, total, err := repo.List(tc.origin, tc.process, tc.keyword, 1, 50)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			got := ids(items)
			if int(total) != len(tc.wantIDs) {
				t.Fatalf("total = %d, want %d (got %v)", total, len(tc.wantIDs), got)
			}
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("rows = %v, want %v", got, tc.wantIDs)
			}
			for i, want := range tc.wantIDs {
				if got[i] != want {
					t.Fatalf("rows = %v, want %v", got, tc.wantIDs)
				}
			}
		})
	}

	// Whitespace-only / empty keyword must not filter anything.
	for _, kw := range []string{"", "   "} {
		_, total, err := repo.List("", "", kw, 1, 50)
		if err != nil {
			t.Fatalf("empty keyword list: %v", err)
		}
		if total != 4 {
			t.Fatalf("keyword=%q total = %d, want 4", kw, total)
		}
	}
}

func TestBeanListPaginationConsistency(t *testing.T) {
	_, repo := newBeanRepoTestDB(t)
	beans := seedBeans(t, repo)

	// Page 1.
	page1, total, err := repo.List("", "", "", 1, 2)
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if total != 4 || len(page1) != 2 {
		t.Fatalf("page1 total=%d len=%d, want 4/2", total, len(page1))
	}
	if got := ids(page1); got[0] != beans[0].ID || got[1] != beans[1].ID {
		t.Fatalf("page1 rows = %v, want first two in id ASC", got)
	}
	// Page 2.
	page2, total2, err := repo.List("", "", "", 2, 2)
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if total2 != 4 || len(page2) != 2 {
		t.Fatalf("page2 total=%d len=%d, want 4/2", total2, len(page2))
	}
	if got := ids(page2); got[0] != beans[2].ID || got[1] != beans[3].ID {
		t.Fatalf("page2 rows = %v, want last two in id ASC", got)
	}

	// A filtered total must stay consistent across pages of the same filter.
	const kw = "甜感"
	_, filteredTotal, err := repo.List("", "", kw, 1, 1)
	if err != nil {
		t.Fatalf("filtered: %v", err)
	}
	all, allTotal, err := repo.List("", "", kw, 1, 50)
	if err != nil {
		t.Fatalf("filtered all: %v", err)
	}
	if filteredTotal != allTotal || int(filteredTotal) != len(all) {
		t.Fatalf("filtered total mismatch: page=%d all=%d rows=%d", filteredTotal, allTotal, len(all))
	}
}

func TestBeanUpdateAndDeleteReadback(t *testing.T) {
	_, repo := newBeanRepoTestDB(t)
	beans := seedBeans(t, repo)
	b4 := beans[3]

	// Admin update: the literal-percent keyword must stop matching after rename,
	// and the new name must be searchable.
	b4.Name = "改名后的豆"
	if err := repo.Update(&b4); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, total, _ := repo.List("", "", "100%", 1, 50); total != 0 {
		t.Fatalf("expected 0 match for old literal-percent name, got %d", total)
	}
	if _, total, _ := repo.List("", "", "改名", 1, 50); total != 1 {
		t.Fatalf("expected new name searchable, got %d", total)
	}

	// Admin de-list: removed bean disappears from every listing/search.
	if err := repo.Delete(b4.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, total, _ := repo.List("", "", "", 1, 50); total != 3 {
		t.Fatalf("expected 3 beans after delete, got %d", total)
	}
	if _, total, _ := repo.List("", "", "改名", 1, 50); total != 0 {
		t.Fatalf("expected deleted bean unsearchable, got %d", total)
	}
}
