package trimming

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// FE7-1: tygo codegen 対象にするため export（BackendTrimming の生成契約ゲート化・FE-refactor.md
// BackendTrimming BLOCKED の解消）。JSON wire 形状は json タグで完全維持しており不変。
type TrimmingOptionSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type TrimmingCourseSummaryResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

// TrimmingResponse は appointments ベースのフラット DTO（BE-119）
type TrimmingResponse struct {
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
	HasDetail      bool      `json:"has_detail"`
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
	// FE7-2: Pet/Staff は petSummaryResponse/staffSummaryResponse（11+18箇所で共有される
	// 非公開型）を参照しており tygo が解決できないため tstype:"-" で生成対象から除外する
	// （JSON wire 形状・json タグは無変更 — 実際のレスポンスには pet/staff は引き続き含まれる。
	// 生成型 TrimmingResponse のみこの2フィールドを欠く）。
	Pet     *petSummaryResponse             `json:"pet,omitempty" tstype:"-"`
	Staff   *staffSummaryResponse           `json:"staff,omitempty" tstype:"-"`
	Course  *TrimmingCourseSummaryResponse  `json:"course,omitempty"`
	Options []TrimmingOptionSummaryResponse `json:"options"`
}

func toTrimmingResponse(appt *model.Reservation) TrimmingResponse {
	resp := TrimmingResponse{
		ID:                appt.ID,
		ClinicID:          appt.ClinicID,
		ReservationTypeID: appt.ReservationTypeID,
		StartTime:         localTime(appt.StartTime),
		EndTime:           localTime(appt.EndTime),
		PetID:             appt.PetID,
		StaffID:           appt.DoctorID,
		Status:            string(appt.Status),
		Source:            string(appt.Source),
		CreatedAt:         localTime(appt.CreatedAt),
		UpdatedAt:         localTime(appt.UpdatedAt),
		Pet:               toPetSummary(appt.Pet),
		Staff:             toStaffSummary(appt.Doctor),
		Options:           make([]TrimmingOptionSummaryResponse, 0),
		// TrimmingDetail が nil の異常データ向けデフォルト（モデルデフォルトと一致）
		BWUnit: "Kg",
	}

	if appt.TrimmingDetail != nil {
		d := appt.TrimmingDetail
		resp.HasDetail = true
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
			resp.Course = &TrimmingCourseSummaryResponse{
				ID:    d.Course.ID,
				Name:  d.Course.Name,
				Price: price,
			}
		}
		for i := range d.Options {
			resp.Options = append(resp.Options, TrimmingOptionSummaryResponse{
				ID:   d.Options[i].ID,
				Name: d.Options[i].Name,
			})
		}
	}

	return resp
}
