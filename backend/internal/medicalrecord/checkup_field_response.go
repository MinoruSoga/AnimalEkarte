package medicalrecord

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// checkupTypeFieldResponse は健診パッケージのフィールド定義レスポンス（FE 動的フォーム用）。
type checkupTypeFieldResponse struct {
	ID            uint64          `json:"id"`
	CheckupTypeID uint64          `json:"checkup_type_id"`
	Name          string          `json:"name"`
	FieldType     string          `json:"field_type"`
	Unit          string          `json:"unit"`
	MinValue      *float64        `json:"min_value,omitempty"`
	MaxValue      *float64        `json:"max_value,omitempty"`
	Options       json.RawMessage `json:"options"`
	IsProvisional bool            `json:"is_provisional"`
	SortOrder     int             `json:"sort_order"`
}

func toCheckupTypeFieldResponse(f *model.CheckupTypeField) checkupTypeFieldResponse {
	opts := json.RawMessage(f.Options)
	if len(opts) == 0 {
		opts = json.RawMessage("[]")
	}
	return checkupTypeFieldResponse{
		ID:            f.ID,
		CheckupTypeID: f.CheckupTypeID,
		Name:          f.Name,
		FieldType:     string(f.FieldType),
		Unit:          f.Unit,
		MinValue:      f.MinValue,
		MaxValue:      f.MaxValue,
		Options:       opts,
		IsProvisional: f.IsProvisional,
		SortOrder:     f.SortOrder,
	}
}

// checkupFieldResultResponse は健診結果値レスポンス。
type checkupFieldResultResponse struct {
	ID                 uint64   `json:"id"`
	CheckupID          uint64   `json:"checkup_id"`
	CheckupTypeFieldID *uint64  `json:"checkup_type_field_id,omitempty"`
	FieldName          string   `json:"field_name"`
	FieldType          string   `json:"field_type"`
	Unit               string   `json:"unit"`
	ValueNumber        *float64 `json:"value_number,omitempty"`
	ValueText          string   `json:"value_text"`
	ValueBool          *bool    `json:"value_bool,omitempty"`
	ValueList          []string `json:"value_list"`
	IsAbnormal         bool     `json:"is_abnormal"`
	Status             string   `json:"status"`
	SortOrder          int      `json:"sort_order"`
}

func toCheckupFieldResultResponse(r *model.CheckupFieldResult) checkupFieldResultResponse {
	list := []string(r.ValueList)
	if list == nil {
		list = []string{}
	}
	return checkupFieldResultResponse{
		ID:                 r.ID,
		CheckupID:          r.CheckupID,
		CheckupTypeFieldID: r.CheckupTypeFieldID,
		FieldName:          r.FieldName,
		FieldType:          string(r.FieldType),
		Unit:               r.Unit,
		ValueNumber:        r.ValueNumber,
		ValueText:          r.ValueText,
		ValueBool:          r.ValueBool,
		ValueList:          list,
		IsAbnormal:         r.IsAbnormal,
		Status:             string(r.Status),
		SortOrder:          r.SortOrder,
	}
}

// petCheckupResultResponse は飼い主レポート用に親 checkup の日付・パッケージ名を付与した結果レスポンス。
type petCheckupResultResponse struct {
	checkupFieldResultResponse
	Date            string `json:"date"`
	CheckupTypeID   uint64 `json:"checkup_type_id"`
	CheckupTypeName string `json:"checkup_type_name"`
}

func toPetCheckupResultResponse(r *model.CheckupFieldResult) petCheckupResultResponse {
	resp := petCheckupResultResponse{checkupFieldResultResponse: toCheckupFieldResultResponse(r)}
	if r.Checkup != nil {
		resp.Date = r.Checkup.Date.In(time.Local).Format(time.DateOnly)
		resp.CheckupTypeID = r.Checkup.CheckupTypeID
		if r.Checkup.CheckupType != nil {
			resp.CheckupTypeName = r.Checkup.CheckupType.Name
		}
	}
	return resp
}
