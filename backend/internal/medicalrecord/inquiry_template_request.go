package medicalrecord

type createInquiryTemplateRequest struct {
	Category  string `json:"category"              binding:"required,min=1,max=255"`
	Title     string `json:"title"                 binding:"required,min=1,max=255"`
	Content   string `json:"content"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

func (r createInquiryTemplateRequest) toServiceInput() *CreateInquiryTemplateInput {
	return &CreateInquiryTemplateInput{
		Category:  r.Category,
		Title:     r.Title,
		Content:   r.Content,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
}

type updateInquiryTemplateRequest struct {
	Category  *string `json:"category"`
	Title     *string `json:"title"`
	Content   *string `json:"content"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}

func (r updateInquiryTemplateRequest) toServiceInput() *UpdateInquiryTemplateInput {
	return &UpdateInquiryTemplateInput{
		Category:  r.Category,
		Title:     r.Title,
		Content:   r.Content,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
}
