package service

import (
	"context"
	"log/slog"
	"maps"
	"strconv"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ─── Input DTOs ───────────────────────────────────────────────────────────────

// CreateTreatmentInput は治療項目作成の入力DTO（HTTPを知らない）
type CreateTreatmentInput struct {
	ItemType       model.TreatmentItemType
	ConsultationID *uint64
	ProcedureID    *uint64
	MedicineID     *uint64
	InventoryID    *uint64
	UnitPrice      int64
	Quantity       float64
	IsSelected     bool
	Status         string
	Content        string
	Memo           string
	AdminRoute     string
	IsInsurance    bool
	DiscountRate   float64
	DiscountAmount int64
	SortOrder      int
	ActorID        *uint64 // #201 監査ログ用: 操作スタッフ ID（nil = システム）
}

// UpdateTreatmentInput は治療項目更新の入力DTO（ポインタ型 = nil は未送信）
type UpdateTreatmentInput struct {
	ItemType       *model.TreatmentItemType
	ConsultationID *uint64
	ProcedureID    *uint64
	MedicineID     *uint64
	InventoryID    *uint64
	UnitPrice      *int64
	Quantity       *float64
	IsSelected     *bool
	Status         *string
	Content        *string
	Memo           *string
	AdminRoute     *string
	IsInsurance    *bool
	DiscountRate   *float64
	DiscountAmount *int64
	SortOrder      *int
	ActorID        *uint64 // #201 監査ログ用: 操作スタッフ ID（nil = システム）
}

// BulkUpdateTreatmentsInput は並び順一括更新の入力DTO
type BulkUpdateTreatmentsInput struct {
	Treatments []BulkTreatmentItem
}

// BulkTreatmentItem は一括更新の個別項目
type BulkTreatmentItem struct {
	ID        uint64
	SortOrder int
}

// ─── Interface ────────────────────────────────────────────────────────────────

// GORMのzero-value問題（false/0/"" がスキップされる）を回避するために使用する。
func buildTreatmentUpdate(input *UpdateTreatmentInput) map[string]any {
	fields := map[string]any{}
	if input.ItemType != nil {
		fields["item_type"] = *input.ItemType
	}
	if input.ConsultationID != nil {
		fields["consultation_id"] = *input.ConsultationID
	}
	if input.ProcedureID != nil {
		fields["procedure_id"] = *input.ProcedureID
	}
	if input.MedicineID != nil {
		fields["medicine_id"] = *input.MedicineID
	}
	if input.InventoryID != nil {
		fields["inventory_id"] = *input.InventoryID
	}
	if input.UnitPrice != nil {
		fields["unit_price"] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields["quantity"] = *input.Quantity
	}
	if input.IsSelected != nil {
		fields["is_selected"] = *input.IsSelected
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Content != nil {
		fields["content"] = *input.Content
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.AdminRoute != nil {
		fields["admin_route"] = *input.AdminRoute
	}
	if input.IsInsurance != nil {
		fields["is_insurance"] = *input.IsInsurance
	}
	if input.DiscountRate != nil {
		fields["discount_rate"] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields["discount_amount"] = *input.DiscountAmount
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}

// TreatmentService は治療項目のビジネスロジックインターフェース
type TreatmentService interface {
	List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	// ListPetHistory は #158/#159 飼主レポート用: ペット単位の治療履歴を medical_records.date 降順で返す。
	ListPetHistory(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error)
	Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error)
	Update(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64) error
	BulkUpdateSortOrder(ctx context.Context, clinicID, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type treatmentService struct {
	repos *repository.Repositories
	// auditSvc は非nilかどうかのみを「逸脱audit機能の有効/無効」フラグとして使う。
	// BE-refactor.md R1-2 (D1): treatment は repos.Transaction(ctx, func(txRepos...)) で tx 境界を
	// 作る（Transactor.WithTx + ctx-txKey とは別機構）。この機構では tx が txRepos（tx にバインドされた
	// *Repositories）に宿り、ctx には伝播しないため、AuditTxLogger.LogEntryTx（dbOrTx が ctx の txKey を
	// 見る）は参加できない — base db に書いてしまい fail-closed にならない。そのため実際の書込は
	// auditDoseDeviationTx が txRepos.Audit.CreateTx を直接使う（tx にバインドされた Audit リポジトリ
	// インスタンス自体が tx 参加を保証する）。auditSvc のメソッドは呼ばない。
	auditSvc AuditService
}

// NewTreatmentService はTreatmentServiceを初期化して返す（audit なし・後方互換）。
// 本番配線は NewTreatmentServiceWithAudit を使う（#201 逸脱 audit のため）。
func NewTreatmentService(repos *repository.Repositories) TreatmentService {
	return &treatmentService{repos: repos}
}

// NewTreatmentServiceWithAudit は逸脱 audit 機能を有効化する（#201 B-2）。
// auditSvc は非nilフラグとしてのみ使う（上記 struct コメント参照）。
func NewTreatmentServiceWithAudit(repos *repository.Repositories, auditSvc AuditService) TreatmentService {
	return &treatmentService{repos: repos, auditSvc: auditSvc}
}

func (s *treatmentService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	treatment, err := s.repos.Treatment.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get treatment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get treatment")
	}
	return treatment, nil
}

func (s *treatmentService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	treatments, err := s.repos.Treatment.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list treatments", "error", err)
		return nil, apperrors.Wrap(err, "failed to list treatments")
	}
	return treatments, nil
}

func (s *treatmentService) ListPetHistory(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error) {
	if filter.ItemType != nil {
		if err := validateTreatmentItemType(*filter.ItemType); err != nil {
			return nil, 0, apperrors.Wrap(err, "failed to validate treatment item type")
		}
	}
	treatments, total, err := s.repos.Treatment.FindHistoryByPetID(ctx, clinicID, petID, filter, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list pet treatment history", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list pet treatment history")
	}
	return treatments, total, nil
}

func (s *treatmentService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateTreatmentInput) (*model.Treatment, error) {
	if err := validateTreatmentItemType(input.ItemType); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate treatment item type")
	}
	if input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	if input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgQuantityPositive)
	}
	if input.DiscountRate < 0 || input.DiscountRate > 100 {
		return nil, apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
	}

	status := model.TreatmentStatusPending
	if input.Status != "" {
		s, err := parseTreatmentStatus(input.Status)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create treatment", "error", err)
			return nil, apperrors.Wrap(err, "failed to create treatment")
		}
		status = s
	}

	// クロステナント write 防止: request 由来の clinic-scoped マスタFK
	// (medicine/procedure/consultation/inventory) が caller の clinic に属することを検証する。
	// 別 clinic のマスタ参照は NotFound で遮断し #124/#125 同型の mislink を防ぐ。
	if err := s.validateTreatmentMasterFKs(ctx, clinicID, input.MedicineID, input.ProcedureID, input.ConsultationID, input.InventoryID); err != nil {
		return nil, err
	}

	var treatment *model.Treatment
	var doseEval *SavedDoseEvaluation

	// ─── Transaction ───
	err := s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
		// テナント所有権 + 確定ロック検証（Update/Delete と対称化・healthcare review CRITICAL）。
		// treatments は自前 clinic_id を持たず medical_records 経由で隔離するため、所有権を Create でも明示検証する。
		// BE-refactor.md X-11: LockByIDForUpdate の行ロックで finalize（medical_record_repository.Update の
		// draft-only WHERE）と直列化し、確定と同時の治療追加が確定済みカルテに混入する競合を防ぐ。
		parent, err := txRepos.MedicalRecord.LockByIDForUpdate(ctx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to verify medical record ownership", "error", err)
			return apperrors.Wrap(err, "failed to verify medical record ownership")
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテには治療を追加できません")
		}

		treatment = &model.Treatment{
			MedicalRecordID: medicalRecordID,
			ItemType:        input.ItemType,
			ConsultationID:  input.ConsultationID,
			ProcedureID:     input.ProcedureID,
			MedicineID:      input.MedicineID,
			InventoryID:     input.InventoryID,
			UnitPrice:       input.UnitPrice,
			Quantity:        input.Quantity,
			IsSelected:      input.IsSelected,
			Status:          status,
			Content:         input.Content,
			Memo:            input.Memo,
			AdminRoute:      input.AdminRoute,
			IsInsurance:     input.IsInsurance,
			DiscountRate:    input.DiscountRate,
			DiscountAmount:  input.DiscountAmount,
			SortOrder:       input.SortOrder,
		}

		// #201 B-2: per_weight 医薬の保存時 BE 再検証＋スナップショット値固定（C1/両上限/species 一致）。
		eval, derr := s.evaluateDoseForSave(ctx, txRepos, clinicID, medicalRecordID, input.ItemType, input.MedicineID, input.Quantity)
		if derr != nil {
			return derr // species 不一致など fail-closed
		}
		if eval != nil {
			doseEval = eval
			applyDoseSnapshotToTreatment(ctx, treatment, eval)
		}

		// 1. Create Treatment
		if err := txRepos.Treatment.Create(ctx, treatment); err != nil {
			return apperrors.Wrap(err, "failed to create treatment")
		}

		// 2. Decrease Stock (if applicable)
		if input.MedicineID != nil || input.InventoryID != nil {
			var targetInvID uint64
			if input.InventoryID != nil {
				targetInvID = *input.InventoryID
			} else {
				targetInvID = *input.MedicineID
			}

			if targetInvID > 0 {
				if err := txRepos.Inventory.DecreaseStock(ctx, targetInvID, input.Quantity); err != nil {
					return apperrors.Wrap(err, "failed to decrease inventory stock")
				}
			}
		}

		// #201 B-2 / BE-refactor.md R1-2: 逸脱（過量/過少/著しい上書き）を監査記録（fail-closed）。
		// tx 内で失敗すると treatment 作成・在庫減算ごとロールバックする。
		if doseEval != nil && input.MedicineID != nil {
			if err := s.auditDoseDeviationTx(ctx, txRepos, clinicID, input.ActorID, treatment.ID, *input.MedicineID, doseEval); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to create treatment", "error", err)
		return nil, apperrors.Wrap(err, "failed to create treatment")
	}

	slog.InfoContext(ctx, "treatment created with atomic inventory sync",
		slog.Uint64("treatment_id", treatment.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return treatment, nil
}

func (s *treatmentService) Update(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64, input *UpdateTreatmentInput) (*model.Treatment, error) {
	// 所属確認（clinic_id + id で検索）
	existing, err := s.repos.Treatment.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get treatment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get treatment")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}

	// HC-004: 親カルテが確定済みの場合は編集拒否
	parent, err := s.repos.MedicalRecord.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict("確定済みカルテの治療は編集できません")
	}

	// クロステナント write 防止: 変更後に貼り替わる clinic-scoped マスタFK の所有権を検証する。
	if err := s.validateTreatmentMasterFKs(ctx, clinicID, input.MedicineID, input.ProcedureID, input.ConsultationID, input.InventoryID); err != nil {
		return nil, err
	}

	if input.ItemType != nil {
		if err := validateTreatmentItemType(*input.ItemType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate treatment item type")
		}
	}
	if input.Status != nil {
		if _, err := parseTreatmentStatus(*input.Status); err != nil {
			slog.ErrorContext(ctx, "failed to update treatment", "error", err)
			return nil, apperrors.Wrap(err, "failed to update treatment")
		}
	}
	if input.Quantity != nil && *input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgQuantityPositive)
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	if input.DiscountRate != nil && (*input.DiscountRate < 0 || *input.DiscountRate > 100) {
		return nil, apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
	}

	fields := buildTreatmentUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}

	// #201 B-2: quantity/medicine が変わる場合のみ保存時 BE 再検証＋スナップショット更新。
	// 再検証の読み出しと UPDATE を同一トランザクションに束ね、並行マスタ変更による
	// スナップショット不整合（TOCTOU）を防ぐ（security review MEDIUM-1）。
	// quantity/medicine/item_type のいずれかが変わると dose スナップショットの再評価対象になる。
	doseRelevant := input.Quantity != nil || input.MedicineID != nil || input.ItemType != nil
	var doseEval *SavedDoseEvaluation
	var doseMedicineID uint64
	if txErr := s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
		// テナント所有権 + 確定ロック検証（Create と対称化・BE-refactor.md H-8a）。
		// :331-338 の事前チェックは fast-fail として維持しつつ、tx 内でも LockByIDForUpdate
		// の行ロックで finalize と直列化し、チェック通過後〜Update 実行前に finalize が
		// 割り込むレースを防ぐ。
		parent, err := txRepos.MedicalRecord.LockByIDForUpdate(ctx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to verify medical record ownership", "error", err)
			return apperrors.Wrap(err, "failed to verify medical record ownership")
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテの治療は編集できません")
		}

		if doseRelevant {
			effItemType := existing.ItemType
			if input.ItemType != nil {
				effItemType = *input.ItemType
			}
			effMedicineID := existing.MedicineID
			if input.MedicineID != nil {
				effMedicineID = input.MedicineID
			}
			effQty := existing.Quantity
			if input.Quantity != nil {
				effQty = *input.Quantity
			}
			eval, derr := s.evaluateDoseForSave(ctx, txRepos, clinicID, medicalRecordID, effItemType, effMedicineID, effQty)
			if derr != nil {
				return derr // species 不一致など fail-closed
			}
			if eval != nil {
				doseEval = eval
				doseMedicineID = *effMedicineID
				maps.Copy(fields, doseSnapshotColumns(ctx, eval))
			} else if treatmentHasDoseSnapshot(existing) {
				// per_weight 対象でなくなった（薬剤/item_type 変更）→ stale スナップショットをクリア（L-3）。
				maps.Copy(fields, clearedDoseColumns())
			}
		}
		if err := txRepos.Treatment.Update(ctx, clinicID, treatmentID, fields); err != nil {
			return err
		}
		// #201 B-2 / BE-refactor.md R1-2: 逸脱監査を fail-closed 化。失敗すると Update ごとロールバックする。
		if doseEval != nil {
			return s.auditDoseDeviationTx(ctx, txRepos, clinicID, input.ActorID, treatmentID, doseMedicineID, doseEval)
		}
		return nil
	}); txErr != nil {
		slog.ErrorContext(ctx, "failed to update treatment", "error", txErr)
		return nil, apperrors.Wrap(txErr, "failed to update treatment")
	}

	slog.InfoContext(ctx, "treatment updated",
		slog.Uint64("treatment_id", treatmentID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID))

	treatment, err := s.repos.Treatment.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get updated treatment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get updated treatment")
	}
	return treatment, nil
}

func (s *treatmentService) Delete(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64) error {
	existing, err := s.repos.Treatment.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get treatment")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}

	// HC-004: 親カルテが確定済みの場合は削除拒否（BE-refactor.md H-8b）。
	// finalized チェックと Delete を同一 tx に束ね、閉包先頭の LockByIDForUpdate の行ロックで
	// finalize（medical_record_repository.Update の draft-only WHERE）と直列化する。
	if err := s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
		parent, err := txRepos.MedicalRecord.LockByIDForUpdate(ctx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to verify medical record ownership", "error", err)
			return apperrors.Wrap(err, "failed to verify medical record ownership")
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテの治療は削除できません")
		}

		if err := txRepos.Treatment.Delete(ctx, clinicID, treatmentID); err != nil {
			return apperrors.Wrap(err, "failed to delete treatment")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete treatment", "error", err, "id", treatmentID, "clinic_id", clinicID)
		return err
	}

	slog.InfoContext(ctx, "treatment deleted",
		slog.Uint64("treatment_id", treatmentID),
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return nil
}

func (s *treatmentService) BulkUpdateSortOrder(ctx context.Context, clinicID, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error {
	// テナント検証: medicalRecordID が clinicID に所属するか確認
	if _, err := s.repos.MedicalRecord.FindByID(ctx, clinicID, medicalRecordID); err != nil {
		return apperrors.Wrap(err, "failed to verify medical record ownership")
	}

	updates := make([]repository.TreatmentSortUpdate, 0, len(input.Treatments))
	for _, item := range input.Treatments {
		updates = append(updates, repository.TreatmentSortUpdate{
			ID:        item.ID,
			ClinicID:  clinicID,
			SortOrder: item.SortOrder,
		})
	}

	if err := s.repos.Treatment.BulkUpdateSortOrder(ctx, updates); err != nil {
		return apperrors.Wrap(err, "failed to bulk update treatment sort order")
	}

	slog.InfoContext(ctx, "treatments bulk sort_order updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Int("count", len(updates)))

	return nil
}

// validateTreatmentMasterFKs は request 由来の clinic-scoped マスタFK
// (medicine/procedure/consultation/inventory) の所有権を検証する。treatments は自前 clinic_id を
// 持たず medical_record 経由で隔離されるため、これらのマスタが別 clinic のものでないことを
// write 前に明示確認しないと cross-tenant の mislink（#124/#125 同型）を作れてしまう。
// 別 clinic のマスタ参照は repo の FindByID が NotFound を返し write を遮断する。
// InventoryID は BE-refactor.md X-14a: DecreaseStock(ctx, targetInvID, qty) が clinicID を
// 取らないため、write 前の所有権検証が唯一のクロステナント防御線になる。
func (s *treatmentService) validateTreatmentMasterFKs(ctx context.Context, clinicID uint64, medicineID, procedureID, consultationID, inventoryID *uint64) error {
	if err := validateOwnedMasterFK(ctx, "medicine", clinicID, medicineID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repos.Medicine.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(ctx, "procedure", clinicID, procedureID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repos.Procedure.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(ctx, "consultation", clinicID, consultationID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repos.Consultation.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(ctx, "inventory", clinicID, inventoryID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repos.Inventory.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func validateTreatmentItemType(t model.TreatmentItemType) error {
	switch t {
	case model.TreatmentItemTypeConsultation,
		model.TreatmentItemTypeProcedure,
		model.TreatmentItemTypeMedicine,
		model.TreatmentItemTypeOther:
		return nil
	}
	return apperrors.WrapInvalidInput("invalid item_type: " + string(t))
}

func parseTreatmentStatus(s string) (model.TreatmentStatus, error) {
	switch model.TreatmentStatus(s) {
	case model.TreatmentStatusPending,
		model.TreatmentStatusCompleted,
		model.TreatmentStatusNotApplicable:
		return model.TreatmentStatus(s), nil
	}
	return "", apperrors.WrapInvalidInput("invalid status: " + s)
}
