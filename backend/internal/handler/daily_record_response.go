package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type vitalRecordResponse struct {
	ID              string    `json:"id"`
	DailyRecordID   string    `json:"daily_record_id"`
	Time            string    `json:"time"`
	Temperature     *float64  `json:"temperature,omitempty"`
	HeartRate       *int      `json:"heart_rate,omitempty"`
	RespirationRate *int      `json:"respiration_rate,omitempty"`
	Weight          *float64  `json:"weight,omitempty"`
	Notes           string    `json:"notes"`
	StaffID         *string   `json:"staff_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type careLogRecordResponse struct {
	ID            string    `json:"id"`
	DailyRecordID string    `json:"daily_record_id"`
	Time          string    `json:"time"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Value         string    `json:"value"`
	StaffID       *string   `json:"staff_id,omitempty"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

type staffNoteRecordResponse struct {
	ID            string    `json:"id"`
	DailyRecordID string    `json:"daily_record_id"`
	Time          string    `json:"time"`
	Content       string    `json:"content"`
	StaffID       *string   `json:"staff_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type dailyRecordResponse struct {
	ID                string                    `json:"id"`
	HospitalizationID string                    `json:"hospitalization_id"`
	Date              time.Time                 `json:"date"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	VitalRecords      []vitalRecordResponse     `json:"vital_records"`
	CareLogRecords    []careLogRecordResponse   `json:"care_logs"`
	StaffNoteRecords  []staffNoteRecordResponse `json:"staff_notes"`
}

func toVitalRecordResponse(vr *model.VitalRecord) vitalRecordResponse {
	dailyRecordID := ""
	if vr.DailyRecordID != nil {
		dailyRecordID = strconv.FormatUint(*vr.DailyRecordID, 10)
	}
	r := vitalRecordResponse{
		ID:              strconv.FormatUint(vr.ID, 10),
		DailyRecordID:   dailyRecordID,
		Time:            vr.RecordedAt.Format("15:04:05"),
		Temperature:     vr.Temperature,
		HeartRate:       vr.HeartRate,
		RespirationRate: vr.RespirationRate,
		Weight:          vr.Weight,
		Notes:           vr.Notes,
		CreatedAt:       vr.CreatedAt,
	}
	if vr.StaffID != nil {
		s := strconv.FormatUint(*vr.StaffID, 10)
		r.StaffID = &s
	}
	return r
}

func toCareLogRecordResponse(cr *model.CareLogRecord) careLogRecordResponse {
	r := careLogRecordResponse{
		ID:            strconv.FormatUint(cr.ID, 10),
		DailyRecordID: strconv.FormatUint(cr.DailyRecordID, 10),
		Time:          cr.Time.Format("15:04:05"),
		Type:          string(cr.Type),
		Status:        string(cr.Status),
		Value:         cr.Value,
		Notes:         cr.Notes,
		CreatedAt:     cr.CreatedAt,
	}
	if cr.StaffID != nil {
		s := strconv.FormatUint(*cr.StaffID, 10)
		r.StaffID = &s
	}
	return r
}

func toStaffNoteRecordResponse(sn *model.StaffNoteRecord) staffNoteRecordResponse {
	r := staffNoteRecordResponse{
		ID:            strconv.FormatUint(sn.ID, 10),
		DailyRecordID: strconv.FormatUint(sn.DailyRecordID, 10),
		Time:          sn.Time.Format("15:04:05"),
		Content:       sn.Content,
		CreatedAt:     sn.CreatedAt,
	}
	if sn.StaffID != nil {
		s := strconv.FormatUint(*sn.StaffID, 10)
		r.StaffID = &s
	}
	return r
}

func toDailyRecordResponse(dr *model.DailyRecord) dailyRecordResponse {
	vitals := make([]vitalRecordResponse, 0, len(dr.VitalRecords))
	for i := range dr.VitalRecords {
		vitals = append(vitals, toVitalRecordResponse(&dr.VitalRecords[i]))
	}

	careLogs := make([]careLogRecordResponse, 0, len(dr.CareLogRecords))
	for i := range dr.CareLogRecords {
		careLogs = append(careLogs, toCareLogRecordResponse(&dr.CareLogRecords[i]))
	}

	staffNotes := make([]staffNoteRecordResponse, 0, len(dr.StaffNoteRecords))
	for i := range dr.StaffNoteRecords {
		staffNotes = append(staffNotes, toStaffNoteRecordResponse(&dr.StaffNoteRecords[i]))
	}

	return dailyRecordResponse{
		ID:                strconv.FormatUint(dr.ID, 10),
		HospitalizationID: strconv.FormatUint(dr.HospitalizationID, 10),
		Date:              dr.Date,
		CreatedAt:         dr.CreatedAt,
		UpdatedAt:         dr.UpdatedAt,
		VitalRecords:      vitals,
		CareLogRecords:    careLogs,
		StaffNoteRecords:  staffNotes,
	}
}
