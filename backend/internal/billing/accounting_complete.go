package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// ComputeCompleteAccountingDigest は complete request の正規化 digest（hex）を返す。
// 同一 payload は常に同一 digest になる決定的方式（JSON marshal of normalized struct）。
func ComputeCompleteAccountingDigest(input *CompleteAccountingInput) (string, error) {
	if input == nil {
		return "", apperrors.WrapInvalidInput("complete input is required")
	}
	type digestItem struct {
		Category              string  `json:"category"`
		Name                  string  `json:"name"`
		UnitPrice             int64   `json:"unit_price"`
		Quantity              float64 `json:"quantity"`
		DiscountRate          float64 `json:"discount_rate"`
		DiscountAmount        int64   `json:"discount_amount"`
		TaxType               string  `json:"tax_type"`
		TaxRate               float64 `json:"tax_rate"`
		IsInsuranceApplicable bool    `json:"is_insurance_applicable"`
		Source                string  `json:"source"`
		OtherReason           string  `json:"other_reason"`
		MerchandiseItemID     uint64  `json:"merchandise_item_id"`
		TreatmentID           uint64  `json:"treatment_id"`
		VaccinationID         uint64  `json:"vaccination_id"`
		AppointmentID         uint64  `json:"appointment_id"`
		TrimmingCourseID      uint64  `json:"trimming_course_id"`
		TrimmingOptionID      uint64  `json:"trimming_option_id"`
		SortOrder             int     `json:"sort_order"`
	}
	type digestSplit struct {
		Method         string `json:"method"`
		Amount         int64  `json:"amount"`
		ReceivedAmount int64  `json:"received_amount"`
		ChangeAmount   int64  `json:"change_amount"`
		ChangeOverride bool   `json:"change_override"`
	}
	type digestRoot struct {
		MedicalRecordID   uint64        `json:"medical_record_id"`
		HospitalizationID uint64        `json:"hospitalization_id"`
		OwnerID           uint64        `json:"owner_id"`
		PetID             uint64        `json:"pet_id"`
		ScheduledDate     string        `json:"scheduled_date"`
		Memo              string        `json:"memo"`
		HasInsurance      bool          `json:"has_insurance"`
		InsuranceRatio    float64       `json:"insurance_ratio"`
		InsuranceName     string        `json:"insurance_name"`
		InsuranceAmount   int64         `json:"insurance_amount"`
		DiscountAmount    int64         `json:"discount_amount"`
		PostCloseReason   string        `json:"post_close_reason"`
		Items             []digestItem  `json:"items"`
		PaymentSplits     []digestSplit `json:"payment_splits"`
	}
	root := digestRoot{
		Memo:          strings.TrimSpace(input.Memo),
		HasInsurance:  input.HasInsurance,
		ScheduledDate: input.ScheduledDate.UTC().Format(time.RFC3339),
		Items:         make([]digestItem, 0, len(input.Items)),
		PaymentSplits: make([]digestSplit, 0, len(input.PaymentSplits)),
	}
	if input.MedicalRecordID != nil {
		root.MedicalRecordID = *input.MedicalRecordID
	}
	if input.HospitalizationID != nil {
		root.HospitalizationID = *input.HospitalizationID
	}
	if input.OwnerID != nil {
		root.OwnerID = *input.OwnerID
	}
	if input.PetID != nil {
		root.PetID = *input.PetID
	}
	if input.InsuranceRatio != nil {
		root.InsuranceRatio = *input.InsuranceRatio
	}
	if input.InsuranceName != nil {
		root.InsuranceName = *input.InsuranceName
	}
	if input.InsuranceAmount != nil {
		root.InsuranceAmount = *input.InsuranceAmount
	}
	if input.DiscountAmount != nil {
		root.DiscountAmount = *input.DiscountAmount
	}
	if input.PostCloseReason != nil {
		root.PostCloseReason = strings.TrimSpace(*input.PostCloseReason)
	}
	for _, it := range input.Items {
		di := digestItem{
			Category:              it.Category,
			Name:                  it.Name,
			UnitPrice:             it.UnitPrice,
			Quantity:              it.Quantity,
			DiscountRate:          it.DiscountRate,
			DiscountAmount:        it.DiscountAmount,
			TaxType:               it.TaxType,
			TaxRate:               it.TaxRate,
			IsInsuranceApplicable: it.IsInsuranceApplicable,
			Source:                it.Source,
			SortOrder:             it.SortOrder,
		}
		if it.OtherReason != nil {
			di.OtherReason = *it.OtherReason
		}
		if it.MerchandiseItemID != nil {
			di.MerchandiseItemID = *it.MerchandiseItemID
		}
		if it.TreatmentID != nil {
			di.TreatmentID = *it.TreatmentID
		}
		if it.VaccinationID != nil {
			di.VaccinationID = *it.VaccinationID
		}
		if it.AppointmentID != nil {
			di.AppointmentID = *it.AppointmentID
		}
		if it.TrimmingCourseID != nil {
			di.TrimmingCourseID = *it.TrimmingCourseID
		}
		if it.TrimmingOptionID != nil {
			di.TrimmingOptionID = *it.TrimmingOptionID
		}
		root.Items = append(root.Items, di)
	}
	for _, sp := range input.PaymentSplits {
		root.PaymentSplits = append(root.PaymentSplits, digestSplit{
			Method:         string(sp.Method),
			Amount:         sp.Amount,
			ReceivedAmount: sp.ReceivedAmount,
			ChangeAmount:   sp.ChangeAmount,
			ChangeOverride: sp.ChangeOverride,
		})
	}
	raw, err := json.Marshal(root)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to marshal complete digest payload")
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
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
		// 並行初回: UNIQUE 競合を避けるため tx 内でも再 lookup。
		if existing, err := s.repo.FindByCompletionRequestID(txCtx, input.ClinicID, input.IdempotencyKey); err != nil {
			return apperrors.Wrap(err, "failed to re-lookup completion request in tx")
		} else if existing != nil {
			replay, err := s.resolveIdempotentReplay(txCtx, input.ClinicID, existing, digest)
			if err != nil {
				return err
			}
			result = replay
			return nil
		}

		// BUG-011: treatment 付き明細があるとき billing.medical_record_id が必須。
		// FE が未送信でも treatment から一意に解決する（明示値は優先・不一致は拒否）。
		medicalRecordID, err := resolveCompleteMedicalRecordID(txCtx, input.ClinicID, input.MedicalRecordID, input.Items)
		if err != nil {
			return err
		}

		if err := s.validateAccountingRelatedFKs(
			txCtx, input.ClinicID,
			medicalRecordID, input.HospitalizationID, input.OwnerID, input.PetID,
		); err != nil {
			return err
		}

		// BUG-013: blocking unbilled を同一 tx 内で再検証（TOCTOU 解消）。
		if s.unbilledGuard != nil && input.PetID != nil {
			if err := s.unbilledGuard.AssertNoBlockingUnbilled(txCtx, input.ClinicID, *input.PetID); err != nil {
				return err
			}
		}

		// Close 境界の write-time 再評価。
		postClose, err := s.resolvePostCloseInTx(txCtx, input.ClinicID, input.ScheduledDate, input.IsPostClose)
		if err != nil {
			return err
		}
		if postClose {
			if input.PostCloseReason == nil || strings.TrimSpace(*input.PostCloseReason) == "" {
				return apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
			}
			input.IsPostClose = true
		}

		reqID := input.IdempotencyKey
		hash := digest
		billing := &model.Billing{
			ClinicID:              input.ClinicID,
			MedicalRecordID:       medicalRecordID,
			HospitalizationID:     input.HospitalizationID,
			OwnerID:               input.OwnerID,
			PetID:                 input.PetID,
			Subtotal:              0,
			TaxTotal:              0,
			TotalAmount:           0,
			HasInsurance:          input.HasInsurance,
			Status:                model.BillingStatusWaiting,
			ScheduledDate:         input.ScheduledDate,
			Memo:                  input.Memo,
			CompletionRequestID:   &reqID,
			CompletionRequestHash: &hash,
		}
		if err := s.repo.Create(txCtx, input.ClinicID, billing); err != nil {
			// Create は UNIQUE を AlreadyExists に変換する（pg 23505 は chain されない）。
			// completion_request_id 衝突時のみ replay。他 UNIQUE は existing==nil のまま元エラー。
			if apperrors.IsAlreadyExists(err) || persistence.IsUniqueConstraintErr(err) {
				existing, lookupErr := s.repo.FindByCompletionRequestID(txCtx, input.ClinicID, input.IdempotencyKey)
				if lookupErr != nil {
					return apperrors.Wrap(lookupErr, "failed to resolve completion unique conflict")
				}
				if existing != nil {
					replay, replayErr := s.resolveIdempotentReplay(txCtx, input.ClinicID, existing, digest)
					if replayErr != nil {
						return replayErr
					}
					result = replay
					return nil
				}
			}
			return apperrors.Wrap(err, "failed to create accounting header for complete")
		}

		// Items: ambient tx 参加。N 番目失敗で全 rollback。
		for i := range input.Items {
			it := input.Items[i]
			itemInput := &CreateBillingItemInput{
				ClinicID:              input.ClinicID,
				BillingID:             billing.ID,
				Category:              it.Category,
				Name:                  it.Name,
				UnitPrice:             it.UnitPrice,
				Quantity:              it.Quantity,
				DiscountRate:          it.DiscountRate,
				DiscountAmount:        it.DiscountAmount,
				TaxType:               it.TaxType,
				TaxRate:               it.TaxRate,
				IsInsuranceApplicable: it.IsInsuranceApplicable,
				Source:                it.Source,
				OtherReason:           it.OtherReason,
				MerchandiseItemID:     it.MerchandiseItemID,
				TreatmentID:           it.TreatmentID,
				VaccinationID:         it.VaccinationID,
				AppointmentID:         it.AppointmentID,
				TrimmingCourseID:      it.TrimmingCourseID,
				TrimmingOptionID:      it.TrimmingOptionID,
				SortOrder:             it.SortOrder,
				StaffID:               input.StaffID,
				// 手入力 other は CreatedBy 必須（legacy item handler が OptionalStaffID をセットするのと同型）。
				CreatedBy: input.StaffID,
			}
			if _, err := s.itemWriter.CreateItemForComplete(txCtx, itemInput); err != nil {
				return apperrors.Wrap(err, fmt.Sprintf("failed to create complete item index=%d", i))
			}
		}

		subtotal, taxTotal, totalAmount, err := s.totalsWriter.RecalculateTotalsForComplete(txCtx, input.ClinicID, billing.ID)
		if err != nil {
			return apperrors.Wrap(err, "failed to recalculate complete totals")
		}

		// Payments / splits: server totals を使う（client total を信頼しない）。
		insuranceAmount := int64(0)
		if input.InsuranceAmount != nil {
			insuranceAmount = *input.InsuranceAmount
		}
		discountAmount := int64(0)
		if input.DiscountAmount != nil {
			discountAmount = *input.DiscountAmount
		}
		billingAmount := totalAmount - insuranceAmount - discountAmount
		if billingAmount < 0 {
			return apperrors.WrapInvalidInput("請求金額が負になります（保険・割引の指定を確認してください）")
		}
		if err := validatePaymentSplits(input.PaymentSplits, &billingAmount); err != nil {
			return err
		}

		updateInput := &UpdateAccountingInput{
			ID:              billing.ID,
			ClinicID:        input.ClinicID,
			StaffID:         input.StaffID,
			Subtotal:        &subtotal,
			TaxTotal:        &taxTotal,
			TotalAmount:     &totalAmount,
			InsuranceRatio:  input.InsuranceRatio,
			InsuranceName:   input.InsuranceName,
			InsuranceAmount: &insuranceAmount,
			DiscountAmount:  &discountAmount,
			BillingAmount:   &billingAmount,
			PaymentSplits:   input.PaymentSplits,
			PostCloseReason: input.PostCloseReason,
			IsPostClose:     input.IsPostClose,
		}
		if len(input.PaymentSplits) > 0 {
			method := representativeMethod(input.PaymentSplits)
			updateInput.PaymentMethod = &method
		}
		payment := buildPaymentFromInput(updateInput)
		if payment.Method != "" {
			pid, err := resolvePaymentMethodMasterID(payment.Method, payment.PaymentMethodID, systemKeyToID)
			if err != nil {
				return err
			}
			payment.PaymentMethodID = pid
		}
		splits := buildPaymentSplits(updateInput)
		for i := range splits {
			pid, err := resolvePaymentMethodMasterID(splits[i].Method, splits[i].PaymentMethodID, systemKeyToID)
			if err != nil {
				return err
			}
			splits[i].PaymentMethodID = pid
		}
		if err := s.repo.SavePayment(txCtx, payment); err != nil {
			return apperrors.Wrap(err, "failed to save payment for complete")
		}
		if err := s.repo.SavePaymentSplits(txCtx, splits); err != nil {
			return apperrors.Wrap(err, "failed to save payment splits for complete")
		}

		now := time.Now()
		completedStatus := model.BillingStatusCompleted
		updated, err := s.repo.Update(txCtx, input.ClinicID, billing.ID, map[string]any{
			"status":       completedStatus,
			"completed_at": now,
			"subtotal":     subtotal,
			"tax_total":    taxTotal,
			"total_amount": totalAmount,
		})
		if err != nil {
			return apperrors.Wrap(err, "failed to mark accounting completed")
		}

		if err := s.completeAccountingAppointments(txCtx, input.ClinicID, updated); err != nil {
			return apperrors.Wrap(err, "failed to complete accounting appointments during complete")
		}

		if input.IsPostClose {
			// New billing: existing total was 0; delta = final total.
			adjInput := &UpdateAccountingInput{
				ID:              billing.ID,
				ClinicID:        input.ClinicID,
				StaffID:         input.StaffID,
				TotalAmount:     &totalAmount,
				PostCloseReason: input.PostCloseReason,
				IsPostClose:     true,
			}
			// writePostCloseAdjustment uses existing.TotalAmount for delta; seed 0 for create path.
			existingForAdj := &model.Billing{ID: billing.ID, TotalAmount: 0, ScheduledDate: input.ScheduledDate}
			if err := s.writePostCloseAdjustment(txCtx, adjInput, existingForAdj); err != nil {
				return err
			}
			if err := s.logPostCloseEdit(txCtx, adjInput); err != nil {
				return err
			}
		}

		// Reload before commit so response failure rolls back the whole complete.
		reloaded, err := s.repo.FindByID(txCtx, input.ClinicID, billing.ID)
		if err != nil {
			return apperrors.Wrap(err, "failed to reload accounting after complete")
		}
		result = &CompleteAccountingResult{Accounting: reloaded, Created: true}
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
