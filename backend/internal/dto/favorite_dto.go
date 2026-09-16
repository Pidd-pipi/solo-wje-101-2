package dto

import (
	"time"

	"github.com/wjecoffeetaste/wjecoffeetaste/internal/model"
)

// BeanCard is a coffee bean augmented with the requesting user's favorite state.
type BeanCard struct {
	model.CoffeeBean
	IsFavored bool `json:"is_favored"`
}

// FavoriteResult is returned by favorite/unfavorite endpoints.
type FavoriteResult struct {
	BeanID     uint `json:"bean_id"`
	Favored    bool `json:"favored"`
	FavoriteID uint `json:"favorite_id,omitempty"`
}

// FavoriteItem is a favored bean with the favorite timestamp.
type FavoriteItem struct {
	ID            uint      `json:"id"`
	BeanID        uint      `json:"bean_id"`
	Name          string    `json:"name"`
	Origin        string    `json:"origin"`
	ProcessMethod string    `json:"process_method"`
	FlavorTags    string    `json:"flavor_tags"`
	FavoredAt     time.Time `json:"favored_at"`
}

// PreferenceItem is a single preference bucket with its count.
type PreferenceItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// TastePreference aggregates a user's roast/process/flavor preferences.
type TastePreference struct {
	Roast   []PreferenceItem `json:"roast"`
	Process []PreferenceItem `json:"process"`
	Flavor  []PreferenceItem `json:"flavor"`
}
