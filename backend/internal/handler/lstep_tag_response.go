package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/service"
)

// lstepTagsResponse は GET /owners/:id/lstep/tags のレスポンス。
type lstepTagsResponse struct {
	LineUserID  *string   `json:"line_user_id"`
	IsLinked    bool      `json:"is_linked"`
	LstepOptOut bool      `json:"lstep_opt_out"`
	Tags        []string  `json:"tags"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// lstepAddTagResponse は POST /owners/:id/lstep/tags のレスポンス。
type lstepAddTagResponse struct {
	TagName string `json:"tag_name"`
	Added   bool   `json:"added"`
}

func toLstepTagsResponse(r *service.OwnerTagsResult) lstepTagsResponse {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return lstepTagsResponse{
		LineUserID:  r.LineUserID,
		IsLinked:    r.IsLinked,
		LstepOptOut: r.LstepOptOut,
		Tags:        tags,
		FetchedAt:   localTime(r.FetchedAt),
	}
}
