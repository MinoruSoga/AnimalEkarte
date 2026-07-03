package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type checkupResponse struct {
	ID              string  `json:"id"`
	MedicalRecordID string  `json:"medical_record_id"`
	CheckupTypeID   string  `json:"checkup_type_id"`
	PetID           *string `json:"pet_id,omitempty"`
	Date            string  `json:"date"`
	NextDate        *string `json:"next_date,omitempty"`
	DoctorID        *string `json:"doctor_id,omitempty"`
	Result          string  `json:"result"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`

	// Nested
	CheckupType *checkupTypeResponse  `json:"checkup_type,omitempty"`
	Doctor      *staffSummaryResponse `json:"doctor,omitempty"`
}

func toCheckupResponse(c *model.Checkup) checkupResponse {
	r := checkupResponse{
		ID:              strconv.FormatUint(c.ID, 10),
		MedicalRecordID: strconv.FormatUint(c.MedicalRecordID, 10),
		CheckupTypeID:   strconv.FormatUint(c.CheckupTypeID, 10),
		Date:            c.Date.In(time.Local).Format("2006-01-02"),
		Result:          c.Result,
		CreatedAt:       c.CreatedAt.In(time.Local).Format(time.RFC3339),
		UpdatedAt:       c.UpdatedAt.In(time.Local).Format(time.RFC3339),
	}
	if c.NextDate != nil {
		nd := c.NextDate.In(time.Local).Format("2006-01-02")
		r.NextDate = &nd
	}
	if c.PetID != nil {
		s := strconv.FormatUint(*c.PetID, 10)
		r.PetID = &s
	}
	if c.DoctorID != nil {
		s := strconv.FormatUint(*c.DoctorID, 10)
		r.DoctorID = &s
	}
	if c.CheckupType != nil {
		ct := toCheckupTypeResponse(c.CheckupType)
		r.CheckupType = &ct
	}
	r.Doctor = toStaffSummary(c.Doctor)
	return r
}

// checkupGlobalResponse はクリニック横断一覧用レスポンス（MedicalRecord.Pet.Owner を含む）
type checkupGlobalResponse struct {
	ID              string  `json:"id"`
	MedicalRecordID string  `json:"medical_record_id"`
	CheckupTypeID   string  `json:"checkup_type_id"`
	PetID           *string `json:"pet_id,omitempty"`
	Date            string  `json:"date"`
	NextDate        *string `json:"next_date,omitempty"`
	DoctorID        *string `json:"doctor_id,omitempty"`
	Result          string  `json:"result"`

	// Nested
	CheckupType *checkupTypeResponse  `json:"checkup_type,omitempty"`
	Doctor      *staffSummaryResponse `json:"doctor,omitempty"`
	PetName     string                `json:"pet_name"`
	OwnerName   string                `json:"owner_name"`
	OwnerID     *string               `json:"owner_id,omitempty"`
}

func toCheckupGlobalResponse(c *model.Checkup) checkupGlobalResponse {
	r := checkupGlobalResponse{
		ID:              strconv.FormatUint(c.ID, 10),
		MedicalRecordID: strconv.FormatUint(c.MedicalRecordID, 10),
		CheckupTypeID:   strconv.FormatUint(c.CheckupTypeID, 10),
		Date:            c.Date.In(time.Local).Format("2006-01-02"),
		Result:          c.Result,
	}
	if c.NextDate != nil {
		nd := c.NextDate.In(time.Local).Format("2006-01-02")
		r.NextDate = &nd
	}
	if c.PetID != nil {
		s := strconv.FormatUint(*c.PetID, 10)
		r.PetID = &s
	}
	if c.DoctorID != nil {
		s := strconv.FormatUint(*c.DoctorID, 10)
		r.DoctorID = &s
	}
	if c.CheckupType != nil {
		ct := toCheckupTypeResponse(c.CheckupType)
		r.CheckupType = &ct
	}
	r.Doctor = toStaffSummary(c.Doctor)
	// Pet/Owner from MedicalRecord relation
	if c.MedicalRecord != nil && c.MedicalRecord.Pet != nil {
		r.PetName = c.MedicalRecord.Pet.Name
		if c.MedicalRecord.Pet.Owner != nil {
			r.OwnerName = c.MedicalRecord.Pet.Owner.Name
			s := strconv.FormatUint(c.MedicalRecord.Pet.Owner.ID, 10)
			r.OwnerID = &s
		}
	}
	return r
}

// ---- Alert response types ----

type checkupAlertItemResponse struct {
	CheckupID       uint64 `json:"checkup_id"`
	PetID           uint64 `json:"pet_id"`
	PetName         string `json:"pet_name"`
	OwnerName       string `json:"owner_name"`
	CheckupTypeName string `json:"checkup_type_name"`
	LastCheckupDate string `json:"last_checkup_date"`
	NextDate        string `json:"next_date"`
	Days            int    `json:"days"` // overdue: 経過日数(正), upcoming: 残り日数(正)
}

type checkupAlertCategoryResponse struct {
	Count int                        `json:"count"`
	Items []checkupAlertItemResponse `json:"items"`
}

type checkupAlertsResponse struct {
	Overdue  checkupAlertCategoryResponse `json:"overdue"`
	Upcoming checkupAlertCategoryResponse `json:"upcoming"`
}

func toCheckupAlertItemResponse(c *model.Checkup, today time.Time) checkupAlertItemResponse {
	var petID uint64
	var petName, ownerName string
	if c.MedicalRecord != nil && c.MedicalRecord.Pet != nil {
		petID = c.MedicalRecord.Pet.ID
		petName = c.MedicalRecord.Pet.Name
		if c.MedicalRecord.Pet.Owner != nil {
			ownerName = c.MedicalRecord.Pet.Owner.Name
		}
	}
	var checkupTypeName string
	if c.CheckupType != nil {
		checkupTypeName = c.CheckupType.Name
	}
	nextDateStr := ""
	var days int
	if c.NextDate != nil {
		nextDateStr = c.NextDate.In(time.Local).Format("2006-01-02")
		diff := int(c.NextDate.Sub(today).Hours() / 24)
		if diff < 0 {
			days = -diff
		} else {
			days = diff
		}
	}
	return checkupAlertItemResponse{
		CheckupID:       c.ID,
		PetID:           petID,
		PetName:         petName,
		OwnerName:       ownerName,
		CheckupTypeName: checkupTypeName,
		LastCheckupDate: c.Date.In(time.Local).Format("2006-01-02"),
		NextDate:        nextDateStr,
		Days:            days,
	}
}

func toCheckupAlertsResponse(result *service.CheckupAlertsResult) checkupAlertsResponse {
	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	overdueItems := make([]checkupAlertItemResponse, 0, len(result.Overdue))
	for i := range result.Overdue {
		overdueItems = append(overdueItems, toCheckupAlertItemResponse(&result.Overdue[i], today))
	}
	upcomingItems := make([]checkupAlertItemResponse, 0, len(result.Upcoming))
	for i := range result.Upcoming {
		upcomingItems = append(upcomingItems, toCheckupAlertItemResponse(&result.Upcoming[i], today))
	}
	return checkupAlertsResponse{
		Overdue:  checkupAlertCategoryResponse{Count: len(overdueItems), Items: overdueItems},
		Upcoming: checkupAlertCategoryResponse{Count: len(upcomingItems), Items: upcomingItems},
	}
}
