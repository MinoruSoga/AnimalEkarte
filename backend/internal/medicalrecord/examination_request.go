package medicalrecord

import (
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type listExaminationQuery struct {
	PetID        string
	OwnerID      string
	Status       string
	StartDate    string
	EndDate      string
	IncludeItems string
}

func newListExaminationQuery(values url.Values) listExaminationQuery {
	return listExaminationQuery{
		PetID:        values.Get("pet_id"),
		OwnerID:      values.Get("owner_id"),
		Status:       values.Get("status"),
		StartDate:    values.Get("start_date"),
		EndDate:      values.Get("end_date"),
		IncludeItems: values.Get("include_items"),
	}
}

type listExaminationFilters struct {
	PetID        *uint64
	OwnerID      *uint64
	Status       *string
	StartDate    *string
	EndDate      *string
	IncludeItems bool
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
	includeItems, err := parseOptionalBooleanQueryFilter(q.IncludeItems, "include_items")
	if err != nil {
		return listExaminationFilters{}, err
	}
	return listExaminationFilters{
		PetID:        petID,
		OwnerID:      ownerID,
		Status:       optionalStringQueryFilter(q.Status),
		StartDate:    startDate,
		EndDate:      endDate,
		IncludeItems: includeItems,
	}, nil
}

func parseOptionalBooleanQueryFilter(value, field string) (bool, error) {
	if value == "" {
		return false, nil
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, apperrors.WrapInvalidInput("invalid " + field)
	}
	return boolValue, nil
}

// createExaminationRequest は検査作成のバインド struct
type createExaminationRequest struct {
	MedicalRecordID *uint64                  `json:"medical_record_id"  binding:"omitempty"`
	PetID           *uint64                  `json:"pet_id"`
	ExamTypeID      uint64                   `json:"exam_type_id"       binding:"required"`
	DoctorID        *uint64                  `json:"doctor_id"`
	Date            time.Time                `json:"date"              binding:"required"`
	ResultSummary   string                   `json:"result_summary"`
	Machine         string                   `json:"machine"`
	Status          string                   `json:"status"             binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
	Items           *[]upsertExamItemRequest `json:"items" binding:"omitempty,dive"`
}

func (r *createExaminationRequest) toServiceInput() *CreateExaminationInput {
	return &CreateExaminationInput{
		MedicalRecordID: r.MedicalRecordID,
		PetID:           r.PetID,
		ExamTypeID:      r.ExamTypeID,
		DoctorID:        r.DoctorID,
		Date:            r.Date,
		ResultSummary:   r.ResultSummary,
		Machine:         r.Machine,
		Status:          model.ExaminationStatus(r.Status),
		Items:           optionalExamItemsToServiceInput(r.Items),
	}
}

// updateExaminationRequest は検査更新のバインド struct
type updateExaminationRequest struct {
	MedicalRecordID *uint64                  `json:"medical_record_id"`
	PetID           *uint64                  `json:"pet_id"`
	ExamTypeID      uint64                   `json:"exam_type_id"`
	DoctorID        *uint64                  `json:"doctor_id"`
	Date            *time.Time               `json:"date"`
	ResultSummary   *string                  `json:"result_summary"`
	Machine         *string                  `json:"machine"`
	Status          *string                  `json:"status"             binding:"omitempty,oneof=pending in_progress result_entered completed confirmed"`
	Items           *[]upsertExamItemRequest `json:"items" binding:"omitempty,dive"`
}

type unconfirmExaminationRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

func (r updateExaminationRequest) toServiceInput() UpdateExaminationInput {
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

	return UpdateExaminationInput{
		MedicalRecordID: r.MedicalRecordID,
		PetID:           r.PetID,
		ExamTypeID:      examTypeID,
		DoctorID:        r.DoctorID,
		Date:            r.Date,
		ResultSummary:   r.ResultSummary,
		Machine:         r.Machine,
		Status:          status,
		Items:           optionalExamItemsToServiceInput(r.Items),
	}
}

// upsertExamItemRequest は検査項目 1 行分のバインド struct。
// status / is_abnormal は受け付けない（サーバ側でマスタ基準値から計算する）。
type upsertExamItemRequest struct {
	ExamTypeFieldID *uint64 `json:"exam_type_field_id"`
	Name            string  `json:"name"             binding:"required"`
	InspectionValue string  `json:"inspection_value"`
	NormalValue     string  `json:"normal_value"`
	Unit            string  `json:"unit"`
	ReferenceValue  string  `json:"reference_value"`
	SortOrder       int     `json:"sort_order"`
}

// replaceExamItemsRequest は検査項目の一括置換 PUT リクエスト。
// items が nil の場合は空配列として扱う（全削除と等価）。
type replaceExamItemsRequest struct {
	Items []upsertExamItemRequest `json:"items" binding:"dive"`
}

func (r replaceExamItemsRequest) toServiceInput() []UpsertExamItemInput {
	inputs := make([]UpsertExamItemInput, 0, len(r.Items))
	for _, item := range r.Items {
		inputs = append(inputs, UpsertExamItemInput{
			ExamTypeFieldID: item.ExamTypeFieldID,
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			NormalValue:     item.NormalValue,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			SortOrder:       item.SortOrder,
		})
	}
	return inputs
}

func optionalExamItemsToServiceInput(items *[]upsertExamItemRequest) *[]UpsertExamItemInput {
	if items == nil {
		return nil
	}
	inputs := (replaceExamItemsRequest{Items: *items}).toServiceInput()
	return &inputs
}
