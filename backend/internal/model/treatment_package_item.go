package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TreatmentPackageItem は package 構成品行（EMR-518 §3.2 / EMR-521）。
// component_key は同一構成品の編集/並替で安定し、削除後別意味へ再利用しない。
// item_kind に応じて typed FK がちょうど 1 つ非 NULL（manual_other は全 FK NULL で
// 正規 manual schema の manual_definition が必須）。裸の polymorphic master_id は持たない。
// position/removed_at は (removed_at IS NULL) = (position IS NOT NULL) の CHECK 対。
// 独立 item 版は置かず、親 package.version を collection CAS とする。物理削除なし。
type TreatmentPackageItem struct {
	ID                uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	PackageID         uint64           `gorm:"column:package_id;type:bigint;not null" json:"package_id"`
	ComponentKey      uuid.UUID        `gorm:"column:component_key;type:uuid;not null" json:"component_key"`
	Position          *int             `gorm:"column:position;type:integer" json:"position"`
	RemovedAt         *time.Time       `gorm:"column:removed_at;type:timestamptz" json:"removed_at"`
	ItemKind          string           `gorm:"column:item_kind;type:text;not null" json:"item_kind"`
	ConsultationID    *uint64          `gorm:"column:consultation_id;type:bigint" json:"consultation_id"`
	ProcedureID       *uint64          `gorm:"column:procedure_id;type:bigint" json:"procedure_id"`
	MedicineID        *uint64          `gorm:"column:medicine_id;type:bigint" json:"medicine_id"`
	ExamTypeID        *uint64          `gorm:"column:exam_type_id;type:bigint" json:"exam_type_id"`
	VaccineID         *uint64          `gorm:"column:vaccine_id;type:bigint" json:"vaccine_id"`
	CheckupTypeID     *uint64          `gorm:"column:checkup_type_id;type:bigint" json:"checkup_type_id"`
	MerchandiseItemID *uint64          `gorm:"column:merchandise_item_id;type:bigint" json:"merchandise_item_id"`
	Quantity          ExactNumeric     `gorm:"column:quantity;type:numeric;not null" json:"quantity"`
	ManualDefinition  *json.RawMessage `gorm:"column:manual_definition;type:jsonb" json:"manual_definition"`
	CreatedAt         time.Time        `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
}

func (TreatmentPackageItem) TableName() string { return "treatment_package_items" }

// item_kind の 8 選択肢（EMR-518 §3.2 exact matrix）。
const (
	TreatmentPackageItemKindConsultation     = "consultation"
	TreatmentPackageItemKindProcedure        = "procedure"
	TreatmentPackageItemKindMedicine         = "medicine"
	TreatmentPackageItemKindExaminationType  = "examination_type"
	TreatmentPackageItemKindVaccine          = "vaccine"
	TreatmentPackageItemKindCheckupType      = "checkup_type"
	TreatmentPackageItemKindMerchandiseItem  = "merchandise_item"
	TreatmentPackageItemKindManualOther      = "manual_other"
)
