package util

import "testing"

func TestRoastText(t *testing.T) {
	cases := map[string]string{"light": "浅烘", "medium": "中烘", "dark": "深烘", "x": "未知"}
	for in, want := range cases {
		if got := RoastText(in); got != want {
			t.Errorf("RoastText(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestProcessText(t *testing.T) {
	if got := ProcessText("anaerobic"); got != "厌氧发酵" {
		t.Errorf("ProcessText = %s", got)
	}
}

func TestFormatScore(t *testing.T) {
	if got := FormatScore(8.5); got != "8.5" {
		t.Errorf("FormatScore = %s", got)
	}
}
