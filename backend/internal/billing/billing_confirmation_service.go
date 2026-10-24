package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ConfirmBillingConfirmationInput は会計医師確認の入力DTO
type ConfirmBillingConfirmationInput struct {
	ConfirmedBy uint64
	Memo        string
}

// ReturnBillingConfirmationInput は会計差し戻しの入力DTO
type ReturnBillingConfirmationInput struct {
	ReturnedBy   uint64
	ReturnReason string
	Memo         string
}

// BillingConfirmationService は会計医師確認の入口。
// held for EMR-462: 旧実装はカルテ単位の billing_confirmations 行（pending / confirmed / returned）を読み書きしていたが、
// 置換先は請求ごとの確認セット（doctor_review_sets / doctor_review_records、臨床世代と digest に束縛）で、
// カルテ単位の行も returned 状態も無い。reader / writer が未接続のため、3 操作とも保存へ触れる前に固定エラー（422）で拒否する。
type BillingConfirmationService interface {
	Get(ctx context.Context, clinicID, medicalRecordID uint64) error
	Confirm(ctx context.Context, clinicID, medicalRecordID uint64, input *ConfirmBillingConfirmationInput) error
	Return(ctx context.Context, clinicID, medicalRecordID uint64, input *ReturnBillingConfirmationInput) error
}

type billingConfirmationService struct{}

// NewBillingConfirmationService は quarantine 中の BillingConfirmationService を返す。
// 引数は composition の呼出し形を保つために受けるが、保存へ触れないため使わない。held for EMR-462。
func NewBillingConfirmationService(_ BillingConfirmationRepository, _ billingMedicalRecordLocker, _ Transactor) BillingConfirmationService {
	return &billingConfirmationService{}
}

func (*billingConfirmationService) Get(_ context.Context, _, _ uint64) error {
	return apperrors.WrapFinanceConsumerPending("会計の医師確認の参照は、医師確認（請求ごとの確認セット）の読取が未接続のため利用できません")
}

func (*billingConfirmationService) Confirm(_ context.Context, _, _ uint64, _ *ConfirmBillingConfirmationInput) error {
	return apperrors.WrapFinanceConsumerPending("会計の医師確認は、医師確認（請求ごとの確認セット）の保存が未接続のため利用できません")
}

func (*billingConfirmationService) Return(_ context.Context, _, _ uint64, _ *ReturnBillingConfirmationInput) error {
	return apperrors.WrapFinanceConsumerPending("会計の差し戻しは、差し戻し状態を持たない新しい医師確認の仕様が未決定のため利用できません")
}
