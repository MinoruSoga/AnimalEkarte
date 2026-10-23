package inventory

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	colMerchandiseItemName      = "name"
	colMerchandiseItemCategory  = "category"
	colMerchandiseItemUnitPrice = "unit_price"
	colMerchandiseItemTaxType   = "tax_type"
	colMerchandiseItemTaxRate   = "tax_rate"
	colMerchandiseItemIsActive  = "is_active"
	colMerchandiseItemSortOrder = "sort_order"
)

// EMR-449 A01: 物販マスタ専有の監査 action/resource（checkup_type_audit.go と同じ専有方式）。
const (
	merchandiseItemAuditResource     = "merchandise_item"
	merchandiseItemAuditActionCreate = "merchandise_item.create"
	merchandiseItemAuditActionUpdate = "merchandise_item.update"
)

// EMR-449 A.4: 編集開始版の wire 検査メッセージ（request 側と共通）。
const (
	errMsgMerchandiseVersionInvalid  = "version には1以上の整数を指定してください"
	errMsgMerchandiseVersionRequired = "is_bill_discount_excluded の更新には version（編集開始時の版）の指定が必要です"
)

func buildMerchandiseItemUpdate(input *UpdateMerchandiseItemInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colMerchandiseItemName] = *input.Name
	}
	if input.Category != nil {
		fields[colMerchandiseItemCategory] = *input.Category
	}
	if input.UnitPrice != nil {
		fields[colMerchandiseItemUnitPrice] = *input.UnitPrice
	}
	if input.TaxType != nil {
		fields[colMerchandiseItemTaxType] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields[colMerchandiseItemTaxRate] = *input.TaxRate
	}
	if input.IsActive != nil {
		fields[colMerchandiseItemIsActive] = *input.IsActive
	}
	if input.SortOrder != nil {
		fields[colMerchandiseItemSortOrder] = *input.SortOrder
	}
	// EMR-449: 明示値だけを SET し、false をゼロ値 skip しない。未送信は既存 NULL/false/true を保持。
	setBillDiscountExcludedField(fields, input.IsBillDiscountExcluded)
	return fields
}

// --- Input DTOs ---

// CreateMerchandiseItemInput は物販品作成の入力DTO
type CreateMerchandiseItemInput struct {
	Name      string
	Category  string
	UnitPrice int64
	TaxType   string              // "" → "excluded" (default)
	TaxRate   *model.ExactNumeric // nil → 0.10 (default)
	IsActive  bool
	SortOrder int

	// IsBillDiscountExcluded は全体割引除外（EMR-449 A01）。nil = 省略（service が明示 false を保存）。
	IsBillDiscountExcluded *bool
	// ActorID は全体割引除外の監査に使う操作スタッフ ID（trusted request context 由来）。
	ActorID *uint64
}

// UpdateMerchandiseItemInput は物販品更新の入力DTO（nil = 未指定）
type UpdateMerchandiseItemInput struct {
	Name      *string
	Category  *string
	UnitPrice *int64
	TaxType   *string
	TaxRate   *model.ExactNumeric
	IsActive  *bool
	SortOrder *int

	// IsBillDiscountExcluded は nil = 変更なし（既存 NULL/false/true を保持）。
	IsBillDiscountExcluded *bool
	// Version は省略可能な編集開始版（EMR-449 A.4）。nil = CAS なし。
	// 送信時は必ず WHERE 照合し、不一致は 409・書込なし。version 列は書かない。
	Version *int64
	// ActorID は全体割引除外の監査に使う操作スタッフ ID（trusted request context 由来）。
	ActorID *uint64
}

// --- DB column constants ---

// buildMerchandiseItemUpdate は UPDATE 用 map を構築する。
// GORM のゼロ値スキップ問題（bool false が無視される等）を回避するために使用する。

// ---- MerchandiseItemService ----

// Transactor is the consumer-side ambient transaction port for atomic merchandise delete.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// MerchandiseItemService は物販品マスタのビジネスロジック
type MerchandiseItemService interface {
	List(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	Create(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// merchandiseItemStore is the use case's consumer-side persistence view.
type merchandiseItemStore interface {
	FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
	CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)
	Create(ctx context.Context, item *model.MerchandiseItem) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateMerchandiseItemInput) (*model.MerchandiseItem, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type merchandiseItemService struct {
	repo       merchandiseItemStore
	transactor Transactor
	audit      AuditTxLogger
}

// NewMerchandiseItemService は物販品サービスを初期化する。
// transactor は Delete の soft-delete + usage 再確認を同一 transaction に載せるために必須
// （BE-ACT-MERCHANDISE-ATOMIC-DELETE）。
// 監査依存は未接続。新設定（is_bill_discount_excluded）の書込は fail-closed になるため、
// 書込を伴う利用には NewMerchandiseItemServiceWithAudit を使う（EMR-449）。
func NewMerchandiseItemService(repo merchandiseItemStore, transactor Transactor) MerchandiseItemService {
	return &merchandiseItemService{repo: repo, transactor: transactor}
}

// NewMerchandiseItemServiceWithAudit は新設定の書込と監査を同一 transaction に載せる構成。
// EMR-449 A01/A.5: create は常に新列を書くため監査依存が必須で、欠落は書込前に fail-closed。
func NewMerchandiseItemServiceWithAudit(repo merchandiseItemStore, transactor Transactor, audit AuditTxLogger) MerchandiseItemService {
	return &merchandiseItemService{repo: repo, transactor: transactor, audit: audit}
}

func (s *merchandiseItemService) List(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
	result, err := s.repo.FindAll(ctx, clinicID, category)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list merchandise items")
	}
	return result, nil
}

func (s *merchandiseItemService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get merchandise item")
	}
	return result, nil
}

func (s *merchandiseItemService) Create(ctx context.Context, clinicID uint64, input *CreateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}

	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		taxType = model.TaxType(input.TaxType)
	}
	taxRate := defaultMerchandiseTaxRate()
	if input.TaxRate != nil {
		if err := validateMerchandiseTaxRate(*input.TaxRate); err != nil {
			return nil, err
		}
		taxRate = *input.TaxRate
	}
	item := &model.MerchandiseItem{
		ClinicID:  clinicID,
		Name:      input.Name,
		Category:  model.ItemCategory(input.Category),
		UnitPrice: input.UnitPrice,
		TaxType:   taxType,
		TaxRate:   taxRate,
		IsActive:  input.IsActive,
		SortOrder: input.SortOrder,
		// EMR-449 A.2: 省略は明示 false を保存し、DB default や Go のゼロ値に任せない。
		IsBillDiscountExcluded: createBillDiscountExcluded(input.IsBillDiscountExcluded),
	}

	// EMR-449 A.5: create は常に新列を書くため、書込と監査を同一最外 TX に載せる。
	// 監査依存の欠落は書込前に fail-closed（500・書込0）。
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("merchandise item transaction dependency is required")
	}
	if s.audit == nil {
		return nil, apperrors.WrapInternalServerError("merchandise item bill discount audit dependency is required")
	}
	var result *model.MerchandiseItem
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, item); err != nil {
			return apperrors.Wrap(err, "failed to create merchandise item")
		}
		// A.4: is_active=false 補償 UPDATE で trigger が版を進めるため、同 TX で再読取し実保存版を返す。
		stored, err := s.repo.FindByID(txCtx, clinicID, item.ID)
		if err != nil {
			return apperrors.Wrap(err, "failed to reload merchandise item after create")
		}
		if err := logBillDiscountAuditTx(txCtx, s.audit, clinicID, input.ActorID,
			merchandiseItemAuditResource, merchandiseItemAuditActionCreate,
			billDiscountChange{ResourceID: stored.ID, BeforeExists: false, After: stored.IsBillDiscountExcluded},
		); err != nil {
			return err
		}
		result = stored
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "merchandise item created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", result.ID),
	)
	return result, nil
}

// merchandiseItemForUpdateFinder は更新 TX の前値読取専用の FOR UPDATE read。
// concrete repo が実装しない構成は fail-closed（EMR-449 A-2/D2）。
type merchandiseItemForUpdateFinder interface {
	FindByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
}

func (s *merchandiseItemService) Update(ctx context.Context, clinicID, id uint64, input *UpdateMerchandiseItemInput) (*model.MerchandiseItem, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	// EMR-449 A.4: 送信された version は必ず検査し無視しない（0/負は 400）。
	if input.Version != nil && *input.Version < 1 {
		return nil, apperrors.WrapInvalidInput(errMsgMerchandiseVersionInvalid)
	}
	// 新設定を含む PATCH は編集開始 version 必須。
	if input.IsBillDiscountExcluded != nil && input.Version == nil {
		return nil, apperrors.WrapInvalidInput(errMsgMerchandiseVersionRequired)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get merchandise item")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	// 旧 wire の binding max=1 に代わる exact 比較による入力範囲検査（EMR-530）。
	if input.TaxRate != nil {
		if err := validateMerchandiseTaxRate(*input.TaxRate); err != nil {
			return nil, err
		}
	}
	fields := buildMerchandiseItemUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}

	if input.IsBillDiscountExcluded != nil {
		// EMR-449 A.5: 新設定の書込は before/after 監査と同一最外 TX。依存欠落は fail-closed。
		if s.transactor == nil {
			return nil, apperrors.WrapInternalServerError("merchandise item transaction dependency is required")
		}
		if s.audit == nil {
			return nil, apperrors.WrapInternalServerError("merchandise item bill discount audit dependency is required")
		}
		var result *model.MerchandiseItem
		if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
			// EMR-449 A-2: 前値読取は FOR UPDATE で行を排他ロックしながら読む。
			// FOR SHARE のままだと同版の並行 PATCH が SHARE を保持したまま UPDATE
			// 昇格を待ち合い 40P01（500）になる。FOR UPDATE を先に取れば敗者は
			// lock 解放後に新しい版を読み、CAS の WHERE 不一致で 409 に収束する。
			forUpdateRepo, ok := s.repo.(merchandiseItemForUpdateFinder)
			if !ok {
				return apperrors.WrapInternalServerError("merchandise item repository must implement FindByIDForUpdate")
			}
			before, err := forUpdateRepo.FindByIDForUpdate(txCtx, clinicID, id)
			if err != nil {
				return apperrors.Wrap(err, "failed to get merchandise item")
			}
			updated, err := s.repo.Update(txCtx, clinicID, id, *input)
			if err != nil {
				return apperrors.Wrap(err, "failed to update merchandise item")
			}
			if err := logBillDiscountAuditTx(txCtx, s.audit, clinicID, input.ActorID,
				merchandiseItemAuditResource, merchandiseItemAuditActionUpdate,
				billDiscountChange{ResourceID: id, BeforeExists: true, Before: before.IsBillDiscountExcluded, After: updated.IsBillDiscountExcluded},
			); err != nil {
				return err
			}
			result = updated
			return nil
		}); err != nil {
			return nil, err
		}
		slog.InfoContext(ctx, "merchandise item updated",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("merchandise_item_id", id),
		)
		return result, nil
	}

	result, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update merchandise item")
	}
	slog.InfoContext(ctx, "merchandise item updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", id),
	)
	return result, nil
}

func (s *merchandiseItemService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder merchandise items")
	}
	slog.InfoContext(ctx, "merchandise items reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

func (s *merchandiseItemService) Delete(ctx context.Context, clinicID, id uint64) error {
	// UX Count is best-effort. Integrity is the repository's clinic_id+id+usage
	// NOT EXISTS DELETE, which serializes with campaign target FOR SHARE.
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("merchandise item transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		count, err := s.repo.CountUsageByMerchandiseItemID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check merchandise item dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この物販品目は請求・見積・キャンペーンで使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete merchandise item")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to delete merchandise item")
	}
	slog.InfoContext(ctx, "merchandise item deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("merchandise_item_id", id),
	)
	return nil
}
