package dto

// RecipeCreateRequest creates a brew recipe.
type RecipeCreateRequest struct {
	Name      string `json:"name" binding:"required,max=128"`
	Device    string `json:"device" binding:"omitempty,max=64"`
	WaterTemp int    `json:"water_temp"`
	GrindSize string `json:"grind_size" binding:"omitempty,max=32"`
	Ratio     string `json:"ratio" binding:"omitempty,max=32"`
	Steps     string `json:"steps"`
}
