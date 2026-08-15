package lstep

// postOwnerLstepTagRequest は POST /owners/:id/lstep/tags のリクエスト。
type postOwnerLstepTagRequest struct {
	TagName string `json:"tag_name" binding:"required"`
	Reason  string `json:"reason"`
}
