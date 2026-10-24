package billing

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PeriodUnpaidOwnerPet は飼主+ペット単位の期間未納繰越集約結果。EMR-188
// LatestScheduled はそのグループの未納会計の MAX(scheduled_date)（YYYY-MM-DD）。EMR-189
type PeriodUnpaidOwnerPet struct {
	OwnerID             uint64  `json:"owner_id"`
	OwnerName           string  `json:"owner_name"`
	PetID               *uint64 `json:"pet_id,omitempty"`
	PetName             string  `json:"pet_name"`
	PrevPeriodCarryover int64   `json:"prev_period_carryover"`
	CurrentPeriodUnpaid int64   `json:"current_period_unpaid"`
	PeriodEndCarryover  int64   `json:"period_end_carryover"`
	LatestScheduled     string  `json:"latest_scheduled"`
}

// PeriodUnpaidSummary は期間未納繰越のサマリー情報。EMR-188
type PeriodUnpaidSummary struct {
	PrevPeriodCarryover int64 `json:"prev_period_carryover"`
	CurrentPeriodUnpaid int64 `json:"current_period_unpaid"`
	PeriodEndCarryover  int64 `json:"period_end_carryover"`
}

// UnpaidOwnerAggregate は飼主単位の未納集約結果
// BUG-370
type UnpaidOwnerAggregate struct {
	OwnerID         uint64 `json:"owner_id"`
	OwnerName       string `json:"owner_name"`
	Count           int64  `json:"count"`
	TotalAmount     int64  `json:"total_amount"`
	OldestScheduled string `json:"oldest_scheduled"`
	LatestScheduled string `json:"latest_scheduled"`
}

// UnpaidSummary は未納者一覧のサマリー情報（売掛金総額）
// BUG-370
type UnpaidSummary struct {
	TotalAmount  int64 `json:"total_amount"`
	BillingCount int64 `json:"billing_count"`
	OwnerCount   int64 `json:"owner_count"`
}

// OwnerUnpaidBalance は飼主単位の未納残高（未入金 billing の合計と件数）。#182
type OwnerUnpaidBalance struct {
	TotalAmount int64 `json:"total_amount"`
	Count       int64 `json:"count"`
}

// validBillingOwnerPetScope excludes a billing whose optional owner or pet
// relation crosses a clinic boundary. NULL owner/pet values remain valid for
// direct billings.
// DEC-27: pets.owner_id is the current owner; billings.owner_id is a snapshot
// at billing time. Do not require equality after pet transfer.
func validBillingOwnerPetScope(db *gorm.DB) *gorm.DB {
	return db.Where(`
		(
			billings.owner_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM owners AS billing_owner
				WHERE billing_owner.id = billings.owner_id
				  AND billing_owner.clinic_id = billings.clinic_id
				  AND billing_owner.deleted_at IS NULL
			)
		)
		AND (
			billings.pet_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM pets AS billing_pet
				WHERE billing_pet.id = billings.pet_id
				  AND billing_pet.clinic_id = billings.clinic_id
				  AND billing_pet.deleted_at IS NULL
			)
		)
	`)
}

// 未納の集計（会計単位・飼主単位・飼主+ペット単位の期間繰越・飼主別の残高）は、held for EMR-506。
// 未収残高を保存する列がなく（billings.status・total_amount・completed_at と payments 系の表は廃止。001_init.sql:10889,10918-10920）、
// 「入金済み」は受領 (payment_funds) と配賦 (fund_consumption_events) から導く定義に変わる
// （docs/spec/domains/accounting/lifecycle.md の FR-accounting-receivable-partial-zero）。
// 請求ごとの最新の算定額 - 配賦済み額を読む reader と、飼主・ペットの氏名 reader（EMR-491）が未接続のため、
// 旧式の CASE 式で 0 や全額を返さず、固定エラー（422 finance_consumer_pending）で拒否する。
// resume: 算定固定の reader と受領・配賦の reader が接続されたとき。
const unpaidPendingMessage = "未納の一覧・残高・期間繰越は、請求ごとの最新の算定額と受領・配賦の読取が未接続のため利用できません"

// FindUnpaidByBilling は未収残高 > 0 の会計を会計単位で返していた。#120 / BUG-007
func (*accountingRepository) FindUnpaidByBilling(_ context.Context, _ uint64, _, _ string, _, _ int) ([]model.Billing, int64, error) {
	return nil, 0, apperrors.WrapFinanceConsumerPending(unpaidPendingMessage)
}

// FindUnpaidByOwner は未収を飼主単位で集約していた。#120 / BUG-007
func (*accountingRepository) FindUnpaidByOwner(_ context.Context, _ uint64, _, _ string, _, _ int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	return nil, 0, UnpaidSummary{}, apperrors.WrapFinanceConsumerPending(unpaidPendingMessage)
}

// SumUnpaidByOwner は指定飼主の未納残高（未収残高合計と件数）を返していた。#182 / BUG-007
func (*accountingRepository) SumUnpaidByOwner(_ context.Context, _, _ uint64) (OwnerUnpaidBalance, error) {
	return OwnerUnpaidBalance{}, apperrors.WrapFinanceConsumerPending(unpaidPendingMessage)
}

// FindPeriodUnpaidCarryover は指定期間の未納繰越（期間前繰越・期間内未納・期末繰越）を飼主+ペット単位で返していた。EMR-188 / BUG-007
func (*accountingRepository) FindPeriodUnpaidCarryover(_ context.Context, _ uint64, _, _ string, _, _ int) ([]PeriodUnpaidOwnerPet, int64, PeriodUnpaidSummary, error) {
	return nil, 0, PeriodUnpaidSummary{}, apperrors.WrapFinanceConsumerPending(unpaidPendingMessage)
}
