package inventory

import (
	"context"
	"encoding/json"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-449 A01: 物販マスタの全体割引除外 is_bill_discount_excluded の inventory ローカル部品。
// medicalrecord/bill_discount_setting.go（EMR-447 共通部品）の mirror であり、
// 保険対象外 is_non_insurance・税区分・明細割引とは別設定。
// 保険用 helper の「create 省略 = true」を転用せず、本設定の service 既定 false を明示する。
// 正本: docs/work/emr/emr-master-update-selected-contract-20261022.md A.2 / A.4 / A.5。
const (
	colIsBillDiscountExcluded = "is_bill_discount_excluded"

	errMsgBillDiscountNullNotAllowed = "is_bill_discount_excluded に null は指定できません"
	errMsgBillDiscountInvalid        = "is_bill_discount_excluded は true または false で指定してください"
)

// AuditEntry は inventory の consumer-side 監査入力（trimming/ports.go と同じ
// field-for-field mirror）。composition root が共有監査カーネル audit.Entry へ adapt する。
type AuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	Metadata   any
}

// AuditTxLogger は共有監査カーネル（tx 内 LogEntryTx）の consumer-side view。
// 監査依存の欠落は本設定の書込を fail-closed にする。
type AuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *AuditEntry) error
}

// billDiscountExcludedField は POST/PATCH DTO で「未送信」と「明示 true/false」を区別する。
// encoding/json の *bool だけでは null と未送信が合流するため set フラグで存在を保持する。
// null と非 boolean はどちらも 400 とし、同梱項目を含めて全体を書き込まない。
type billDiscountExcludedField struct {
	set   bool
	value bool
}

func (f *billDiscountExcludedField) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return apperrors.WrapInvalidInput(errMsgBillDiscountNullNotAllowed)
	}
	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput(errMsgBillDiscountInvalid)
	}
	f.set = true
	f.value = value
	return nil
}

// toCreateInput は POST の未送信を nil にする。service が false を明示保存する。
func (f billDiscountExcludedField) toCreateInput() *bool {
	return f.toUpdateInput()
}

// toUpdateInput は PATCH の未送信を nil（変更なし）にする。
func (f billDiscountExcludedField) toUpdateInput() *bool {
	if !f.set {
		return nil
	}
	value := f.value
	return &value
}

// createBillDiscountExcluded は新規登録で保存する値を返す。
// 未送信は false を明示保存し、DB default や Go のゼロ値に任せない。
func createBillDiscountExcluded(input *bool) *bool {
	value := false
	if input != nil {
		value = *input
	}
	return &value
}

// setBillDiscountExcludedField は明示値だけを UPDATE map へ入れる。
// false もゼロ値として読み飛ばさずに SET し、未送信なら既存の NULL/false/true を保持する。
func setBillDiscountExcludedField(fields map[string]any, value *bool) {
	if value == nil {
		return
	}
	fields[colIsBillDiscountExcluded] = *value
}

// billDiscountChange は1回の書込で記録する設定の前後値。
// BeforeExists=false は create（前の行なし）で、既存行の raw NULL（Before=nil）とは区別する。
type billDiscountChange struct {
	ResourceID   uint64
	BeforeExists bool
	Before       *bool
	After        *bool
}

func billDiscountAuditValue(setting *bool) map[string]any {
	var value any
	if setting != nil {
		value = *setting
	}
	return map[string]any{colIsBillDiscountExcluded: value}
}

// logBillDiscountAuditTx は設定の前後値を業務書込と同じ transaction の監査 sink へ記録する。
// 監査依存の欠落と監査失敗はエラーを返し、呼出し側の transaction 全体を rollback させる。
func logBillDiscountAuditTx(
	ctx context.Context, audit AuditTxLogger, clinicID uint64, actorID *uint64,
	resource, action string, change billDiscountChange,
) error {
	if audit == nil {
		return apperrors.WrapInternalServerError("bill discount exclusion audit dependency is required")
	}
	if actorID == nil || *actorID == 0 || clinicID == 0 || change.ResourceID == 0 {
		return apperrors.WrapInvalidInput("authenticated clinic, staff actor and resource are required for bill discount exclusion")
	}
	var before any
	if change.BeforeExists {
		before = billDiscountAuditValue(change.Before)
	}
	resourceID := change.ResourceID
	if err := audit.LogEntryTx(ctx, &AuditEntry{
		ClinicID: &clinicID, ActorID: actorID, ActorType: model.AuditActorTypeStaff,
		Action: action, Resource: resource, ResourceID: &resourceID,
		OldValue: before, NewValue: billDiscountAuditValue(change.After),
		Metadata: map[string]any{"field": colIsBillDiscountExcluded},
	}); err != nil {
		return apperrors.Wrap(err, "failed to write bill discount exclusion audit")
	}
	return nil
}
