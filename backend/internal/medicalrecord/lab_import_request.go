package medicalrecord

import (
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ------------------------------------
// Request types
// ------------------------------------

// labImportPreviewRequest は preview エンドポイントのリクエストボディ。
// PHI・接続情報・raw デバイスペイロードは受け付けない。
type labImportPreviewRequest struct {
	SourceType        string                  `json:"source_type"        binding:"required"`
	SourceFingerprint string                  `json:"source_fingerprint"`
	ResultRows        []labImportResultRowReq `json:"result_rows"`
}

// labImportRevertRequest は compensating revert のボディ（TASK-032）。
// reason 必須。Idempotency-Key はヘッダで受け取る。
type labImportRevertRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// labImportCommitRequest は commit エンドポイントのリクエストボディ。
// Inputs は呼び出し元がバッチ行から解決した LabExamPersistInput スライス。
// clinic_id はリクエストボディで受け取らず JWT コンテキストから注入する（改ざん防止）。
type labImportCommitRequest struct {
	Batch  labImportBatchReq `json:"batch"  binding:"required"`
	Inputs []labExamInputReq `json:"inputs"`
}

type labImportBatchReq struct {
	SourceType        string                  `json:"source_type"        binding:"required"`
	SourceFingerprint string                  `json:"source_fingerprint"`
	ReceivedAt        string                  `json:"received_at"`
	ResultRows        []labImportResultRowReq `json:"result_rows"`
}

type labImportResultRowReq struct {
	OldPetKey      string `json:"old_pet_key" binding:"omitempty,max=255"`
	OldChartKey    string `json:"old_chart_key" binding:"omitempty,max=255"`
	OldRowKey      string `json:"old_row_key" binding:"omitempty,max=255"`
	ExamDate       string `json:"exam_date" binding:"omitempty,max=32"`
	ExamCode       string `json:"exam_code" binding:"omitempty,max=64"`
	ExamName       string `json:"exam_name" binding:"omitempty,max=255"`
	ItemName       string `json:"item_name" binding:"omitempty,max=255"`
	DisplayValue   string `json:"display_value" binding:"omitempty,max=255"`
	ReferenceValue string `json:"reference_value" binding:"omitempty,max=255"`
}

// labExamInputReq はハンドラー境界での exam 入力。
// clinic_id は JWT コンテキストから注入するため省略する。
type labExamInputReq struct {
	PetID           *uint64          `json:"pet_id"`
	MedicalRecordID *uint64          `json:"medical_record_id"`
	ExamTypeID      uint64           `json:"exam_type_id" binding:"required"`
	Date            string           `json:"date"         binding:"required"`
	Machine         string           `json:"machine"`
	Items           []labExamItemReq `json:"items"`
}

type labExamItemReq struct {
	// MRC-08: boundary validation for free-text import fields and numeric ref range.
	Name            string   `json:"name" binding:"omitempty,max=255"`
	InspectionValue string   `json:"inspection_value" binding:"omitempty,max=255"`
	Unit            string   `json:"unit" binding:"omitempty,max=64"`
	ReferenceValue  string   `json:"reference_value" binding:"omitempty,max=255"`
	RefMin          *float64 `json:"ref_min" binding:"omitempty"`
	RefMax          *float64 `json:"ref_max" binding:"omitempty"`
	SortOrder       int      `json:"sort_order"`
}

// ------------------------------------
// Helper: request → model conversion
// ------------------------------------

func toBatch(req labImportBatchReq) (model.LabInboundBatch, error) {
	var receivedAt time.Time
	if req.ReceivedAt != "" {
		t, err := time.Parse(time.RFC3339, req.ReceivedAt)
		if err != nil {
			return model.LabInboundBatch{}, apperrors.WrapInvalidInput("received_at は RFC3339 形式で指定してください")
		}
		receivedAt = t
	} else {
		receivedAt = time.Now()
	}

	rows := make([]model.LabInboundResultRow, len(req.ResultRows))
	for i := range req.ResultRows {
		rows[i] = model.LabInboundResultRow{
			OldPetKey:      req.ResultRows[i].OldPetKey,
			OldChartKey:    req.ResultRows[i].OldChartKey,
			OldRowKey:      req.ResultRows[i].OldRowKey,
			ExamDate:       req.ResultRows[i].ExamDate,
			ExamCode:       req.ResultRows[i].ExamCode,
			ExamName:       req.ResultRows[i].ExamName,
			ItemName:       req.ResultRows[i].ItemName,
			DisplayValue:   req.ResultRows[i].DisplayValue,
			ReferenceValue: req.ResultRows[i].ReferenceValue,
		}
	}
	return model.LabInboundBatch{
		SourceType:        model.LabImportSourceType(req.SourceType),
		SourceFingerprint: req.SourceFingerprint,
		ReceivedAt:        receivedAt,
		ResultRows:        rows,
	}, nil
}

func toExamInputs(clinicID uint64, reqs []labExamInputReq) ([]LabExamPersistInput, error) {
	inputs := make([]LabExamPersistInput, 0, len(reqs))
	for i, r := range reqs {
		if r.ExamTypeID == 0 {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("inputs[%d].exam_type_id は必須です", i))
		}
		if r.Date == "" {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("inputs[%d].date は必須です", i))
		}
		d, err := time.Parse(time.DateOnly, r.Date)
		if err != nil {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("inputs[%d].date は YYYY-MM-DD 形式で指定してください", i))
		}
		// UTC date-only 正規化 (IsDuplicate 候補フィルタ / write path 一致の要件)
		d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)

		items := make([]LabExamItemInput, len(r.Items))
		for j, it := range r.Items {
			// MRC-08: reject inverted reference ranges at the conversion boundary.
			if it.RefMin != nil && it.RefMax != nil && *it.RefMin > *it.RefMax {
				return nil, apperrors.WrapInvalidInput(fmt.Sprintf(
					"inputs[%d].items[%d].ref_min は ref_max 以下である必要があります", i, j,
				))
			}
			items[j] = LabExamItemInput{
				Name:            it.Name,
				InspectionValue: it.InspectionValue,
				Unit:            it.Unit,
				ReferenceValue:  it.ReferenceValue,
				RefMin:          it.RefMin,
				RefMax:          it.RefMax,
				SortOrder:       it.SortOrder,
			}
		}
		inputs = append(inputs, LabExamPersistInput{
			ClinicID:        clinicID,
			PetID:           r.PetID,
			MedicalRecordID: r.MedicalRecordID,
			ExamTypeID:      r.ExamTypeID,
			Date:            d,
			Machine:         r.Machine,
			Items:           items,
		})
	}
	return inputs, nil
}
