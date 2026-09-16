package integration

import (
	"net/http"
	"testing"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/testsupport"
)

// fullProfile captures every statistic shown on the profile page, including the
// tasting history and recent favorites, so readback after recovery can be proven
// identical to the pre-failure snapshot (no statistic silently becomes zero).
type fullProfile struct {
	User struct {
		ID uint `json:"id"`
	} `json:"user"`
	NoteCount     int      `json:"note_count"`
	AvgScore      float64  `json:"avg_score"`
	TopOrigins    []string `json:"top_origins"`
	Followers     int64    `json:"followers"`
	Following     int64    `json:"following"`
	LikesReceived int64    `json:"likes_received"`
	FavoriteCount int64    `json:"favorite_count"`
	Notes         []struct {
		ID uint `json:"id"`
	} `json:"notes"`
	RecentFavorites []favItem  `json:"recent_favorites"`
	Preference      preference `json:"preference"`
}

func getFullProfile(t *testing.T, h *testsupport.Harness, token string, uid uint) fullProfile {
	t.Helper()
	status, env := h.Do(http.MethodGet, "/api/v1/users/"+itoa(uid)+"/profile", token, nil)
	if status != http.StatusOK {
		t.Fatalf("GET profile uid=%d status=%d env=%+v", uid, status, env)
	}
	var p fullProfile
	h.Decode(env, &p)
	return p
}

func noteIDs(p fullProfile) []uint {
	ids := make([]uint, 0, len(p.Notes))
	for _, n := range p.Notes {
		ids = append(ids, n.ID)
	}
	return ids
}

func favIDs(p fullProfile) []uint {
	ids := make([]uint, 0, len(p.RecentFavorites))
	for _, f := range p.RecentFavorites {
		ids = append(ids, f.BeanID)
	}
	return ids
}

// assertSameStats proves a recovered profile is identical to the pre-failure
// snapshot across every statistic, the tasting history and recent favorites.
func assertSameStats(t *testing.T, stage string, base, got fullProfile) {
	t.Helper()
	if got.NoteCount != base.NoteCount {
		t.Fatalf("%s: note_count=%d want %d", stage, got.NoteCount, base.NoteCount)
	}
	if got.AvgScore != base.AvgScore {
		t.Fatalf("%s: avg_score=%v want %v", stage, got.AvgScore, base.AvgScore)
	}
	if !sameStrings(got.TopOrigins, base.TopOrigins) {
		t.Fatalf("%s: top_origins=%v want %v", stage, got.TopOrigins, base.TopOrigins)
	}
	if got.Followers != base.Followers || got.Following != base.Following {
		t.Fatalf("%s: followers/following=%d/%d want %d/%d",
			stage, got.Followers, got.Following, base.Followers, base.Following)
	}
	if got.LikesReceived != base.LikesReceived {
		t.Fatalf("%s: likes_received=%d want %d", stage, got.LikesReceived, base.LikesReceived)
	}
	if got.FavoriteCount != base.FavoriteCount {
		t.Fatalf("%s: favorite_count=%d want %d", stage, got.FavoriteCount, base.FavoriteCount)
	}
	if !sameUints(noteIDs(got), noteIDs(base)) {
		t.Fatalf("%s: tasting history=%v want %v", stage, noteIDs(got), noteIDs(base))
	}
	if !sameUints(favIDs(got), favIDs(base)) {
		t.Fatalf("%s: recent_favorites=%v want %v", stage, favIDs(got), favIDs(base))
	}
	if !samePrefs(got.Preference, base.Preference) {
		t.Fatalf("%s: preference drifted: got=%+v want=%+v", stage, got.Preference, base.Preference)
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameUints(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func samePrefs(a, b preference) bool {
	if len(a.Roast) != len(b.Roast) || len(a.Process) != len(b.Process) || len(a.Flavor) != len(b.Flavor) {
		return false
	}
	for i := range a.Roast {
		if a.Roast[i] != b.Roast[i] {
			return false
		}
	}
	for i := range a.Process {
		if a.Process[i] != b.Process[i] {
			return false
		}
	}
	for i := range a.Flavor {
		if a.Flavor[i] != b.Flavor[i] {
			return false
		}
	}
	return true
}

// TestProfileStatsAnyReadFailureFailsWholeProfile builds a rich profile, then for
// EACH statistic read (notes/avg/origins/follow/likes) injects a recoverable
// failure and asserts the WHOLE profile fails (never a 200 with that statistic
// zeroed). After recovery a retry must return every statistic, the tasting
// history and recent favorites exactly as before.
func TestProfileStatsAnyReadFailureFailsWholeProfile(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	admin, userA, userB := h.Token(0), h.Token(1), h.Token(2)
	uidB := h.Users[2].ID

	// Two tasting notes by userB with distinct origins and scores 8 and 6.
	createNoteAs(t, h, userB, "日晒花魁", "埃塞俄比亚", "natural", "light", `["莓果"]`, 8.0)
	note2 := createNoteAs(t, h, userB, "曼特宁手冲", "印度尼西亚", "washed", "dark", `["黑巧克力"]`, 6.0)

	// Likes received on userB's notes (from userA and admin).
	for _, tok := range []string{userA, admin} {
		if _, env := h.Do(http.MethodPost, "/api/v1/notes/"+itoa(note2)+"/like", tok, nil); env.Code != 0 {
			t.Fatalf("like setup: %+v", env)
		}
	}
	// Also like the first note as userA so both notes have likes.
	// (note1 id = note2-1 given sequential creation.)
	if _, env := h.Do(http.MethodPost, "/api/v1/notes/"+itoa(note2-1)+"/like", userA, nil); env.Code != 0 {
		t.Fatalf("like setup 2: %+v", env)
	}

	// Follow graph around userB: userA and admin follow userB (followers=2);
	// userB follows userA (following=1).
	if _, env := h.Do(http.MethodPost, "/api/v1/users/"+itoa(uidB)+"/follow", userA, nil); env.Code != 0 {
		t.Fatalf("follow setup: %+v", env)
	}
	if _, env := h.Do(http.MethodPost, "/api/v1/users/"+itoa(uidB)+"/follow", admin, nil); env.Code != 0 {
		t.Fatalf("follow setup admin: %+v", env)
	}
	if _, env := h.Do(http.MethodPost, "/api/v1/users/"+itoa(h.Users[1].ID)+"/follow", userB, nil); env.Code != 0 {
		t.Fatalf("following setup: %+v", env)
	}

	// Two favorites by userB.
	for _, beanIdx := range []uint{0, 2} {
		if _, env := h.Do(http.MethodPost, "/api/v1/beans/"+itoa(h.Beans[beanIdx].ID)+"/favorite", userB, nil); env.Code != 0 {
			t.Fatalf("favorite setup: %+v", env)
		}
	}

	base := getFullProfile(t, h, userB, uidB)
	// Sanity-check the baseline is actually populated (otherwise the test would
	// trivially "pass" against zeros).
	if base.NoteCount != 2 || base.AvgScore != 7.0 ||
		base.Followers != 2 || base.Following != 1 || base.LikesReceived != 3 ||
		base.FavoriteCount != 2 || len(base.Notes) != 2 || len(base.RecentFavorites) != 2 ||
		len(base.TopOrigins) != 2 || len(base.Preference.Roast) != 2 {
		t.Fatalf("baseline not populated as expected: %+v", base)
	}

	faults := []struct {
		name string
		kind int32
	}{
		{"notes", testsupport.ProfileFaultNotes},
		{"avg", testsupport.ProfileFaultAvg},
		{"origins", testsupport.ProfileFaultOrigins},
		{"follow", testsupport.ProfileFaultFollow},
		{"likes", testsupport.ProfileFaultLikes},
	}
	for _, f := range faults {
		t.Run(f.name+" fault fails whole profile", func(t *testing.T) {
			h.Fault.SetProfileStatFault(f.kind)
			status, env := h.Do(http.MethodGet, "/api/v1/users/"+itoa(uidB)+"/profile", userB, nil)
			if status != http.StatusInternalServerError {
				t.Fatalf("%s fault: profile status=%d want 500 (env=%+v)", f.name, status, env)
			}
			if env.Code == 0 {
				t.Fatalf("%s fault: returned success envelope (partial/zeroed profile risk)", f.name)
			}
			// Recover and retry: every statistic/history/favorite must match baseline.
			h.Fault.SetProfileStatFault(testsupport.ProfileFaultNone)
			got := getFullProfile(t, h, userB, uidB)
			assertSameStats(t, f.name+" recovered", base, got)
		})
	}
}

// TestProfileAnonymousForbiddenAndMissingKeepsBehavior locks the existing
// non-failure behavior: anonymous profile reads stay public (200), protected
// endpoints still require auth (401), admin-only actions still reject users
// (403), and an unknown account is 404 — all unaffected by the profile change.
func TestProfileAnonymousForbiddenAndMissingKeepsBehavior(t *testing.T) {
	h := testsupport.NewHarness(t)
	defer h.Close()
	uid := h.Users[1].ID

	// Anonymous profile access stays public and returns real stats, not zeros.
	status, env := h.Do(http.MethodGet, "/api/v1/users/"+itoa(uid)+"/profile", "", nil)
	if status != http.StatusOK {
		t.Fatalf("anonymous profile status=%d want 200", status)
	}
	var p fullProfile
	h.Decode(env, &p)
	if p.User.ID != uid {
		t.Fatalf("anonymous profile returned wrong user %d", p.User.ID)
	}

	// Unknown account -> 404 (existing behavior preserved).
	if status, _ := h.Do(http.MethodGet, "/api/v1/users/999999/profile", "", nil); status != http.StatusNotFound {
		t.Fatalf("missing user profile status=%d want 404", status)
	}

	// Protected endpoint without token -> 401.
	if status, _ := h.Do(http.MethodGet, "/api/v1/users/me", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("anonymous /users/me status=%d want 401", status)
	}

	// Privilege escalation: a normal user cannot perform an admin bean update.
	body := map[string]any{"name": "x", "process_method": "washed"}
	if status, _ := h.Do(http.MethodPut, "/api/v1/beans/"+itoa(h.Beans[0].ID), h.Token(1), body); status != http.StatusForbidden {
		t.Fatalf("non-admin bean update status=%d want 403", status)
	}
}

// createNoteAs creates a note with the given token and returns its id.
func createNoteAs(t *testing.T, h *testsupport.Harness, token, coffeeName, origin, _process, roast, tags string, score float64) uint {
	t.Helper()
	body := map[string]any{
		"coffee_name":   coffeeName,
		"origin":        origin,
		"roast_level":   roast,
		"flavor_tags":   tags,
		"aroma_score":   score,
		"acidity_score": score,
		"body_score":    score,
		"overall_score": score,
		"brew_method":   "手冲",
	}
	status, env := h.Do(http.MethodPost, "/api/v1/notes", token, body)
	if status != http.StatusCreated {
		t.Fatalf("create note %s status=%d env=%+v", coffeeName, status, env)
	}
	var created struct {
		ID uint `json:"id"`
	}
	h.Decode(env, &created)
	if created.ID == 0 {
		t.Fatalf("create note %s returned no id", coffeeName)
	}
	return created.ID
}
