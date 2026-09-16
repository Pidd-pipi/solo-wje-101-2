package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/testsupport"
)

func urlQuery(s string) string { return url.QueryEscape(s) }

func itoa(v uint) string { return strconv.FormatUint(uint64(v), 10) }

func pageQuery(page, size int) string {
	return fmt.Sprintf("page=%d&page_size=%d", page, size)
}

func containsLiteral(haystack, needle string) bool { return strings.Contains(haystack, needle) }

func errfav(format string, args ...any) error { return fmt.Errorf(format, args...) }

// errs2 records an assertion failure from a non-test goroutine (t.Fatalf is only
// safe in the test goroutine; t.Errorf is safe to call concurrently).
func errs2(t *testing.T, format string, args ...any) {
	t.Helper()
	t.Errorf(format, args...)
}

func favored(list []pageBean, id uint) bool {
	for _, b := range list {
		if b.ID == id {
			return b.IsFavored
		}
	}
	return false
}

func getProfile(t *testing.T, h *testsupport.Harness, token string, uid uint) profileData {
	t.Helper()
	status, env := h.Do(http.MethodGet, "/api/v1/users/"+itoa(uid)+"/profile", token, nil)
	if status != http.StatusOK {
		t.Fatalf("GET profile status=%d env=%+v", status, env)
	}
	var p profileData
	h.Decode(env, &p)
	return p
}

func mustNote(t *testing.T, h *testsupport.Harness, token, coffeeName, roast, tags string, score float64) {
	t.Helper()
	body := map[string]any{
		"coffee_name":   coffeeName,
		"origin":        "测试产地",
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
}

func prefMap(items []prefItem) map[string]int {
	m := make(map[string]int, len(items))
	for _, it := range items {
		m[it.Key] = it.Count
	}
	return m
}
