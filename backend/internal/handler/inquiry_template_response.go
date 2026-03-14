package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type inquiryTemplateResponse struct {
	ID        string    `json:"id"`
	ClinicID  string    `json:"clinic_id"`
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
		ID:        strconv.FormatUint(it.ID, 10),
		ClinicID:  strconv.FormatUint(it.ClinicID, 10),
		Category:  it.Category,
		Title:     it.Title,
		Content:   it.Content,
		IsActive:  it.IsActive,
		SortOrder: it.SortOrder,
		CreatedAt: it.CreatedAt,
		UpdatedAt: it.UpdatedAt,
	}
}

func toInquiryTemplateResponseList(items []model.InquiryTemplate) []inquiryTemplateResponse {
	list := make([]inquiryTemplateResponse, 0, len(items))
	for i := range items {
		list = append(list, toInquiryTemplateResponse(&items[i]))
	}
	return list
}
