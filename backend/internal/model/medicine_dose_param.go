package model

import (
	"gorm.io/gorm"
	"time"
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
	ID                  uint64                `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID            uint64                `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	MedicineID          uint64                `gorm:"column:medicine_id;type:bigint;not null" json:"medicine_id"`
	Species             MedicineDoseSpecies   `gorm:"column:species;type:medicine_dose_species;not null" json:"species"`
	DoseBasis           MedicineDoseBasis     `gorm:"column:dose_basis;type:medicine_dose_basis;not null;default:'per_administration'" json:"dose_basis"`
	DosePerKg           ExactNumeric          `gorm:"column:dose_per_kg;type:numeric(10,6);not null" json:"dose_per_kg"`
	MinMgPerKg          *ExactNumeric         `gorm:"column:min_mg_per_kg;type:numeric(10,6)" json:"min_mg_per_kg,omitempty"`
	MaxMgPerKg          *ExactNumeric         `gorm:"column:max_mg_per_kg;type:numeric(10,6)" json:"max_mg_per_kg,omitempty"`
	AbsoluteMaxDose     *ExactNumeric         `gorm:"column:absolute_max_dose;type:numeric(10,4)" json:"absolute_max_dose,omitempty"`
	RoundingStep        *ExactNumeric         `gorm:"column:rounding_step;type:numeric(10,4)" json:"rounding_step,omitempty"`
	RoundingMode        *MedicineRoundingMode `gorm:"column:rounding_mode;type:medicine_rounding_mode" json:"rounding_mode,omitempty"`
	Notes               string                `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	CreatedAt           time.Time             `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time             `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt        `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	PatternName         *string               `gorm:"column:pattern_name;type:text" json:"-"`
	Purpose             *string               `gorm:"column:purpose;type:text" json:"-"`
	AdministrationRoute *string               `gorm:"column:administration_route;type:text" json:"-"`
	Version             int64                 `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	Provenance          *string               `gorm:"column:provenance;type:text" json:"-"`
}

func (MedicineDoseParam) TableName() string { return "medicine_dose_params" }
