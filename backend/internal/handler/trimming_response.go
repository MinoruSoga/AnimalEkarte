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

// trimmingResponse は appointments ベースのフラット DTO（BE-119）
type trimmingResponse struct {
	ID                uint64    `json:"id"`
	ClinicID          uint64    `json:"clinic_id"`
	ReservationTypeID uint64    `json:"reservation_type_id"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	PetID             *uint64   `json:"pet_id,omitempty"`
	StaffID           *uint64   `json:"staff_id,omitempty"` // doctor_id をマップ
	Status            string    `json:"status"`
	Source            string    `json:"source"`
	// トリミング詳細（appointment_trimming_details から flat 化）
	CourseID       *uint64   `json:"course_id,omitempty"`
	StyleRequest   string    `json:"style_request"`
	BW             *float64  `json:"bw,omitempty"`
	BWUnit         string    `json:"bw_unit"`
	BT             *float64  `json:"bt,omitempty"`
	UsedShampoo    string    `json:"used_shampoo"`
	UsedRibbon     string    `json:"used_ribbon"`
	Remarks        string    `json:"remarks"`
	StyleImage     string    `json:"style_image"`
	CompletedImage string    `json:"completed_image"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// リレーション
	Pet     *petSummaryResponse             `json:"pet,omitempty"`
	Staff   *staffSummaryResponse           `json:"staff,omitempty"`
	Course  *trimmingCourseSummaryResponse  `json:"course,omitempty"`
	Options []trimmingOptionSummaryResponse `json:"options"`
}

func toTrimmingResponse(appt *model.Reservation) trimmingResponse {
	resp := trimmingResponse{
		ID:                appt.ID,
		ClinicID:          appt.ClinicID,
		ReservationTypeID: appt.ReservationTypeID,
		StartTime:         appt.StartTime,
		EndTime:           appt.EndTime,
		PetID:             appt.PetID,
		StaffID:           appt.DoctorID,
		Status:            string(appt.Status),
		Source:            string(appt.Source),
		CreatedAt:         appt.CreatedAt,
		UpdatedAt:         appt.UpdatedAt,
		Pet:               toPetSummary(appt.Pet),
		Staff:             toStaffSummary(appt.Doctor),
		Options:           make([]trimmingOptionSummaryResponse, 0),
		// TrimmingDetail が nil の異常データ向けデフォルト（モデルデフォルトと一致）
		BWUnit: "Kg",
	}

	if appt.TrimmingDetail != nil {
		d := appt.TrimmingDetail
		resp.CourseID = d.CourseID
		resp.StyleRequest = d.StyleRequest
		resp.BW = d.BodyWeight
		resp.BWUnit = string(d.BWUnit)
		resp.BT = d.BodyTemperature
		resp.UsedShampoo = d.UsedShampoo
		resp.UsedRibbon = d.UsedRibbon
		resp.Remarks = d.Remarks
		resp.StyleImage = d.StyleImage
		resp.CompletedImage = d.CompletedImage

		if d.Course != nil {
			var price int64
			if d.Course.Price != nil {
				price = *d.Course.Price
			}
			resp.Course = &trimmingCourseSummaryResponse{
				ID:    d.Course.ID,
				Name:  d.Course.Name,
				Price: price,
			}
		}
		for i := range d.Options {
			resp.Options = append(resp.Options, trimmingOptionSummaryResponse{
				ID:   d.Options[i].ID,
				Name: d.Options[i].Name,
			})
		}
	}

	return resp
}
