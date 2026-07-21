package sharedkernel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// DefaultTaxRate はデフォルト税率（外税10%・billing/discharge 系の共有定数）。
const DefaultTaxRate = 0.10

// MasterNameMaxLength は各マスタ name カラム（VARCHAR(255)）に合わせた上限（BUG-379）。
const MasterNameMaxLength = 255

// 共有エラーメッセージ定数（BUG-385）。service/medicalrecord の両方が同文言を使う subset のみ
// ここに置く（片側でしか使わない文言は各 package に残す）。
const (
	ErrMsgAtLeastOneField  = "少なくとも1つのフィールドを指定してください"
	ErrMsgIDsNotEmpty      = "並び順のIDリストが空です"
	ErrMsgInputNotNil      = "更新内容が指定されていません"
	ErrMsgPriceZeroOrMore  = "金額は0以上を入力してください"
	ErrMsgQuantityPositive = "数量は0より大きい値を入力してください"
)

// ValidateRequiredName は必須名称の共通バリデーション（空白 trim・長さ上限・制御文字拒否）。
func ValidateRequiredName(name string) error {
	if strings.TrimSpace(name) == "" {
		return apperrors.WrapInvalidInput("名前を入力してください")
	}
	if utf8.RuneCountInString(name) > MasterNameMaxLength {
		return apperrors.WrapInvalidInput(fmt.Sprintf("名前は%d文字以内で入力してください", MasterNameMaxLength))
	}
	for _, r := range name {
		if r == '\u0000' {
			return apperrors.WrapInvalidInput("名前に無効な文字が含まれています")
		}
		if r < 0x20 && r != '\t' && r != '\n' {
			return apperrors.WrapInvalidInput("名前に無効な文字が含まれています")
		}
	}
	return nil
}

// ValidateOptionalName は nil 許容の名称バリデーション。
// PATCH 系で nil の場合は更新しない意味なのでスキップ、非 nil の場合のみ検証する。
func ValidateOptionalName(name *string) error {
	if name == nil {
		return nil
	}
	return ValidateRequiredName(*name)
}

// ValidateOwnedMasterFK は request 由来の clinic-scoped master FK の所有権を検証する
// (R24: マスタFKガードの複数イディオムを共有ヘルパーへ統一)。
// find は対象マスタの FindByID を包んだアダプタ。nil の FK はスキップ（optional FK 規約）。
func ValidateOwnedMasterFK(ctx context.Context, entity string, clinicID uint64, id *uint64,
	find func(ctx context.Context, clinicID, id uint64) error) error {
	if id == nil {
		return nil
	}
	if err := find(ctx, clinicID, *id); err != nil {
		slog.ErrorContext(ctx, "failed to verify "+entity+" ownership",
			"error", err, "id", *id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to verify "+entity+" ownership")
	}
	return nil
}

// ValidateOwnedMasterFKs は複数 FK の一括所有権検証（ValidateOwnedMasterFK のループ）。
func ValidateOwnedMasterFKs(ctx context.Context, entity string, clinicID uint64, ids []uint64,
	find func(ctx context.Context, clinicID, id uint64) error) error {
	for i := range ids {
		if err := ValidateOwnedMasterFK(ctx, entity, clinicID, &ids[i], find); err != nil {
			return err
		}
	}
	return nil
}

// SetNullableUint64Field は PATCH 系 update fields 構築の nullable FK 共通処理。
// clearAssoc は関連の明示クリア（fields[col] = nil）、非 nil id は設定、いずれでもなければ
// フィールドに触れない（省略 = 変更なしの PATCH 規約）。
func SetNullableUint64Field(fields map[string]any, col string, clearAssoc bool, id *uint64) {
	switch {
	case clearAssoc:
		fields[col] = nil
	case id != nil:
		fields[col] = *id
	}
}

// ValidateNonNegativePrice は金額（nil 許容）の非負検証。
func ValidateNonNegativePrice(price *int64) error {
	if price == nil {
		return nil
	}
	if *price < 0 {
		return apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	return nil
}

// ValidateDiscountRate は割引率（0〜100）の範囲検証。
func ValidateDiscountRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
	}
	return nil
}

// ValidateTimeRange は end_time > start_time を確認する共通バリデーション
// （BE9-2C R①: service/reservation_service.go から昇格。reservation/trimming の恒久ドメイン跨ぎ）。
func ValidateTimeRange(startTime, endTime time.Time) error {
	if !endTime.After(startTime) {
		return apperrors.WrapInvalidInput("end_time must be after start_time")
	}
	return nil
}
