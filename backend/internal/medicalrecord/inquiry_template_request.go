package medicalrecord

// createInquiryTemplateRequest is the presence-aware create body for inquiry templates.
// IsActive is *bool so JSON binding can distinguish omitted / false / true.
type createInquiryTemplateRequest struct {
	Category  string `json:"category"              binding:"required,min=1,max=255"`
	Title     string `json:"title"                 binding:"required,min=1,max=255"`
	Content   string `json:"content"`
	IsActive  *bool  `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

// toServiceInput maps the request to the medicalrecord use-case input.
// Omitted is_active (nil) resolves to true; explicit false/true are preserved.
func (r createInquiryTemplateRequest) toServiceInput() *CreateInquiryTemplateInput {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &CreateInquiryTemplateInput{
		Category:  r.Category,
		Title:     r.Title,
		Content:   r.Content,
		IsActive:  isActive,
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
