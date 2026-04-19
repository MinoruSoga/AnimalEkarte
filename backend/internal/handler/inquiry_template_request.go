package handler

type createInquiryTemplateRequest struct {
	Category  string `json:"category"              binding:"required,min=1,max=255"`
	Title     string `json:"title"                 binding:"required,min=1,max=255"`
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
