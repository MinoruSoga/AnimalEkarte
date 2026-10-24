package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	colExamTypeName           = "name"
	colExamTypePrice          = "price"
	colExamTypeIsActive       = "is_active"
	colExamTypeDescription    = "description"
	colExamTypeParentID       = "parent_id"
	colExamTypeSortOrder      = "sort_order"
	colExamTypeIsNonInsurance = "is_non_insurance"
)

func buildExamTypeUpdate(input *UpdateExamTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colExamTypeName] = *input.Name
	}
	if input.Price != nil {
		fields[colExamTypePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colExamTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colExamTypeDescription] = *input.Description
	}
	setNullableUint64Field(fields, colExamTypeParentID, input.ClearParentID, input.ParentID)
	if input.SortOrder != nil {
		fields[colExamTypeSortOrder] = *input.SortOrder
	}
	if input.IsNonInsurance != nil {
		fields[colExamTypeIsNonInsurance] = *input.IsNonInsurance
	}
	// EMR-447 A.2: 明示送信時だけ対象列を SET。省略は UPDATE map へ入れず既存値を保持する。
	setBillDiscountExcludedField(fields, input.IsBillDiscountExcluded)
	return fields
}

// ---- ExaminationTypeService ----

// CreateExamTypeInput は検査種別作成の入力DTO
type CreateExamTypeInput struct {
	Name           string
	Price          *int64
	IsActive       bool
	Description    string
	ParentID       *uint64
	SortOrder      int
	IsNonInsurance bool
	// IsBillDiscountExcluded は nil=省略（service が false を明示保存・EMR-447 A.2）。
	IsBillDiscountExcluded *bool
	// ActorID は全体割引除外設定の監査に使う操作スタッフ ID。
	ActorID *uint64
}

// UpdateExamTypeInput は検査種別更新のサービス入力 DTO
type UpdateExamTypeInput struct {
	Name           *string
	Price          *int64
	IsActive       *bool
	Description    *string
	ParentID       *uint64
	ClearParentID  bool
	SortOrder      *int
	IsNonInsurance *bool
	// IsBillDiscountExcluded は nil=省略（変更なし）。送信時は Version が必須（EMR-447 A.4）。
	IsBillDiscountExcluded *bool
	// Version は任意送信の期待版。非 nil なら >0 必須で repository が WHERE 照合する。
	Version *int64
	// ActorID は全体割引除外設定の監査に使う操作スタッフ ID。
	ActorID *uint64
}

type ExaminationTypeService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CreateField(ctx context.Context, clinicID uint64, command *CreateExamTypeFieldCommand) (*model.ExamTypeField, error)
	UpdateField(ctx context.Context, clinicID, examTypeID, fieldID uint64, input *UpdateExamTypeFieldInput) (*ExamTypeFieldResult, error)
	DeleteField(ctx context.Context, clinicID, examTypeID, fieldID uint64) error
	ReorderFields(ctx context.Context, clinicID, examTypeID uint64, ids []uint64) error
	ReplaceReferenceRanges(ctx context.Context, clinicID, examTypeID uint64, command *ReplaceReferenceRangesCommand) (*ExamTypeFieldResult, error)
	ListReferenceRanges(ctx context.Context, clinicID uint64, fieldIDs []uint64) (map[uint64][]model.ExamReferenceRange, error)
}

type examTypeService struct {
	repo       ExamTypeRepository
	transactor Transactor
	auditTx    AuditTxLogger
}

// EMR-447 A01: 全体割引除外設定の前後値監査に使う action/resource（checkup_type_audit.go と同型）。
const (
	examTypeAuditActionCreate = "examination_type.create"
	examTypeAuditActionUpdate = "examination_type.update"
	examTypeAuditResource     = "examination_type"
)

// errMsgExamTypeVersionConflict は楽観ロックの版不一致（再読込を促す）メッセージ。
const errMsgExamTypeVersionConflict = "他のユーザーがこの検査種別を変更しました。再読み込みしてください"

// errMsgExamTypeVersionRequired は is_bill_discount_excluded を含む更新で
// client 版が未指定のときの検証エラー（EMR-447 A.4: 新設定を含む PATCH は編集開始版必須）。
const errMsgExamTypeVersionRequired = "version を指定してください（編集開始時の版）"

// errMsgExamTypeVersionInvalid は送信された期待版が 0 以下のときの検証エラー。
const errMsgExamTypeVersionInvalid = "version は正の整数で指定してください"

// NewExamTypeService constructs ExaminationTypeService without the audit sink.
// transactor is required so Create/Update/Delete and field writes share ambient transactions
// (MRB-07: compile-time wiring; nil is fail-closed as 500 via withTx).
// auditTx が無い構成では is_bill_discount_excluded を書く操作（create は常時、update は
// 送信時のみ）が logBillDiscountAuditTx 経由で fail-closed になる（EMR-447 A.5）。
func NewExamTypeService(repo ExamTypeRepository, transactor Transactor) ExaminationTypeService {
	return &examTypeService{repo: repo, transactor: transactor}
}

// NewExamTypeServiceWithAudit は監査依存を注入する構成子（EMR-447 A.5）。
// auditTx は業務書込と同じ tx での監査に必須。nil なら新 field の書込は fail-closed。
func NewExamTypeServiceWithAudit(repo ExamTypeRepository, transactor Transactor, auditTx AuditTxLogger) ExaminationTypeService {
	return &examTypeService{repo: repo, transactor: transactor, auditTx: auditTx}
}

func (s *examTypeService) List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list exam types")
	}
	return items, nil
}
func (s *examTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get exam type")
	}
	return result, nil
}
func (s *examTypeService) Create(ctx context.Context, clinicID uint64, input *CreateExamTypeInput) (*model.ExaminationType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	// MRB-08 / X-05: parent_id 検証と Create を同一 WithTx に収め、FindByID の FOR SHARE を発火させる。
	exType := &model.ExaminationType{
		ClinicID:       clinicID,
		Name:           input.Name,
		Price:          input.Price,
		IsActive:       input.IsActive,
		Description:    input.Description,
		ParentID:       input.ParentID,
		SortOrder:      sharedkernel.Ptr(input.SortOrder),
		IsNonInsurance: input.IsNonInsurance,
		// EMR-447 A.2: 省略は false を明示保存し、DB default / Go ゼロ値に任せない。
		IsBillDiscountExcluded: createBillDiscountExcluded(input.IsBillDiscountExcluded),
	}
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.validateParentOwnership(txCtx, clinicID, input.ParentID); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, exType); err != nil {
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				input.Name,
				apperrors.ConstraintExamTypeName,
				apperrors.CodeExamTypeNameConflict,
			); conflict != nil {
				return conflict
			}
			return apperrors.Wrap(err, "failed to create exam type")
		}
		// EMR-447 A.5: create は常に新 field を書くため、前値なしの監査を同じ tx で必須記録する。
		return logBillDiscountAuditTx(txCtx, s.auditTx, clinicID, input.ActorID,
			examTypeAuditResource, examTypeAuditActionCreate,
			billDiscountChange{ResourceID: exType.ID, After: exType.IsBillDiscountExcluded})
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "exam type created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("exam_type_id", exType.ID))
	return exType, nil
}

// examTypeForUpdateFinder は更新 TX の前値読取専用の FOR UPDATE read。
// concrete repo が実装しない構成は fail-closed（master_parent_write_lock と同型）。
type examTypeForUpdateFinder interface {
	FindByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
}

func (s *examTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateExamTypeInput) (*model.ExaminationType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	// EMR-447 A.4: 新設定を含む PATCH は編集開始版必須。送信された版は >0 でなければならず、
	// 新設定を含まない旧 PATCH の version 省略は従来互換（CAS なし）を維持する。
	if input.IsBillDiscountExcluded != nil && input.Version == nil {
		return nil, apperrors.WrapInvalidInput(errMsgExamTypeVersionRequired)
	}
	if input.Version != nil && *input.Version <= 0 {
		return nil, apperrors.WrapInvalidInput(errMsgExamTypeVersionInvalid)
	}
	fields := buildExamTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	// MRB-08: 存在確認・parent 検証・Update を同一 tx に収める。
	var exType *model.ExaminationType
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		// EMR-447 A-2: 前値読取は FOR UPDATE で行を排他ロックしながら読む。
		// FOR SHARE のままだと同版の並行 PATCH が SHARE を保持したまま UPDATE
		// 昇格を待ち合い 40P01（500）になる。FOR UPDATE を先に取れば敗者は
		// lock 解放後に新しい版を読み、CAS の WHERE 不一致で 409 に収束する。
		forUpdateRepo, ok := s.repo.(examTypeForUpdateFinder)
		if !ok {
			return apperrors.WrapInternalServerError("exam type repository must implement FindByIDForUpdate")
		}
		before, err := forUpdateRepo.FindByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get exam type")
		}
		if err := s.validateParentOwnership(txCtx, clinicID, input.ParentID); err != nil {
			return err
		}
		updated, err := s.repo.Update(txCtx, clinicID, id, *input)
		if err != nil {
			nameForConflict := ""
			if input.Name != nil {
				nameForConflict = *input.Name
			}
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				nameForConflict,
				apperrors.ConstraintExamTypeName,
				apperrors.CodeExamTypeNameConflict,
			); conflict != nil {
				return conflict
			}
			return apperrors.Wrap(err, "failed to update exam type")
		}
		// EMR-447 A.5: 新 field が送信された時だけ、raw NULL/false/true を区別した
		// 前後値監査を業務書込と同じ tx で記録する（is_non_insurance 監査とは別 field）。
		if input.IsBillDiscountExcluded != nil {
			if err := logBillDiscountAuditTx(txCtx, s.auditTx, clinicID, input.ActorID,
				examTypeAuditResource, examTypeAuditActionUpdate,
				billDiscountChange{ResourceID: updated.ID, BeforeExists: true, Before: before.IsBillDiscountExcluded, After: updated.IsBillDiscountExcluded}); err != nil {
				return err
			}
		}
		exType = updated
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "exam type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return exType, nil
}
func (s *examTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	// MRB-08: 依存 count と Delete を同一 tx に収め、並行使用中の silent 削除を防ぐ。
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.FindByID(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to find exam type")
		}
		childCount, err := s.repo.CountChildrenByParentID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check exam type children")
		}
		if childCount > 0 {
			return apperrors.WrapConflict("この検査種別にはサブ種別が登録されているため削除できません")
		}
		count, err := s.repo.CountUsageByExamTypeID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check exam type dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この検査種別は検査記録で使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete exam type")
		}
		return nil
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "exam type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("exam_type_id", id))
	return nil
}

func (s *examTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder exam types")
	}
	slog.InfoContext(ctx, "exam type reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

// validateParentOwnership verifies a request-supplied parent_id belongs to the caller's
// clinic before it is persisted (X-14 self-ref master FK guard).
func (s *examTypeService) validateParentOwnership(ctx context.Context, clinicID uint64, parentID *uint64) error {
	return validateOwnedMasterFK(ctx, "parent exam type", clinicID, parentID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.repo.FindByID(actx, cid, mid)
			return err
		})
}
