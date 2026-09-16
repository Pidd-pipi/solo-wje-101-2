package dto

// BeanCreateRequest creates/updates a coffee bean.
type BeanCreateRequest struct {
	Name          string `json:"name" binding:"required,max=128"`
	Origin        string `json:"origin" binding:"omitempty,max=128"`
	ProcessMethod string `json:"process_method" binding:"required"`
	FlavorTags    string `json:"flavor_tags"`
	Description   string `json:"description"`
}
