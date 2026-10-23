package clinic

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/animal-ekarte/backend/internal/model"
)

// 共通医院時間 GET の wire 型（api.yaml ClinicOperatingHoursResponse / ClinicCommonHolidayMonthResponse）。
// ID と版は正規十進文字列。unconfigured は version/settings を null にする。

const (
	operatingHoursStateConfigured   = "configured"
	operatingHoursStateUnconfigured = "unconfigured"
)

// ClinicOperatingHoursResponse is GET /clinics/{clinic_id}/operating-hours.
type ClinicOperatingHoursResponse struct {
	SchemaVersion string          `json:"schema_version"`
	ClinicID      string          `json:"clinic_id"`
	State         string          `json:"state"`
	Version       *string         `json:"version"`
	Settings      json.RawMessage `json:"settings"`
}

// ClinicClosedDateResponse is one individual closed date.
type ClinicClosedDateResponse struct {
	Date   string `json:"date"`
	Reason string `json:"reason"`
}

// ClinicCommonHolidayMonthResponse is GET /clinic-holidays/common and GET /closing-settings/holidays/common.
// 月内の休診日と common 版だけで、全営業時間・曜日・祝日方針は含めない。
type ClinicCommonHolidayMonthResponse struct {
	ClinicID    string                     `json:"clinic_id"`
	YearMonth   string                     `json:"year_month"`
	State       string                     `json:"state"`
	Version     *string                    `json:"version"`
	ClosedDates []ClinicClosedDateResponse `json:"closed_dates"`
}

func versionText(current CommonOperatingHours) *string {
	if !current.Configured {
		return nil
	}
	v := strconv.FormatInt(current.Version, 10)
	return &v
}

// NewClinicOperatingHoursResponse maps the current state to the wire. Settings are encoded only by
// OperatingHoursSettings.MarshalJSON (the canonical bytes); an invalid configured value is an error,
// never an "all closed" success.
func NewClinicOperatingHoursResponse(clinicID uint64, current CommonOperatingHours) (ClinicOperatingHoursResponse, error) {
	resp := ClinicOperatingHoursResponse{
		SchemaVersion: model.ClinicOperatingHoursSchemaV1,
		ClinicID:      strconv.FormatUint(clinicID, 10),
		State:         operatingHoursStateUnconfigured,
		Version:       nil,
		Settings:      json.RawMessage("null"),
	}
	if !current.Configured {
		return resp, nil
	}
	if current.Version <= 0 {
		return ClinicOperatingHoursResponse{}, errors.New("configured operating hours must carry a positive version")
	}
	settings, err := current.Settings.MarshalJSON()
	if err != nil {
		return ClinicOperatingHoursResponse{}, err
	}
	resp.State = operatingHoursStateConfigured
	resp.Version = versionText(current)
	resp.Settings = json.RawMessage(settings)
	return resp, nil
}

// NewClinicCommonHolidayMonthResponse maps the current state and month to the wire.
func NewClinicCommonHolidayMonthResponse(clinicID uint64, month YearMonth, current CommonOperatingHours) (ClinicCommonHolidayMonthResponse, error) {
	if current.Configured && current.Version <= 0 {
		return ClinicCommonHolidayMonthResponse{}, errors.New("configured operating hours must carry a positive version")
	}
	dates, err := current.ClosedDatesIn(month)
	if err != nil {
		return ClinicCommonHolidayMonthResponse{}, err
	}
	state := operatingHoursStateUnconfigured
	if current.Configured {
		state = operatingHoursStateConfigured
	}
	resp := ClinicCommonHolidayMonthResponse{
		ClinicID:    strconv.FormatUint(clinicID, 10),
		YearMonth:   month.String(),
		State:       state,
		Version:     versionText(current),
		ClosedDates: make([]ClinicClosedDateResponse, 0, len(dates)),
	}
	for _, cd := range dates {
		resp.ClosedDates = append(resp.ClosedDates, ClinicClosedDateResponse{Date: cd.Date.String(), Reason: cd.Reason})
	}
	return resp, nil
}
