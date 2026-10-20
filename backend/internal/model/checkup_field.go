package model

import (
	"encoding/json"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
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
	ID              uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	CheckupTypeID   uint64           `gorm:"column:checkup_type_id;type:bigint;not null" json:"checkup_type_id"`
	Name            string           `gorm:"column:name;type:text;not null" json:"name"`
	FieldType       CheckupFieldType `gorm:"column:field_type;type:checkup_field_type;not null" json:"field_type"`
	Unit            string           `gorm:"column:unit;type:text;not null;default:''" json:"unit"`
	MinValue        *ExactNumeric    `gorm:"column:min_value;type:decimal(10,4)" json:"min_value,omitempty"`
	MaxValue        *ExactNumeric    `gorm:"column:max_value;type:decimal(10,4)" json:"max_value,omitempty"`
	Options         json.RawMessage  `gorm:"column:options;type:jsonb;not null;default:'[]'::jsonb" json:"options"`
	IsProvisional   bool             `gorm:"column:is_provisional;type:boolean;not null;default:false" json:"is_provisional"`
	SortOrder       int              `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	CreatedAt       time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time        `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	ImportNamespace *string          `gorm:"column:import_namespace;type:text" json:"import_namespace,omitempty"`
	ImportKey       *string          `gorm:"column:import_key;type:text" json:"import_key,omitempty"`
	Version         int64            `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
}

func (CheckupTypeField) TableName() string { return "checkup_type_fields" }

// CheckupFieldResult は健診記録（checkup）に紐づく型付き結果値（checkups の純粋従属子）。
// field 定義削除後も自己記述的であるよう field_name/field_type/unit を非正規化保持する
// （exam_results.exam_type_field_id と同型: nullable FK + snapshot）。
type CheckupFieldResult struct {
	ID                 uint64                  `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID           uint64                  `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	CheckupID          uint64                  `gorm:"column:checkup_id;type:bigint;not null" json:"checkup_id"`
	CheckupTypeFieldID *uint64                 `gorm:"column:checkup_type_field_id;type:bigint" json:"checkup_type_field_id,omitempty"`
	FieldName          string                  `gorm:"column:field_name;type:text;not null;default:''" json:"field_name"`
	FieldType          CheckupFieldType        `gorm:"column:field_type;type:checkup_field_type;not null" json:"field_type"`
	Unit               string                  `gorm:"column:unit;type:text;not null;default:''" json:"unit"`
	ValueNumber        *ExactNumeric           `gorm:"column:value_number;type:decimal(10,4)" json:"value_number,omitempty"`
	ValueText          string                  `gorm:"column:value_text;type:text;not null;default:''" json:"value_text"`
	ValueBool          *bool                   `gorm:"column:value_bool;type:boolean" json:"value_bool,omitempty"`
	ValueList          pq.StringArray          `gorm:"column:value_list;type:text[];not null;default:'{}'" json:"value_list"`
	RefMin             *ExactNumeric           `gorm:"column:ref_min;type:decimal(10,4)" json:"ref_min,omitempty"`
	RefMax             *ExactNumeric           `gorm:"column:ref_max;type:decimal(10,4)" json:"ref_max,omitempty"`
	IsAbnormal         bool                    `gorm:"column:is_abnormal;type:boolean;not null;default:false" json:"is_abnormal"`
	Status             ExaminationResultStatus `gorm:"column:status;type:exam_result_status;not null;default:'normal'" json:"status"`
	SortOrder          int                     `gorm:"column:sort_order;type:integer;not null;default:0" json:"sort_order"`
	CreatedAt          time.Time               `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time               `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	CheckupTypeField   *CheckupTypeField       `gorm:"foreignKey:CheckupTypeFieldID,ClinicID;references:ID,ClinicID" json:"checkup_type_field,omitempty"`
	Checkup            *Checkup                `gorm:"foreignKey:CheckupID;references:ID" json:"checkup,omitempty"`
}

func (CheckupFieldResult) TableName() string { return "checkup_field_results" }
