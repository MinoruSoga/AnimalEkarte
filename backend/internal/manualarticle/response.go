package manualarticle

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ManualArticleResponse はマニュアル記事の HTTP レスポンス
type ManualArticleResponse struct {
	ID               uint64    `json:"id"`
	Category         string    `json:"category"`
	Slug             string    `json:"slug"`
	Title            string    `json:"title"`
	OrderValue       float64   `json:"order_value"`
	Section          string    `json:"section"`
	BodyMarkdown     string    `json:"body_markdown"`
	UpdatedByStaffID *uint64   `json:"updated_by_staff_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ManualArticleListResponse はマニュアル記事一覧の HTTP レスポンス
type ManualArticleListResponse struct {
	Data []ManualArticleResponse `json:"data"`
}

// ManualArticleVersionListResponse は編集履歴一覧の HTTP レスポンス
type ManualArticleVersionListResponse struct {
	Data []ManualArticleVersionResponse `json:"data"`
}

// ManualArticleVersionResponse は編集履歴のレスポンス
type ManualArticleVersionResponse struct {
	ID              uint64    `json:"id"`
	ArticleID       uint64    `json:"article_id"`
	Title           string    `json:"title"`
	OrderValue      float64   `json:"order_value"`
	Section         string    `json:"section"`
	BodyMarkdown    string    `json:"body_markdown"`
	EditedByStaffID *uint64   `json:"edited_by_staff_id,omitempty"`
	EditedAt        time.Time `json:"edited_at"`
}

func toManualArticleListResponse(articles []model.ManualArticle) ManualArticleListResponse {
	data := httpapi.MapSlice(articles, toManualArticleResponse)
	return ManualArticleListResponse{Data: data}
}

func toManualArticleVersionListResponse(versions []model.ManualArticleVersion) ManualArticleVersionListResponse {
	data := httpapi.MapSlice(versions, toManualArticleVersionResponse)
	return ManualArticleVersionListResponse{Data: data}
}

func toManualArticleResponse(a *model.ManualArticle) ManualArticleResponse {
	return ManualArticleResponse{
		ID:               a.ID,
		Category:         string(a.Category),
		Slug:             a.Slug,
		Title:            a.Title,
		OrderValue:       a.OrderValue,
		Section:          a.Section,
		BodyMarkdown:     a.BodyMarkdown,
		UpdatedByStaffID: a.UpdatedByStaffID,
		CreatedAt:        httpapi.LocalTime(a.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(a.UpdatedAt),
	}
}

func toManualArticleVersionResponse(v *model.ManualArticleVersion) ManualArticleVersionResponse {
	return ManualArticleVersionResponse{
		ID:              v.ID,
		ArticleID:       v.ArticleID,
		Title:           v.Title,
		OrderValue:      v.OrderValue,
		Section:         v.Section,
		BodyMarkdown:    v.BodyMarkdown,
		EditedByStaffID: v.EditedByStaffID,
		EditedAt:        httpapi.LocalTime(v.EditedAt),
	}
}
