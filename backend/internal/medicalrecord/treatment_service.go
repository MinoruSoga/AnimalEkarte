package medicalrecord

import (
	"context"
	"log/slog"
	"maps"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	// DoseDeviationReason は TASK-377: 上限内の下限割れ/著しい乖離で必須の free-text 理由。
	// reason-required でない経路では無視され snapshot に残さない。
	DoseDeviationReason string
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
	// DoseDeviationReason は TASK-377: reason-required 時に必須。nil = 未送信（欠落と同義）。
	DoseDeviationReason *string
	// DiscountEditAllowed is set by the HTTP boundary from discount:edit RBAC.
	// Service rechecks discount fields against the FOR UPDATE locked row (SEC-CS-F09).
	DiscountEditAllowed bool
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
	treatmentRepo     TreatmentRepository
	medicalRecordRepo treatmentMedicalRecordRepo
	medicineRepo      medicineFinder
	procedureRepo     procedureFinder
	consultationRepo  consultationFinder
	inventoryRepo     treatmentInventoryRepo
	vitalRepo         VitalRepository
	doseParamRepo     doseParamFinder
	transactor        Transactor
	// auditTx は非nilかどうかを「逸脱audit機能の有効/無効」フラグとして併用する。
	// BE9-2D ④b: 旧実装は repos.Transaction(ctx, func(txRepos)) 機構（tx が txRepos に宿り ctx に
	// 伝播しない）だったため LogEntryTx（dbOrTx が ctx の txKey を見る）が tx に参加できず、
	// txRepos.Audit.CreateTx を直接使っていた。Transactor.WithTx + treatment/vital repo の dbOrTx 化
	// により LogEntryTx が同一 tx へ参加できるため、medicine/dose-param と同じ AuditTxLogger 経由の
	// fail-closed 監査（R1-2 (D1)・#211/refund パターン）へ統一した。
	auditTx AuditTxLogger
}

// NewTreatmentServiceWithAudit は逸脱 audit 機能を有効化する（#201 B-2）。
// auditTx は非nilフラグとしてのみ使う（上記 struct コメント参照）。
// BE9-2D ④b: Phase 1 で集約依存を個別注入へ分解済み。本 package への縦移動で cross-package 依存を
// service_deps.go の consumer-side view（treatmentMedicalRecordRepo/medicineFinder 等）へ、
// treatment/vital repo を in-package 具象型へ差し替えた。
func NewTreatmentServiceWithAudit(
	treatmentRepo TreatmentRepository,
	medicalRecordRepo treatmentMedicalRecordRepo,
	medicineRepo medicineFinder,
	procedureRepo procedureFinder,
	consultationRepo consultationFinder,
	inventoryRepo treatmentInventoryRepo,
	vitalRepo VitalRepository,
	doseParamRepo doseParamFinder,
	transactor Transactor,
	auditTx AuditTxLogger,
) TreatmentService {
	return &treatmentService{
		treatmentRepo:     treatmentRepo,
		medicalRecordRepo: medicalRecordRepo,
		medicineRepo:      medicineRepo,
		procedureRepo:     procedureRepo,
		consultationRepo:  consultationRepo,
		inventoryRepo:     inventoryRepo,
		vitalRepo:         vitalRepo,
		doseParamRepo:     doseParamRepo,
		transactor:        transactor,
		auditTx:           auditTx,
	}
}

func (s *treatmentService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	treatment, err := s.treatmentRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get treatment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get treatment")
	}
	return treatment, nil
}

func (s *treatmentService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	treatments, err := s.treatmentRepo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
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
	treatments, total, err := s.treatmentRepo.FindHistoryByPetID(ctx, clinicID, petID, filter, page, limit)
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
		return nil, apperrors.WrapInvalidInput(errMsgPriceZeroOrMore)
	}
	if input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput(errMsgQuantityPositive)
	}
	if err := validateDiscountRate(input.DiscountRate); err != nil {
		return nil, err
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
	// BE9-2D ④b: repos.Transaction（tx-bound clone）→ Transactor.WithTx（ctx-txKey）へ変換。
	// 閉包内の read/write は各 repo の dbOrTx が txCtx の ambient tx へ参加する（挙動は旧機構と等価）。
	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// テナント所有権 + 確定ロック検証（Update/Delete と対称化・healthcare review CRITICAL）。
		// treatments は自前 clinic_id を持たず medical_records 経由で隔離するため、所有権を Create でも明示検証する。
		// BE-refactor.md X-11/E-5: lockDraftMedicalRecord に行ロック+finalizedガードを集約。
		if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
			"failed to verify medical record ownership", "確定済みカルテには治療を追加できません"); err != nil {
			return err
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
		// TASK-377: reason-required では理由 validation → strict snapshot → audit を同一 tx で fail-closed。
		eval, derr := s.evaluateDoseForSave(txCtx, clinicID, medicalRecordID, input.ItemType, input.MedicineID, input.Quantity)
		if derr != nil {
			return derr // species 不一致など fail-closed
		}
		if eval != nil && eval.ExceedsCapSaved {
			return apperrors.WrapInvalidInput("投与量がマスタで設定された絶対上限を超えているため保存できません")
		}
		if eval != nil {
			if err := applyDeviationReasonToEval(eval, input.DoseDeviationReason); err != nil {
				return err
			}
			// reason-required の actor / audit 依存は treatment write 前に fail-closed する
			// （tx rollback に依存せず write ゼロを保証する）。
			if err := s.ensureDoseDeviationAuditReady(eval, input.ActorID); err != nil {
				return err
			}
			doseEval = eval
			if err := applyDoseSnapshotToTreatment(txCtx, treatment, eval); err != nil {
				return err
			}
		}

		// 1. Create Treatment
		if err := s.treatmentRepo.Create(txCtx, treatment); err != nil {
			return apperrors.Wrap(err, "failed to create treatment")
		}

		// 2. Decrease Stock (if applicable)
		// BE-refactor.md FOLLOWUP-X14A: MedicineID を inventory id として DecreaseStock へ渡す
		// フォールバックは書込 IDOR（medicines.id と inventory_items.id の採番衝突で他クリニックの
		// 在庫を減算しうる）だったため廃止した。減算は InventoryID が明示指定された場合のみ行う。
		if input.InventoryID != nil && *input.InventoryID > 0 {
			if err := s.inventoryRepo.DecreaseStock(txCtx, clinicID, *input.InventoryID, input.Quantity); err != nil {
				return apperrors.Wrap(err, "failed to decrease inventory stock")
			}
		}

		// #201 B-2 / BE-refactor.md R1-2: 逸脱（過量/過少/著しい上書き）を監査記録（fail-closed）。
		// tx 内で失敗すると treatment 作成・在庫減算ごとロールバックする。
		if doseEval != nil && input.MedicineID != nil {
			if err := s.auditDoseDeviationTx(txCtx, clinicID, input.ActorID, treatment.ID, *input.MedicineID, doseEval); err != nil {
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
	existing, err := s.treatmentRepo.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get treatment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get treatment")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}

	// HC-004: 親カルテが確定済みの場合は編集拒否
	parent, err := s.medicalRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
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
		return nil, apperrors.WrapInvalidInput(errMsgQuantityPositive)
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(errMsgPriceZeroOrMore)
	}
	if input.DiscountRate != nil {
		if err := validateDiscountRate(*input.DiscountRate); err != nil {
			return nil, err
		}
	}

	fields := buildTreatmentUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}

	// #201 B-2: quantity/medicine が変わる場合のみ保存時 BE 再検証＋スナップショット更新。
	// 再検証の読み出しと UPDATE を同一トランザクションに束ね、並行マスタ変更による
	// スナップショット不整合（TOCTOU）を防ぐ（security review MEDIUM-1）。
	// quantity/medicine/item_type のいずれかが変わると dose スナップショットの再評価対象になる。
	doseRelevant := input.Quantity != nil || input.MedicineID != nil || input.ItemType != nil
	var doseEval *SavedDoseEvaluation
	var doseMedicineID uint64
	if txErr := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// テナント所有権 + 確定ロック検証（Create と対称化・BE-refactor.md H-8a）。
		// 冒頭の事前チェックは fast-fail として維持しつつ、tx 内でも lockDraftMedicalRecord
		// の行ロックで finalize と直列化し、チェック通過後〜Update 実行前に finalize が
		// 割り込むレースを防ぐ（BE-refactor.md E-5）。
		if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
			"failed to verify medical record ownership", "確定済みカルテの治療は編集できません"); err != nil {
			return err
		}

		// SEC-CS-F09: lock treatment row and recheck discount against the locked snapshot.
		// Handler early check uses a pre-TX GetByID; stale equality must not authorize overwrite.
		current, err := s.treatmentRepo.LockByIDForUpdate(txCtx, clinicID, treatmentID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock treatment for update")
		}
		if current.MedicalRecordID != medicalRecordID {
			return apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
		}
		if err := applyTreatmentDiscountGuard(current, input, fields); err != nil {
			return err
		}
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput(errMsgAtLeastOneField)
		}

		if doseRelevant {
			// 親行 + treatment 行ロック取得後の locked snapshot で dose を評価する。
			effItemType, effMedicineID, effQty := effectiveDoseInputs(current, input)
			eval, derr := s.evaluateDoseForSave(txCtx, clinicID, medicalRecordID, effItemType, effMedicineID, effQty)
			if derr != nil {
				return derr // species 不一致など fail-closed
			}
			applied, err := s.applyLockedTreatmentDoseFields(txCtx, current, input, eval, effMedicineID, fields)
			if err != nil {
				return err
			}
			doseEval = applied.eval
			doseMedicineID = applied.medicineID
		}
		if err := s.treatmentRepo.Update(txCtx, clinicID, treatmentID, fields); err != nil {
			return err
		}
		// #201 B-2 / BE-refactor.md R1-2: 逸脱監査を fail-closed 化。失敗すると Update ごとロールバックする。
		if doseEval != nil {
			return s.auditDoseDeviationTx(txCtx, clinicID, input.ActorID, treatmentID, doseMedicineID, doseEval)
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

	treatment, err := s.treatmentRepo.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get updated treatment", "error", err)
		return nil, apperrors.Wrap(err, "failed to get updated treatment")
	}
	return treatment, nil
}

func (s *treatmentService) Delete(ctx context.Context, clinicID, medicalRecordID, treatmentID uint64) error {
	existing, err := s.treatmentRepo.FindByID(ctx, clinicID, treatmentID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get treatment")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}

	// HC-004: 親カルテが確定済みの場合は削除拒否（BE-refactor.md H-8b）。
	// finalized チェックと Delete を同一 tx に束ね、閉包先頭の LockByIDForUpdate の行ロックで
	// finalize（medical_record_repository.Update の draft-only WHERE）と直列化する。
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
			"failed to verify medical record ownership", "確定済みカルテの治療は削除できません"); err != nil {
			return err
		}

		if err := s.treatmentRepo.Delete(txCtx, clinicID, treatmentID); err != nil {
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
	updates := make([]TreatmentSortUpdate, 0, len(input.Treatments))
	for _, item := range input.Treatments {
		updates = append(updates, TreatmentSortUpdate{
			ID:              item.ID,
			ClinicID:        clinicID,
			MedicalRecordID: medicalRecordID,
			SortOrder:       item.SortOrder,
		})
	}

	// HC-004: 親カルテが確定済みの場合は並び順変更を拒否（BE-refactor.md H-8c）。
	// テナント所有権確認（LockByIDForUpdate）と finalized チェックを同一 tx に束ね、
	// 行ロックで finalize（medical_record_repository.Update の draft-only WHERE）と直列化する。
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
			"failed to verify medical record ownership", "確定済みカルテの治療は編集できません"); err != nil {
			return err
		}

		if err := s.treatmentRepo.BulkUpdateSortOrder(txCtx, updates); err != nil {
			return apperrors.Wrap(err, "failed to bulk update treatment sort order")
		}
		return nil
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "treatments bulk sort_order updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medical_record_id", medicalRecordID),
		slog.Int("count", len(updates)))

	return nil
}

type lockedTreatmentDoseApply struct {
	eval       *SavedDoseEvaluation
	medicineID uint64
}

func (s *treatmentService) applyLockedTreatmentDoseFields(
	ctx context.Context,
	current *model.Treatment,
	input *UpdateTreatmentInput,
	eval *SavedDoseEvaluation,
	effMedicineID *uint64,
	fields map[string]any,
) (lockedTreatmentDoseApply, error) {
	if eval != nil && eval.ExceedsCapSaved {
		return lockedTreatmentDoseApply{}, apperrors.WrapInvalidInput("投与量がマスタで設定された絶対上限を超えているため保存できません")
	}
	if eval == nil {
		if treatmentHasDoseSnapshot(current) {
			maps.Copy(fields, clearedDoseColumns())
		}
		return lockedTreatmentDoseApply{}, nil
	}
	rawReason := ""
	if input.DoseDeviationReason != nil {
		rawReason = *input.DoseDeviationReason
	}
	if err := applyDeviationReasonToEval(eval, rawReason); err != nil {
		return lockedTreatmentDoseApply{}, err
	}
	if err := s.ensureDoseDeviationAuditReady(eval, input.ActorID); err != nil {
		return lockedTreatmentDoseApply{}, err
	}
	snapCols, err := doseSnapshotColumns(ctx, eval)
	if err != nil {
		return lockedTreatmentDoseApply{}, err
	}
	maps.Copy(fields, snapCols)
	return lockedTreatmentDoseApply{eval: eval, medicineID: *effMedicineID}, nil
}
