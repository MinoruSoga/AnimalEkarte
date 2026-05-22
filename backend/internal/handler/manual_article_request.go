package handler

// UpsertManualArticleRequest はマニュアル upsert の HTTP リクエスト
type UpsertManualArticleRequest struct {
	Title        string  `json:"title"         binding:"required"`
	OrderValue   float64 `json:"order_value"`
	Section      string  `json:"section"       binding:"required"`
	BodyMarkdown string  `json:"body_markdown" binding:"required"`
}
