package repository

import (
	"encoding/json"
	"sort"
)

// CountItem is a generic grouped key/count row reused by preference stats.
type CountItem struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// countFlavorTags parses raw JSON flavor-tag arrays and counts occurrences,
// returning items sorted by count desc then key asc with a stable order.
func countFlavorTags(raws []string) []CountItem {
	counts := make(map[string]int64)
	for _, raw := range raws {
		var tags []string
		if err := json.Unmarshal([]byte(raw), &tags); err != nil {
			continue
		}
		for _, t := range tags {
			if t != "" {
				counts[t]++
			}
		}
	}
	items := make([]CountItem, 0, len(counts))
	for k, v := range counts {
		items = append(items, CountItem{Key: k, Count: v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].Key < items[j].Key
	})
	return items
}

// topCountItems keeps at most n items.
func topCountItems(items []CountItem, n int) []CountItem {
	if len(items) <= n {
		return items
	}
	return items[:n]
}
