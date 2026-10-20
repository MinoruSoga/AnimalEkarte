package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- CheckupTypeService ----

// CreateCheckupTypeInput は定期健診種別作成の入力DTO
type CreateCheckupTypeInput struct {
	Name        string
	Price       *int64
	IsActive    bool
	Description string
	Interval    string
	TargetAge   string
	ParentID    *uint64
	SortOrder   int
	// IsNonInsurance は nil=省略（新規初期値 true を明示保存）。
	// 明示 null は request の binding 層で 400 になるため service には届かない（EMR-456）。
	IsNonInsurance *bool
	// ActorID は保険設定の監査に使う操作スタッフ ID。
	ActorID *uint64
}

// UpdateCheckupTypeInput はチェックアップ種別更新のサービス入力 DTO
type UpdateCheckupTypeInput struct {
	Name          *string
	Price         *int64
	IsActive      *bool
	Description   *string
	Interval      *string
	TargetAge     *string
	ParentID      *uint64
	ClearParentID bool
	SortOrder     *int
	// IsNonInsurance は nil=省略（変更なし）。NULL（未保存）へ戻す経路は持たない（EMR-456）。
	// 明示 null は request の binding 層で 400 になるため service には届かない。
	IsNonInsurance *bool
	// Version は楽観ロックの期待版。更新要求では必須（nil は検証エラー・review 4320626569）。
	Version *int64
	// ActorID は保険設定の監査に使う操作スタッフ ID。
	ActorID *uint64
	// expectedVersion は service が算出し repository が WHERE に使う（HTTP/client からは設定しない）。
	// version 列は書かない: checkup_types の BEFORE UPDATE トリガ（app_private.entity_version_guard）が
	// 内容変更時だけ OLD.version+1 にし、内容不変の UPDATE での version 変更は 23514 で拒否するため。
	expectedVersion *int64
}

const (
	colCheckupTypeName        = "name"
	colCheckupTypePrice       = "price"
	colCheckupTypeIsActive    = "is_active"
	colCheckupTypeDescription = "description"
	colCheckupTypeInterval    = "interval"
	colCheckupTypeTargetAge   = "target_age"
	colCheckupTypeParentID    = "parent_id"
	colCheckupTypeSortOrder   = "sort_order"
	colCheckupTypeNonIns      = "is_non_insurance"
)

// newCheckupTypeIsNonInsuranceDefault は新規登録時の保険設定初期値（保険対象外）。
// FR-accounting-item-insurance-setting の新規初期値であり、旧欠落値の補完には使わない。
const newCheckupTypeIsNonInsuranceDefault = true

// errMsgCheckupTypeVersionConflict は楽観ロックの版不一致（再読込を促す）メッセージ。
const errMsgCheckupTypeVersionConflict = "他のユーザーがこの定期健診種別を変更しました。再読み込みしてください"

// errMsgCheckupTypeVersionRequired は更新要求で client 版が未指定のときの検証エラー。
// service が読取時点の版で補う fallback は、画面を開いた後の他者更新を検出できず無音上書きになるため廃止した。
const errMsgCheckupTypeVersionRequired = "version を指定してください（編集開始時の版）"

func buildCheckupTypeUpdate(input *UpdateCheckupTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colCheckupTypeName] = *input.Name
	}
	if input.Price != nil {
		fields[colCheckupTypePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colCheckupTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colCheckupTypeDescription] = *input.Description
	}
	if input.Interval != nil {
		fields[colCheckupTypeInterval] = *input.Interval
	}
	if input.TargetAge != nil {
		fields[colCheckupTypeTargetAge] = *input.TargetAge
	}
	setNullableUint64Field(fields, colCheckupTypeParentID, input.ClearParentID, input.ParentID)
	if input.SortOrder != nil {
		fields[colCheckupTypeSortOrder] = *input.SortOrder
	}
	if input.IsNonInsurance != nil {
		fields[colCheckupTypeNonIns] = *input.IsNonInsurance
	}
	return fields
}

type CheckupTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type checkupTypeService struct {
	repo       CheckupTypeRepository
	transactor Transactor
	auditTx    AuditTxLogger
}

// NewCheckupTypeService constructs CheckupTypeService.
// transactor と auditTx は Create/Update の書込と保険設定監査を同一 tx に収めるため必須
// （EMR-456。nil は composition 不備として 500 で fail-closed）。
func NewCheckupTypeService(repo CheckupTypeRepository, transactor Transactor, auditTx AuditTxLogger) CheckupTypeService {
	return &checkupTypeService{repo: repo, transactor: transactor, auditTx: auditTx}
}

func (s *checkupTypeService) List(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list checkup types")
	}
	return items, nil
}
func (s *checkupTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get checkup type")
	}
	return result, nil
}
func (s *checkupTypeService) Create(ctx context.Context, clinicID uint64, input *CreateCheckupTypeInput) (*model.CheckupType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	isNonInsurance := newCheckupTypeIsNonInsuranceDefault
	if input.IsNonInsurance != nil {
		isNonInsurance = *input.IsNonInsurance
	}
	checkupType := &model.CheckupType{
		ClinicID:       clinicID,
		Name:           input.Name,
		Price:          input.Price,
		IsActive:       input.IsActive,
		Description:    input.Description,
		Interval:       input.Interval,
		TargetAge:      input.TargetAge,
		ParentID:       input.ParentID,
		SortOrder:      &input.SortOrder,
		IsNonInsurance: &isNonInsurance,
	}
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.validateParentOwnership(txCtx, clinicID, input.ParentID); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, checkupType); err != nil {
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				input.Name,
				apperrors.ConstraintCheckupTypeName,
				apperrors.CodeCheckupTypeNameConflict,
			); conflict != nil {
				return conflict
			}
			return apperrors.Wrap(err, "failed to create checkup type")
		}
		return s.auditInsuranceTx(txCtx, clinicID, input.ActorID, checkupTypeAuditActionCreate, nil, checkupType)
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "checkup type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("checkup_type_id", checkupType.ID))
	return checkupType, nil
}
func (s *checkupTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCheckupTypeInput) (*model.CheckupType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	if len(buildCheckupTypeUpdate(input)) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	// EMR-456: 更新要求は編集開始時の client 版で必須。service が読取時点の版で補う fallback は、
	// 画面を開いた後の他者更新を検出できず無音上書きになるため廃止した（review 4320626569・チケット回答 2026-10-20）。
	if input.Version == nil {
		return nil, apperrors.WrapInvalidInput(errMsgCheckupTypeVersionRequired)
	}
	var checkupType *model.CheckupType
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		before, err := s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get checkup type")
		}
		if err := s.validateParentOwnership(txCtx, clinicID, input.ParentID); err != nil {
			return err
		}
		// EMR-456: 版は client 提示の期待版で必ず WHERE 照合する（未指定は上述の 400）。
		// version 列は書かない: BEFORE UPDATE トリガ（app_private.entity_version_guard）が
		// 内容変更時だけ OLD.version+1 にし、内容不変の UPDATE での版変更は 23514 で拒否する。
		expected := *input.Version
		effective := *input
		effective.expectedVersion = &expected
		updated, err := s.repo.Update(txCtx, clinicID, id, effective)
		if err != nil {
			nameForConflict := ""
			if input.Name != nil {
				nameForConflict = *input.Name
			}
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				nameForConflict,
				apperrors.ConstraintCheckupTypeName,
				apperrors.CodeCheckupTypeNameConflict,
			); conflict != nil {
				return conflict
			}
			return apperrors.Wrap(err, "failed to update checkup type")
		}
		if !sameOptionalBool(before.IsNonInsurance, updated.IsNonInsurance) {
			if err := s.auditInsuranceTx(txCtx, clinicID, input.ActorID, checkupTypeAuditActionUpdate, before, updated); err != nil {
				return err
			}
		}
		checkupType = updated
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "checkup type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("checkup_type_id", id))
	return checkupType, nil
}
func (s *checkupTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get checkup type")
	}
	childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check checkup type children")
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この定期健診種別にはサブ種別が登録されているため削除できません")
	}
	count, err := s.repo.CountUsageByCheckupTypeID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check checkup type dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この定期健診種別は健診記録で使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete checkup type")
	}
	slog.InfoContext(ctx, "checkup type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("checkup_type_id", id))
	return nil
}

func (s *checkupTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder checkup types")
	}
	slog.InfoContext(ctx, "checkup type reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// validateParentOwnership verifies a request-supplied parent_id belongs to the caller's
// clinic before it is persisted (X-14 self-ref master FK guard).
func (s *checkupTypeService) validateParentOwnership(ctx context.Context, clinicID uint64, parentID *uint64) error {
	return validateOwnedMasterFK(ctx, "parent checkup type", clinicID, parentID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repo.FindByID(actx, cid, mid)
			return err
		})
}
