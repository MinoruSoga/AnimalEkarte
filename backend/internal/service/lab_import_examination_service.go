package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// LabExamPersistInput は lab import バッチ 1 行分の変換済み入力。
// raw PHI・接続文字列・認証情報は含めない。
type LabExamPersistInput struct {
	ClinicID        uint64
	PetID           *uint64
	MedicalRecordID *uint64
	ExamTypeID      uint64
	Date            time.Time
	Machine         string
	JobID           uuid.UUID
	Items           []LabExamItemInput
}

// LabExamItemInput は exam_results 1 行分の変換済み入力。
//
// 定性値（"+"・"陰性"・"TNTC" 等）は InspectionValue にそのまま格納する。
// 定性値は ParseFloat 不能なため computeExamResultStatus が (normal, false) を返す点に注意。
// RefMin / RefMax が nil の場合は範囲比較をスキップする。
type LabExamItemInput struct {
	Name            string
	InspectionValue string
	Unit            string
	ReferenceValue  string
	RefMin          *float64
	RefMax          *float64
	SortOrder       int
}

// LabExamPersistResult は 1 exam の保存結果サマリ。
// PersistBatch は個別行でこの型に失敗を記録する（関数エラーでは返さない）。
type LabExamPersistResult struct {
	ExamID    uint64
	ItemCount int
	Duplicate bool
	// RowError は persistExam が失敗した場合に PersistBatch が記録する。
	// nil なら成功（Duplicate=true を含む）。
	RowError error
	JobID    uuid.UUID
}

// LabImportExaminationService は lab import バッチを exams / exam_results へ永続化する。
//
// Phase 1 スコープ:
//   - fixture ソースからの synthetic 入力を受け付ける
//   - clinic_id 隔離を保証する（examRepo.Create の ClinicID + ReplaceItemsByExamID の JOIN 検証）
//   - (clinic_id, exam_type_id, date, pet_id) の複合キーによる重複スキップを実装する
//   - DB レベルの unique violation も重複として扱う（TOCTOU 安全ネット）
//   - lab_import_jobs との接続点を JobID フィールドで保持する
//   - HC-005: lab import は MedicalRecord の status を検証しない
//     （lab 結果は確定済みカルテへの追記を許容する — 手動作成経路と異なる意図的な設計差異）
//
// Phase BLOCKED:
//   - Dr.Wan MDB 接続 / raw デバイスペイロード解析（外部スキーマ未確認）
//   - 外部 credentials / ネットワーク I/O
//
// DB 注意: exam_results.exam_id に複合 index が存在しない場合、
// 大量バッチでの ReplaceItemsByExamID はテーブルスキャンになる。
// Phase 2 以前に `idx_exam_results_exam_id` migration を適用すること。
// buildExamResults は LabExamItemInput スライスを model.ExamResult スライスへ変換する。
// status / is_abnormal はサービス層で計算し、呼び出し元から受け付けない（信頼境界保護）。
// 定性値（ParseFloat 不能な文字列）は (normal, false) として格納される。
func buildExamResults(examID uint64, items []LabExamItemInput) []model.ExamResult {
	out := make([]model.ExamResult, 0, len(items))
	for _, item := range items {
		status, isAbnormal := computeExamResultStatus(item.InspectionValue, item.RefMin, item.RefMax)
		out = append(out, model.ExamResult{
			ExamID:          examID,
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			RefMin:          item.RefMin,
			RefMax:          item.RefMax,
			IsAbnormal:      isAbnormal,
			Status:          status,
			SortOrder:       item.SortOrder,
		})
	}
	return out
}

type LabImportExaminationService interface {
	// PersistBatch は複数 exam を順次永続化し全行の結果を返す。
	// 個別行のエラーは LabExamPersistResult.RowError に記録する。
	// バッチ全体のエラー（context キャンセル等）は関数エラーとして返す。
	PersistBatch(ctx context.Context, inputs []LabExamPersistInput) ([]*LabExamPersistResult, error)
}

// LabImportDuplicateChecker は (clinic_id, exam_type_id, date, pet_id) キーによる重複チェック。
//
// 実装は (clinic_id, exam_type_id, date, pet_id) で exams テーブルを検索する。
// DB レベルの unique constraint がない場合は TOCTOU 競合で重複が発生しうる。
// persistExam では AlreadyExists エラーも Duplicate として処理する安全ネットを設ける。
// Phase 2 で DB unique 制約を追加する場合は IsDuplicate を削除可能。
type LabImportDuplicateChecker interface {
	// IsDuplicate は (clinic_id, exam_type_id, date, pet_id) に一致する exam が既存かを返す。
	IsDuplicate(ctx context.Context, clinicID, examTypeID uint64, date time.Time, petID *uint64) (bool, error)
}

// labImportExaminationService は LabImportExaminationService の実装。
type labImportExaminationService struct {
	examRepo          repository.ExaminationRepository
	dupChecker        LabImportDuplicateChecker
	examTypeRepo      repository.ExamTypeRepository
	petRepo           repository.PetRepository
	medicalRecordRepo repository.MedicalRecordRepository
}

// NewLabImportExaminationService は LabImportExaminationService を初期化して返す。
func NewLabImportExaminationService(
	examRepo repository.ExaminationRepository,
	dupChecker LabImportDuplicateChecker,
	examTypeRepo repository.ExamTypeRepository,
	petRepo repository.PetRepository,
	medicalRecordRepo repository.MedicalRecordRepository,
) LabImportExaminationService {
	return &labImportExaminationService{
		examRepo:          examRepo,
		dupChecker:        dupChecker,
		examTypeRepo:      examTypeRepo,
		petRepo:           petRepo,
		medicalRecordRepo: medicalRecordRepo,
	}
}

func (s *labImportExaminationService) persistExam(ctx context.Context, input LabExamPersistInput) (*LabExamPersistResult, error) {
	if input.ClinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	if input.ExamTypeID == 0 {
		return nil, apperrors.WrapInvalidInput("exam_type_id is required")
	}
	if input.Date.IsZero() {
		return nil, apperrors.WrapInvalidInput("date is required")
	}

	// クロステナント write 防止: 別 clinic の exam_type を紐付けると、その exam_type が持つ
	// 異常値判定の基準値/単位（exam_type_fields）が検査記録に混入する（#124 同型）。所有権を検証する。
	if _, err := s.examTypeRepo.FindByID(ctx, input.ClinicID, input.ExamTypeID); err != nil {
		slog.ErrorContext(ctx, "failed to verify exam type ownership",
			"error", err,
			"exam_type_id", input.ExamTypeID,
			"clinic_id", input.ClinicID,
		)
		return nil, apperrors.Wrap(err, "failed to verify exam type ownership")
	}

	// クロステナント write 防止 (P1-2, PR #186 review): 別 clinic の pet/medical_record を
	// 紐付けると、他院の患者データが lab import 経由で漏洩・混入する。所有権を検証する。
	if input.PetID != nil {
		if _, err := s.petRepo.FindByID(ctx, input.ClinicID, *input.PetID); err != nil {
			slog.ErrorContext(ctx, "failed to verify pet ownership",
				"error", err,
				"pet_id", *input.PetID,
				"clinic_id", input.ClinicID,
			)
			return nil, apperrors.Wrap(err, "failed to verify pet ownership")
		}
	}
	if input.MedicalRecordID != nil {
		if _, err := s.medicalRecordRepo.FindByID(ctx, input.ClinicID, *input.MedicalRecordID); err != nil {
			slog.ErrorContext(ctx, "failed to verify medical record ownership",
				"error", err,
				"medical_record_id", *input.MedicalRecordID,
				"clinic_id", input.ClinicID,
			)
			return nil, apperrors.Wrap(err, "failed to verify medical record ownership")
		}
	}

	dup, err := s.dupChecker.IsDuplicate(ctx, input.ClinicID, input.ExamTypeID, input.Date, input.PetID)
	if err != nil {
		slog.ErrorContext(ctx, "lab import duplicate check failed",
			"error", err,
			"clinic_id", input.ClinicID,
			"exam_type_id", input.ExamTypeID,
			"job_id", input.JobID.String(),
		)
		return nil, apperrors.Wrap(err, "failed to check duplicate exam")
	}
	if dup {
		slog.InfoContext(ctx, "lab import exam skipped (duplicate)",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("exam_type_id", input.ExamTypeID),
			slog.String("job_id", input.JobID.String()),
		)
		return &LabExamPersistResult{
			Duplicate: true,
			JobID:     input.JobID,
		}, nil
	}

	jobID := input.JobID
	exam := &model.Examination{
		ClinicID:        input.ClinicID,
		PetID:           input.PetID,
		MedicalRecordID: input.MedicalRecordID,
		ExamTypeID:      input.ExamTypeID,
		JobID:           &jobID,
		Date:            input.Date,
		Machine:         input.Machine,
		// result_entered: lab 結果は到着時点で値が確定しているため pending/in_progress をスキップ
		Status: model.ExaminationStatusResultEntered,
	}
	if err := s.examRepo.Create(ctx, exam); err != nil {
		// DB unique 制約違反（TOCTOU 安全ネット）: 重複として扱う
		if apperrors.IsAlreadyExists(err) {
			slog.InfoContext(ctx, "lab import exam skipped (db duplicate on create)",
				slog.Uint64("clinic_id", input.ClinicID),
				slog.Uint64("exam_type_id", input.ExamTypeID),
				slog.String("job_id", input.JobID.String()),
			)
			return &LabExamPersistResult{
				Duplicate: true,
				JobID:     input.JobID,
			}, nil
		}
		slog.ErrorContext(ctx, "lab import exam create failed",
			"error", err,
			"clinic_id", input.ClinicID,
			"job_id", input.JobID.String(),
		)
		return nil, apperrors.Wrap(err, "failed to create exam from lab import")
	}

	if len(input.Items) > 0 {
		items := buildExamResults(exam.ID, input.Items)
		if _, _, err := s.examRepo.ReplaceItemsByExamID(ctx, input.ClinicID, exam.ID, items); err != nil {
			slog.ErrorContext(ctx, "lab import exam items save failed",
				"error", err,
				"clinic_id", input.ClinicID,
				"exam_id", exam.ID,
				"job_id", input.JobID.String(),
			)
			// P2-7 (PR #186 review, id 3572087707): item 保存失敗のまま exam を残すと、
			// 結果なしの孤児 exam が (clinic_id, exam_type_id, date, pet_id) を占有し、
			// retry が dupChecker.IsDuplicate に誤ってスキップされ自己修復できなくなる。
			// 孤児 exam を削除して retry が同一キーで再試行できるようにする
			// （IsDuplicate は GORM soft-delete スコープで deleted_at IS NULL のみ対象）。
			if delErr := s.examRepo.Delete(ctx, input.ClinicID, exam.ID); delErr != nil {
				slog.ErrorContext(ctx, "lab import orphan exam cleanup failed",
					"error", delErr,
					"clinic_id", input.ClinicID,
					"exam_id", exam.ID,
					"job_id", input.JobID.String(),
				)
			}
			return nil, apperrors.Wrap(err, fmt.Sprintf("failed to save exam items for exam %d", exam.ID))
		}
	}

	slog.InfoContext(ctx, "lab import exam persisted",
		slog.Uint64("clinic_id", input.ClinicID),
		slog.Uint64("exam_id", exam.ID),
		slog.Int("item_count", len(input.Items)),
		slog.String("job_id", input.JobID.String()),
	)

	return &LabExamPersistResult{
		ExamID:    exam.ID,
		ItemCount: len(input.Items),
		Duplicate: false,
		JobID:     input.JobID,
	}, nil
}

// PersistBatch は全行を処理し、個別行エラーを LabExamPersistResult.RowError に記録する。
// バッチ全体のシステムエラー（context キャンセル等）のみ関数エラーとして返す。
// 呼び出し元は RowError を集計して LabImportJob のカウンタを更新する。
func (s *labImportExaminationService) PersistBatch(ctx context.Context, inputs []LabExamPersistInput) ([]*LabExamPersistResult, error) {
	results := make([]*LabExamPersistResult, 0, len(inputs))
	for _, input := range inputs {
		// context キャンセルはバッチ全体を中断する
		if err := ctx.Err(); err != nil {
			return results, apperrors.Wrap(err, "lab import batch context cancelled")
		}
		res, err := s.persistExam(ctx, input)
		if err != nil {
			// 個別行エラー: RowError に記録して継続（バッチ中断しない）
			results = append(results, &LabExamPersistResult{
				RowError: err,
				JobID:    input.JobID,
			})
			continue
		}
		results = append(results, res)
	}
	return results, nil
}
