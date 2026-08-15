package medicalrecord

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/timeutil"
)

// LabReportQueryService は clinic-scoped で検査帳票向け read-only DTO を返す。
//
// Phase 4B.2 スコープ:
//   - exams + exam_results + exam_types + lab_import_jobs を参照する。
//   - 全クエリで clinic_id scope を適用する。
//   - PII-safe: owner 名・pet 名・result_summary・raw デバイスペイロードを返さない。
//
// Phase BLOCKED:
//   - 帳票テンプレート選択・PDF/CSV 生成
//   - crosswalk / 業務コード表
//   - Dr.Wan / MDB 接続
//   - result_summary（PHI 分類確認後に詳細 DTO へ追加可能）
type LabReportQueryService interface {
	// ListJobReportSummaries は job_id に紐づく exam サマリ一覧を返す（clinic scope 必須）。
	ListJobReportSummaries(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]model.LabExamReportSummary, error)

	// GetExamReport は exam_id の詳細 DTO を返す（clinic scope 必須）。
	GetExamReport(ctx context.Context, clinicID uint64, examID uint64) (*model.LabExamReportDetail, error)
}

type labReportQueryService struct {
	examRepo     examinationReportRepo
	usageTracker LabImportUsageTracker
}

// NewLabReportQueryService は LabReportQueryService を初期化して返す。
func NewLabReportQueryService(examRepo examinationReportRepo, usageTracker ...LabImportUsageTracker) LabReportQueryService {
	var tracker LabImportUsageTracker
	if len(usageTracker) > 0 {
		tracker = usageTracker[0]
	}
	return &labReportQueryService{examRepo: examRepo, usageTracker: tracker}
}

func (s *labReportQueryService) usage() LabImportUsageTracker {
	if s.usageTracker != nil {
		return s.usageTracker
	}
	return noopLabImportUsageTracker{}
}

func (s *labReportQueryService) ListJobReportSummaries(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]model.LabExamReportSummary, error) {
	if clinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	exams, err := s.examRepo.FindByJobID(ctx, clinicID, jobID)
	if err != nil {
		return nil, apperrors.Wrap(err, fmt.Sprintf("failed to list report summaries for job %s", jobID))
	}
	out := make([]model.LabExamReportSummary, 0, len(exams))
	for _, e := range exams {
		out = append(out, toLabExamReportSummary(e))
	}
	return out, nil
}

func (s *labReportQueryService) GetExamReport(ctx context.Context, clinicID, examID uint64) (*model.LabExamReportDetail, error) {
	if clinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	exam, err := s.examRepo.FindByID(ctx, clinicID, examID)
	if err != nil {
		return nil, apperrors.Wrap(err, fmt.Sprintf("failed to get exam report %d", examID))
	}
	// TASK-032: record usage receipt before returning clinical payload.
	if err := s.usage().RecordClinicalUse(ctx, clinicID, exam, model.LabImportUsageKindLabReport, nil); err != nil {
		return nil, err
	}
	return toLabExamReportDetail(exam), nil
}

// ------------------------------------
// DTO conversion helpers
// ------------------------------------

func toLabExamReportSummary(e *model.Examination) model.LabExamReportSummary {
	typeName := ""
	if e.ExaminationType != nil {
		typeName = e.ExaminationType.Name
	}
	resultCount := len(e.Items)
	abnormalCount := 0
	for i := range e.Items {
		if e.Items[i].IsAbnormal {
			abnormalCount++
		}
	}
	return model.LabExamReportSummary{
		ExamID:        e.ID,
		ClinicID:      strconv.FormatUint(e.ClinicID, 10),
		JobID:         e.JobID,
		Date:          e.Date.Format(time.DateOnly),
		ExamTypeName:  typeName,
		Status:        string(e.Status),
		ResultCount:   resultCount,
		AbnormalCount: abnormalCount,
		Machine:       e.Machine,
		CreatedAt:     timeutil.LocalRFC3339(e.CreatedAt),
	}
}

func toLabExamReportDetail(e *model.Examination) *model.LabExamReportDetail {
	typeName := ""
	if e.ExaminationType != nil {
		typeName = e.ExaminationType.Name
	}
	items := make([]model.LabExamResultItem, 0, len(e.Items))
	for i := range e.Items {
		r := &e.Items[i]
		items = append(items, model.LabExamResultItem{
			Name:            r.Name,
			InspectionValue: r.InspectionValue,
			NormalValue:     r.NormalValue,
			Unit:            r.Unit,
			ReferenceValue:  r.ReferenceValue,
			RefMin:          r.RefMin,
			RefMax:          r.RefMax,
			QualitativeMin:  r.QualitativeMin,
			QualitativeMax:  r.QualitativeMax,
			IsAbnormal:      r.IsAbnormal,
			Status:          string(r.Status),
			SortOrder:       r.SortOrder,
		})
	}
	return &model.LabExamReportDetail{
		ExamID:          e.ID,
		ClinicID:        strconv.FormatUint(e.ClinicID, 10),
		JobID:           e.JobID,
		PetID:           e.PetID,
		MedicalRecordID: e.MedicalRecordID,
		DoctorID:        e.DoctorID,
		Date:            e.Date.Format(time.DateOnly),
		ExamTypeName:    typeName,
		Status:          string(e.Status),
		Machine:         e.Machine,
		Items:           items,
		CreatedAt:       timeutil.LocalRFC3339(e.CreatedAt),
		UpdatedAt:       timeutil.LocalRFC3339(e.UpdatedAt),
	}
}
