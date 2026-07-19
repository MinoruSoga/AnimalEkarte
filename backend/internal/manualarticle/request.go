package manualarticle

import "github.com/animal-ekarte/backend/internal/model"

// UpsertManualArticleRequest はマニュアル upsert の HTTP リクエスト
type UpsertManualArticleRequest struct {
	Title        string  `json:"title"         binding:"required"`
	OrderValue   float64 `json:"order_value"`
	Section      string  `json:"section"       binding:"required"`
	BodyMarkdown string  `json:"body_markdown" binding:"required"`
}

func (r UpsertManualArticleRequest) toServiceInput(category model.ManualCategory, slug string) *UpsertManualArticleInput {
	return &UpsertManualArticleInput{
		Category:     category,
		Slug:         slug,
		Title:        r.Title,
		OrderValue:   r.OrderValue,
		Section:      r.Section,
		BodyMarkdown: r.BodyMarkdown,
	}
}
