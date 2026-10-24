package staff

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// ShiftMutationReceipt.intent の値。
const (
	shiftIntentCreate = "create"
	shiftIntentUpdate = "update"
)

// shiftBreakResponse は休憩のレスポンス（api.yaml ShiftWorkBreakView）。
// 区間が未解決（NULL を含む）でも判明している日時は保持し、interval_state=unknown で示す。
type shiftBreakResponse struct {
	StartsAt      *string `json:"starts_at"`
	EndsAt        *string `json:"ends_at"`
	IntervalState string  `json:"interval_state"`
}

// shiftResponse は勤務のレスポンス（api.yaml ShiftEntry）。start_time / end_time や DB 行は返さない。
// date は勤務開始側（starts_at）の JST 日付で、日跨ぎ勤務を分割・自動補完しない。
type shiftResponse struct {
	ID            string               `json:"id"`
	ClinicID      string               `json:"clinic_id"`
	StaffID       string               `json:"staff_id"`
	StaffName     string               `json:"staff_name,omitempty"`
	Date          string               `json:"date"`
	ShiftType     model.ShiftType      `json:"shift_type"`
	StartsAt      *string              `json:"starts_at"`
	EndsAt        *string              `json:"ends_at"`
	Version       int64                `json:"version"`
	IntervalState string               `json:"interval_state"`
	Notes         string               `json:"notes"`
	Breaks        []shiftBreakResponse `json:"breaks"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

// shiftMutationReceiptResponse は作成・更新の確定結果（api.yaml ShiftMutationReceipt）。
// 冪等再送（同 key の確定済み receipt の再返却）は未実装のため、この操作自身の結果だけを返す。
type shiftMutationReceiptResponse struct {
	OperationKey     string        `json:"operation_key"`
	Intent           string        `json:"intent"`
	ShiftID          string        `json:"shift_id"`
	ResultingVersion int64         `json:"resulting_version"`
	Shift            shiftResponse `json:"shift"`
}

// shiftTimestampOrNil は明示日時を JST の RFC 3339 で返す。nil は nil（JSON null）。
func shiftTimestampOrNil(ts *time.Time) *string {
	if ts == nil {
		return nil
	}
	formatted := ts.In(config.JST).Format(time.RFC3339Nano)
	return &formatted
}

// shiftResponseDate は応答の date を返す。勤務は starts_at の JST 暦日、
// 終日非勤務・区間未解決（starts_at が NULL）は保存された勤務日。
func shiftResponseDate(s *model.ShiftEntry) string {
	if s.StartsAt != nil && sharedkernel.RequiresTimeSlot(s.ShiftType) {
		return s.StartsAt.In(config.JST).Format(time.DateOnly)
	}
	return s.Date.In(config.JST).Format(time.DateOnly)
}

func toShiftResponse(s *model.ShiftEntry) shiftResponse {
	breaks := make([]shiftBreakResponse, 0, len(s.Breaks))
	for _, b := range s.Breaks {
		breaks = append(breaks, shiftBreakResponse{
			StartsAt:      shiftTimestampOrNil(b.StartsAt),
			EndsAt:        shiftTimestampOrNil(b.EndsAt),
			IntervalState: sharedkernel.ShiftBreakIntervalState(b),
		})
	}
	r := shiftResponse{
		ID:            strconv.FormatUint(s.ID, 10),
		ClinicID:      strconv.FormatUint(s.ClinicID, 10),
		StaffID:       strconv.FormatUint(s.StaffID, 10),
		Date:          shiftResponseDate(s),
		ShiftType:     s.ShiftType,
		StartsAt:      shiftTimestampOrNil(s.StartsAt),
		EndsAt:        shiftTimestampOrNil(s.EndsAt),
		Version:       s.Version,
		IntervalState: sharedkernel.ShiftEntryIntervalState(s, s.Breaks),
		Notes:         s.Notes,
		Breaks:        breaks,
		CreatedAt:     localTimeRFC3339(s.CreatedAt),
		UpdatedAt:     localTimeRFC3339(s.UpdatedAt),
	}
	if s.Staff != nil && s.Staff.ID != 0 {
		r.StaffName = s.Staff.Name
	}
	return r
}

// toShiftMutationReceipt は作成・更新の確定結果を ShiftMutationReceipt へ写す。
func toShiftMutationReceipt(operationKey, intent string, s *model.ShiftEntry) shiftMutationReceiptResponse {
	return shiftMutationReceiptResponse{
		OperationKey:     operationKey,
		Intent:           intent,
		ShiftID:          strconv.FormatUint(s.ID, 10),
		ResultingVersion: s.Version,
		Shift:            toShiftResponse(s),
	}
}
