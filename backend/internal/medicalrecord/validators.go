package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// 共有カーネル昇格batch（BE9-2D ④後）: ①以来の internal/service との documented duplicate は
// sharedkernel へ一本化した。本ファイルは既存呼び出し面（package-private 名）互換の 1行 delegate
// と定数 alias のみ — ロジック・文言の正本は sharedkernel（複製ドリフト排除）。呼び出し側の
// sharedkernel 直参照への切替は各 slice の後続リファクタで行う。

// Error message constants（BUG-385）— 正本は sharedkernel。
const (
	errMsgAtLeastOneField = sharedkernel.ErrMsgAtLeastOneField
	errMsgIDsNotEmpty     = sharedkernel.ErrMsgIDsNotEmpty
	errMsgInputNotNil     = sharedkernel.ErrMsgInputNotNil
)

func validateRequiredName(name string) error {
	return sharedkernel.ValidateRequiredName(name)
}

func validateOptionalName(name *string) error {
	return sharedkernel.ValidateOptionalName(name)
}

func validateOwnedMasterFK(ctx context.Context, entity string, clinicID uint64, id *uint64,
	find func(ctx context.Context, clinicID, id uint64) error) error {
	return sharedkernel.ValidateOwnedMasterFK(ctx, entity, clinicID, id, find)
}

func setNullableUint64Field(fields map[string]any, col string, clearAssoc bool, id *uint64) { //nolint:unparam // col is the parent FK name; callers pass distinct constants that currently share "parent_id"
	sharedkernel.SetNullableUint64Field(fields, col, clearAssoc, id)
}
