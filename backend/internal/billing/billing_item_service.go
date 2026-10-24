// Package billing provides billing item use cases.
package billing

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// --- DB column constants ---

const (
	colBillingItemName                  = "name"
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
// （IsInsuranceApplicable・SortOrder は request が非 pointer のため「省略 = false / 0」の旧来の挙動を維持している）
type CreateBillingItemInput struct {
	ClinicID              uint64
	BillingID             uint64
	Category              string // Service 内で model.ItemCategory に変換
	Name                  string
	UnitPrice             model.ExactNumeric
	Quantity              model.ExactNumeric
	DiscountRate          model.ExactNumeric
	DiscountAmount        int64
	TaxType               string              // "" = デフォルト "excluded"
	TaxRate               *model.ExactNumeric // nil または 0 は旧来の互換でデフォルト 0.10（明示 0% は保存できない既存挙動）
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
	// EMR-229: 社販処理の目安として項目名に年月等を追記できるよう name 更新を許可する。
	Name                  *string
	UnitPrice             *model.ExactNumeric
	Quantity              *model.ExactNumeric
	DiscountRate          *model.ExactNumeric
	DiscountAmount        *int64
	TaxType               *model.TaxType
	TaxRate               *model.ExactNumeric
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
	if input.Name != nil {
		fields[colBillingItemName] = strings.TrimSpace(*input.Name)
	}
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

// billingSourceItemPendingMessage は請求元（provenance）を持つ明細の作成を拒否する固定エラーの文言（EMR-505 / Q7）。
const billingSourceItemPendingMessage = "診療・予防接種・検査・トリミング・入院に由来する請求明細の作成は、請求元の紐付け（billing_source_bindings）が未接続のため利用できません。手入力の明細は作成できます"

// rejectSourceBearingCreate は請求元を持つ明細の作成を、読み書きの前に固定エラー（422 finance_consumer_pending）で拒否する。
// 対象は treatment・vaccination・exam・appointment・trimming のいずれかの id、または manual 以外の source。
// 請求元の保存は billing_source_bindings へ移り（明細側の列は 001_init.sql:10891 で廃止）、その writer が未接続（EMR-481）のため、
// 廃止列へ書くことも、由来のない明細として黙って作ることもしない。resume: 481 の treatment subset が merge されたとき。
func rejectSourceBearingCreate(input *CreateBillingItemInput) error {
	hasProvenance := input.TreatmentID != nil ||
		input.VaccinationID != nil ||
		input.ExamID != nil ||
		input.AppointmentID != nil ||
		input.TrimmingCourseID != nil ||
		input.TrimmingOptionID != nil
	if hasProvenance || (input.Source != "" && input.Source != string(model.ItemSourceManual)) {
		return apperrors.WrapFinanceConsumerPending(billingSourceItemPendingMessage)
	}
	return nil
}

// exactInRange は n が有限で lower 以上 upper 以下かを返す。比較は exact で、丸めない。
func exactInRange(n model.ExactNumeric, lower, upper int64) bool {
	low, lowOK := n.Cmp(model.ExactNumericFromInt64(lower))
	high, highOK := n.Cmp(model.ExactNumericFromInt64(upper))
	return lowOK && highOK && low >= 0 && high <= 0
}

// validateBillingItemPrice は単価が有限の 0 以上であることを検査する。
func validateBillingItemPrice(unitPrice model.ExactNumeric) error {
	if sign, ok := unitPrice.Cmp(model.ExactNumericFromInt64(0)); !ok || sign < 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	return nil
}

// validateBillingItemQuantity は数量が有限の正の値であることを検査する。
func validateBillingItemQuantity(quantity model.ExactNumeric) error {
	if sign, ok := quantity.Cmp(model.ExactNumericFromInt64(0)); !ok || sign <= 0 {
		return apperrors.WrapInvalidInput("数量は正の値である必要があります")
	}
	return nil
}

// validateBillingItemDiscountRate は割引率（百分率）が有限の 0 以上 100 以下であることを検査する。
func validateBillingItemDiscountRate(rate model.ExactNumeric) error {
	if !exactInRange(rate, 0, 100) {
		return apperrors.WrapInvalidInput("割引率は0以上100以下で指定してください")
	}
	return nil
}

// validateBillingItemTaxRate は税率（割合）が有限の 0 以上 1 以下であることを検査する。
func validateBillingItemTaxRate(rate model.ExactNumeric) error {
	if !exactInRange(rate, 0, 1) {
		return apperrors.WrapInvalidInput("税率は0以上1以下で指定してください")
	}
	return nil
}

// validateUpdateBillingItemDecimals は update で指定された（非 nil の）10進フィールドを検査する。
func validateUpdateBillingItemDecimals(input *UpdateBillingItemInput) error {
	if input.UnitPrice != nil {
		if err := validateBillingItemPrice(*input.UnitPrice); err != nil {
			return err
		}
	}
	if input.Quantity != nil {
		if err := validateBillingItemQuantity(*input.Quantity); err != nil {
			return err
		}
	}
	if input.DiscountRate != nil {
		if err := validateBillingItemDiscountRate(*input.DiscountRate); err != nil {
			return err
		}
	}
	if input.TaxRate != nil {
		if err := validateBillingItemTaxRate(*input.TaxRate); err != nil {
			return err
		}
	}
	return nil
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
	if err := validateBillingItemPrice(input.UnitPrice); err != nil {
		return err
	}
	if err := validateBillingItemQuantity(input.Quantity); err != nil {
		return err
	}
	if err := validateBillingItemDiscountRate(input.DiscountRate); err != nil {
		return err
	}
	if input.TaxRate != nil {
		if err := validateBillingItemTaxRate(*input.TaxRate); err != nil {
			return err
		}
	}
	return rejectSourceBearingCreate(input)
}

func resolveBillingItemDefaults(input *CreateBillingItemInput) (model.TaxType, model.ExactNumeric, model.ItemSource, error) {
	// カテゴリバリデーション
	if err := validateItemCategory(input.Category); err != nil {
		return "", model.ExactNumeric{}, "", apperrors.Wrap(err, "failed to validate item category")
	}

	// TaxType デフォルト設定とバリデーション
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		if err := sharedkernel.ValidateTaxType(input.TaxType); err != nil {
			return "", model.ExactNumeric{}, "", apperrors.Wrap(err, "failed to validate tax type")
		}
		taxType = model.TaxType(input.TaxType)
	}

	// TaxRate: nil または 0 は 0.10 にする。旧来の互換（旧実装は `input.TaxRate > 0` でなければ既定 0.10）をそのまま
	// 維持しており、明示された 0% の税率は保存できない。業務上の誤りの可能性があるため仕様判断が出るまで変更しない。
	// 0.10 は float64 の sharedkernel.DefaultTaxRate ではなく exact で作る（EMR-505 フォローアップ）。
	taxRate := model.ExactNumericFromScaledInt64(10, 2)
	if input.TaxRate != nil && !input.TaxRate.IsZero() {
		taxRate = *input.TaxRate
	}

	// Source デフォルト設定とバリデーション
	source := model.ItemSourceManual
	if input.Source != "" {
		if err := validateItemSource(input.Source); err != nil {
			return "", model.ExactNumeric{}, "", apperrors.Wrap(err, "failed to validate item source")
		}
		source = model.ItemSource(input.Source)
	}

	return taxType, taxRate, source, nil
}

func applyBillingItemOtherMetadata(input *CreateBillingItemInput, item *model.BillingItem) error {
	item.OtherReason = nil
	item.CreatedBy = nil
	if item.Source == nil || *item.Source != model.ItemSourceManual ||
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
	// DeleteItemsForComplete は EMR-253 Complete takeover の ambient-tx collaborator。
	// takeover 対象 waiting billing の既存明細を一括 soft-delete する（request items で置換する前提）。
	DeleteItemsForComplete(ctx context.Context, clinicID, billingID uint64) error
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
	// AssertUnbilledForComplete は blocking warning に加えて、complete の
	// expected_unbilled_revision と現在の集約版が一致しなければ Conflict を返す（EMR-196②）。
	// expectedRevision は必須（空文字は必ず不一致になる）。
	AssertUnbilledForComplete(ctx context.Context, clinicID, petID uint64, expectedRevision string) error
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
// Revision は items+warnings の決定的フィンガープリント（EMR-196② の楽観ロック token）。
// OwnerDiscountRate は pet→主飼主マスタの割引率(%)（EMR-229: 手入力追加行の初期表示用）。
type UnbilledDetails struct {
	Items             []model.BillingItem
	Warnings          []UnbilledWarning
	Revision          string
	OwnerDiscountRate model.ExactNumeric
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
	// petOwnerFinder は EMR-229: 未請求候補集約で pet→主飼主を解決する（任意 DI・nil なら飼主割引 0）。
	petOwnerFinder billingPetOwnerFinder
	auditTx        billingAuditTxLogger // #115 / BUG-463: 締め後編集 fail-closed 監査
	// closeRepo は W-013 締め後明細変更の append-only adjustment 台帳用（任意 DI）。
	closeRepo CashRegisterCloseRepository
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

// WithBillingItemPetOwnerFinder は未請求候補集約での pet→飼主割引率解決に使う finder を配線する（EMR-229）。
func WithBillingItemPetOwnerFinder(finder billingPetOwnerFinder) billingItemServiceOption {
	return func(s *billingItemService) {
		s.petOwnerFinder = finder
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

// 請求の claim_state（billings.claim_state）。001_init.sql:9454 の
// CHECK (claim_state IN ('draft','finalized','cancelled')) に対応する。列は NULL を許す（同行）ため、
// NULL・未知の値は「状態不明」として扱い、明細の変更を許可しない。
const (
	billingClaimStateDraft     = "draft"
	billingClaimStateFinalized = "finalized"
	billingClaimStateCancelled = "cancelled"
)

// billingClaimState は請求の claim_state を返す。NULL または請求なしは空文字（状態不明）。
func billingClaimState(billing *model.Billing) string {
	if billing == nil || billing.ClaimState == nil {
		return ""
	}
	return *billing.ClaimState
}

// rejectIfBillingFinalized は draft 以外の請求への明細変更を Conflict で拒否する（BUG-463 / DeleteItem と同型）。
// finalized・cancelled は旧来の completed・cancelled と同じ文言。NULL・未知の claim_state は許可せず、状態不明として拒否する。
func rejectIfBillingFinalized(billing *model.Billing, action string) error {
	if billing == nil {
		return apperrors.WrapInternalServerError("locked billing is nil")
	}
	switch billingClaimState(billing) {
	case billingClaimStateDraft:
		return nil
	case billingClaimStateFinalized, billingClaimStateCancelled:
		return apperrors.WrapConflict(fmt.Sprintf("確定済みまたは取消済みの会計明細は%sできません", action))
	default:
		return apperrors.WrapConflict(fmt.Sprintf("会計の請求状態（claim_state）を確認できないため、明細を%sできません", action))
	}
}

// rejectItemUpdateUnlessDraft は UpdateItem 用。draft のときだけ更新を許可する。
// 確定済み（finalized）の明細訂正は、変更イベントと新しい算定版を作る correct の操作（EMR-483 / 505 / 506）にあたり、
// その writer が未接続のため、旧来の「理由付き訂正」（BUG-009）は固定エラー（422）で拒否する。
// cancelled と状態不明（NULL・未知）は Conflict。
func rejectItemUpdateUnlessDraft(billing *model.Billing, action string) error {
	if billing == nil {
		return apperrors.WrapInternalServerError("locked billing is nil")
	}
	switch billingClaimState(billing) {
	case billingClaimStateDraft:
		return nil
	case billingClaimStateFinalized:
		return apperrors.WrapFinanceConsumerPending(fmt.Sprintf("確定済み会計の明細%sは、訂正の変更イベントと新しい算定版（correct）が未接続のため利用できません", action))
	case billingClaimStateCancelled:
		return apperrors.WrapConflict(fmt.Sprintf("確定済みまたは取消済みの会計明細は%sできません", action))
	default:
		return apperrors.WrapConflict(fmt.Sprintf("会計の請求状態（claim_state）を確認できないため、明細を%sできません", action))
	}
}

func requirePostCloseReason(isPostClose bool, reason *string) error {
	if isPostClose && (reason == nil || *reason == "") {
		return apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}
	return nil
}

// createItemAmbientOpts は CreateItem / Complete 共有の ambient tx 内明細作成オプション。
type createItemAmbientOpts struct {
	recordPostClose bool // false のとき締め後監査は command 単位で Complete が記録する
}
