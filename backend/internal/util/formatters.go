package util

import (
	"fmt"
	"time"
)

// Shared formatters: date, score, roast/process/role text.

// FormatDate renders YYYY-MM-DD.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDateTime renders YYYY-MM-DD HH:mm.
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

// FormatScore renders a score out of 10 with one decimal.
func FormatScore(s float64) string {
	return fmt.Sprintf("%.1f", s)
}

// RoastText maps a roast level to Chinese text.
func RoastText(s string) string {
	switch s {
	case "light":
		return "浅烘"
	case "medium":
		return "中烘"
	case "dark":
		return "深烘"
	default:
		return "未知"
	}
}

// ProcessText maps a processing method to Chinese text.
func ProcessText(s string) string {
	switch s {
	case "washed":
		return "水洗"
	case "natural":
		return "日晒"
	case "honey":
		return "蜜处理"
	case "anaerobic":
		return "厌氧发酵"
	default:
		return "未知"
	}
}

// RoleText maps a role to Chinese text.
func RoleText(r string) string {
	switch r {
	case "user":
		return "普通用户"
	case "admin":
		return "管理员"
	default:
		return "未知"
	}
}
