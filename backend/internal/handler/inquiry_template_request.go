package handler

type createInquiryTemplateRequest struct {
	Category  string `json:"category"`
	Title     string `json:"title"     binding:"required"`
	Content   string `json:"content"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

type updateInquiryTemplateRequest struct {
	Category  *string `json:"category"`
	Title     *string `json:"title"`
	Content   *string `json:"content"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}
