package model

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CheckupFieldType は健診パッケージのフィールド型（#211・6種）。
type CheckupFieldType string

const (
	CheckupFieldTypeNumber       CheckupFieldType = "number"
	CheckupFieldTypeSingleSelect CheckupFieldType = "single_select"
	CheckupFieldTypeMultiSelect  CheckupFieldType = "multi_select"
	CheckupFieldTypeBoolean      CheckupFieldType = "boolean"
	CheckupFieldTypeChecklist    CheckupFieldType = "checklist"
	CheckupFieldTypeText         CheckupFieldType = "text"
)

// IsValid は field_type が既知の値かを返す（request 境界の検証用）。
func (t CheckupFieldType) IsValid() bool {
	switch t {
	case CheckupFieldTypeNumber, CheckupFieldTypeSingleSelect, CheckupFieldTypeMultiSelect,
		CheckupFieldTypeBoolean, CheckupFieldTypeChecklist, CheckupFieldTypeText:
		return true
	default:
		return false
	}
}

// CheckupTypeField は健診パッケージ（checkup_type）の型付きフィールド定義マスタ。
// examination の exam_type_fields に相当するが、field_type / options / 異常値基準を持つ。
type CheckupTypeField struct {
	ID              uint64           `gorm:"primaryKey;autoIncrement"            json:"id"`
	ClinicID        uint64           `gorm:"not null"                            json:"clinic_id"`
	CheckupTypeID   uint64           `gorm:"not null"                            json:"checkup_type_id"`
	Name            string           `gorm:"not null"                            json:"name"`
	FieldType       CheckupFieldType `gorm:"type:checkup_field_type;not null"    json:"field_type"`
	Unit            string           `gorm:"not null;default:''"                 json:"unit"`
	MinValue        *float64         `gorm:"type:decimal(10,4)"                  json:"min_value,omitempty"`
	MaxValue        *float64         `gorm:"type:decimal(10,4)"                  json:"max_value,omitempty"`
	Options         datatypes.JSON   `gorm:"type:jsonb;not null;default:'[]'"    json:"options"`
	IsProvisional   bool             `gorm:"not null;default:false"              json:"is_provisional"`
	// ImportNamespace / ImportKey are nullable stable keys for versioned package import (TASK-374).
	ImportNamespace *string          `gorm:"column:import_namespace"             json:"import_namespace,omitempty"`
	ImportKey       *string          `gorm:"column:import_key"                   json:"import_key,omitempty"`
	SortOrder       int              `gorm:"type:integer;default:0"              json:"sort_order"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"                      json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"                      json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `                                           json:"-"`
}

func (CheckupTypeField) TableName() string { return "checkup_type_fields" }

// CheckupFieldResult は健診記録（checkup）に紐づく型付き結果値（checkups の純粋従属子）。
// field 定義削除後も自己記述的であるよう field_name/field_type/unit を非正規化保持する
// （exam_results.exam_type_field_id と同型: nullable FK + snapshot）。
type CheckupFieldResult struct {
	ID                 uint64                  `gorm:"primaryKey;autoIncrement"                 json:"id"`
	ClinicID           uint64                  `gorm:"not null"                                 json:"clinic_id"`
	CheckupID          uint64                  `gorm:"not null"                                 json:"checkup_id"`
	CheckupTypeFieldID *uint64                 `                                                json:"checkup_type_field_id,omitempty"`
	FieldName          string                  `gorm:"not null;default:''"                      json:"field_name"`
	FieldType          CheckupFieldType        `gorm:"type:checkup_field_type;not null"         json:"field_type"`
	Unit               string                  `gorm:"not null;default:''"                      json:"unit"`
	ValueNumber        *float64                `gorm:"type:decimal(10,4)"                       json:"value_number,omitempty"`
	ValueText          string                  `gorm:"not null;default:''"                      json:"value_text"`
	ValueBool          *bool                   `                                                json:"value_bool,omitempty"`
	ValueList          pq.StringArray          `gorm:"type:text[];not null;default:'{}'"        json:"value_list"`
	RefMin             *float64                `gorm:"type:decimal(10,4)"                       json:"ref_min,omitempty"`
	RefMax             *float64                `gorm:"type:decimal(10,4)"                       json:"ref_max,omitempty"`
	IsAbnormal         bool                    `gorm:"default:false"                            json:"is_abnormal"`
	Status             ExaminationResultStatus `gorm:"type:exam_result_status;default:'normal'" json:"status"`
	SortOrder          int                     `gorm:"type:integer;default:0"                   json:"sort_order"`
	CreatedAt          time.Time               `gorm:"autoCreateTime"                            json:"created_at"`
	UpdatedAt          time.Time               `gorm:"autoUpdateTime"                            json:"updated_at"`

	// Relations — 関連名はモデル名と一致させる（P3.1 read lint の assoc キーが
	// clinicScopedMasterAssoc["CheckupTypeField"] で解決されるため）。
	CheckupTypeField *CheckupTypeField `gorm:"foreignKey:CheckupTypeFieldID" json:"checkup_type_field,omitempty"`
	// 飼い主レポート用に親 checkup（日付・種別）を解決する。Checkup は記録であり
	// マスタではないため P3.1 lint 対象外だが、ネストの CheckupType は対象（clinic_id 述語を付ける）。
	Checkup *Checkup `gorm:"foreignKey:CheckupID" json:"checkup,omitempty"`
}

func (CheckupFieldResult) TableName() string { return "checkup_field_results" }
