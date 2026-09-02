package billing

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CompleteAccountingItemInput は complete command の明細1行入力。
// FK 検証は CreateItemForComplete（ValidateCreateReferences）が tx 内で行う。
// 価格の master 再解決は既存 CreateItem と同型（source-linked 価格の server 上書きは別 packet）。
type CompleteAccountingItemInput struct {
	Category              string
	Name                  string
	UnitPrice             int64
	Quantity              float64
	DiscountRate          float64
	DiscountAmount        int64
	TaxType               string
	TaxRate               float64
	IsInsuranceApplicable bool
	Source                string
	OtherReason           *string
	MerchandiseItemID     *uint64
	TreatmentID           *uint64
	VaccinationID         *uint64
	ExamID                *uint64
	AppointmentID         *uint64
	TrimmingCourseID      *uint64
	TrimmingOptionID      *uint64
	SortOrder             int
}

// CompleteAccountingInput は POST /accountings/complete のサービス入力。
// clinic_id / actor は handler が認証 context から注入する。client total は受け取らない。
type CompleteAccountingInput struct {
	ClinicID          uint64
	StaffID           *uint64
	IdempotencyKey    string
	MedicalRecordID   *uint64
	HospitalizationID *uint64
	OwnerID           *uint64
	PetID             *uint64
	ScheduledDate     time.Time
	Memo              string
	HasInsurance      bool
	InsuranceRatio    *float64
	InsuranceName     *string
	InsuranceAmount   *int64
	DiscountAmount    *int64
	Items             []CompleteAccountingItemInput
	PaymentSplits     []PaymentSplitInput
	PostCloseReason   *string
	// IsPostClose は handler の候補 read（write 時に resolvePostCloseInTx で再評価）。
	IsPostClose bool
}

// CompleteAccountingResult は complete の結果。Created=true は 201、false は idempotent replay 200。
type CompleteAccountingResult struct {
	Accounting *model.Billing
	Created    bool
}

// completeItemWriter は ambient tx 内で明細を作成する collaborator（WithTx を開始しない）。
type completeItemWriter interface {
	CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
}

// completeTotalsWriter は ambient tx 内で totals を再計算して billings に書く collaborator。
type completeTotalsWriter interface {
	RecalculateTotalsForComplete(ctx context.Context, clinicID, billingID uint64) (subtotal, taxTotal, totalAmount int64, err error)
}

// WithCompleteItemWriter は BUG-018 complete 用の明細 writer を配線する。
func WithCompleteItemWriter(w completeItemWriter) accountingServiceOption {
	return func(s *accountingService) {
		s.itemWriter = w
	}
}

// WithCompleteTotalsWriter は BUG-018 complete 用の totals writer を配線する。
func WithCompleteTotalsWriter(w completeTotalsWriter) accountingServiceOption {
	return func(s *accountingService) {
		s.totalsWriter = w
	}
}

// ValidateIdempotencyKeyUUID は Idempotency-Key が UUID であることを検証する。
func ValidateIdempotencyKeyUUID(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return apperrors.WrapInvalidInput("Idempotency-Key header is required")
	}
	if _, err := uuid.Parse(key); err != nil {
		return apperrors.WrapInvalidInput("Idempotency-Key must be a valid UUID")
	}
	return nil
}

// resolveCompleteMedicalRecordID は BUG-011: complete ヘッダの medical_record_id を確定する。
// - 明示値がある場合は treatment 由来の親カルテと一致することを検証する
// - 未指定でも treatment 付き明細があれば treatments.medical_record_id から一意に解決する
// - treatment が複数カルテにまたがる場合は参照組み合わせ不正
// - treatment が無い場合は tx 不要（明示値をそのまま返す）
func resolveCompleteMedicalRecordID(
	ctx context.Context,
	clinicID uint64,
	explicit *uint64,
	items []CompleteAccountingItemInput,
) (*uint64, error) {
	hasTreatment := false
	for i := range items {
		if items[i].TreatmentID != nil {
			hasTreatment = true
			break
		}
	}
	if !hasTreatment {
		return explicit, nil
	}

	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError("complete medical_record resolution requires an active transaction")
	}
	tx = tx.WithContext(ctx)

	var resolvedFromTreatments *uint64
	for i := range items {
		if items[i].TreatmentID == nil {
			continue
		}
		var treatmentRef struct {
			MedicalRecordID uint64
		}
		if err := tx.
			Table("treatments").
			Select("treatments.medical_record_id").
			Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
			Where("treatments.id = ? AND treatments.deleted_at IS NULL", *items[i].TreatmentID).
			Take(&treatmentRef).Error; err != nil {
			return nil, apperrors.FromGORM(err, "treatment", fmt.Sprintf("%d", *items[i].TreatmentID))
		}
		if resolvedFromTreatments == nil {
			id := treatmentRef.MedicalRecordID
			resolvedFromTreatments = &id
			continue
		}
		if *resolvedFromTreatments != treatmentRef.MedicalRecordID {
			return nil, invalidBillingItemReferenceCombination()
		}
	}

	if explicit != nil {
		if resolvedFromTreatments != nil && *explicit != *resolvedFromTreatments {
			return nil, invalidBillingItemReferenceCombination()
		}
		return explicit, nil
	}
	return resolvedFromTreatments, nil
}

// Complete は header/items/totals/payments/splits/reservation/audit を単一 tx で全 commit または全 rollback する（BUG-018）。
func (s *accountingService) Complete(ctx context.Context, input *CompleteAccountingInput) (*CompleteAccountingResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("complete input is required")
	}
	if err := ValidateIdempotencyKeyUUID(input.IdempotencyKey); err != nil {
		return nil, err
	}
	if input.ScheduledDate.IsZero() {
		return nil, apperrors.WrapInvalidInput("scheduled_date is required")
	}
	if len(input.Items) == 0 {
		return nil, apperrors.WrapInvalidInput("items are required")
	}
	if s.itemWriter == nil {
		return nil, apperrors.WrapInternalServerError("complete item writer is not configured")
	}
	if s.totalsWriter == nil {
		return nil, apperrors.WrapInternalServerError("complete totals writer is not configured")
	}

	digest, err := ComputeCompleteAccountingDigest(input)
	if err != nil {
		return nil, err
	}

	// Pre-tx idempotent lookup（同一 key の replay / 異 digest 409）。
	// soft-deleted 行も含むため key 再利用不可。
	if existing, err := s.repo.FindByCompletionRequestID(ctx, input.ClinicID, input.IdempotencyKey); err != nil {
		return nil, apperrors.Wrap(err, "failed to lookup completion request")
	} else if existing != nil {
		return s.resolveIdempotentReplay(ctx, input.ClinicID, existing, digest)
	}

	// Payment method master 解決は tx 外の低コスト読取。
	systemKeyToID, err := s.loadPaymentMethodSystemKeyToID(ctx, input.ClinicID)
	if err != nil {
		return nil, err
	}

	var result *CompleteAccountingResult
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		completeResult, err := s.completeInTx(txCtx, input, digest, systemKeyToID)
		if err != nil {
			return err
		}
		result = completeResult
		return nil
	}); err != nil {

		return nil, apperrors.Wrap(err, "failed to complete accounting in transaction")
	}

	if result != nil && result.Created && result.Accounting != nil {
		s.syncCPMStageTag(ctx, input.ClinicID, result.Accounting)
		slog.InfoContext(ctx, "accounting completed atomically",
			slog.Uint64("billing_id", result.Accounting.ID),
			slog.Uint64("clinic_id", input.ClinicID))
	}
	return result, nil
}

func (s *accountingService) resolveIdempotentReplay(ctx context.Context, clinicID uint64, existing *model.Billing, digest string) (*CompleteAccountingResult, error) {
	if existing.DeletedAt.Valid {
		return nil, apperrors.WrapConflict("Idempotency-Key was already used by a deleted accounting and cannot be reused")
	}
	storedHash := ""
	if existing.CompletionRequestHash != nil {
		storedHash = *existing.CompletionRequestHash
	}
	if storedHash != digest {
		return nil, apperrors.WrapConflict("Idempotency-Key was reused with a different request payload")
	}
	// Replay: return current accounting without re-writing payment/audit.
	// Prefer fully preloaded FindByID when not soft-deleted.
	reloaded, err := s.repo.FindByID(ctx, clinicID, existing.ID)
	if err != nil {
		// Fall back to existing snapshot if reload fails for any non-notfound reason after cancel race.
		return nil, apperrors.Wrap(err, "failed to reload accounting for idempotent replay")
	}
	return &CompleteAccountingResult{Accounting: reloaded, Created: false}, nil
}
