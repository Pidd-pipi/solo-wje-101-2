package dto

// NoteCreateRequest creates/updates a tasting note.
type NoteCreateRequest struct {
	CoffeeName   string  `json:"coffee_name" binding:"required,max=128"`
	Origin       string  `json:"origin" binding:"omitempty,max=128"`
	RoastLevel   string  `json:"roast_level" binding:"required"`
	FlavorTags   string  `json:"flavor_tags"`
	AromaScore   float64 `json:"aroma_score"`
	AcidityScore float64 `json:"acidity_score"`
	BodyScore    float64 `json:"body_score"`
	OverallScore float64 `json:"overall_score"`
	BrewMethod   string  `json:"brew_method" binding:"omitempty,max=64"`
	BrewRecipeID uint    `json:"brew_recipe_id"`
	NotesText    string  `json:"notes_text"`
	ImageURL     string  `json:"image_url" binding:"omitempty,max=255"`
}
