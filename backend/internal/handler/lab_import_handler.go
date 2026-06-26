package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
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
	OldPetKey      string `json:"old_pet_key"`
	OldChartKey    string `json:"old_chart_key"`
	OldRowKey      string `json:"old_row_key"`
	ExamDate       string `json:"exam_date"`
	ExamCode       string `json:"exam_code"`
	ExamName       string `json:"exam_name"`
	ItemName       string `json:"item_name"`
	DisplayValue   string `json:"display_value"`
	ReferenceValue string `json:"reference_value"`
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
	Name            string   `json:"name"`
	InspectionValue string   `json:"inspection_value"`
	Unit            string   `json:"unit"`
	ReferenceValue  string   `json:"reference_value"`
	RefMin          *float64 `json:"ref_min"`
	RefMax          *float64 `json:"ref_max"`
	SortOrder       int      `json:"sort_order"`
}

// ------------------------------------
// Response types
// ------------------------------------

type labImportJobResponse struct {
	ID                string  `json:"id"`
	ClinicID          string  `json:"clinic_id"`
	SourceType        string  `json:"source_type"`
	SourceFingerprint string  `json:"source_fingerprint"`
	Status            string  `json:"status"`
	RowCount          int     `json:"row_count"`
	PersistedCount    int     `json:"persisted_count"`
	DuplicateCount    int     `json:"duplicate_count"`
	NeedsReviewCount  int     `json:"needs_review_count"`
	FailedCount       int     `json:"failed_count"`
	ErrorCode         *string `json:"error_code,omitempty"`
	StartedAt         *string `json:"started_at,omitempty"`
	FinishedAt        *string `json:"finished_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type labImportEventResponse struct {
	ID               uint64  `json:"id"`
	ClinicID         string  `json:"clinic_id"`
	JobID            string  `json:"job_id"`
	EventType        string  `json:"event_type"`
	FromStatus       *string `json:"from_status,omitempty"`
	ToStatus         *string `json:"to_status,omitempty"`
	RowCount         int     `json:"row_count"`
	PersistedCount   int     `json:"persisted_count"`
	DuplicateCount   int     `json:"duplicate_count"`
	NeedsReviewCount int     `json:"needs_review_count"`
	ErrorCode        *string `json:"error_code,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

// labImportPreviewResponse は preview エンドポイントのハンドラー境界 DTO（P7 準拠）。
type labImportPreviewResponse struct {
	RowCount        int      `json:"row_count"`
	MappingWarnings []string `json:"mapping_warnings"`
	BlockedReasons  []string `json:"blocked_reasons"`
}

// labImportCommitResponse は commit エンドポイントのハンドラー境界 DTO（P7 準拠）。
// JobID を string に変換してフロントエンドの uint64 精度問題を回避する。
type labImportCommitResponse struct {
	JobID            string `json:"job_id"`
	PersistedCount   int    `json:"persisted_count"`
	DuplicateCount   int    `json:"duplicate_count"`
	NeedsReviewCount int    `json:"needs_review_count"`
	FailedCount      int    `json:"failed_count"`
}

// ------------------------------------
// Conversion functions (P18)
// ------------------------------------

func toLabImportPreviewResponse(r *model.LabImportPreviewResponse) labImportPreviewResponse {
	warnings := r.MappingWarnings
	if warnings == nil {
		warnings = []string{}
	}
	blocked := r.BlockedReasons
	if blocked == nil {
		blocked = []string{}
	}
	return labImportPreviewResponse{
		RowCount:        r.RowCount,
		MappingWarnings: warnings,
		BlockedReasons:  blocked,
	}
}

func toLabImportCommitResponse(r *model.LabImportCommitResponse) labImportCommitResponse {
	return labImportCommitResponse{
		JobID:            r.JobID.String(),
		PersistedCount:   r.PersistedCount,
		DuplicateCount:   r.DuplicateCount,
		NeedsReviewCount: r.NeedsReviewCount,
		FailedCount:      r.FailedCount,
	}
}

func toLabImportJobResponse(j *model.LabImportJob) labImportJobResponse {
	r := labImportJobResponse{
		ID:                j.ID.String(),
		ClinicID:          strconv.FormatUint(j.ClinicID, 10),
		SourceType:        string(j.SourceType),
		SourceFingerprint: j.SourceFingerprint,
		Status:            string(j.Status),
		RowCount:          j.RowCount,
		PersistedCount:    j.PersistedCount,
		DuplicateCount:    j.DuplicateCount,
		NeedsReviewCount:  j.NeedsReviewCount,
		FailedCount:       j.FailedCount,
		ErrorCode:         j.ErrorCode,
		CreatedAt:         j.CreatedAt.In(time.Local).Format(time.RFC3339),
		UpdatedAt:         j.UpdatedAt.In(time.Local).Format(time.RFC3339),
	}
	if j.StartedAt != nil {
		s := j.StartedAt.In(time.Local).Format(time.RFC3339)
		r.StartedAt = &s
	}
	if j.FinishedAt != nil {
		s := j.FinishedAt.In(time.Local).Format(time.RFC3339)
		r.FinishedAt = &s
	}
	return r
}

func toLabImportEventResponse(e *model.LabImportEvent) labImportEventResponse {
	r := labImportEventResponse{
		ID:               e.ID,
		ClinicID:         strconv.FormatUint(e.ClinicID, 10),
		JobID:            e.JobID.String(),
		EventType:        string(e.EventType),
		RowCount:         e.RowCount,
		PersistedCount:   e.PersistedCount,
		DuplicateCount:   e.DuplicateCount,
		NeedsReviewCount: e.NeedsReviewCount,
		ErrorCode:        e.ErrorCode,
		CreatedAt:        e.CreatedAt.In(time.Local).Format(time.RFC3339),
	}
	if e.FromStatus != nil {
		s := string(*e.FromStatus)
		r.FromStatus = &s
	}
	if e.ToStatus != nil {
		s := string(*e.ToStatus)
		r.ToStatus = &s
	}
	return r
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
	for i, r := range req.ResultRows {
		rows[i] = model.LabInboundResultRow{
			OldPetKey:      r.OldPetKey,
			OldChartKey:    r.OldChartKey,
			OldRowKey:      r.OldRowKey,
			ExamDate:       r.ExamDate,
			ExamCode:       r.ExamCode,
			ExamName:       r.ExamName,
			ItemName:       r.ItemName,
			DisplayValue:   r.DisplayValue,
			ReferenceValue: r.ReferenceValue,
		}
	}
	return model.LabInboundBatch{
		SourceType:        model.LabImportSourceType(req.SourceType),
		SourceFingerprint: req.SourceFingerprint,
		ReceivedAt:        receivedAt,
		ResultRows:        rows,
	}, nil
}

func toExamInputs(clinicID uint64, reqs []labExamInputReq) ([]service.LabExamPersistInput, error) {
	inputs := make([]service.LabExamPersistInput, 0, len(reqs))
	for i, r := range reqs {
		if r.ExamTypeID == 0 {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("inputs[%d].exam_type_id は必須です", i))
		}
		if r.Date == "" {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("inputs[%d].date は必須です", i))
		}
		d, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("inputs[%d].date は YYYY-MM-DD 形式で指定してください", i))
		}
		// UTC date-only 正規化 (IsDuplicate の要件)
		d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)

		items := make([]service.LabExamItemInput, len(r.Items))
		for j, it := range r.Items {
			items[j] = service.LabExamItemInput{
				Name:            it.Name,
				InspectionValue: it.InspectionValue,
				Unit:            it.Unit,
				ReferenceValue:  it.ReferenceValue,
				RefMin:          it.RefMin,
				RefMax:          it.RefMax,
				SortOrder:       it.SortOrder,
			}
		}
		inputs = append(inputs, service.LabExamPersistInput{
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

// ------------------------------------
// Handler methods
// ------------------------------------

// PostLabImportPreview godoc
// POST /api/v1/lab-imports/preview — バッチ内容を検証してサマリを返す（DB 書き込みなし）。
func (h *Handler) PostLabImportPreview(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req labImportPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	batch := model.LabInboundBatch{
		SourceType:        model.LabImportSourceType(req.SourceType),
		SourceFingerprint: req.SourceFingerprint,
		ReceivedAt:        time.Now(),
	}
	for _, r := range req.ResultRows {
		batch.ResultRows = append(batch.ResultRows, model.LabInboundResultRow{
			OldPetKey:      r.OldPetKey,
			OldChartKey:    r.OldChartKey,
			OldRowKey:      r.OldRowKey,
			ExamDate:       r.ExamDate,
			ExamCode:       r.ExamCode,
			ExamName:       r.ExamName,
			ItemName:       r.ItemName,
			DisplayValue:   r.DisplayValue,
			ReferenceValue: r.ReferenceValue,
		})
	}

	preview, err := h.svc.LabResultImport.Preview(c.Request.Context(), clinicID, batch)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabImportPreviewResponse(preview))
}

// PostLabImportCommit godoc
// POST /api/v1/lab-imports — fixture バッチを commit して exams + job を永続化する。
// source_type=fixture 以外は 400 を返す（Phase 3 制約）。
func (h *Handler) PostLabImportCommit(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var req labImportCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	batch, err := toBatch(req.Batch)
	if err != nil {
		RespondError(c, err)
		return
	}

	inputs, err := toExamInputs(clinicID, req.Inputs)
	if err != nil {
		RespondError(c, err)
		return
	}

	result, err := h.svc.LabResultImport.Commit(c.Request.Context(), clinicID, batch, inputs)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("/api/v1/lab-imports/%s", result.JobID.String()))
	c.JSON(http.StatusCreated, toLabImportCommitResponse(result))
}

// GetLabImportJob godoc
// GET /api/v1/lab-imports/:job_id — ジョブ詳細を返す。
func (h *Handler) GetLabImportJob(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := parseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	job, err := h.svc.LabImportJob.GetJob(c.Request.Context(), clinicID, jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toLabImportJobResponse(job))
}

// ListLabImportEvents godoc
// GET /api/v1/lab-imports/:job_id/events — ジョブのイベント一覧を返す（作成昇順）。
func (h *Handler) ListLabImportEvents(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	jobID, ok := parseUUIDParam(c, "job_id")
	if !ok {
		return
	}

	events, err := h.svc.LabImportJob.ListEvents(c.Request.Context(), clinicID, jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	resp := make([]labImportEventResponse, len(events))
	for i, e := range events {
		resp[i] = toLabImportEventResponse(e)
	}
	c.JSON(http.StatusOK, resp)
}

// RegisterLabImportRoutes は lab import エンドポイントのルートを登録する。
// 全エンドポイントに RequirePermission を付与する（P5 準拠）。
func (h *Handler) RegisterLabImportRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/lab-imports")
	perm := h.RequirePermission
	g.POST("/preview", perm(string(model.ResourceLabImport), "create"), h.PostLabImportPreview)
	g.POST("", perm(string(model.ResourceLabImport), "create"), h.PostLabImportCommit)
	g.GET("/:job_id", perm(string(model.ResourceLabImport), "view"), h.GetLabImportJob)
	g.GET("/:job_id/events", perm(string(model.ResourceLabImport), "view"), h.ListLabImportEvents)
}
