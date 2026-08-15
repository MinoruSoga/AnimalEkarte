package manualarticle

import "github.com/animal-ekarte/backend/internal/model"

// UpsertManualArticleRequest はマニュアル upsert の HTTP リクエスト
type UpsertManualArticleRequest struct {
	Title        string  `json:"title"         binding:"required,max=255"`
	OrderValue   float64 `json:"order_value"`
	Section      string  `json:"section"       binding:"required,max=255"`
	BodyMarkdown string  `json:"body_markdown" binding:"required,max=100000"`
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
