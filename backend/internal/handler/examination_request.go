package handler

import (
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type listExaminationQuery struct {
	PetID     string
	OwnerID   string
	Status    string
	StartDate string
	EndDate   string
}

func newListExaminationQuery(values url.Values) listExaminationQuery {
	return listExaminationQuery{
		PetID:     values.Get("pet_id"),
		OwnerID:   values.Get("owner_id"),
		Status:    values.Get("status"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}

type listExaminationFilters struct {
	PetID     *uint64
	OwnerID   *uint64
	Status    *string
	StartDate *string
	EndDate   *string
}

func (q *listExaminationQuery) toServiceFilters() (listExaminationFilters, error) {
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return listExaminationFilters{}, err
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listExaminationFilters{}, err
	}
	startDate, err := parseOptionalDateQueryFilter(q.StartDate, "start_date")
	if err != nil {
		return listExaminationFilters{}, err
	}
	endDate, err := parseOptionalDateQueryFilter(q.EndDate, "end_date")
	if err != nil {
		return listExaminationFilters{}, err
	}
	return listExaminationFilters{
		PetID:     petID,
		OwnerID:   ownerID,
		Status:    optionalStringQueryFilter(q.Status),
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// createExaminationRequest は検査作成のバインド struct
type createExaminationRequest struct {
	MedicalRecordID *uint64   `json:"medical_record_id"  binding:"omitempty"`
	PetID           *uint64   `json:"pet_id"`
	ExamTypeID      uint64    `json:"exam_type_id"       binding:"required"`
	DoctorID        *uint64   `json:"doctor_id"`
	Date            time.Time `json:"date"              binding:"required"`
	ResultSummary   string    `json:"result_summary"`
	Machine         string    `json:"machine"`
	Status          string    `json:"status"             binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
}

func (r *createExaminationRequest) toServiceInput() *service.CreateExaminationInput {
	return &service.CreateExaminationInput{
		MedicalRecordID: r.MedicalRecordID,
		PetID:           r.PetID,
		ExamTypeID:      r.ExamTypeID,
		DoctorID:        r.DoctorID,
		Date:            r.Date,
		ResultSummary:   r.ResultSummary,
		Machine:         r.Machine,
		Status:          model.ExaminationStatus(r.Status),
	}
}

// updateExaminationRequest は検査更新のバインド struct
type updateExaminationRequest struct {
	MedicalRecordID *uint64    `json:"medical_record_id"`
	PetID           *uint64    `json:"pet_id"`
	ExamTypeID      uint64     `json:"exam_type_id"`
	DoctorID        *uint64    `json:"doctor_id"`
	Date            *time.Time `json:"date"`
	ResultSummary   *string    `json:"result_summary"`
	Machine         *string    `json:"machine"`
	Status          *string    `json:"status"             binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
}

func (r updateExaminationRequest) toServiceInput() service.UpdateExaminationInput {
	var status *model.ExaminationStatus
	if r.Status != nil {
		v := model.ExaminationStatus(*r.Status)
		status = &v
	}

	var examTypeID *uint64
	if r.ExamTypeID != 0 {
		v := r.ExamTypeID
		examTypeID = &v
	}

	return service.UpdateExaminationInput{
		MedicalRecordID: r.MedicalRecordID,
		PetID:           r.PetID,
		ExamTypeID:      examTypeID,
		DoctorID:        r.DoctorID,
		Date:            r.Date,
		ResultSummary:   r.ResultSummary,
		Machine:         r.Machine,
		Status:          status,
	}
}

// upsertExamItemRequest は検査項目 1 行分のバインド struct。
// status / is_abnormal は受け付けない（サーバ側で ref_min/ref_max から計算する）。
type upsertExamItemRequest struct {
	ExamTypeFieldID *uint64  `json:"exam_type_field_id"`
	Name            string   `json:"name"             binding:"required"`
	InspectionValue string   `json:"inspection_value"`
	NormalValue     string   `json:"normal_value"`
	Unit            string   `json:"unit"`
	ReferenceValue  string   `json:"reference_value"`
	RefMin          *float64 `json:"ref_min"`
	RefMax          *float64 `json:"ref_max"`
	SortOrder       int      `json:"sort_order"`
}

// replaceExamItemsRequest は検査項目の一括置換 PUT リクエスト。
// items が nil の場合は空配列として扱う（全削除と等価）。
type replaceExamItemsRequest struct {
	Items []upsertExamItemRequest `json:"items" binding:"dive"`
}

func (r replaceExamItemsRequest) toServiceInput() []service.UpsertExamItemInput {
	inputs := make([]service.UpsertExamItemInput, 0, len(r.Items))
	for _, item := range r.Items {
		inputs = append(inputs, service.UpsertExamItemInput{
			ExamTypeFieldID: item.ExamTypeFieldID,
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			NormalValue:     item.NormalValue,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			RefMin:          item.RefMin,
			RefMax:          item.RefMax,
			SortOrder:       item.SortOrder,
		})
	}
	return inputs
}
