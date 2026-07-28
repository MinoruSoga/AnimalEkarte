package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type inquiryTemplateResponse struct {
	ID        uint64    `json:"id"`
	ClinicID  uint64    `json:"clinic_id"`
	Category  string    `json:"category"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toInquiryTemplateResponse(it *model.InquiryTemplate) inquiryTemplateResponse {
	return inquiryTemplateResponse{
		ID:        it.ID,
		ClinicID:  it.ClinicID,
		Category:  it.Category,
		Title:     it.Title,
		Content:   it.Content,
		IsActive:  it.IsActive,
		SortOrder: it.SortOrder,
		CreatedAt: httpapi.LocalTime(it.CreatedAt),
		UpdatedAt: httpapi.LocalTime(it.UpdatedAt),
	}
}
