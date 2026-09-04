package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
	ExamTypeFieldID *uint64
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
//   - pet_id と medical_record_id を同時に受け取る場合は同一患者相関を fail-closed で検証する
//   - 完全同一ペイロード（header + items）の再インポートのみスキップする（Issue #249 R-3）
//   - 同日・同検査種別でも内容が異なれば新規 exam として保存する
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
			ExamTypeItemID:  item.ExamTypeFieldID,
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
	PersistExam(ctx context.Context, input LabExamPersistInput) (*LabExamPersistResult, error)
	// PersistBatch は複数 exam を順次永続化し全行の結果を返す。
	// 個別行のエラーは LabExamPersistResult.RowError に記録する。
	// バッチ全体のエラー（context キャンセル等）は関数エラーとして返す。
	PersistBatch(ctx context.Context, inputs []LabExamPersistInput) ([]*LabExamPersistResult, error)
}

// LabImportDuplicateChecker は lab import の完全同一ペイロード再インポートを検知する。
//
// 候補は (clinic_id, exam_type_id, date, pet_id) で絞り、その中で medical_record_id・
// machine・exam_results ペイロード項目がすべて一致する既存 exam がある場合のみ true。
// 日付粒度の 4-col 一致だけでは重複にしない（Issue #249 R-3 / PO ruling）。
// DB レベルの unique constraint がない場合は TOCTOU 競合で重複が発生しうる。
// persistExam では AlreadyExists エラーも Duplicate として処理する安全ネットを設ける。
type LabImportDuplicateChecker interface {
	// IsDuplicate は入力ペイロードと完全一致する既存 exam がある場合に true を返す。
	// JobID・Status・派生 IsAbnormal/Status・id/timestamps は identity に含めない。
	IsDuplicate(ctx context.Context, input LabExamPersistInput) (bool, error)
}

// labImportExaminationService は LabImportExaminationService の実装。
type labImportExaminationService struct {
	examRepo          examinationImportRepo
	dupChecker        LabImportDuplicateChecker
	examTypeRepo      ExamTypeRepository
	petRepo           petFinder
	medicalRecordRepo medicalRecordFinder
	transactor        Transactor
}

// NewLabImportExaminationService は LabImportExaminationService を初期化して返す。
// transactor は exam Create と exam_results 置換を同一 transaction に収める（BE-refactor.md MRC-05 / X-06）。
func NewLabImportExaminationService(
	examRepo examinationImportRepo,
	dupChecker LabImportDuplicateChecker,
	examTypeRepo ExamTypeRepository,
	petRepo petFinder,
	medicalRecordRepo medicalRecordFinder,
	transactor Transactor,
) LabImportExaminationService {
	return &labImportExaminationService{
		examRepo:          examRepo,
		dupChecker:        dupChecker,
		examTypeRepo:      examTypeRepo,
		petRepo:           petRepo,
		medicalRecordRepo: medicalRecordRepo,
		transactor:        transactor,
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

	if err := s.verifyLabExamPersistOwnership(ctx, input); err != nil {
		return nil, err
	}

	dup, err := s.dupChecker.IsDuplicate(ctx, input)
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
		return skippedDuplicateLabExam(input), nil
	}

	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("lab import examination transaction dependency is required")
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

	// exam 本体と exam_results は 1 つの検査結果 business graph なので同一 transaction で原子的に書く
	// （BE-refactor.md MRC-05 / X-06）。Replace 失敗時は rollback により孤児 exam を残さない。
	// 既に ambient tx がある（device receive/attach）ときは内側で新規 Transaction を開かない。
	duplicateOnCreate, writeErr := s.persistLabExamWrite(ctx, input, exam)
	if writeErr != nil {
		return nil, writeErr
	}
	if duplicateOnCreate {
		slog.InfoContext(ctx, "lab import exam skipped (db duplicate on create)",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("exam_type_id", input.ExamTypeID),
			slog.String("job_id", input.JobID.String()),
		)
		return skippedDuplicateLabExam(input), nil
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

func skippedDuplicateLabExam(input LabExamPersistInput) *LabExamPersistResult {
	return &LabExamPersistResult{
		Duplicate: true,
		JobID:     input.JobID,
	}
}

func (s *labImportExaminationService) persistLabExamWrite(
	ctx context.Context,
	input LabExamPersistInput,
	exam *model.Examination,
) (bool, error) {
	var duplicateOnCreate bool
	write := func(txCtx context.Context) error {
		dup, err := s.writeLabExamGraph(txCtx, input, exam)
		if err != nil {
			return err
		}
		duplicateOnCreate = dup
		return nil
	}
	if persistence.TxFromContext(ctx) != nil {
		err := write(ctx)
		return duplicateOnCreate, err
	}
	err := s.transactor.WithTx(ctx, write)
	return duplicateOnCreate, err
}

func (s *labImportExaminationService) PersistExam(ctx context.Context, input LabExamPersistInput) (*LabExamPersistResult, error) {
	return s.persistExam(ctx, input)
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

func requireOwnedExamTypeFields(examType *model.ExaminationType, items []LabExamItemInput) error {
	owned := make(map[uint64]struct{})
	if examType != nil {
		for i := range examType.Items {
			owned[examType.Items[i].ID] = struct{}{}
		}
	}
	for _, item := range items {
		if item.ExamTypeFieldID == nil {
			continue
		}
		if _, ok := owned[*item.ExamTypeFieldID]; !ok {
			return apperrors.WrapInvalidInput("exam_type_field が当該検査種別に属していません（別クリニック/別種別の項目は紐付けできません）")
		}
	}
	return nil
}

// computeExamResultStatus: ③で複製していたが⑦の examination_service 移動で原本に統合（計画通りの自己解消）。
