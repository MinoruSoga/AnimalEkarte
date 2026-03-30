package handler

type updateInquiryRequest struct {
	ChiefComplaint           *string `json:"chief_complaint"`
	ChiefComplaintCategoryID *uint64 `json:"chief_complaint_category_id"`
	Notes                    *string `json:"notes"`
}
