package medicalrecord

import (
	"fmt"
	"time"
)

const dailyRecordTimeLayout = "15:04:05"

// addVitalRecordRequest はバイタル記録追加のバインド struct
type addVitalRecordRequest struct {
	Time            string   `json:"time"             binding:"required"`
	Temperature     *float64 `json:"temperature"`
	HeartRate       *int     `json:"heart_rate"`
	RespirationRate *int     `json:"respiration_rate"`
	Weight          *float64 `json:"weight"`
	Notes           string   `json:"notes"            binding:"max=1000"`
	StaffID         *uint64  `json:"staff_id"`
}

func (r addVitalRecordRequest) toServiceInput() (*CreateVitalRecordInput, error) {
	recordedAt, err := time.ParseInLocation(dailyRecordTimeLayout, r.Time, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid time format, expected HH:MM:SS")
	}

	return &CreateVitalRecordInput{
		Time:            recordedAt,
		Temperature:     r.Temperature,
		HeartRate:       r.HeartRate,
		RespirationRate: r.RespirationRate,
		Weight:          r.Weight,
		Notes:           r.Notes,
		StaffID:         r.StaffID,
	}, nil
}

// addCareLogRequest はケアログ記録追加のバインド struct
type addCareLogRequest struct {
	Time    string  `json:"time"    binding:"required"`
	Type    string  `json:"type"    binding:"required,oneof=food excretion medicine treatment other"`
	Status  string  `json:"status"  binding:"omitempty,oneof=completed partial skipped"`
	Value   string  `json:"value"   binding:"max=1000"`
	StaffID *uint64 `json:"staff_id"`
	Notes   string  `json:"notes"   binding:"max=1000"`
}

func (r *addCareLogRequest) toServiceInput() (*CreateCareLogInput, error) {
	if _, err := time.ParseInLocation(dailyRecordTimeLayout, r.Time, time.Local); err != nil {
		return nil, fmt.Errorf("invalid time format, expected HH:MM:SS")
	}

	return &CreateCareLogInput{
		Time:    r.Time,
		Type:    r.Type,
		Status:  r.Status,
		Value:   r.Value,
		StaffID: r.StaffID,
		Notes:   r.Notes,
	}, nil
}

// addStaffNoteRequest はスタッフメモ記録追加のバインド struct
type addStaffNoteRequest struct {
	Time    string  `json:"time"    binding:"required"`
	Content string  `json:"content" binding:"required,max=1000"`
	StaffID *uint64 `json:"staff_id"`
}

func (r addStaffNoteRequest) toServiceInput() (*CreateStaffNoteInput, error) {
	if _, err := time.ParseInLocation(dailyRecordTimeLayout, r.Time, time.Local); err != nil {
		return nil, fmt.Errorf("invalid time format, expected HH:MM:SS")
	}

	return &CreateStaffNoteInput{
		Time:    r.Time,
		Content: r.Content,
		StaffID: r.StaffID,
	}, nil
}
