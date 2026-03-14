package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type jobTitleResponse struct {
	ID        string    `json:"id"`
	ClinicID  string    `json:"clinic_id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toJobTitleResponse(jt *model.JobTitle) jobTitleResponse {
	return jobTitleResponse{
		ID:        strconv.FormatUint(jt.ID, 10),
		ClinicID:  strconv.FormatUint(jt.ClinicID, 10),
		Name:      jt.Name,
		IsActive:  jt.IsActive,
		SortOrder: jt.SortOrder,
		CreatedAt: jt.CreatedAt,
		UpdatedAt: jt.UpdatedAt,
	}
}

func toJobTitleResponseList(items []model.JobTitle) []jobTitleResponse {
	list := make([]jobTitleResponse, 0, len(items))
	for i := range items {
		list = append(list, toJobTitleResponse(&items[i]))
	}
	return list
}
