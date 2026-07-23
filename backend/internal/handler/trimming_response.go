package handler

import (
	"time"
)

// BE9-2E codegen compatibility carriers: tygo still pins this legacy file path.
// Move that configuration to internal/trimming and delete these aliases in BE9-2F.
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
