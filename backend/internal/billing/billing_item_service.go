// Package service provides business logic implementations for BillingItem entity.
package billing

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// --- DB column constants ---

const (
	colBillingItemUnitPrice             = "unit_price"
	colBillingItemQuantity              = "quantity"
	colBillingItemDiscountRate          = "discount_rate"
	colBillingItemDiscountAmount        = "discount_amount"
	colBillingItemTaxType               = "tax_type"
	colBillingItemTaxRate               = "tax_rate"
	colBillingItemIsInsuranceApplicable = "is_insurance_applicable"
	otherReasonMaxRunes                 = 500
)

// --- Input DTOs ---

// CreateBillingItemInput は billing_item 作成の入力DTO
type CreateBillingItemInput struct {
	ClinicID              uint64
	BillingID             uint64
	Category              string // Service 内で model.ItemCategory に変換
	Name                  string
	UnitPrice             int64
	Quantity              float64
	DiscountRate          float64
	DiscountAmount        int64
	TaxType               string // "" = デフォルト "excluded"
	TaxRate               float64
	IsInsuranceApplicable bool
	Source                string // "" = デフォルト "manual"
	OtherReason           *string
	CreatedBy             *uint64
	TreatmentID           *uint64
	VaccinationID         *uint64
	ExamID                *uint64
	AppointmentID         *uint64
	TrimmingCourseID      *uint64
	TrimmingOptionID      *uint64
	MerchandiseItemID     *uint64 // #81: 個別商品指定によるキャンペーンマッチング
	SortOrder             int
	// #115 / BUG-463: 締め後編集（handler がレジ締め済み判定を注入）
	StaffID         *uint64
	PostCloseReason *string
	IsPostClose     bool
}

// UpdateBillingItemInput は billing_item 更新の入力DTO（nil = 未指定）
type UpdateBillingItemInput struct {
	UnitPrice             *int64
	Quantity              *float64
	DiscountRate          *float64
	DiscountAmount        *int64
	TaxType               *model.TaxType
	TaxRate               *float64
	IsInsuranceApplicable *bool
	// #115 / BUG-463: 締め後編集（handler がレジ締め済み判定を注入）
	StaffID         *uint64
	PostCloseReason *string
	IsPostClose     bool
}

// DeleteBillingItemInput は明細削除の入力DTO。
// StaffID は vaccination claim 解放監査 actor（BUG-440）および締め後編集監査 actor。
// IsPostClose / PostCloseReason はレジ締め後削除ゲート（BUG-463 residual）。
// nil input は非締め後削除として扱う。
type DeleteBillingItemInput struct {
	StaffID         *uint64
	PostCloseReason *string
	IsPostClose     bool
}

func buildBillingItemUpdate(input *UpdateBillingItemInput) map[string]any {
	fields := make(map[string]any)
	if input.UnitPrice != nil {
		fields[colBillingItemUnitPrice] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields[colBillingItemQuantity] = *input.Quantity
	}
	if input.DiscountRate != nil {
		fields[colBillingItemDiscountRate] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields[colBillingItemDiscountAmount] = *input.DiscountAmount
	}
	if input.TaxType != nil {
		fields[colBillingItemTaxType] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields[colBillingItemTaxRate] = *input.TaxRate
	}
	if input.IsInsuranceApplicable != nil {
		fields[colBillingItemIsInsuranceApplicable] = *input.IsInsuranceApplicable
	}
	return fields
}

func validateCreateBillingItemInput(input *CreateBillingItemInput) error {
	if input == nil {
		return apperrors.WrapInvalidInput("請求明細は必須です")
	}
	if input.BillingID == 0 {
		return apperrors.WrapInvalidInput("請求IDは必須です")
	}
	if input.Name == "" {
		return apperrors.WrapInvalidInput("商品名は必須です")
	}
	if input.UnitPrice < 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	if input.Quantity <= 0 {
		return apperrors.WrapInvalidInput("数量は正の値である必要があります")
	}
	return nil
}

func resolveBillingItemDefaults(input *CreateBillingItemInput) (model.TaxType, float64, model.ItemSource, error) {
	// カテゴリバリデーション
	if err := validateItemCategory(input.Category); err != nil {
		return "", 0, "", apperrors.Wrap(err, "failed to validate item category")
	}

	// TaxType デフォルト設定とバリデーション
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		if err := sharedkernel.ValidateTaxType(input.TaxType); err != nil {
			return "", 0, "", apperrors.Wrap(err, "failed to validate tax type")
		}
		taxType = model.TaxType(input.TaxType)
	}

	// TaxRate デフォルト設定
	taxRate := sharedkernel.DefaultTaxRate
	if input.TaxRate > 0 {
		taxRate = input.TaxRate
	}

	// Source デフォルト設定とバリデーション
	source := model.ItemSourceManual
	if input.Source != "" {
		if err := validateItemSource(input.Source); err != nil {
			return "", 0, "", apperrors.Wrap(err, "failed to validate item source")
		}
		source = model.ItemSource(input.Source)
	}

	return taxType, taxRate, source, nil
}

func applyBillingItemOtherMetadata(input *CreateBillingItemInput, item *model.BillingItem) error {
	item.OtherReason = nil
	item.CreatedBy = nil
	if item.Source != model.ItemSourceManual ||
		item.Category != model.ItemCategoryOther ||
		input.MerchandiseItemID != nil {
		return nil
	}
	if input.OtherReason == nil {
		return apperrors.WrapInvalidInput("その他カテゴリの手入力明細は理由を入力してください")
	}
	trimmed := strings.TrimSpace(*input.OtherReason)
	if trimmed == "" {
		return apperrors.WrapInvalidInput("その他カテゴリの手入力明細は理由を入力してください")
	}
	if utf8.RuneCountInString(trimmed) > otherReasonMaxRunes {
		return apperrors.WrapInvalidInput("その他理由は500文字以内で入力してください")
	}
	if input.CreatedBy == nil || *input.CreatedBy == 0 {
		return apperrors.WrapInvalidInput("その他カテゴリの手入力明細は操作者が必要です")
	}
	item.OtherReason = &trimmed
	item.CreatedBy = input.CreatedBy
	return nil
}

// ---- BillingItemService ----

// BillingItemService は billing_items の CRUD とトータル再計算を担うインターフェース
type BillingItemService interface {
	CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	// CreateItemForComplete / RecalculateTotalsForComplete は BUG-018 Complete の ambient-tx collaborator。
	CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	RecalculateTotalsForComplete(ctx context.Context, clinicID, billingID uint64) (subtotal, taxTotal, totalAmount int64, err error)
	UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error)
	// DeleteItem は明細を soft-delete する。
	// input.StaffID は vaccination claim 解放監査 actor（BUG-440）および締め後編集監査 actor。
	// input.IsPostClose 時は post_close_reason 必須 + 同 tx fail-closed 監査（BUG-463 residual）。
	// nil input は非締め後削除として扱う。
	DeleteItem(ctx context.Context, clinicID, id uint64, input *DeleteBillingItemInput) error
	GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
	// GetUnbilledItemDetails は未請求候補 items と typed blocking warnings を返す（BUG-013）。
	GetUnbilledItemDetails(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error)
	// AssertNoBlockingUnbilled は pet に blocking unbilled warning がある場合 Conflict を返す（write-time fail-closed）。
	AssertNoBlockingUnbilled(ctx context.Context, clinicID, petID uint64) error
	// GetUngroupedSameDaySummary は同日同ペットの未会計対象化項目(診察/トリミング)の件数を返す(#77 取り残し警告)。
	GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error)
	// GetDiscountSuggestions は指定明細に適用可能な割引候補を返す（#81 Q-I スタッフ選択）。
	// campaignRepo 未配線の場合は飼主割引のみ。
	GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error)
	// GetBilling は締め後編集判定用に請求ヘッダを返す（handler が ScheduledDate を参照）。
	GetBilling(ctx context.Context, clinicID, billingID uint64) (*model.Billing, error)
	// GetBillingForItem は明細 ID から親請求ヘッダを返す（締め後編集判定用）。
	GetBillingForItem(ctx context.Context, clinicID, itemID uint64) (*model.Billing, error)
}

// UngroupedSameDaySummary は #77 取り残し警告用の未会計対象化件数サマリ。
type UngroupedSameDaySummary struct {
	MedicalRecordCount int64
	TrimmingCount      int64
}

// Unbilled warning codes / sources (BUG-013). Keep payload minimal — no IDs, names, prices, SQL.
const (
	UnbilledWarningSourceVaccination               = "vaccination"
	UnbilledWarningCodeVaccinationMasterUnbillable = "vaccination_master_unbillable"
)

// UnbilledWarning は未請求集約の data-quality 警告（response 公開契約）。
type UnbilledWarning struct {
	Source   string `json:"source"`
	Code     string `json:"code"`
	Count    int    `json:"count"`
	Blocking bool   `json:"blocking"`
}

// UnbilledDetails は additive GET /billing-items/unbilled-details の結果。
type UnbilledDetails struct {
	Items    []model.BillingItem
	Warnings []UnbilledWarning
}

type billingItemService struct {
	repo               BillingItemRepository
	billingRepo        accountingBillingView
	treatmentRepo      treatmentBillingReader
	transactor         Transactor
	trimmingCourseRepo trimmingCourseFinder // X-4: クロステナント write 防止用の所有権検証
	trimmingOptionRepo trimmingOptionFinder // X-4: クロステナント write 防止用の所有権検証
	campaignRepo       CampaignRepository   // #81 段階2b: nil の場合は自動割引なし
	ownerRepo          billingOwnerReader   // #81 段階2b: 飼主割引取得用
	auditTx            billingAuditTxLogger // #115 / BUG-463: 締め後編集 fail-closed 監査
	// closeRepo は W-013 締め後明細変更の append-only adjustment 台帳用（任意 DI）。
	closeRepo CashRegisterCloseRepository
}

type unbilledTrimmingItemFinder interface {
	FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
}

type ungroupedTrimmingCounter interface {
	CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

// billingItemServiceOption は BillingItemService 構築時の任意依存注入。
type billingItemServiceOption func(*billingItemService)

// WithBillingItemAuditTx は締め後編集の fail-closed 監査 logger を配線する（#115 / BUG-463）。
func WithBillingItemAuditTx(auditTx billingAuditTxLogger) billingItemServiceOption {
	return func(s *billingItemService) {
		s.auditTx = auditTx
	}
}

// WithBillingItemCloseRepository は締め後明細変更時の cash_register_close_adjustments 追記に使う close repo を配線する（W-013）。
func WithBillingItemCloseRepository(repo CashRegisterCloseRepository) billingItemServiceOption {
	return func(s *billingItemService) {
		s.closeRepo = repo
	}
}

// NewBillingItemServiceWithCampaign は #81 段階2b: キャンペーン/飼主割引の自動適用を有効にした BillingItemService を返す。
func NewBillingItemServiceWithCampaign(repo BillingItemRepository, billingRepo accountingBillingView, treatmentRepo treatmentBillingReader, transactor Transactor, trimmingCourseRepo trimmingCourseFinder, trimmingOptionRepo trimmingOptionFinder, campaignRepo CampaignRepository, ownerRepo billingOwnerReader, opts ...billingItemServiceOption) BillingItemService {
	s := &billingItemService{repo: repo, billingRepo: billingRepo, treatmentRepo: treatmentRepo, transactor: transactor, trimmingCourseRepo: trimmingCourseRepo, trimmingOptionRepo: trimmingOptionRepo, campaignRepo: campaignRepo, ownerRepo: ownerRepo}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// rejectIfBillingFinalized は確定/取消済み会計への明細変更を Conflict で拒否する（BUG-463 / DeleteItem と同型）。
func rejectIfBillingFinalized(billing *model.Billing, action string) error {
	if billing == nil {
		return apperrors.WrapInternalServerError("locked billing is nil")
	}
	if billing.Status == model.BillingStatusCompleted ||
		billing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict(fmt.Sprintf("確定済みまたは取消済みの会計明細は%sできません", action))
	}
	return nil
}

// rejectUnlessCompletedCorrectionAllowed は UpdateItem 用。
// cancelled は常に拒否。completed は post_close_reason 付きの訂正のみ許可（BUG-009）。
func rejectUnlessCompletedCorrectionAllowed(billing *model.Billing, action string, reason *string) error {
	if billing == nil {
		return apperrors.WrapInternalServerError("locked billing is nil")
	}
	if billing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict(fmt.Sprintf("確定済みまたは取消済みの会計明細は%sできません", action))
	}
	if billing.Status == model.BillingStatusCompleted {
		if reason == nil || strings.TrimSpace(*reason) == "" {
			return apperrors.WrapInvalidInput("確定済み会計の明細修正には修正理由（post_close_reason）が必要です")
		}
		return nil
	}
	return nil
}

func requirePostCloseReason(isPostClose bool, reason *string) error {
	if isPostClose && (reason == nil || *reason == "") {
		return apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}
	return nil
}

// validateBillingItemOwnership は billing および trimming マスタFKのテナント所有権を検証する。
func (s *billingItemService) validateBillingItemOwnership(ctx context.Context, input *CreateBillingItemInput) error {
	// テナント所有権確認: billing が同一クリニックに属することを確認
	if _, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID); err != nil {
		slog.ErrorContext(ctx, "billing not found or belongs to different clinic", "error", err)
		return apperrors.Wrap(err, "billing not found or belongs to different clinic")
	}

	// X-4: クロステナント write 防止: trimming_course_id/trimming_option_id が
	// caller の clinic に属することを検証する(#124/#125 と同型の master FK 所有権チェック)。
	if input.TrimmingCourseID != nil {
		if _, err := s.trimmingCourseRepo.FindByID(ctx, input.ClinicID, *input.TrimmingCourseID); err != nil {
			slog.ErrorContext(ctx, "trimming course not found or belongs to different clinic", "error", err)
			return apperrors.Wrap(err, "failed to verify trimming course ownership")
		}
	}
	if input.TrimmingOptionID != nil {
		if _, err := s.trimmingOptionRepo.FindByID(ctx, input.ClinicID, *input.TrimmingOptionID); err != nil {
			slog.ErrorContext(ctx, "trimming option not found or belongs to different clinic", "error", err)
			return apperrors.Wrap(err, "failed to verify trimming option ownership")
		}
	}
	return nil
}

// resolveOwnerDiscountRate は会計に紐付く飼主の割引率を返す。ownerRepo 未配線または取得失敗は 0。
func (s *billingItemService) resolveOwnerDiscountRate(ctx context.Context, clinicID uint64, ownerID *uint64) float64 {
	if ownerID == nil || s.ownerRepo == nil {
		return 0
	}
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, *ownerID)
	if err != nil || owner == nil {
		return 0
	}
	return owner.DiscountRate
}

// resolveAutoDiscount は #81 段階2b: 明細に適用するキャンペーン/飼主割引額を算出する(best-effort)。
// campaignRepo 未配線時は 0。会計日(billing.ScheduledDate)・明細カテゴリ・個別商品IDで該当キャンペーンを検索し、
// 飼主割引と高い方を採用する(CalculateItemCampaignDiscount)。
func (s *billingItemService) resolveAutoDiscount(
	ctx context.Context,
	input *CreateBillingItemInput,
	category model.ItemCategory,
	unitPrice int64,
	quantity float64,
) int64 {
	if s.campaignRepo == nil {
		return 0
	}
	billing, err := s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID)
	if err != nil || billing == nil {
		return 0
	}
	ownerRate := s.resolveOwnerDiscountRate(ctx, input.ClinicID, billing.OwnerID)
	campaign, cerr := s.campaignRepo.FindApplicableForItem(ctx, input.ClinicID, billing.ScheduledDate, category, input.MerchandiseItemID)
	if cerr != nil {
		// A-4: best-effort 継続自体は妥当（自動割引はあくまで補助機能）だが、クエリ障害で
		// 自動割引が静かに止まると運用から不可視になるため Warn ログを追加する。
		slog.WarnContext(ctx, "campaign lookup failed; skipping auto discount", "error", cerr, "clinic_id", input.ClinicID, "billing_id", input.BillingID)
		campaign = nil // best-effort: キャンペーン検索失敗は割引なしで続行
	}
	itemSubtotal := int64(float64(unitPrice) * quantity)
	return CalculateItemCampaignDiscount(itemSubtotal, campaign, ownerRate)
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
			recalculate:     true,
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

// createItemAmbientOpts は CreateItem / Complete 共有の ambient tx 内明細作成オプション。
type createItemAmbientOpts struct {
	recalculate     bool // false のとき totals 再計算を呼び出し側に委ねる（Complete が最後に一括再計算）
	recordPostClose bool // false のとき締め後監査は command 単位で Complete が記録する
}

// createItemInAmbientTx は ambient tx 内で明細1件を作成する（WithTx を開始しない）。
// BUG-018 Complete 経路から呼び出され、独立 tx の部分 commit を防ぐ。
func (s *billingItemService) createItemInAmbientTx(ctx context.Context, input *CreateBillingItemInput, opts createItemAmbientOpts) (*model.BillingItem, error) {
	taxType, taxRate, source, err := resolveBillingItemDefaults(input)
	if err != nil {
		return nil, err
	}

	item := &model.BillingItem{
		BillingID:             input.BillingID,
		Category:              model.ItemCategory(input.Category),
		Name:                  input.Name,
		UnitPrice:             input.UnitPrice,
		Quantity:              input.Quantity,
		DiscountRate:          input.DiscountRate,
		DiscountAmount:        input.DiscountAmount,
		TaxType:               taxType,
		TaxRate:               taxRate,
		IsInsuranceApplicable: input.IsInsuranceApplicable,
		Source:                source,
		MerchandiseItemID:     input.MerchandiseItemID,
		TreatmentID:           input.TreatmentID,
		VaccinationID:         input.VaccinationID,
		ExamID:                input.ExamID,
		AppointmentID:         input.AppointmentID,
		TrimmingCourseID:      input.TrimmingCourseID,
		TrimmingOptionID:      input.TrimmingOptionID,
		SortOrder:             input.SortOrder,
	}

	// BUG-463: lock parent billing and reject finalized status before any mutation
	billing, err := s.billingRepo.LockAndFindByID(ctx, input.ClinicID, input.BillingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock billing before creating item")
	}
	if err := rejectIfBillingFinalized(billing, "登録"); err != nil {
		return nil, err
	}

	category, err := s.repo.ValidateCreateReferences(
		ctx,
		input.ClinicID,
		input.BillingID,
		input.MerchandiseItemID,
		input.TreatmentID,
		input.AppointmentID,
		input.TrimmingCourseID,
		input.TrimmingOptionID,
	)
	if err != nil {
		slog.WarnContext(ctx, "billing item references rejected", "error", err)
		return nil, apperrors.Wrap(err, "failed to validate billing item references")
	}
	if input.MerchandiseItemID != nil {
		item.Category = category
	}
	if input.ExamID != nil && input.VaccinationID != nil {
		return nil, invalidBillingItemReferenceCombination()
	}
	if err := applyExamBillingProvenance(ctx, s.repo, input, item); err != nil {
		return nil, err
	}
	if input.VaccinationID != nil {
		if input.MerchandiseItemID != nil ||
			input.TreatmentID != nil ||
			input.AppointmentID != nil ||
			input.TrimmingCourseID != nil ||
			input.TrimmingOptionID != nil ||
			input.ExamID != nil {
			return nil, invalidBillingItemReferenceCombination()
		}
		_, err := s.repo.ValidateVaccinationCreateReference(
			ctx,
			input.ClinicID,
			input.BillingID,
			*input.VaccinationID,
		)
		if err != nil {
			return nil, err
		}
		item.Category = model.ItemCategoryVaccine
		item.Source = model.ItemSourceMedicalRecord
		item.VaccinationID = input.VaccinationID
		clinicID := input.ClinicID
		item.ClinicID = &clinicID
	}

	if err := applyBillingItemOtherMetadata(input, item); err != nil {
		return nil, err
	}
	if item.CreatedBy != nil {
		if err := s.repo.LockActiveStaffAssignment(ctx, input.ClinicID, *item.CreatedBy); err != nil {
			return nil, err
		}
	}

	// #81 段階2b: 明示的な割引指定が無ければキャンペーン/飼主割引を自動適用(best-effort)。
	if item.DiscountAmount == 0 && input.VaccinationID == nil && input.ExamID == nil {
		item.DiscountAmount = s.resolveAutoDiscount(
			ctx,
			input,
			item.Category,
			item.UnitPrice,
			item.Quantity,
		)
	}

	if err := s.repo.Create(ctx, item); err != nil {
		slog.ErrorContext(ctx, "failed to create billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to create billing item")
	}

	if opts.recalculate {
		if err := s.recalculateTotals(ctx, input.ClinicID, input.BillingID, false); err != nil {
			slog.ErrorContext(ctx, "failed to recalculate billing totals after create",
				slog.Uint64("billing_id", input.BillingID),
				slog.String("error", err.Error()),
			)
			return nil, apperrors.Wrap(err, "failed to recalculate billing totals")
		}
	}

	if opts.recordPostClose {
		if err := s.recordBillingItemPostClose(ctx, input.IsPostClose, input.ClinicID, input.BillingID, billing, input.StaffID, input.PostCloseReason, "create", &item.ID); err != nil {
			return nil, err
		}
	}

	return item, nil
}

// CreateItemForComplete は BUG-018 Complete 用の ambient-tx 明細作成。
// unbilled / post-close は command 側が一度だけ行う。totals 再計算は skip し caller が最後にまとめる。
func (s *billingItemService) CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error) {
	if err := validateCreateBillingItemInput(input); err != nil {
		return nil, err
	}
	return s.createItemInAmbientTx(ctx, input, createItemAmbientOpts{
		recalculate:     false,
		recordPostClose: false,
	})
}

// RecalculateTotalsForComplete は Complete 用に items から totals を再計算して billings に書く。
func (s *billingItemService) RecalculateTotalsForComplete(ctx context.Context, clinicID, billingID uint64) (subtotal, taxTotal, totalAmount int64, err error) {
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		return 0, 0, 0, apperrors.Wrap(err, "failed to find billing items for complete totals")
	}
	subtotal, taxTotal, totalAmount = CalculateBillingTotals(items)
	if err := s.repo.UpdateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount); err != nil {
		return 0, 0, 0, apperrors.Wrap(err, "failed to update billing totals for complete")
	}
	return subtotal, taxTotal, totalAmount, nil
}

func (s *billingItemService) UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("請求明細の更新内容は必須です")
	}
	// #115 / BUG-463: 締め後編集は理由必須（handler 迂回経路にも強制）
	if err := requirePostCloseReason(input.IsPostClose, input.PostCloseReason); err != nil {
		return nil, err
	}
	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get billing item", "error", err)
		return nil, apperrors.Wrap(err, "failed to get billing item")
	}
	if input.UnitPrice != nil && *input.UnitPrice < 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	if input.Quantity != nil && *input.Quantity <= 0 {
		return nil, apperrors.WrapInvalidInput("数量は正の値である必要があります")
	}

	fields := buildBillingItemUpdate(input)
	if len(fields) == 0 {
		return item, nil
	}

	var updated *model.BillingItem
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// BUG-463 / BUG-009: lock parent; cancelled 拒否、completed は理由付き訂正のみ許可
		billing, err := s.billingRepo.LockAndFindByID(txCtx, clinicID, item.BillingID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock billing before updating item")
		}
		if err := rejectUnlessCompletedCorrectionAllowed(billing, "更新", input.PostCloseReason); err != nil {
			return err
		}

		if err := s.repo.Update(txCtx, clinicID, id, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update billing item", "error", err)
			return apperrors.Wrap(err, "failed to update billing item")
		}

		completedCorrection := billing.Status == model.BillingStatusCompleted
		if err := s.recalculateTotals(txCtx, clinicID, item.BillingID, completedCorrection); err != nil {
			slog.ErrorContext(txCtx, "failed to recalculate billing totals after update",
				slog.Uint64("billing_id", item.BillingID),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, "failed to recalculate billing totals")
		}

		itemID := id
		if completedCorrection {
			// BUG-009: 確定済み訂正は常に監査。締め済み日なら adjustment 台帳も同 tx。
			if err := s.recordCompletedItemCorrection(txCtx, clinicID, item.BillingID, billing, input.StaffID, input.PostCloseReason, "update", &itemID); err != nil {
				return err
			}
		} else if err := s.recordBillingItemPostClose(txCtx, input.IsPostClose, clinicID, item.BillingID, billing, input.StaffID, input.PostCloseReason, "update", &itemID); err != nil {
			return err
		}

		slog.InfoContext(txCtx, "billing item updated",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("item_id", id),
			slog.Uint64("billing_id", item.BillingID),
		)
		updated, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get billing item after update", "error", err)
			return apperrors.Wrap(err, "failed to get billing item after update")
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to update billing item in transaction")
	}

	return updated, nil
}

func (s *billingItemService) DeleteItem(ctx context.Context, clinicID, id uint64, input *DeleteBillingItemInput) error {
	if input == nil {
		input = &DeleteBillingItemInput{}
	}
	// #115 / BUG-463 residual: 締め後削除は理由必須（handler 迂回経路にも強制）
	if err := requirePostCloseReason(input.IsPostClose, input.PostCloseReason); err != nil {
		return err
	}

	item, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get billing item")
	}
	billingID := item.BillingID
	// Delete が vaccination_id を NULL 化する前に claim 解放対象を保持する（BUG-440）。
	releasedVaccinationID := item.VaccinationID

	return s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		billing, err := s.billingRepo.LockAndFindByID(txCtx, clinicID, billingID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock billing before deleting item")
		}
		if err := rejectIfBillingFinalized(billing, "削除"); err != nil {
			return err
		}

		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete billing item", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete billing item")
		}

		if err := s.recalculateTotals(txCtx, clinicID, billingID, false); err != nil {
			slog.ErrorContext(txCtx, "failed to recalculate billing totals after delete",
				slog.Uint64("billing_id", billingID),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, "failed to recalculate billing totals")
		}

		// BUG-463 residual / W-013: 締め後削除は adjustment 台帳 + 監査を同 tx fail-closed で記録する。
		// vaccination claim 解放監査（BUG-440）と両立し、両方発火し得る。
		itemID := id
		if err := s.recordBillingItemPostClose(txCtx, input.IsPostClose, clinicID, billingID, billing, input.StaffID, input.PostCloseReason, "delete", &itemID); err != nil {
			return err
		}

		// BUG-440: vaccination claim 解放時のみ immutable actor 監査（同 tx fail-closed）。
		// 非 vaccination 明細削除は claim-release 監査しない（ledger 方針: claim 解放の追跡が目的）。
		if releasedVaccinationID != nil {
			if err := s.logVaccinationClaimRelease(txCtx, clinicID, billingID, id, *releasedVaccinationID, input.StaffID); err != nil {
				return err
			}
		}

		slog.InfoContext(txCtx, "billing item deleted",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("item_id", id),
			slog.Uint64("billing_id", billingID),
		)
		return nil
	})
}

// GetDiscountSuggestions は指定明細に適用可能な割引候補を返す (#81 Q-I スタッフ選択)。
func (s *billingItemService) GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error) {
	item, err := s.repo.FindByID(ctx, clinicID, itemID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing item")
	}
	billing, err := s.billingRepo.FindByID(ctx, clinicID, item.BillingID)
	if err != nil || billing == nil {
		return nil, apperrors.Wrap(err, "failed to find billing")
	}
	ownerRate := s.resolveOwnerDiscountRate(ctx, clinicID, billing.OwnerID)
	var campaigns []*model.Campaign
	if s.campaignRepo != nil {
		campaigns, err = s.campaignRepo.FindAllApplicableForItem(ctx, clinicID, billing.ScheduledDate, item.Category, item.MerchandiseItemID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find applicable campaigns for suggestions", "error", err)
			campaigns = nil // best-effort
		}
	}
	itemSubtotal := int64(float64(item.UnitPrice) * item.Quantity)
	return BuildDiscountSuggestions(itemSubtotal, campaigns, ownerRate), nil
}

// recalculateTotals は billing の全明細から subtotal/tax_total/total_amount を再計算して保存する。
// allowCompleted は BUG-009 の確定済み明細訂正経路でのみ true。
func (s *billingItemService) recalculateTotals(ctx context.Context, clinicID, billingID uint64, allowCompleted bool) error {
	items, err := s.repo.FindByBillingID(ctx, clinicID, billingID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find billing items")
	}
	subtotal, taxTotal, totalAmount := CalculateBillingTotals(items)
	var writeErr error
	if allowCompleted {
		writeErr = s.repo.UpdateBillingTotalsForCompletedCorrection(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount)
	} else {
		writeErr = s.repo.UpdateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount)
	}
	if writeErr != nil {
		return apperrors.Wrap(writeErr, "failed to update billing totals")
	}
	return nil
}

// recordCompletedItemCorrection は確定済み明細の理由付き訂正を監査し、締め済み日なら adjustment も書く（BUG-009）。
func (s *billingItemService) recordCompletedItemCorrection(
	ctx context.Context,
	clinicID, billingID uint64,
	billing *model.Billing,
	staffID *uint64,
	reason *string,
	operation string,
	itemID *uint64,
) error {
	if reason == nil || strings.TrimSpace(*reason) == "" {
		return apperrors.WrapInvalidInput("確定済み会計の明細修正には修正理由（post_close_reason）が必要です")
	}
	if billing == nil {
		return apperrors.WrapInternalServerError("billing is required for completed item correction")
	}
	// 締め済み日のみ台帳追記（close 無しで createPostCloseAdjustment すると Conflict になる）
	if s.closeRepo != nil {
		closed, err := s.closeRepo.HasCloseOnDate(ctx, clinicID, billing.ScheduledDate)
		if err != nil {
			return apperrors.Wrap(err, "failed to re-check cash register close state for completed item correction")
		}
		if closed {
			if err := createPostCloseAdjustment(ctx, s.closeRepo, clinicID, billingID, billing.ScheduledDate, strings.TrimSpace(*reason), staffID, 0); err != nil {
				return err
			}
		}
	}
	return s.logBillingItemPostCloseEdit(ctx, clinicID, billingID, itemID, staffID, reason, operation)
}

// GetBilling は締め後編集判定用に請求ヘッダを返す。
func (s *billingItemService) GetBilling(ctx context.Context, clinicID, billingID uint64) (*model.Billing, error) {
	billing, err := s.billingRepo.FindByID(ctx, clinicID, billingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing")
	}
	return billing, nil
}

// GetBillingForItem は明細 ID から親請求ヘッダを返す。
func (s *billingItemService) GetBillingForItem(ctx context.Context, clinicID, itemID uint64) (*model.Billing, error) {
	item, err := s.repo.FindByID(ctx, clinicID, itemID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing item")
	}
	billing, err := s.billingRepo.FindByID(ctx, clinicID, item.BillingID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find billing")
	}
	return billing, nil
}

// recordBillingItemPostClose は write 時に締め状態を再評価し、adjustment 台帳 + 監査を同一 tx で残す（W-013）。
// handler の IsPostClose フラグは候補 read に過ぎないため HasCloseOnDate で再評価する。
func (s *billingItemService) recordBillingItemPostClose(
	ctx context.Context,
	handlerFlag bool,
	clinicID, billingID uint64,
	billing *model.Billing,
	staffID *uint64,
	reason *string,
	operation string,
	itemID *uint64,
) error {
	postClose := handlerFlag
	if s.closeRepo != nil && billing != nil {
		closed, err := s.closeRepo.HasCloseOnDate(ctx, clinicID, billing.ScheduledDate)
		if err != nil {
			return apperrors.Wrap(err, "failed to re-check cash register close state for billing item")
		}
		if closed {
			postClose = true
		}
	}
	if !postClose {
		return nil
	}
	if reason == nil || strings.TrimSpace(*reason) == "" {
		return apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}
	// 明細経路は total 再計算後でも delta を行単位で確定しづらいため 0 とし、reason で追跡する。
	if err := createPostCloseAdjustment(ctx, s.closeRepo, clinicID, billingID, billing.ScheduledDate, *reason, staffID, 0); err != nil {
		return err
	}
	return s.logBillingItemPostCloseEdit(ctx, clinicID, billingID, itemID, staffID, reason, operation)
}

// logBillingItemPostCloseEdit はレジ締め済み期間の明細編集監査ログを記録する（#115 / BUG-463）。
// ambient tx に参加する LogEntryTx を使い、監査失敗時は明細変更ごとロールバックする（fail-closed）。
func (s *billingItemService) logBillingItemPostCloseEdit(
	ctx context.Context,
	clinicID, billingID uint64,
	itemID *uint64,
	staffID *uint64,
	reason *string,
	operation string,
) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("billing audit dependency is required for post-close edits")
	}
	meta := map[string]any{
		"billing_id": billingID,
		"operation":  operation,
	}
	if itemID != nil {
		meta["item_id"] = *itemID
	}
	if reason != nil {
		meta["reason"] = *reason
	}
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    staffID,
		ActorType:  sharedkernel.AuditActorTypeFor(staffID),
		Action:     model.AuditActionBillingPostCloseEdit,
		Resource:   "billing_item",
		ResourceID: itemID,
		Metadata:   meta,
	}); err != nil {
		return apperrors.Wrap(err, "failed to write post_close_edit audit log")
	}
	return nil
}

// logVaccinationClaimRelease は明細削除による予防接種 claim 解放を監査する（BUG-440 / DEC-28 A）。
// ambient tx の LogEntryTx で fail-closed: 監査失敗時は delete + totals 再計算ごとロールバックする。
// metadata は ID のみ（PII 非格納）。
func (s *billingItemService) logVaccinationClaimRelease(
	ctx context.Context,
	clinicID, billingID, itemID, vaccinationID uint64,
	staffID *uint64,
) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("billing audit dependency is required for vaccination claim release")
	}
	meta := map[string]any{
		"billing_id":     billingID,
		"item_id":        itemID,
		"vaccination_id": vaccinationID,
		"reason":         "billing_item_delete",
	}
	resourceID := itemID
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    staffID,
		ActorType:  sharedkernel.AuditActorTypeFor(staffID),
		Action:     model.AuditActionBillingVaccinationClaimRelease,
		Resource:   "billing_item",
		ResourceID: &resourceID,
		Metadata:   meta,
	}); err != nil {
		return apperrors.Wrap(err, "failed to write vaccination claim release audit log")
	}
	return nil
}
