package handler

type updateInquiryRequest struct {
	ChiefComplaint           *string `json:"chief_complaint"`
	ChiefComplaintTypeID *uint64 `json:"chief_complaint_type_id"`
	Notes                    *string `json:"notes"`
}
