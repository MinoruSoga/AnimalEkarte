// Package billing provides billing item create use cases.
package billing

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// validateBillingItemOwnership は billing および trimming マスタFKのテナント所有権を検証する。
func (s *billingItemService) validateBillingItemOwnership(ctx context.Context, input *CreateBillingItemInput) error {
	// テナント所有権確認: billing が同一クリニックに属することを確認
	if _, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID); err != nil {
		return apperrors.Wrap(err, "billing not found or belongs to different clinic")
	}

	// trimming_course_id / trimming_option_id を伴う明細の作成は validateCreateBillingItemInput で拒否済み（EMR-505 / Q7）のため、
	// trimming マスタの所有権検証（X-4）は不要になった。
	return nil
}

// resolveOwnerDiscountRate は会計に紐付く飼主の割引率(%)を ExactNumeric のまま返す（float64 を経由しない）。
// ownerRepo 未配線・飼主未解決・取得失敗は 0（従来どおりの best-effort。割引なしで続行する）。
func (s *billingItemService) resolveOwnerDiscountRate(ctx context.Context, clinicID uint64, ownerID *uint64) model.ExactNumeric {
	noDiscount := model.ExactNumericFromInt64(0)
	if ownerID == nil || s.ownerRepo == nil {
		return noDiscount
	}
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, *ownerID)
	if err != nil || owner == nil {
		return noDiscount
	}
	return owner.DiscountRate
}

// resolveAutoDiscount は #81 段階2b: 明細に適用するキャンペーン/飼主割引を算出する(best-effort)。
// 会計日(billing.ScheduledDate)・明細カテゴリ・個別商品IDで該当キャンペーンを検索し、
// 飼主割引と高い方を採用する(ResolveItemDiscount)。
// EMR-229: 適用割引率も返す。campaignRepo 未配線でも飼主割引は適用する（従来は campaignRepo
// 未配線だと飼主割引まで 0 になっていた）。
func (s *billingItemService) resolveAutoDiscount(
	ctx context.Context,
	clinicID uint64,
	billing *model.Billing,
	category model.ItemCategory,
	merchandiseItemID *uint64,
	unitPrice, quantity model.ExactNumeric,
) (int64, model.ExactNumeric) {
	ownerRate := s.resolveOwnerDiscountRate(ctx, clinicID, billing.OwnerID)
	var campaign *model.Campaign
	if s.campaignRepo != nil {
		c, cerr := s.campaignRepo.FindApplicableForItem(ctx, clinicID, billing.ScheduledDate, category, merchandiseItemID)
		if cerr != nil {
			// A-4: best-effort 継続自体は妥当（自動割引はあくまで補助機能）だが、クエリ障害で
			// 自動割引が静かに止まると運用から不可視になるため Warn ログを追加する。
			slog.WarnContext(ctx, "campaign lookup failed; skipping auto discount", "error", cerr, "clinic_id", clinicID, "billing_id", billing.ID)
		} else {
			campaign = c
		}
	}
	itemSubtotal, err := discountBaseAmount(unitPrice, quantity)
	if err != nil {
		// 割引の基数を確定できない行には自動割引を付けない（best-effort。割引額 0 として続行する）。
		slog.WarnContext(ctx, "discount base unavailable; skipping auto discount", "error", err, "clinic_id", clinicID, "billing_id", billing.ID)
		return 0, model.ExactNumericFromInt64(0)
	}
	return ResolveItemDiscount(itemSubtotal, campaign, ownerRate)
}

func (s *billingItemService) CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if err := validateCreateBillingItemInput(input); err != nil {
		return nil, err
	}
	// #115 / BUG-463: 締め後編集は理由必須（handler 迂回経路にも強制）
	if err := requirePostCloseReason(input.IsPostClose, input.PostCloseReason); err != nil {
		return nil, err
	}
	if err := s.validateBillingItemOwnership(ctx, input); err != nil {
		return nil, err
	}
	// 未請求予防接種の blocking warning は会計確定と会計作成だけを止める（BUG-015）。
	// 明細追加に流用すると物販・その他が常時 409 になり会計業務が停止する。

	var item *model.BillingItem
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		created, err := s.createItemInAmbientTx(txCtx, input, createItemAmbientOpts{
			recordPostClose: true,
		})
		if err != nil {
			return err
		}
		item = created
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create billing item in transaction")
	}

	slog.InfoContext(ctx, "billing item created",
		slog.Uint64("clinic_id", input.ClinicID),
		slog.Uint64("billing_id", input.BillingID),
		slog.Uint64("item_id", item.ID),
	)
	return item, nil
}

// createItemInAmbientTx は ambient tx 内で明細1件を作成する（WithTx を開始しない）。
// BUG-018 Complete 経路から呼び出され、独立 tx の部分 commit を防ぐ。
func (s *billingItemService) createItemInAmbientTx(ctx context.Context, input *CreateBillingItemInput, opts createItemAmbientOpts) (*model.BillingItem, error) {
	taxType, taxRate, source, err := resolveBillingItemDefaults(input)
	if err != nil {
		return nil, err
	}

	item := newBillingItemFromCreateInput(input, taxType, taxRate, source)

	billing, err := s.billingRepo.LockAndFindByID(ctx, input.ClinicID, input.BillingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock billing before creating item")
	}
	if err := rejectIfBillingFinalized(billing, "登録"); err != nil {
		return nil, err
	}

	// 請求元（treatment・appointment・trimming）を伴う作成は validateCreateBillingItemInput で拒否済み（Q7）のため、
	// ここで検証する参照は merchandise_item_id だけ。provenance の引数には nil を渡す。
	category, err := s.repo.ValidateCreateReferences(ctx, input.ClinicID, input.BillingID, input.MerchandiseItemID, nil, nil, nil, nil)
	if err != nil {
		slog.WarnContext(ctx, "billing item references rejected", "error", err)
		return nil, apperrors.Wrap(err, "failed to validate billing item references")
	}
	if input.MerchandiseItemID != nil {
		item.Category = category
	}

	if err := applyBillingItemOtherMetadata(input, item); err != nil {
		return nil, err
	}
	if item.CreatedBy != nil {
		if err := s.repo.LockActiveStaffAssignment(ctx, input.ClinicID, *item.CreatedBy); err != nil {
			return nil, err
		}
	}

	if item.DiscountAmount == 0 {
		// EMR-229: 自動割引が掛かった場合は適用率も discount_rate に記録し、
		// 画面で「飼主マスタの割引率が入っている」状態を再現できるようにする。
		if amount, appliedRate := s.resolveAutoDiscount(
			ctx,
			input.ClinicID,
			billing,
			item.Category,
			input.MerchandiseItemID,
			item.UnitPrice,
			item.Quantity,
		); amount > 0 {
			item.DiscountAmount = amount
			item.DiscountRate = appliedRate
		}
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, apperrors.Wrap(err, "failed to create billing item")
	}

	if opts.recordPostClose {
		if err := s.recordBillingItemPostClose(ctx, input.IsPostClose, input.ClinicID, input.BillingID, billing, input.StaffID, input.PostCloseReason, "create", &item.ID); err != nil {
			return nil, err
		}
	}

	return item, nil
}

func newBillingItemFromCreateInput(
	input *CreateBillingItemInput,
	taxType model.TaxType,
	taxRate model.ExactNumeric,
	source model.ItemSource,
) *model.BillingItem {
	// request の IsInsuranceApplicable・SortOrder は非 pointer（省略 = false / 0 の旧来の挙動）。明細へは指定された値として書く。
	isInsuranceApplicable := input.IsInsuranceApplicable
	sortOrder := input.SortOrder
	return &model.BillingItem{
		BillingID:             input.BillingID,
		Category:              model.ItemCategory(input.Category),
		Name:                  input.Name,
		UnitPrice:             input.UnitPrice,
		Quantity:              input.Quantity,
		DiscountRate:          input.DiscountRate,
		DiscountAmount:        input.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               &taxRate,
		IsInsuranceApplicable: &isInsuranceApplicable,
		Source:                &source,
		MerchandiseItemID:     input.MerchandiseItemID,
		SortOrder:             &sortOrder,
		ClinicID:              input.ClinicID,
	}
}

// CreateItemForComplete は BUG-018 Complete 用の ambient-tx 明細作成。
// unbilled / post-close は command 側が一度だけ行う。合計は保存しない（RecalculateTotalsForComplete は算定のみ）。
func (s *billingItemService) CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if err := validateCreateBillingItemInput(input); err != nil {
		return nil, err
	}
	return s.createItemInAmbientTx(ctx, input, createItemAmbientOpts{
		recordPostClose: false,
	})
}

// DeleteItemsForComplete は EMR-253: takeover 対象 waiting billing の既存明細を一括削除する。
// ambient tx（呼出し元の complete tx）に参加し、失敗時は呼出し側が全体 rollback する。
func (s *billingItemService) DeleteItemsForComplete(ctx context.Context, clinicID, billingID uint64) error {
	if err := s.repo.DeleteAllByBillingID(ctx, clinicID, billingID); err != nil {
		return apperrors.Wrap(err, "failed to delete billing items for complete takeover")
	}
	return nil
}

// RecalculateTotalsForComplete は Complete 用に items から合計を算定して返す。
// billings のヘッダ合計列は廃止された（001_init.sql:10889-10891）ため保存せず、算定のみを行う。算定結果の保存は
// 算定固定（snapshot）の writer が担う。held for EMR-462/506: このインターフェース（accounting_complete.go:136）の整理は 506 が行う。
// 内税・非課税など算定器が扱えない明細は calculation_unsupported（422）を返し、0 で埋めない。
func (s *billingItemService) RecalculateTotalsForComplete(ctx context.Context, clinicID, billingID uint64) (subtotal, taxTotal, totalAmount int64, err error) {
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		return 0, 0, 0, apperrors.Wrap(err, "failed to find billing items for complete totals")
	}
	result, err := CalculateItemsTotals(items)
	if err != nil {
		return 0, 0, 0, err
	}
	return result.Subtotal, result.TaxTotal, result.TotalAmount, nil
}
