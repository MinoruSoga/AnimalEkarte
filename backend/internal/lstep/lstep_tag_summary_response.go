package lstep

import (
	"strconv"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// tagSummaryItemResponse はタグ集計1件のJSONレスポンス。
type tagSummaryItemResponse struct {
	TagName    string `json:"tag_name"`
	OwnerCount int64  `json:"owner_count"`
	Category   string `json:"category"`
}

// tagSummaryResponse は GET /lstep/tag-summary のJSONレスポンス。
type tagSummaryResponse struct {
	Tags                 []tagSummaryItemResponse `json:"tags"`
	TotalOwnersWithLstep int64                    `json:"total_owners_with_lstep"`
	AsOf                 string                   `json:"as_of"`
}

// tagOwnerItemResponse は GET /lstep/owners 1件のJSONレスポンス。
type tagOwnerItemResponse struct {
	OwnerID          string   `json:"owner_id"`
	OwnerName        string   `json:"owner_name"`
	LineUserIDMasked *string  `json:"line_user_id_masked,omitempty"`
	LastVisitDate    *string  `json:"last_visit_date,omitempty"`
	AllTags          []string `json:"all_tags"`
	Reason           *string  `json:"reason,omitempty"`
}

// tagOwnerListResponse は GET /lstep/owners のJSONレスポンス。
type tagOwnerListResponse struct {
	Owners  []tagOwnerItemResponse `json:"owners"`
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	PerPage int                    `json:"per_page"`
}

func toTagSummaryResponse(r TagSummaryResponse) tagSummaryResponse {
	items := make([]tagSummaryItemResponse, len(r.Tags))
	for i, t := range r.Tags {
		items[i] = tagSummaryItemResponse(t)
	}
	return tagSummaryResponse{
		Tags:                 items,
		TotalOwnersWithLstep: r.TotalOwnersWithLstep,
		AsOf:                 httpapi.LocalTimeRFC3339(r.AsOf),
	}
}

func toTagOwnerListResponse(r TagOwnerListResponse) tagOwnerListResponse {
	owners := make([]tagOwnerItemResponse, len(r.Owners))
	for i, o := range r.Owners {
		tags := o.AllTags
		if tags == nil {
			tags = []string{}
		}
		owners[i] = tagOwnerItemResponse{
			OwnerID:       strconv.FormatUint(o.OwnerID, 10),
			OwnerName:     o.OwnerName,
			LastVisitDate: o.LastVisitDate,
			AllTags:       tags,
			Reason:        o.Reason,
		}
	}
	return tagOwnerListResponse{Owners: owners, Total: r.Total, Page: r.Page, PerPage: r.PerPage}
}
