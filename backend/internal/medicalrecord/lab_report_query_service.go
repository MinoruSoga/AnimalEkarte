package medicalrecord

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
		summary, convErr := toLabExamReportSummary(e)
		if convErr != nil {
			return nil, convErr
		}
		out = append(out, summary)
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
	detail, err := toLabExamReportDetail(exam)
	if err != nil {
		return nil, err
	}
	// TASK-032: record usage receipt before returning clinical payload.
	if err := s.usage().RecordClinicalUse(ctx, clinicID, exam, model.LabImportUsageKindLabReport, nil); err != nil {
		return nil, err
	}
	return detail, nil
}

// ------------------------------------
// DTO conversion helpers
// ------------------------------------

func toLabExamReportSummary(e *model.Examination) (model.LabExamReportSummary, error) {
	typeName := ""
	if e.ExaminationType != nil {
		typeName = e.ExaminationType.Name
	}
	// date / is_abnormal は NULL 許容列。未記録を空日付や正常扱いで埋めず、帳票全体を拒否する。
	date, err := sharedkernel.RequireStored(e.Date, "exams.date")
	if err != nil {
		return model.LabExamReportSummary{}, err
	}
	resultCount := len(e.Items)
	abnormalCount := 0
	for i := range e.Items {
		isAbnormal, abnormalErr := sharedkernel.RequireStored(e.Items[i].IsAbnormal, "exam_results.is_abnormal")
		if abnormalErr != nil {
			return model.LabExamReportSummary{}, abnormalErr
		}
		if isAbnormal {
			abnormalCount++
		}
	}
	return model.LabExamReportSummary{
		ExamID:        e.ID,
		ClinicID:      strconv.FormatUint(e.ClinicID, 10),
		JobID:         e.JobID,
		Date:          date.Format(time.DateOnly),
		ExamTypeName:  typeName,
		Status:        sharedkernel.TextOrEmpty(e.Status),
		ResultCount:   resultCount,
		AbnormalCount: abnormalCount,
		Machine:       e.Machine,
		CreatedAt:     timeutil.LocalRFC3339(e.CreatedAt),
	}, nil
}

func toLabExamReportDetail(e *model.Examination) (*model.LabExamReportDetail, error) {
	typeName := ""
	if e.ExaminationType != nil {
		typeName = e.ExaminationType.Name
	}
	// date / is_abnormal は NULL 許容列。未記録を空日付や正常扱いで埋めず、帳票全体を拒否する。
	date, err := sharedkernel.RequireStored(e.Date, "exams.date")
	if err != nil {
		return nil, err
	}
	items := make([]model.LabExamResultItem, 0, len(e.Items))
	for i := range e.Items {
		r := &e.Items[i]
		isAbnormal, abnormalErr := sharedkernel.RequireStored(r.IsAbnormal, "exam_results.is_abnormal")
		if abnormalErr != nil {
			return nil, abnormalErr
		}
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
			IsAbnormal:      isAbnormal,
			Status:          sharedkernel.TextOrEmpty(r.Status),
			SortOrder:       sharedkernel.SortOrderValue(r.SortOrder),
		})
	}
	return &model.LabExamReportDetail{
		ExamID:          e.ID,
		ClinicID:        strconv.FormatUint(e.ClinicID, 10),
		JobID:           e.JobID,
		PetID:           e.PetID,
		MedicalRecordID: e.MedicalRecordID,
		DoctorID:        e.DoctorID,
		Date:            date.Format(time.DateOnly),
		ExamTypeName:    typeName,
		Status:          sharedkernel.TextOrEmpty(e.Status),
		Machine:         e.Machine,
		Items:           items,
		CreatedAt:       timeutil.LocalRFC3339(e.CreatedAt),
		UpdatedAt:       timeutil.LocalRFC3339(e.UpdatedAt),
	}, nil
}
