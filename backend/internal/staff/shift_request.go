package staff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

type listShiftEntriesQuery struct {
	Date    string
	StaffID string
}

func newListShiftEntriesQuery(values url.Values) listShiftEntriesQuery {
	return listShiftEntriesQuery{
		Date:    values.Get("date"),
		StaffID: values.Get("staff_id"),
	}
}

type listShiftEntriesFilters struct {
	YearMonth string
	StaffID   *uint64
}

type onDutyStaffsQuery struct {
	Date string
}

func newOnDutyStaffsQuery(values url.Values) onDutyStaffsQuery {
	return onDutyStaffsQuery{Date: values.Get("date")}
}

func (q onDutyStaffsQuery) toDate() (time.Time, error) {
	if q.Date == "" {
		return time.Time{}, fmt.Errorf("date query parameter is required (YYYY-MM-DD)")
	}
	date, err := time.ParseInLocation(time.DateOnly, q.Date, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: expected YYYY-MM-DD")
	}
	return date, nil
}

func (q listShiftEntriesQuery) toServiceFilters() (listShiftEntriesFilters, error) {
	staffID, err := parseOptionalUintQueryFilter(q.StaffID, "staff_id")
	if err != nil {
		return listShiftEntriesFilters{}, err
	}
	return listShiftEntriesFilters{
		YearMonth: q.Date,
		StaffID:   staffID,
	}, nil
}

// ---- 勤務（ShiftEntry）の作成・更新リクエスト ----
// api.yaml の CreateShiftRequest / UpdateShiftRequest（EMR-493 / EMR-549）に対応する。
// 旧 wire（start_time / end_time / break_start / break_end の時刻文字列）は受け付けず 400 で拒否する。

// minUpdateShiftRequestProperties は UpdateShiftRequest の minProperties
// （operation_key と expected_version に加えて業務 field を 1 つ以上指定する）。
const minUpdateShiftRequestProperties = 3

// legacyShiftRequestKeys は廃止した旧 wire のトップレベルキー。黙って無視せず 400 にする。
var legacyShiftRequestKeys = []string{"start_time", "end_time"}

// shiftOptionalTime は starts_at / ends_at の「キー無し」「明示 null」「date-time」を区別する。
// Set は JSON にキーがあったことを、Value=nil は明示 null を表す。
type shiftOptionalTime struct {
	Set   bool
	Value *time.Time
}

// UnmarshalJSON は RFC 3339 の date-time（オフセット付き）だけを受け付ける。
// オフセットの無い文字列は時刻から日付を推測することになるため拒否する。
func (o *shiftOptionalTime) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Value = nil
		return nil
	}
	var parsed time.Time
	if err := json.Unmarshal(data, &parsed); err != nil {
		return apperrors.WrapInvalidInput("starts_at and ends_at must be RFC 3339 date-time with an offset")
	}
	o.Value = &parsed
	return nil
}

// shiftBreakRequest は休憩の明示日時（api.yaml ShiftWorkBreak）。starts_at / ends_at は必須。
type shiftBreakRequest struct {
	StartsAt *time.Time `json:"starts_at" binding:"required"`
	EndsAt   *time.Time `json:"ends_at"   binding:"required"`
}

func toShiftBreakInputs(reqs []shiftBreakRequest) ([]ShiftBreakInput, error) {
	breaks := make([]ShiftBreakInput, 0, len(reqs))
	for _, req := range reqs {
		if req.StartsAt == nil || req.EndsAt == nil {
			return nil, apperrors.WrapInvalidInput("break starts_at and ends_at are required")
		}
		breaks = append(breaks, ShiftBreakInput{StartsAt: *req.StartsAt, EndsAt: *req.EndsAt})
	}
	return breaks, nil
}

// shiftRequestKeys は勤務リクエスト本文のトップレベルキー規則。
type shiftRequestKeys struct {
	allowed  []string
	required []string
	nonNull  []string
}

func (k shiftRequestKeys) isAllowed(key string) bool {
	return slices.Contains(k.allowed, key)
}

// createShiftRequestKeys は CreateShiftRequest の規則。全入力を明示させ、
// starts_at / ends_at だけが明示 null を許す（off / paid_leave の終日非勤務）。
func createShiftRequestKeys() shiftRequestKeys {
	all := []string{"operation_key", "staff_id", "date", "shift_type", "starts_at", "ends_at", "breaks", "notes"}
	return shiftRequestKeys{
		allowed:  all,
		required: all,
		nonNull:  []string{"operation_key", "staff_id", "date", "shift_type", "breaks", "notes"},
	}
}

// updateShiftRequestKeys は UpdateShiftRequest の規則。明示 null を許すのは starts_at / ends_at だけで、
// shift_type / breaks / notes の null は 400。staff_id / date / clinic_id の付替えは未知キーとして拒否する。
func updateShiftRequestKeys() shiftRequestKeys {
	return shiftRequestKeys{
		allowed:  []string{"operation_key", "expected_version", "shift_type", "starts_at", "ends_at", "breaks", "notes"},
		required: []string{"operation_key", "expected_version"},
		nonNull:  []string{"operation_key", "expected_version", "shift_type", "breaks", "notes"},
	}
}

// decodeShiftRequestBody は本文を、旧 wire キー・未知キーの拒否、必須キーの存在、null 禁止キーの順に検査してから dst へ復号する。
// 復号は未知フィールドを拒否するため、休憩要素の旧キー（break_start / break_end）も 400 になる。本文のキー集合を返す。
func decodeShiftRequestBody(data []byte, keys shiftRequestKeys, dst any) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, apperrors.WrapInvalidInput(staffJSONInvalidMessage)
	}
	for _, legacy := range legacyShiftRequestKeys {
		if _, found := raw[legacy]; found {
			return nil, apperrors.WrapInvalidInput(legacy + " is no longer supported; use starts_at and ends_at")
		}
	}
	for key := range raw {
		if !keys.isAllowed(key) {
			return nil, apperrors.WrapInvalidInput("request body contains an unsupported field")
		}
	}
	for _, key := range keys.required {
		if _, found := raw[key]; !found {
			return nil, apperrors.WrapInvalidInput(key + " is required")
		}
	}
	for _, key := range keys.nonNull {
		if value, found := raw[key]; found && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, apperrors.WrapInvalidInput(key + " must not be null")
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, apperrors.Wrap(err, "invalid shift request body")
		}
		return nil, apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
	}
	return raw, nil
}

// createShiftRequest はシフト登録リクエスト（api.yaml CreateShiftRequest）。全入力を明示させる。
type createShiftRequest struct {
	OperationKey string               `json:"operation_key" binding:"required,max=36"`
	StaffID      string               `json:"staff_id"      binding:"required,max=19"`
	Date         string               `json:"date"          binding:"required,max=10"`
	ShiftType    string               `json:"shift_type"    binding:"required,max=10,oneof=full morning afternoon off paid_leave"`
	StartsAt     shiftOptionalTime    `json:"starts_at"`
	EndsAt       shiftOptionalTime    `json:"ends_at"`
	Breaks       *[]shiftBreakRequest `json:"breaks"        binding:"required,max=50,dive"`
	Notes        *string              `json:"notes"         binding:"required,max=2000"`
}

// UnmarshalJSON は本文を shift 契約のキー規則で検査して復号する。
func (r *createShiftRequest) UnmarshalJSON(data []byte) error {
	type plain createShiftRequest
	var decoded plain
	if _, err := decodeShiftRequestBody(data, createShiftRequestKeys(), &decoded); err != nil {
		return err
	}
	*r = createShiftRequest(decoded)
	return nil
}

// parseShiftStaffID は staff_id（正数十進文字列、1..9223372036854775807）を厳密に解析する。
// 先頭ゼロ・符号・空白・範囲外は 400 にする。JS Number へ丸められた値を受け付けるための補正はしない。
func parseShiftStaffID(raw string) (uint64, error) {
	if raw == "" || raw[0] < '1' || raw[0] > '9' {
		return 0, apperrors.WrapInvalidInput("staff_id must be a positive decimal string")
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '9' {
			return 0, apperrors.WrapInvalidInput("staff_id must be a positive decimal string")
		}
	}
	id, err := strconv.ParseUint(raw, 10, 63)
	if err != nil {
		return 0, apperrors.WrapInvalidInput("staff_id is out of range")
	}
	return id, nil
}

// toServiceInput は検証済みの作成リクエストを service 入力へ写す。時刻から日付や翌日を推測しない。
func (r *createShiftRequest) toServiceInput() (*CreateShiftEntryInput, error) {
	staffID, err := parseShiftStaffID(r.StaffID)
	if err != nil {
		return nil, err
	}
	date, err := time.ParseInLocation(time.DateOnly, r.Date, config.JST)
	if err != nil {
		return nil, fmt.Errorf("invalid date: use YYYY-MM-DD")
	}
	if !r.StartsAt.Set || !r.EndsAt.Set {
		return nil, apperrors.WrapInvalidInput("starts_at and ends_at must be specified explicitly")
	}
	if r.Breaks == nil || r.Notes == nil {
		return nil, apperrors.WrapInvalidInput("breaks and notes must be specified explicitly")
	}
	breaks, err := toShiftBreakInputs(*r.Breaks)
	if err != nil {
		return nil, err
	}
	return &CreateShiftEntryInput{
		StaffID:   staffID,
		Date:      date,
		ShiftType: r.ShiftType,
		StartsAt:  r.StartsAt.Value,
		EndsAt:    r.EndsAt.Value,
		Notes:     *r.Notes,
		Breaks:    breaks,
	}, nil
}

// updateShiftRequest はシフト更新リクエスト（PATCH, api.yaml UpdateShiftRequest）。
type updateShiftRequest struct {
	OperationKey    string               `json:"operation_key"    binding:"required,max=36"`
	ExpectedVersion int64                `json:"expected_version" binding:"required,min=1"`
	ShiftType       *string              `json:"shift_type"       binding:"omitempty,max=10,oneof=full morning afternoon off paid_leave"`
	StartsAt        shiftOptionalTime    `json:"starts_at"`
	EndsAt          shiftOptionalTime    `json:"ends_at"`
	Notes           *string              `json:"notes"            binding:"omitempty,max=2000"`
	Breaks          *[]shiftBreakRequest `json:"breaks"           binding:"omitempty,max=50,dive"`
}

// UnmarshalJSON は本文を shift 契約のキー規則で検査して復号する。
func (r *updateShiftRequest) UnmarshalJSON(data []byte) error {
	type plain updateShiftRequest
	var decoded plain
	raw, err := decodeShiftRequestBody(data, updateShiftRequestKeys(), &decoded)
	if err != nil {
		return err
	}
	// minProperties: 3（operation_key と expected_version 以外に業務 field を 1 つ以上指定する）
	if len(raw) < minUpdateShiftRequestProperties {
		return apperrors.WrapInvalidInput("at least one shift field must be provided")
	}
	*r = updateShiftRequest(decoded)
	return nil
}

// toServiceInput は検証済みの更新リクエストを service 入力へ写す。未指定のキーは保存値を維持する。
func (r *updateShiftRequest) toServiceInput() (*UpdateShiftEntryInput, error) {
	input := &UpdateShiftEntryInput{
		ShiftType: r.ShiftType,
		StartsAt:  OptionalShiftTime(r.StartsAt),
		EndsAt:    OptionalShiftTime(r.EndsAt),
		Notes:     r.Notes,
	}
	expectedVersion := r.ExpectedVersion
	input.ExpectedVersion = &expectedVersion
	if r.Breaks != nil {
		breaks, err := toShiftBreakInputs(*r.Breaks)
		if err != nil {
			return nil, err
		}
		input.Breaks = &breaks
	}
	return input, nil
}
