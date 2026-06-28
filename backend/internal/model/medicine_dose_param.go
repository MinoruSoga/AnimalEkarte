package model

import (
	"time"

	"gorm.io/gorm"
)

// MedicineCalculationType は投与量計算方式（#201）。
// none = 手動（既定・default-deny）/ per_weight = mg/kg 線形自動計算。
// 非線形（CRI/IU/濃度/BSA 等）は将来この ENUM を拡張して名前付き計算式を追加する。
// 自由入力式は不採用（任意コード実行・型非安全・式ミスの全患者波及）。
type MedicineCalculationType string

const (
	MedicineCalculationTypeNone      MedicineCalculationType = "none"
	MedicineCalculationTypePerWeight MedicineCalculationType = "per_weight"
)

// MedicineDoseBasis は dose_per_kg の基準。
// per_administration = 1回投与あたりの mg/kg / per_day = 1日あたりの mg/kg（1回量は frequency_per_day で按分）。
type MedicineDoseBasis string

const (
	MedicineDoseBasisPerAdministration MedicineDoseBasis = "per_administration"
	MedicineDoseBasisPerDay            MedicineDoseBasis = "per_day"
)

// MedicineRoundingMode は丸め方向。臨床ソースは丸め規則を定義しないため運用前提（NULL=丸めなし）。
type MedicineRoundingMode string

const (
	MedicineRoundingModeUp      MedicineRoundingMode = "up"
	MedicineRoundingModeDown    MedicineRoundingMode = "down"
	MedicineRoundingModeNearest MedicineRoundingMode = "nearest"
)

// MedicineDoseSpecies は計算対象の患者種。mg/kg は犬・猫で網羅的に異なり 'both' は持たない
// （vaccine_species と区別）。free-text animal_species から正規化し、マップ不能種は fail-closed。
type MedicineDoseSpecies string

const (
	MedicineDoseSpeciesDog MedicineDoseSpecies = "dog"
	MedicineDoseSpeciesCat MedicineDoseSpecies = "cat"
)

// validMedicineDoseSpecies は許可された種の集合。新しい種を追加したらここにも追加すること。
var validMedicineDoseSpecies = map[MedicineDoseSpecies]struct{}{
	MedicineDoseSpeciesDog: {},
	MedicineDoseSpeciesCat: {},
}

// ValidMedicineDoseSpecies は s が許可済み定数のいずれかと一致するか検証する。
// 任意の文字列キャスト（MedicineDoseSpecies("rabbit")）は false を返す。
func ValidMedicineDoseSpecies(s MedicineDoseSpecies) bool {
	if s == "" {
		return false
	}
	_, ok := validMedicineDoseSpecies[s]
	return ok
}

// MedicineDoseParam は薬剤 × 種の体重あたり投与量パラメータ（#201 per_weight 自動計算用）。
// strength は製品軸（medicines）、dose_per_kg は種軸（この子テーブル）に分離する。
type MedicineDoseParam struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID   uint64 `gorm:"not null"                 json:"clinic_id"` // 親 medicines から非正規化（clinicScope 直適用）
	MedicineID uint64 `gorm:"not null"                json:"medicine_id"`

	Species   MedicineDoseSpecies `gorm:"type:medicine_dose_species;not null"                    json:"species"`
	DoseBasis MedicineDoseBasis   `gorm:"type:medicine_dose_basis;not null;default:per_administration" json:"dose_basis"`

	DosePerKg       float64  `gorm:"type:numeric(10,6);not null" json:"dose_per_kg"`                 // mg/kg
	MinMgPerKg      *float64 `gorm:"type:numeric(10,6)"          json:"min_mg_per_kg,omitempty"`     // 安全域下限
	MaxMgPerKg      *float64 `gorm:"type:numeric(10,6)"          json:"max_mg_per_kg,omitempty"`     // 体重連動上限
	AbsoluteMaxDose *float64 `gorm:"type:numeric(10,4)"          json:"absolute_max_dose,omitempty"` // 体重非依存 mg/head 上限

	RoundingStep *float64              `gorm:"type:numeric(10,4)"     json:"rounding_step,omitempty"` // NULL=丸めなし
	RoundingMode *MedicineRoundingMode `gorm:"type:medicine_rounding_mode" json:"rounding_mode,omitempty"`

	Notes     string         `gorm:"not null;default:''" json:"notes"`
	CreatedAt time.Time      `gorm:"autoCreateTime"      json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"      json:"updated_at"`
	DeletedAt gorm.DeletedAt `                           json:"-"`
}

func (MedicineDoseParam) TableName() string { return "medicine_dose_params" }
