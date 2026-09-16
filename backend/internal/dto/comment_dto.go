package dto

// CommentCreateRequest adds a comment.
type CommentCreateRequest struct {
	Content string `json:"content" binding:"required,max=1000"`
}
