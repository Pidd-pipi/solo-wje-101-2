// Package integration holds repeatable, real-persistence tests. They run the full
// Gin router against an on-disk, pooled WAL database (never :memory:, never a
// single serialized connection), drive it with concurrent HTTP requests, inject
// recoverable faults, and verify readback and restart persistence.
package integration

import (
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/testsupport"
)

// pageBean is one list row (CoffeeBean fields promoted + is_favored).
type pageBean struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Origin        string `json:"origin"`
	ProcessMethod string `json:"process_method"`
	Description   string `json:"description"`
	FlavorTags    string `json:"flavor_tags"`
	IsFavored     bool   `json:"is_favored"`
}

type beanPage struct {
	List     []pageBean `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type favoriteResult struct {
	BeanID  uint `json:"bean_id"`
	Favored bool `json:"favored"`
}

type preference struct {
	Roast   []prefItem `json:"roast"`
	Process []prefItem `json:"process"`
	Flavor  []prefItem `json:"flavor"`
}

type prefItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type favItem struct {
	BeanID    uint      `json:"bean_id"`
	Name      string    `json:"name"`
	FavoredAt time.Time `json:"favored_at"`
}

type profileData struct {
	User struct {
		ID uint `json:"id"`
	} `json:"user"`
	NoteCount       int        `json:"note_count"`
	FavoriteCount   int64      `json:"favorite_count"`
	RecentFavorites []favItem  `json:"recent_favorites"`
	Preference      preference `json:"preference"`
}

func listBeans(t *testing.T, h *testsupport.Harness, token, query string) (int, beanPage) {
	t.Helper()
	status, env := h.Do(http.MethodGet, "/api/v1/beans?"+query, token, nil)
	if status != http.StatusOK {
		t.Fatalf("GET /beans?%s -> status=%d env=%+v", query, status, env)
	}
	var page beanPage
	h.Decode(env, &page)
	return status, page
}

// TestBeanSearchRealDisk covers name/description/flavor matching, origin+process
// combination, pagination total and non-overlapping pages, and literal handling
// of %, _, backslash and surrounding spaces.
func TestBeanSearchRealDisk(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	token := h.Token(1) // logged-in barista

	t.Run("name/description/flavor matching", func(t *testing.T) {
		cases := []struct {
			keyword string
			min     int
			must    []string
		}{
			{"耶加", 1, []string{"耶加雪菲"}},               // name (Chinese)
			{"甜感", 2, []string{"哥伦比亚慧兰", "哥斯达黎加蜜处理"}}, // description
			{"茉莉", 1, []string{"耶加雪菲"}},               // flavor tag JSON text
			{"100%_特调", 1, []string{"100%阿拉比卡拼配"}},    // literal % and _ inside a tag
			{`批次 a_b\c`, 1, []string{`日晒\批次标注`}},      // literal _ and backslash in description
		}
		for _, tc := range cases {
			_, page := listBeans(t, h, token, "keyword="+urlQuery(tc.keyword))
			if int(page.Total) < tc.min {
				t.Fatalf("keyword=%q total=%d want>=%d", tc.keyword, page.Total, tc.min)
			}
			names := map[string]bool{}
			for _, b := range page.List {
				names[b.Name] = true
			}
			for _, want := range tc.must {
				if !names[want] {
					t.Fatalf("keyword=%q expected to match %q; got %v", tc.keyword, want, names)
				}
			}
		}
	})

	t.Run("wildcards are literals", func(t *testing.T) {
		// '%' alone must NOT act as "match all"; underscore must not match one char.
		for _, kw := range []string{"%", "_", "100%", "AA_top", `a\b`} {
			_, page := listBeans(t, h, token, "keyword="+urlQuery(kw))
			if page.Total == int64(len(h.Beans)) {
				t.Fatalf("keyword=%q behaved like a wildcard: returned all %d beans", kw, page.Total)
			}
			for _, b := range page.List {
				if !containsLiteral(b.Name, kw) && !containsLiteral(b.Description, kw) && !containsLiteral(b.FlavorTags, kw) {
					t.Fatalf("keyword=%q matched bean %q without a literal occurrence", kw, b.Name)
				}
			}
		}
	})

	t.Run("surrounding spaces trimmed", func(t *testing.T) {
		_, trimmed := listBeans(t, h, token, "keyword="+urlQuery("  耶加  "))
		_, exact := listBeans(t, h, token, "keyword="+urlQuery("耶加"))
		if trimmed.Total != exact.Total {
			t.Fatalf("trimmed keyword total=%d differs from exact=%d", trimmed.Total, exact.Total)
		}
		// Pure whitespace must not filter anything.
		_, blank := listBeans(t, h, token, "keyword="+urlQuery("    "))
		if blank.Total != int64(len(h.Beans)) {
			t.Fatalf("whitespace keyword total=%d want all %d", blank.Total, len(h.Beans))
		}
	})

	t.Run("origin and process combination", func(t *testing.T) {
		// 埃塞俄比亚 + washed: 耶加雪菲 and AA_top微批次.
		_, page := listBeans(t, h, token, "origin="+urlQuery("埃塞俄比亚")+"&process=washed")
		if page.Total != 2 {
			t.Fatalf("origin+process total=%d want 2 (%v)", page.Total, page.List)
		}
		for _, b := range page.List {
			if b.Origin != "埃塞俄比亚" || b.ProcessMethod != "washed" {
				t.Fatalf("combined filter returned %+v", b)
			}
		}
		// Combination that matches nothing must still be a valid empty page.
		_, none := listBeans(t, h, token, "origin=巴西&process=honey")
		if none.Total != 0 || len(none.List) != 0 {
			t.Fatalf("impossible combination total=%d list=%d want empty", none.Total, len(none.List))
		}
		// Combination + keyword.
		_, kw := listBeans(t, h, token, "origin="+urlQuery("埃塞俄比亚")+"&process=washed&keyword="+urlQuery("微批次"))
		if kw.Total != 1 || kw.List[0].Name != "AA_top微批次" {
			t.Fatalf("origin+process+keyword total=%d list=%v", kw.Total, kw.List)
		}
	})

	t.Run("pagination total and non-overlapping pages", func(t *testing.T) {
		const size = 5
		want := int64(len(h.Beans))
		allIDs := map[uint]bool{}
		var seen int64
		for pageNo := 1; pageNo <= 3; pageNo++ {
			_, page := listBeans(t, h, token, pageQuery(pageNo, size))
			if page.Total != want {
				t.Fatalf("page %d total=%d want %d", pageNo, page.Total, want)
			}
			for _, b := range page.List {
				if allIDs[b.ID] {
					t.Fatalf("bean id=%d appeared on more than one page", b.ID)
				}
				allIDs[b.ID] = true
				seen++
			}
		}
		if seen != want {
			t.Fatalf("paged ids=%d want %d (overlap or missing)", seen, want)
		}
	})

	t.Run("anonymous list readable", func(t *testing.T) {
		status, env := h.Do(http.MethodGet, "/api/v1/beans?keyword="+urlQuery("耶加"), "", nil)
		if status != http.StatusOK {
			t.Fatalf("anonymous list status=%d env=%+v", status, env)
		}
		var page beanPage
		h.Decode(env, &page)
		if page.Total == 0 {
			t.Fatal("anonymous search expected results")
		}
		for _, b := range page.List {
			if b.IsFavored {
				t.Fatal("anonymous list must never report is_favored=true")
			}
		}
	})
}

// TestFavoriteStateConcurrency drives real concurrent favorite/unfavorite and
// list requests through the pooled on-disk DB, then asserts readback consistency.
func TestFavoriteStateConcurrency(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	userA, userB := h.Token(1), h.Token(2)
	beanID := h.Beans[0].ID // 耶加雪菲

	// Concurrently favorite the SAME bean many times from one user: exactly one row.
	var wg sync.WaitGroup
	errs := make(chan error, 32)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, env := h.Do(http.MethodPost,
				"/api/v1/beans/"+itoa(beanID)+"/favorite", userA, nil)
			if status != http.StatusCreated {
				errs <- errfav("concurrent favorite status=%d code=%d msg=%s", status, env.Code, env.Message)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Fatal(e)
	}

	// Profile count must read back exactly 1 once the dust settles.
	prof := getProfile(t, h, userA, h.Users[1].ID)
	if prof.FavoriteCount != 1 {
		t.Fatalf("after concurrent favorite, favorite_count=%d want 1 (duplicate rows?)", prof.FavoriteCount)
	}

	// Concurrent reads from both users during a steady state are consistent.
	const readers = 24
	var wg2 sync.WaitGroup
	barrier := make(chan struct{})
	for i := 0; i < readers; i++ {
		wg2.Add(1)
		go func(idx int) {
			defer wg2.Done()
			<-barrier
			tok := userA
			wantFav := true
			if idx%2 == 1 {
				tok, wantFav = userB, false
			}
			_, page := listBeans(t, h, tok, "")
			var found bool
			for _, b := range page.List {
				if b.ID == beanID {
					found = true
					if b.IsFavored != wantFav {
						errs2(t, "concurrent reader is_favored=%v want %v", b.IsFavored, wantFav)
					}
				}
			}
			if !found {
				errs2(t, "favored bean missing from list page")
			}
		}(i)
	}
	close(barrier)
	wg2.Wait()

	// Per-user isolation: userB favoriting must not leak into userA's list state.
	if status, env := h.Do(http.MethodPost, "/api/v1/beans/"+itoa(beanID)+"/favorite", userB, nil); status != http.StatusCreated {
		t.Fatalf("userB favorite status=%d env=%+v", status, env)
	}
	_, pageA := listBeans(t, h, userA, "")
	_, pageB := listBeans(t, h, userB, "")
	if !favored(pageA.List, beanID) || !favored(pageB.List, beanID) {
		t.Fatal("both users should independently read is_favored=true after each favored")
	}

	// UserA cancels once: reads back false for A, remains true for B.
	if status, env := h.Do(http.MethodDelete, "/api/v1/beans/"+itoa(beanID)+"/favorite", userA, nil); status != http.StatusOK {
		t.Fatalf("unfavorite status=%d env=%+v", status, env)
	}
	_, afterA := listBeans(t, h, userA, "")
	_, afterB := listBeans(t, h, userB, "")
	if favored(afterA.List, beanID) {
		t.Fatal("userA readback still favored after cancel (state not synced)")
	}
	if !favored(afterB.List, beanID) {
		t.Fatal("userB favorite lost when userA canceled (cross-user corruption)")
	}
	profA := getProfile(t, h, userA, h.Users[1].ID)
	if profA.FavoriteCount != 0 {
		t.Fatalf("userA favorite_count=%d want 0 after cancel", profA.FavoriteCount)
	}
	// Cancel twice is idempotent.
	if status, _ := h.Do(http.MethodDelete, "/api/v1/beans/"+itoa(beanID)+"/favorite", userA, nil); status != http.StatusOK {
		t.Fatalf("second unfavorite should be idempotent status=200, got %d", status)
	}
}

// TestFavoriteStateFaultFailsWholeRequest verifies that when the favorite-state
// read fails, the logged-in bean list fails as a whole (never returns a partial
// page with is_favored=false), while the anonymous list stays readable; recovery
// restores correct readback.
func TestFavoriteStateFaultFailsWholeRequest(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	userA := h.Token(1)
	beanID := h.Beans[1].ID
	if _, env := h.Do(http.MethodPost, "/api/v1/beans/"+itoa(beanID)+"/favorite", userA, nil); env.Code != 0 {
		t.Fatalf("favorite setup failed: %+v", env)
	}

	baseline := func(stage string, expectFavored bool) {
		status, env := h.Do(http.MethodGet, "/api/v1/beans", userA, nil)
		switch {
		case expectFavored && status == http.StatusInternalServerError:
			t.Fatalf("%s: logged-in list returned 500 after recovery (readback broken)", stage)
		case expectFavored && status == http.StatusOK:
			var page beanPage
			h.Decode(env, &page)
			if !favored(page.List, beanID) {
				t.Fatalf("%s: readback shows un-favored but favorite exists", stage)
			}
		}
	}

	baseline("before fault", true)

	// Stage 1: inject favorite-state fault.
	h.Fault.SetFavoriteStateFault(true)
	t.Run("stage:fault-active", func(t *testing.T) {
		status, env := h.Do(http.MethodGet, "/api/v1/beans", userA, nil)
		if status != http.StatusInternalServerError {
			t.Fatalf("logged-in list during fault status=%d want 500", status)
		}
		if env.Code == 0 {
			t.Fatal("logged-in list during fault returned success envelope (half page risk)")
		}
		// Anonymous list must remain readable (no favorite-state read).
		anonStatus, anonEnv := h.Do(http.MethodGet, "/api/v1/beans", "", nil)
		if anonStatus != http.StatusOK {
			t.Fatalf("anonymous list during fault status=%d want 200", anonStatus)
		}
		var ap beanPage
		h.Decode(anonEnv, &ap)
		if len(ap.List) == 0 {
			t.Fatal("anonymous list during fault returned no rows")
		}
	})

	// Stage 2: recover. Readback must be correct again and favorite must persist.
	h.Fault.SetFavoriteStateFault(false)
	status, env := h.Do(http.MethodGet, "/api/v1/beans", userA, nil)
	if status != http.StatusOK {
		t.Fatalf("after recovery logged-in list status=%d want 200", status)
	}
	var recovered beanPage
	h.Decode(env, &recovered)
	if !favored(recovered.List, beanID) {
		t.Fatal("after recovery the existing favorite reads back as un-favored (state damaged by fault)")
	}
	baseline("after recovery", true)
}

// TestProfilePreferenceAggregatesAndFaults builds notes + favorites and verifies
// the roast/process/flavor profile, then injects a preference fault (the whole
// profile must fail) and recovers.
func TestProfilePreferenceAggregatesAndFaults(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	userA := h.Token(1)
	uid := h.Users[1].ID

	// Two light-roast notes, one medium.
	mustNote(t, h, userA, "日晒豆A", "light", `["柑橘","茉莉"]`, 8.0)
	mustNote(t, h, userA, "水洗豆B", "light", `["柑橘"]`, 7.5)
	mustNote(t, h, userA, "深烘豆C", "dark", `["黑巧克力"]`, 8.2)
	// Favorites: one washed bean, one honey bean.
	favBeans := []uint{h.Beans[0].ID, h.Beans[2].ID}
	for _, id := range favBeans {
		if _, env := h.Do(http.MethodPost, "/api/v1/beans/"+itoa(id)+"/favorite", userA, nil); env.Code != 0 {
			t.Fatalf("favorite %d: %+v", id, env)
		}
	}

	prof := getProfile(t, h, userA, uid)
	if prof.FavoriteCount != 2 {
		t.Fatalf("favorite_count=%d want 2", prof.FavoriteCount)
	}
	if len(prof.RecentFavorites) != 2 {
		t.Fatalf("recent_favorites=%d want 2", len(prof.RecentFavorites))
	}
	// Recent favorites must be newest first.
	if !sort.SliceIsSorted(prof.RecentFavorites, func(i, j int) bool {
		return prof.RecentFavorites[i].FavoredAt.After(prof.RecentFavorites[j].FavoredAt)
	}) {
		t.Fatal("recent_favorites not ordered newest-first")
	}

	roast := prefMap(prof.Preference.Roast)
	if roast["light"] != 2 || roast["dark"] != 1 {
		t.Fatalf("roast preference=%v want light=2 dark=1", roast)
	}
	proc := prefMap(prof.Preference.Process)
	if proc["washed"] != 1 || proc["honey"] != 1 {
		t.Fatalf("process preference=%v want washed=1 honey=1", proc)
	}
	flav := prefMap(prof.Preference.Flavor)
	// 柑橘 appears in two notes AND one favored bean -> merged count 3.
	if flav["柑橘"] != 3 {
		t.Fatalf("flavor 柑橘=%d want 3 (merged across notes+favorite beans)", flav["柑橘"])
	}
	if _, ok := flav["黑巧克力"]; !ok {
		t.Fatalf("flavor preference missing 黑巧克力: %v", flav)
	}

	// Cancel a favorite -> profile readback updates count and process preference.
	if _, env := h.Do(http.MethodDelete, "/api/v1/beans/"+itoa(h.Beans[2].ID)+"/favorite", userA, nil); env.Code != 0 {
		t.Fatalf("unfavorite: %+v", env)
	}
	prof2 := getProfile(t, h, userA, uid)
	if prof2.FavoriteCount != 1 {
		t.Fatalf("after cancel favorite_count=%d want 1", prof2.FavoriteCount)
	}
	if prefMap(prof2.Preference.Process)["honey"] != 0 {
		t.Fatalf("honey preference must drop to 0 after unfavoring the honey bean: %+v", prof2.Preference.Process)
	}

	// Fault on the preference read: whole profile fails.
	h.Fault.SetProfilePreferenceFault(true)
	status, env := h.Do(http.MethodGet, "/api/v1/users/"+itoa(uid)+"/profile", userA, nil)
	if status != http.StatusInternalServerError {
		t.Fatalf("profile during preference fault status=%d want 500", status)
	}
	if env.Code == 0 {
		t.Fatal("profile during preference fault returned success (partial profile risk)")
	}
	h.Fault.SetProfilePreferenceFault(false)
	prof3 := getProfile(t, h, userA, uid) // must recover and read back correctly
	if prof3.FavoriteCount != 1 || prefMap(prof3.Preference.Roast)["light"] != 2 {
		t.Fatalf("post-recovery profile inconsistent: count=%d roast=%+v", prof3.FavoriteCount, prof3.Preference.Roast)
	}
}

// TestAdminBeanUpdateAndDelistingReadback verifies that admin updates keep
// favorite relations (and are searchable by the new name), de-listing cascades
// favorites for every user, and non-admins are rejected. All assertions are
// read back through the real HTTP API against the on-disk database.
func TestAdminBeanUpdateAndDelistingReadback(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	admin, userA, userB := h.Token(0), h.Token(1), h.Token(2)
	uidA, uidB := h.Users[1].ID, h.Users[2].ID
	beanID := h.Beans[0].ID // 耶加雪菲
	otherID := h.Beans[2].ID

	// Two users favorite the bean that will be renamed (update must keep relations).
	for _, tok := range []string{userA, userB} {
		if _, env := h.Do(http.MethodPost, "/api/v1/beans/"+itoa(beanID)+"/favorite", tok, nil); env.Code != 0 {
			t.Fatalf("favorite setup: %+v", env)
		}
	}
	// And both favorite the bean that will be de-listed.
	for _, tok := range []string{userA, userB} {
		if _, env := h.Do(http.MethodPost, "/api/v1/beans/"+itoa(otherID)+"/favorite", tok, nil); env.Code != 0 {
			t.Fatalf("favorite other setup: %+v", env)
		}
	}

	// Non-admin must not update beans.
	if status, _ := h.Do(http.MethodPut, "/api/v1/beans/"+itoa(beanID), userA,
		map[string]any{"name": "x", "process_method": "washed"}); status != http.StatusForbidden {
		t.Fatalf("non-admin update status=%d want 403", status)
	}

	// Admin update (rename + new description + new tag), same bean id.
	update := map[string]any{
		"name": "白桃微批次2024", "origin": "埃塞俄比亚", "process_method": "washed",
		"flavor_tags": `["柑橘","白桃"]`, "description": "新产季白桃调",
	}
	if status, env := h.Do(http.MethodPut, "/api/v1/beans/"+itoa(beanID), admin, update); status != http.StatusOK {
		t.Fatalf("admin update status=%d env=%+v", status, env)
	}

	// Readback: searchable by new name/tag/description, old name no longer matches.
	_, byNew := listBeans(t, h, userA, "keyword="+urlQuery("白桃"))
	if byNew.Total != 1 || byNew.List[0].ID != beanID {
		t.Fatalf("updated bean not found by new tag: total=%d list=%v", byNew.Total, byNew.List)
	}
	_, byOld := listBeans(t, h, userA, "keyword="+urlQuery("耶加雪菲"))
	if byOld.Total != 0 {
		t.Fatalf("old name still matches after rename: total=%d", byOld.Total)
	}
	// Favorite relations must survive the update for BOTH users.
	if _, pA := listBeans(t, h, userA, "keyword="+urlQuery("白桃")); !favored(pA.List, beanID) {
		t.Fatal("userA favorite lost after bean update")
	}
	if _, pB := listBeans(t, h, userB, "keyword="+urlQuery("白桃")); !favored(pB.List, beanID) {
		t.Fatal("userB favorite lost after bean update")
	}
	if p := getProfile(t, h, userA, uidA); p.FavoriteCount != 2 {
		t.Fatalf("userA count=%d want 2 after update (relations retained)", p.FavoriteCount)
	}

	// Published tasting notes must remain intact (they are stored independently).
	noteBody := map[string]any{
		"coffee_name": "旧名称豆", "roast_level": "light", "flavor_tags": `["柑橘"]`,
		"origin": "埃塞俄比亚", "overall_score": 7.5, "aroma_score": 7.5,
		"acidity_score": 7.5, "body_score": 7.5,
	}
	if status, env := h.Do(http.MethodPost, "/api/v1/notes", userA, noteBody); status != http.StatusCreated {
		t.Fatalf("create note setup status=%d env=%+v", status, env)
	}

	// Admin de-lists otherID: its favorites cascade away for every user.
	if status, env := h.Do(http.MethodDelete, "/api/v1/beans/"+itoa(otherID), admin, nil); status != http.StatusOK {
		t.Fatalf("admin delete status=%d env=%+v", status, env)
	}
	if _, p := listBeans(t, h, userA, ""); favored(p.List, otherID) {
		t.Fatal("de-listed bean still present in list")
	}
	if _, p := listBeans(t, h, userA, "keyword="+urlQuery("哥斯达黎加")); p.Total != 0 {
		t.Fatalf("de-listed bean still searchable: total=%d", p.Total)
	}
	if p := getProfile(t, h, userA, uidA); p.FavoriteCount != 1 {
		t.Fatalf("userA count=%d want 1 after de-list cascade", p.FavoriteCount)
	}
	if p := getProfile(t, h, userB, uidB); p.FavoriteCount != 1 {
		t.Fatalf("userB count=%d want 1 after de-list cascade", p.FavoriteCount)
	}
	// The renamed bean (still favored) remains for both users.
	if p := getProfile(t, h, userA, uidA); len(p.RecentFavorites) != 1 || p.RecentFavorites[0].BeanID != beanID {
		t.Fatalf("userA recent favorites wrong after de-list: %+v", p.RecentFavorites)
	}
}
func TestPersistenceAcrossRestart(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	userA := h.Token(1)
	beanID := h.Beans[0].ID
	if _, env := h.Do(http.MethodPost, "/api/v1/beans/"+itoa(beanID)+"/favorite", userA, nil); env.Code != 0 {
		t.Fatalf("favorite before restart: %+v", env)
	}

	h2 := h.Restart()
	defer h2.Close()
	userA2 := h2.Token(1)
	prof := getProfile(t, h2, userA2, h2.Users[1].ID)
	if prof.FavoriteCount != 1 {
		t.Fatalf("after restart favorite_count=%d want 1 (not persisted)", prof.FavoriteCount)
	}
	status, env := h2.Do(http.MethodGet, "/api/v1/beans", userA2, nil)
	if status != http.StatusOK {
		t.Fatalf("list after restart status=%d", status)
	}
	var page beanPage
	h2.Decode(env, &page)
	if !favored(page.List, beanID) {
		t.Fatal("after restart is_favored readback is false (relation lost)")
	}
}
