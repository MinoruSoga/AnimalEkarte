package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingOptionSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type trimmingCourseSummaryResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

type trimmingResponse struct {
	ID             uint64                          `json:"id"`
	ClinicID       uint64                          `json:"clinic_id"`
	Date           time.Time                       `json:"date"`
	PetID          *uint64                         `json:"pet_id,omitempty"`
	StaffID        *uint64                         `json:"staff_id,omitempty"`
	CourseID       *uint64                         `json:"course_id,omitempty"`
	Status         string                          `json:"status"`
	StyleRequest   string                          `json:"style_request"`
	BW             *float64                        `json:"bw,omitempty"`
	BWUnit         string                          `json:"bw_unit"`
	BT             *float64                        `json:"bt,omitempty"`
	UsedShampoo    string                          `json:"used_shampoo"`
	UsedRibbon     string                          `json:"used_ribbon"`
	Remarks        string                          `json:"remarks"`
	StyleImage     string                          `json:"style_image"`
	CompletedImage string                          `json:"completed_image"`
	CreatedAt      time.Time                       `json:"created_at"`
	UpdatedAt      time.Time                       `json:"updated_at"`
	Pet            *petSummaryResponse             `json:"pet,omitempty"`
	Staff          *staffSummaryResponse           `json:"staff,omitempty"`
	Course         *trimmingCourseSummaryResponse  `json:"course,omitempty"`
	Options        []trimmingOptionSummaryResponse `json:"options"`
}

func toTrimmingResponse(t *model.TrimmingRecord) trimmingResponse {
	options := make([]trimmingOptionSummaryResponse, 0, len(t.Options))
	for i := range t.Options {
		options = append(options, trimmingOptionSummaryResponse{
			ID:   t.Options[i].ID,
			Name: t.Options[i].Name,
		})
	}

	var course *trimmingCourseSummaryResponse
	if t.Course != nil {
		var price int64
		if t.Course.Price != nil {
			price = *t.Course.Price
		}
		course = &trimmingCourseSummaryResponse{
			ID:    t.Course.ID,
			Name:  t.Course.Name,
			Price: price,
		}
	}

	return trimmingResponse{
		ID:             t.ID,
		ClinicID:       t.ClinicID,
		Date:           t.Date,
		PetID:          t.PetID,
		StaffID:        t.StaffID,
		CourseID:       t.CourseID,
		Status:         string(t.Status),
		StyleRequest:   t.StyleRequest,
		BW:             t.BodyWeight,
		BWUnit:         string(t.BWUnit),
		BT:             t.BodyTemperature,
		UsedShampoo:    t.UsedShampoo,
		UsedRibbon:     t.UsedRibbon,
		Remarks:        t.Remarks,
		StyleImage:     t.StyleImage,
		CompletedImage: t.CompletedImage,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		Pet:            toPetSummary(t.Pet),
		Staff:          toStaffSummary(t.Staff),
		Course:         course,
		Options:        options,
	}
}
