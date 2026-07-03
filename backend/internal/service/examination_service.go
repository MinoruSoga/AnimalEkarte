package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateExaminationInput は検査作成の入力DTO
type CreateExaminationInput struct {
	MedicalRecordID *uint64
	PetID           *uint64
	ExamTypeID      uint64
	DoctorID        *uint64
	Date            time.Time
	ResultSummary   string
	Machine         string
	Status          model.ExaminationStatus
}

// UpdateExaminationInput は検査更新のサービス入力 DTO
type UpdateExaminationInput struct {
	MedicalRecordID *uint64
	PetID           *uint64
	ExamTypeID      *uint64
	DoctorID        *uint64
	Date            *time.Time
	ResultSummary   *string
	Machine         *string
	Status          *model.ExaminationStatus
}

// UpsertExamItemInput は検査項目（exam_results）の一括登録入力 DTO。
// status / is_abnormal はサーバ側で計算するため受け付けない（信頼境界はサーバ）。
type UpsertExamItemInput struct {
	ExamTypeFieldID *uint64
	Name            string
	InspectionValue string
	NormalValue     string
	Unit            string
	ReferenceValue  string
	RefMin          *float64
	RefMax          *float64
	SortOrder       int
}

// computeExamResultStatus は inspection_value を float としてパースし、
// ref_min / ref_max と比較して status と is_abnormal を導出する。
//
// 仕様:
//   - inspection_value が空・パース不能 → (normal, false)
//   - ref_min が指定され v < ref_min → (low, true)
//   - ref_max が指定され v > ref_max → (high, true)
//   - 範囲内 → (normal, false)
//   - ref_min == ref_max == nil → (normal, false)（比較できない）
func computeExamResultStatus(inspectionValue string, refMin, refMax *float64) (model.ExaminationResultStatus, bool) {
	trimmed := strings.TrimSpace(inspectionValue)
	if trimmed == "" {
		return model.ExaminationResultStatusNormal, false
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return model.ExaminationResultStatusNormal, false
	}
	if refMin != nil && v < *refMin {
		return model.ExaminationResultStatusLow, true
	}
	if refMax != nil && v > *refMax {
		return model.ExaminationResultStatusHigh, true
	}
	return model.ExaminationResultStatusNormal, false
}

func buildExaminationUpdate(input UpdateExaminationInput) map[string]any {
	fields := make(map[string]any)
	if input.MedicalRecordID != nil {
		fields["medical_record_id"] = *input.MedicalRecordID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.ExamTypeID != nil {
		fields["exam_type_id"] = *input.ExamTypeID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.ResultSummary != nil {
		fields["result_summary"] = *input.ResultSummary
	}
	if input.Machine != nil {
		fields["machine"] = *input.Machine
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	return fields
}

type ExaminationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	Create(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error)
	Update(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	ListItems(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error)
	// ReplaceItems は検査項目を一括置換する（PUT セマンティクス）。actorID は監査ログ用の操作スタッフ ID
	// （nil = システム実行）。BE-refactor.md R1-2: 実削除が発生した場合は同一 tx 内で fail-closed 監査する。
	ReplaceItems(ctx context.Context, clinicID, examID uint64, actorID *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error)
}

type examinationService struct {
	repo         repository.ExaminationRepository
	medRec       repository.MedicalRecordRepository
	examTypeRepo repository.ExamTypeRepository
	auditTx      AuditTxLogger
	transactor   repository.Transactor
}

func NewExaminationService(repo repository.ExaminationRepository, medRec repository.MedicalRecordRepository, examTypeRepo repository.ExamTypeRepository, auditTx AuditTxLogger, transactor repository.Transactor) ExaminationService {
	return &examinationService{repo: repo, medRec: medRec, examTypeRepo: examTypeRepo, auditTx: auditTx, transactor: transactor}
}

func (s *examinationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list examinations", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list examinations")
	}
	return items, total, nil
}

func (s *examinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get examination", "error", err)
		return nil, apperrors.Wrap(err, "failed to get examination")
	}
	return result, nil
}

func (s *examinationService) Create(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error) {
	// HC-005: 親カルテが確定済みの場合は作成拒否
	if input.MedicalRecordID != nil {
		parent, err := s.medRec.FindByID(ctx, clinicID, *input.MedicalRecordID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find medical record", "error", err)
			return nil, apperrors.Wrap(err, "failed to find medical record")
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return nil, apperrors.WrapConflict("確定済みカルテに検査を追加できません")
		}
	}

	// クロステナント write 防止: 別 clinic の exam_type を紐付けると、その exam_type が持つ
	// 異常値判定の基準値/単位（exam_type_fields）が検査記録に混入する（#124 同型）。所有権を検証する。
	if input.ExamTypeID != 0 {
		if _, err := s.examTypeRepo.FindByID(ctx, clinicID, input.ExamTypeID); err != nil {
			return nil, apperrors.Wrap(err, "failed to verify exam type ownership")
		}
	}

	status := input.Status
	if status == "" {
		status = model.ExaminationStatusPending
	}
	exam := &model.Examination{
		ClinicID:        clinicID,
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		ExamTypeID:      input.ExamTypeID,
		DoctorID:        input.DoctorID,
		Date:            input.Date,
		ResultSummary:   input.ResultSummary,
		Machine:         input.Machine,
		Status:          status,
	}
	if err := s.repo.Create(ctx, exam); err != nil {
		slog.ErrorContext(ctx, "failed to create examination", "error", err)
		return nil, apperrors.Wrap(err, "failed to create examination")
	}
	slog.InfoContext(ctx, "examination created", slog.Uint64("clinic_id", clinicID), slog.Uint64("examination_id", exam.ID))
	return exam, nil
}

func (s *examinationService) Update(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get examination", "error", err)
		return nil, apperrors.Wrap(err, "failed to get examination")
	}
	if existing.Status == model.ExaminationStatusConfirmed {
		return nil, apperrors.WrapInvalidInput("確定済みの検査は編集できません")
	}

	// HC-003: 親カルテが確定済みの場合は編集拒否
	if existing.MedicalRecordID != nil {
		parent, err := s.medRec.FindByID(ctx, clinicID, *existing.MedicalRecordID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find medical record", "error", err)
			return nil, apperrors.Wrap(err, "failed to find medical record")
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return nil, apperrors.WrapConflict("確定済みカルテの検査は編集できません")
		}
	}

	// クロステナント write 防止: 貼り替え先 exam_type が caller の clinic に属することを検証する。
	if input.ExamTypeID != nil {
		if _, err := s.examTypeRepo.FindByID(ctx, clinicID, *input.ExamTypeID); err != nil {
			return nil, apperrors.Wrap(err, "failed to verify exam type ownership")
		}
	}

	fields := buildExaminationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	exam, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update examination", "error", err)
		return nil, apperrors.Wrap(err, "failed to update examination")
	}
	slog.InfoContext(ctx, "examination updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("examination_id", id))
	return exam, nil
}

// ListItems は検査項目一覧を返す。clinic_id 隔離は repository の JOIN 条件で保証する。
// 親 exam の存在確認は FindByID で先行する（404 を返すため）。
func (s *examinationService) ListItems(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, examID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find examination")
	}
	items, err := s.repo.FindAllItemsByExamID(ctx, clinicID, examID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list examination items", "error", err, "exam_id", examID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list examination items")
	}
	return items, nil
}

// ReplaceItems は検査項目を一括置換する（PUT セマンティクス）。
//
// 仕様:
//  1. 親 exam の存在を FindByID で確認（P1）
//  2. 親 exam が confirmed の場合は 400 で拒否（既存 Update と同方針）
//  3. 各 input の inspection_value と ref_min / ref_max から status / is_abnormal を導出（FE 送信値は無視）
//  4. repository の ReplaceItemsByExamID（トランザクション内で全削除→一括挿入）に委譲
//  5. 実削除が発生した場合（deletedCount > 0）は同一 tx 内で監査ログを書き込む。監査書込が失敗したら
//     tx を rollback する（best-effort ではなく fail-closed。BE-refactor.md R1-2・#211 と同方針）。
func (s *examinationService) ReplaceItems(ctx context.Context, clinicID, examID uint64, actorID *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, examID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find examination")
	}
	if existing.Status == model.ExaminationStatusConfirmed {
		return nil, apperrors.WrapInvalidInput("確定済みの検査は編集できません")
	}

	// #124 防止: request の exam_type_field が caller の clinic に属する当該検査種別の
	// フィールドであることを検証する。別 clinic / 別種別のフィールドを紐付けると、その
	// フィールドが持つ基準値・単位が検査結果に誤適用される（#124 実害と同型・クロステナント）。
	// 03bf1cb5 は親 exam_type_id のみ検証しており、この field 経路は未検証だった。
	hasFieldRef := false
	for _, in := range inputs {
		if in.ExamTypeFieldID != nil {
			hasFieldRef = true
			break
		}
	}
	if hasFieldRef {
		examType, err := s.examTypeRepo.FindByID(ctx, clinicID, existing.ExamTypeID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to verify exam type ownership")
		}
		validFieldIDs := make(map[uint64]struct{}, len(examType.Items))
		for i := range examType.Items {
			validFieldIDs[examType.Items[i].ID] = struct{}{}
		}
		for _, in := range inputs {
			if in.ExamTypeFieldID != nil {
				if _, ok := validFieldIDs[*in.ExamTypeFieldID]; !ok {
					return nil, apperrors.WrapInvalidInput("exam_type_field が当該検査種別に属していません（別クリニック/別種別の項目は紐付けできません）")
				}
			}
		}
	}

	items := make([]model.ExamResult, 0, len(inputs))
	for _, in := range inputs {
		status, isAbnormal := computeExamResultStatus(in.InspectionValue, in.RefMin, in.RefMax)
		items = append(items, model.ExamResult{
			ExamID:          examID,
			ExamTypeItemID:  in.ExamTypeFieldID,
			Name:            in.Name,
			InspectionValue: in.InspectionValue,
			NormalValue:     in.NormalValue,
			Unit:            in.Unit,
			ReferenceValue:  in.ReferenceValue,
			RefMin:          in.RefMin,
			RefMax:          in.RefMax,
			IsAbnormal:      isAbnormal,
			Status:          status,
			SortOrder:       in.SortOrder,
		})
	}

	// #211/R1-2 tx 内監査による原子的置換: スナップショット読取→削除/挿入→削除監査 を単一トランザクションで
	// 実行する。監査書込が失敗したら tx 全体を rollback し、削除・挿入も巻き戻す（監査なしの検査結果削除を
	// 許さない＝fail-closed）。exam_results は hard-delete のため old_value が唯一の耐久記録であり、
	// 旧コードでは「置換 commit 後に監査を書く」経路自体が存在しなかった（audit_tx_inventory_lint_test.go
	// が発見した無監査ギャップ）。スナップショットも同一 tx 内で取得し TOCTOU 窓を作らない。
	var saved []model.ExamResult
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		before, err := s.repo.FindAllItemsByExamID(txCtx, clinicID, examID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to snapshot existing examination items before replace", "error", err, "exam_id", examID, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to load existing examination items")
		}

		replaced, deletedCount, err := s.repo.ReplaceItemsByExamID(txCtx, clinicID, examID, items)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to replace examination items", "error", err, "exam_id", examID, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to replace examination items")
		}
		saved = replaced

		// 実際に削除が発生した場合のみ監査する（純粋な新規挿入は削除を伴わない）。ゲートはスナップショット
		// 件数でなく DELETE の実削除数（deletedCount）に基づく（#211 security MEDIUM-1 と同方針: 並行 INSERT
		// 競合下でスナップショット 0 件でも実削除>0 を取りこぼさない）。監査書込失敗は tx を rollback する。
		if deletedCount > 0 {
			actorType := model.AuditActorTypeSystem
			if actorID != nil {
				actorType = model.AuditActorTypeStaff
			}
			if err := s.auditTx.LogEntryTx(txCtx, &AuditLogInput{
				ClinicID:   &clinicID,
				ActorID:    actorID,
				ActorType:  actorType,
				Action:     model.AuditActionExamResultReplace,
				Resource:   model.AuditResourceExamResult,
				ResourceID: &examID,
				OldValue:   extractExamResultsAudit(before),
				NewValue:   extractExamResultsAudit(saved),
				Metadata: map[string]any{
					"exam_id":       examID,
					"deleted_count": deletedCount,
					"new_count":     len(saved),
				},
			}); err != nil {
				slog.ErrorContext(txCtx, "audit log failed for examination items replace; rolling back deletion", "error", err, "exam_id", examID, "clinic_id", clinicID)
				return apperrors.Wrap(err, "failed to write examination items deletion audit")
			}
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to replace examination items in transaction")
	}

	slog.InfoContext(ctx, "examination items replaced",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("examination_id", examID),
		slog.Int("item_count", len(saved)),
	)
	return saved, nil
}

// extractExamResultsAudit は監査ログの old_value/new_value に格納する検査結果値のスナップショットを構築する。
// 飼主/患者の識別情報は含まず、行 ID・フィールド定義参照・検査値のみを記録する
// （extractCheckupFieldResultsAudit と同方針）。
func extractExamResultsAudit(results []model.ExamResult) []map[string]any {
	if len(results) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(results))
	for i := range results {
		r := results[i]
		out = append(out, map[string]any{
			"id":                 r.ID,
			"exam_type_field_id": r.ExamTypeItemID,
			"name":               r.Name,
			"inspection_value":   r.InspectionValue,
			"is_abnormal":        r.IsAbnormal,
			"status":             string(r.Status),
		})
	}
	return out
}

func (s *examinationService) Delete(ctx context.Context, clinicID, id uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find examination")
	}

	// HC-003: 親カルテが確定済みの場合は削除拒否
	if existing.MedicalRecordID != nil {
		parent, err := s.medRec.FindByID(ctx, clinicID, *existing.MedicalRecordID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find medical record", "error", err)
			return apperrors.Wrap(err, "failed to find medical record")
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテの検査は削除できません")
		}
	}

	// FK依存チェック: 検査に紐付く検査明細が存在する場合は削除を拒否
	itemCount, err := s.repo.CountItemsByExamID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check examination item dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check examination item dependencies")
	}
	if itemCount > 0 {
		return apperrors.WrapConflict("検査結果が紐付いているため削除できません。先に検査結果を削除してください")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete examination", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete examination")
	}

	slog.InfoContext(ctx, "examination deleted",
		slog.Uint64("examination_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}
