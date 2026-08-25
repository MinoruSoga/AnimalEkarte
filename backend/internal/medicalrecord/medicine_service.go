// Package service provides business logic implementations for Medicine entity.
package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colMedicineName            = "name"
	colMedicineParentID        = "parent_id"
	colMedicinePrice           = "price"
	colMedicineIsActive        = "is_active"
	colMedicineDescription     = "description"
	colMedicineDosageForm      = "dosage_form"
	colMedicineMedicineUnit    = "medicine_unit"
	colMedicineInventoryID     = "inventory_id"
	colMedicineDefaultQuantity = "default_quantity"
	colMedicineSortOrder       = "sort_order"
	colMedicineTaxType         = "tax_type"
	colMedicineTaxRate         = "tax_rate"
	colMedicineIsNonInsurance  = "is_non_insurance"
	// #201 投与量計算（製品軸）
	colMedicineCalculationType     = "calculation_type"
	colMedicineStrength            = "strength"
	colMedicineFrequencyPerDay     = "frequency_per_day"
	colMedicineDefaultDurationDays = "default_duration_days"
)

func buildMedicineUpdate(input *UpdateMedicineInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colMedicineName] = *input.Name
	}
	setNullableUint64Field(fields, colMedicineParentID, input.ClearParentID, input.ParentID)
	if input.Price != nil {
		fields[colMedicinePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colMedicineIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colMedicineDescription] = *input.Description
	}
	if input.DosageForm != nil {
		if *input.DosageForm == "" {
			fields[colMedicineDosageForm] = nil
		} else {
			fields[colMedicineDosageForm] = *input.DosageForm
		}
	}
	if input.MedicineUnit != nil {
		if *input.MedicineUnit == "" {
			fields[colMedicineMedicineUnit] = nil
		} else {
			fields[colMedicineMedicineUnit] = *input.MedicineUnit
		}
	}
	if input.InventoryID != nil {
		fields[colMedicineInventoryID] = *input.InventoryID
	}
	if input.DefaultQuantity != nil {
		fields[colMedicineDefaultQuantity] = *input.DefaultQuantity
	}
	if input.SortOrder != nil {
		fields[colMedicineSortOrder] = *input.SortOrder
	}
	if input.TaxType != nil {
		fields[colMedicineTaxType] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields[colMedicineTaxRate] = *input.TaxRate
	}
	if input.IsNonInsurance != nil {
		fields[colMedicineIsNonInsurance] = *input.IsNonInsurance
	}
	// #201 投与量計算（製品軸）
	if input.CalculationType != nil {
		fields[colMedicineCalculationType] = *input.CalculationType
	}
	if input.ClearStrength {
		fields[colMedicineStrength] = nil
	} else if input.Strength != nil {
		fields[colMedicineStrength] = *input.Strength
	}
	if input.FrequencyPerDay != nil {
		fields[colMedicineFrequencyPerDay] = *input.FrequencyPerDay
	}
	if input.DefaultDurationDays != nil {
		fields[colMedicineDefaultDurationDays] = *input.DefaultDurationDays
	}
	return fields
}

// --- Input DTOs ---

// CreateMedicineInput は薬剤作成の入力DTO
type CreateMedicineInput struct {
	Name            string
	ParentID        *uint64
	Price           *int64
	IsActive        bool
	Description     string
	DosageForm      *string // nil = 未指定, "tablet" 等 = 値セット
	MedicineUnit    *string // nil = 未指定, "per_ml" 等 = 値セット
	InventoryID     *uint64
	DefaultQuantity float64
	SortOrder       int
	TaxType         *string  // nil = "excluded" (default)
	TaxRate         *float64 // nil = 0.10 (default)
	IsNonInsurance  bool

	// #201 投与量計算（製品軸）。CalculationType nil/"" = none（default-deny）。
	CalculationType     *string
	Strength            *float64
	FrequencyPerDay     *int
	DefaultDurationDays *int
	ActorID             *uint64 // #201 監査ログ用: per_weight 有効化の実施者
}

// UpdateMedicineInput は薬剤更新の入力DTO（nil = 未指定）
type UpdateMedicineInput struct {
	Name            *string
	ParentID        *uint64 // nil = 未指定（ClearParentID=false 時）
	ClearParentID   bool    // true = parent_id を NULL にクリア
	Price           *int64
	IsActive        *bool
	Description     *string
	DosageForm      *string // nil = 未指定, "" = NULL クリア, "tablet" = 値セット
	MedicineUnit    *string // nil = 未指定, "" = NULL クリア, "per_ml" = 値セット
	InventoryID     *uint64 // nil = 未指定, non-nil = 値セット
	DefaultQuantity *float64
	SortOrder       *int
	TaxType         *string
	TaxRate         *float64
	IsNonInsurance  *bool

	// #201 投与量計算（製品軸）。nil = 未指定。
	CalculationType     *string
	Strength            *float64
	ClearStrength       bool // true = strength を NULL クリア
	FrequencyPerDay     *int
	DefaultDurationDays *int
	ActorID             *uint64 // #201 監査ログ用: per_weight 有効化の実施者
}

// --- DB column constants ---

// buildMedicineUpdate は UpdateMedicineInput から map[string]any を構築する。
// GORM のゼロ値スキップ問題（bool false が無視される等）を回避するために使用する。

// ---- MedicineService ----

type MedicineService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type medicineService struct {
	repo          MedicineRepository
	inventoryRepo medicineInventoryRepo
	transactor    Transactor
	auditTx       AuditTxLogger // #201 B-2 / BE-refactor.md R1-2: per_weight 有効化の監査（nil 可・後方互換）
}

// NewMedicineServiceWithAudit は AuditTxLogger を注入する（#201 B-2: per_weight 有効化の監査記録）。
// BE-refactor.md R1-2 (D1): per_weight 有効化は薬剤作成/更新の tx 内で LogEntryTx を使い fail-closed 化する。
func NewMedicineServiceWithAudit(repo MedicineRepository, inventoryRepo medicineInventoryRepo, transactor Transactor, auditTx AuditTxLogger) MedicineService {
	return &medicineService{repo: repo, inventoryRepo: inventoryRepo, transactor: transactor, auditTx: auditTx}
}

func (s *medicineService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	result, total, err := s.repo.FindAll(ctx, clinicID, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list medicines", "error", err, "clinic_id", clinicID)
		return nil, 0, apperrors.Wrap(err, "failed to list medicines")
	}
	return result, total, nil
}

func (s *medicineService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medicine", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get medicine")
	}
	return result, nil
}

// toMedicineUnitPtr は *string を *model.MedicineUnit に変換する（nil/"" → nil）。
func toMedicineUnitPtr(s *string) *model.MedicineUnit {
	if s == nil || *s == "" {
		return nil
	}
	mu := model.MedicineUnit(*s)
	return &mu
}

// resolveCalculationType は *string を計算方式に解決する（nil/"" → none・default-deny）。
func resolveCalculationType(s *string) model.MedicineCalculationType {
	if s == nil || *s == "" {
		return model.MedicineCalculationTypeNone
	}
	return model.MedicineCalculationType(*s)
}

// validateDoseConfigAfterUpdate は #201 投与量計算設定の書込検証。dose 関連フィールドが変わる時
// のみ、更新後の実効設定を検証する（per_weight 誤設定・含量欠落を拒否。dose 非変更の Update には
// 影響しない＝後方互換）（BE-refactor.md E-2）。
func validateDoseConfigAfterUpdate(existing *model.Medicine, input *UpdateMedicineInput) error {
	doseFieldsChanged := input.CalculationType != nil || input.Strength != nil || input.ClearStrength ||
		input.FrequencyPerDay != nil || input.DefaultDurationDays != nil || input.MedicineUnit != nil
	if !doseFieldsChanged {
		return nil
	}
	effCalcType := existing.CalculationType
	if input.CalculationType != nil {
		effCalcType = resolveCalculationType(input.CalculationType)
	}
	if effCalcType == "" {
		effCalcType = model.MedicineCalculationTypeNone
	}
	effUnit := existing.MedicineUnit
	if input.MedicineUnit != nil {
		effUnit = toMedicineUnitPtr(input.MedicineUnit)
	}
	effStrength := existing.Strength
	if input.ClearStrength {
		effStrength = nil
	} else if input.Strength != nil {
		effStrength = input.Strength
	}
	effFreq := existing.FrequencyPerDay
	if input.FrequencyPerDay != nil {
		effFreq = input.FrequencyPerDay
	}
	effDuration := existing.DefaultDurationDays
	if input.DefaultDurationDays != nil {
		effDuration = input.DefaultDurationDays
	}
	return ValidateMedicineDoseConfig(effCalcType, effUnit, effStrength, effFreq, effDuration)
}

func (s *medicineService) Create(ctx context.Context, clinicID uint64, input *CreateMedicineInput) (*model.Medicine, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	if err := s.validateInventoryOwnership(ctx, clinicID, input.InventoryID); err != nil {
		return nil, err
	}

	// #201 投与量計算設定の書込検証（default-deny / per_weight 誤設定拒否）。
	calcType := resolveCalculationType(input.CalculationType)
	if err := ValidateMedicineDoseConfig(calcType, toMedicineUnitPtr(input.MedicineUnit), input.Strength, input.FrequencyPerDay, input.DefaultDurationDays); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate dose config")
	}

	taxType := model.TaxTypeExcluded
	if input.TaxType != nil && *input.TaxType != "" {
		taxType = model.TaxType(*input.TaxType)
	}
	taxRate := DefaultTaxRate
	if input.TaxRate != nil {
		taxRate = *input.TaxRate
	}
	medicine := &model.Medicine{
		ClinicID:            clinicID,
		Name:                input.Name,
		ParentID:            input.ParentID,
		Price:               input.Price,
		IsActive:            input.IsActive,
		Description:         input.Description,
		InventoryID:         input.InventoryID,
		DefaultQuantity:     input.DefaultQuantity,
		SortOrder:           input.SortOrder,
		TaxType:             taxType,
		TaxRate:             taxRate,
		IsNonInsurance:      input.IsNonInsurance,
		CalculationType:     calcType,
		Strength:            input.Strength,
		FrequencyPerDay:     input.FrequencyPerDay,
		DefaultDurationDays: input.DefaultDurationDays,
	}
	if input.DosageForm != nil && *input.DosageForm != "" {
		df := model.DosageForm(*input.DosageForm)
		medicine.DosageForm = &df
	}
	if input.MedicineUnit != nil && *input.MedicineUnit != "" {
		mu := model.MedicineUnit(*input.MedicineUnit)
		medicine.MedicineUnit = &mu
	}

	// BUG-429: 薬剤作成と在庫アイテム自動作成をトランザクションでアトミックに実行
	// BE-refactor.md R1-2 (D1): per_weight 有効化監査も同一 tx に統合する（fail-closed）。
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, medicine); err != nil {
			slog.ErrorContext(txCtx, "failed to create medicine", "error", err, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to create medicine")
		}
		// BUG-320: 薬品作成時に在庫アイテムを自動作成
		inventoryItem := &model.InventoryItem{
			ClinicID:      clinicID,
			Name:          medicine.Name,
			Category:      model.InventoryCategoryMedicine,
			Quantity:      0,
			Unit:          "錠", // デフォルト
			MinStockLevel: 0,
			Status:        model.InventoryStatusSufficient,
		}
		if err := s.inventoryRepo.Create(txCtx, clinicID, inventoryItem); err != nil {
			slog.ErrorContext(txCtx, "failed to create inventory item for medicine", "error", err, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to create inventory item for medicine")
		}
		// MRC-02: persist inventory_id so delete/rename use id not fragile name matching.
		if inventoryItem.ID != 0 {
			invID := inventoryItem.ID
			medicine.InventoryID = &invID
			if _, err := s.repo.Update(txCtx, clinicID, medicine.ID, map[string]any{colMedicineInventoryID: invID}); err != nil {
				slog.ErrorContext(txCtx, "failed to link inventory item to medicine", "error", err, "clinic_id", clinicID, "medicine_id", medicine.ID)
				return apperrors.Wrap(err, "failed to link inventory item to medicine")
			}
		}
		// #201 B-2: per_weight 有効化は安全クリティカル設定変更 → 監査（fail-closed）。
		if calcType == model.MedicineCalculationTypePerWeight {
			if err := s.auditPerWeightEnableTx(txCtx, clinicID, input.ActorID, medicine.ID, nil, medicine); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create medicine", "error", err)
		return nil, apperrors.Wrap(err, "failed to create medicine")
	}

	slog.InfoContext(ctx, "medicine created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medicine_id", medicine.ID),
		slog.String("name", medicine.Name),
	)
	return medicine, nil
}

// auditPerWeightEnableTx は per_weight 有効化（none→per_weight 含む）を監査記録する（fail-closed）。
// BE-refactor.md R1-2: 呼び出し元の ambient tx に参加する LogEntryTx を使う。失敗時は呼び出し元の
// WithTx が rollback し、薬剤作成/更新自体も無効になる（#211/refund パターン踏襲）。
func (s *medicineService) auditPerWeightEnableTx(ctx context.Context, clinicID uint64, actorID *uint64, medicineID uint64, before, after *model.Medicine) error {
	if s.auditTx == nil {
		return nil
	}
	actorType := auditActorTypeFor(actorID)
	newVal := map[string]any{"calculation_type": string(model.MedicineCalculationTypePerWeight)}
	if after != nil && after.Strength != nil {
		newVal["strength"] = *after.Strength
	}
	var oldVal map[string]any
	if before != nil {
		oldVal = map[string]any{"calculation_type": string(before.CalculationType)}
	}
	input := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     model.AuditActionMedicinePerWeightEnable,
		Resource:   model.AuditResourceMedicine,
		ResourceID: &medicineID,
		OldValue:   oldVal,
		NewValue:   newVal,
	}
	if err := s.auditTx.LogEntryTx(ctx, input); err != nil {
		return apperrors.Wrap(err, "failed to audit per_weight enable")
	}
	return nil
}

func (s *medicineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "medicine not found", "error", err)
		return nil, apperrors.Wrap(err, "medicine not found")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}

	if err := validateDoseConfigAfterUpdate(existing, input); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate dose config")
	}

	if err := s.validateParentOwnership(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	if err := s.validateInventoryOwnership(ctx, clinicID, input.InventoryID); err != nil {
		return nil, err
	}

	fields := buildMedicineUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}

	nameChanged := input.Name != nil && *input.Name != existing.Name
	var oldName, newName string
	if nameChanged {
		// TASK-215: 薬剤名変更時に連携在庫の name を同期する
		oldName = existing.Name
		newName = *input.Name
	}
	// #201 B-2: none→per_weight への有効化を監査（fail-closed。BE-refactor.md R1-2）。
	perWeightEnabling := input.CalculationType != nil &&
		resolveCalculationType(input.CalculationType) == model.MedicineCalculationTypePerWeight &&
		existing.CalculationType != model.MedicineCalculationTypePerWeight

	// BE-refactor.md R1-2 (D1): fields 更新・連携在庫名同期・per_weight 有効化監査を単一 tx に統合する。
	// 従来は名前変更時のみ tx 化され、監査は tx 外 best-effort だった。統合後は「本体書込と監査が原子」になる。
	var result *model.Medicine
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		var txErr error
		result, txErr = s.repo.Update(txCtx, clinicID, id, fields)
		if txErr != nil {
			slog.ErrorContext(txCtx, "failed to update medicine", "error", txErr, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(txErr, "failed to update medicine")
		}
		if nameChanged {
			if txErr = s.inventoryRepo.UpdateNameByMedicineCategory(txCtx, clinicID, oldName, newName); txErr != nil {
				slog.ErrorContext(txCtx, "failed to sync inventory item name", "error", txErr, "clinic_id", clinicID)
				return apperrors.Wrap(txErr, "failed to sync inventory item name")
			}
		}
		if perWeightEnabling {
			if auditErr := s.auditPerWeightEnableTx(txCtx, clinicID, input.ActorID, id, existing, result); auditErr != nil {
				return auditErr
			}
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to update medicine", "error", err)
		return nil, apperrors.Wrap(err, "failed to update medicine")
	}
	slog.InfoContext(ctx, "medicine updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medicine_id", id),
	)
	return result, nil
}

func (s *medicineService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder medicines", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder medicines")
	}
	slog.InfoContext(ctx, "medicines reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

func (s *medicineService) Delete(ctx context.Context, clinicID, id uint64) error {
	m, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get medicine")
	}

	// カテゴリ（parent_id = NULL）の場合、子アイテムが存在すれば削除を拒否する
	if m.ParentID == nil {
		count, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
		if err != nil {
			slog.ErrorContext(ctx, "failed to count medicine children", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to count medicine children")
		}
		if count > 0 {
			return apperrors.WrapConflict(
				fmt.Sprintf("このカテゴリには%d件の薬剤が含まれています。先に薬剤を移動または削除してください", count),
			)
		}
	} else {
		// 薬剤アイテムの場合、治療や入院ケアプランで使用中であれば削除を拒否する（BUG-108）
		usageCount, err := s.repo.CountUsageByMedicineID(ctx, clinicID, id)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check medicine usage", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check medicine usage")
		}
		if usageCount > 0 {
			return apperrors.WrapConflict("この薬剤は診療記録で使用中のため削除できません")
		}
	}

	// BUG-429: 薬剤削除と連携在庫削除をトランザクションでアトミックに実行
	// MRC-02: prefer medicines.inventory_id; fall back to name only for pre-link rows.
	// MRC-02: fail-closed delete audit in the same transaction.
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("medicine delete audit dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Re-check usage inside tx to shrink MRC-07 TOCTOU window for medicine deletes.
		if m.ParentID != nil {
			usageCount, err := s.repo.CountUsageByMedicineID(txCtx, clinicID, id)
			if err != nil {
				return apperrors.Wrap(err, "failed to re-check medicine usage")
			}
			if usageCount > 0 {
				return apperrors.WrapConflict("この薬剤は診療記録で使用中のため削除できません")
			}
		} else {
			count, err := s.repo.CountChildrenByParentID(txCtx, clinicID, id)
			if err != nil {
				return apperrors.Wrap(err, "failed to re-count medicine children")
			}
			if count > 0 {
				return apperrors.WrapConflict(
					fmt.Sprintf("このカテゴリには%d件の薬剤が含まれています。先に薬剤を移動または削除してください", count),
				)
			}
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete medicine", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete medicine")
		}
		if m.InventoryID != nil && *m.InventoryID != 0 {
			if err := s.inventoryRepo.Delete(txCtx, clinicID, *m.InventoryID); err != nil {
				// NotFound is acceptable for already-orphaned inventory; other errors fail closed.
				if !apperrors.IsNotFound(err) {
					slog.ErrorContext(txCtx, "failed to delete linked inventory for medicine", "error", err, "clinic_id", clinicID, "inventory_id", *m.InventoryID)
					return apperrors.Wrap(err, "failed to delete linked inventory for medicine")
				}
			}
		} else if err := s.inventoryRepo.DeleteByNameAndMedicineCategory(txCtx, clinicID, m.Name); err != nil {
			slog.ErrorContext(txCtx, "failed to delete linked inventory for medicine by name", "error", err, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete linked inventory for medicine")
		}
		resourceID := id
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorType:  auditActorTypeFor(nil),
			Action:     "medicine.delete",
			Resource:   model.AuditResourceMedicine,
			ResourceID: &resourceID,
			OldValue: map[string]any{
				"id":           m.ID,
				"name":         m.Name,
				"inventory_id": m.InventoryID,
			},
		}); err != nil {
			return apperrors.Wrap(err, "failed to audit medicine delete")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete medicine", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete medicine")
	}
	slog.InfoContext(ctx, "medicine deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("medicine_id", id),
	)
	return nil
}

// validateParentOwnership verifies a request-supplied self-ref parent_id belongs to the
// caller's clinic before it is persisted (X-14 self-ref master FK guard batch U2).
func (s *medicineService) validateParentOwnership(ctx context.Context, clinicID uint64, parentID *uint64) error {
	return validateOwnedMasterFK(ctx, "parent medicine", clinicID, parentID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repo.FindByID(actx, cid, mid)
			return err
		})
}

// validateInventoryOwnership verifies a request-supplied inventory_id belongs to the
// caller's clinic before it is persisted (X-14 master FK guard batch U2).
func (s *medicineService) validateInventoryOwnership(ctx context.Context, clinicID uint64, inventoryID *uint64) error {
	return validateOwnedMasterFK(ctx, "inventory", clinicID, inventoryID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.inventoryRepo.FindByID(actx, cid, mid)
			return err
		})
}
