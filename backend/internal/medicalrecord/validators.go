package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// This file is a deliberate, documented duplication — not a copy/paste accident. Its source
// of truth is internal/service's shared validation kernel (validators.go, validators_name.go,
// update_fields.go — boundary map §4.2, "維持": fan-in 9 domains, kept in place because
// promoting it to its own package is out of this batch's scope). ADR-006's 5-point BE9-2C
// pattern (see BE-refactor.md "BE9-2B実績からの申し送り") forbids medicalrecord from
// importing internal/service — that constraint exists for the *aggregator* (service.Services)
// and its cross-cutting *stateful* dependencies (Audit, Permission), which get a
// consumer-side interface + main.go adapter. These functions are different: pure,
// stateless, no I/O — declaring an interface for them would be speculative generality
// (YAGNI) for zero benefit. Duplicating ~60 lines of pure validation logic here (matching
// internal/service's byte-for-byte, so behavior is provably identical) is the pragmatic move.
// Every future BE9-2C/2D medicalrecord batch reuses this same local copy — no repeated
// duplication. Follow-up: promote the shared kernel to its own package (same treatment
// boundary map §4.3 recommends for the audit kernel) once a second migrated domain needs it;
// then both internal/service's copy and this one collapse onto the new package.

// masterNameMaxLength mirrors internal/service.MasterNameMaxLength (BUG-379).
const masterNameMaxLength = 255

// Error message constants mirroring internal/service's ErrMsg* (BUG-385) — only the subset
// this slice's services actually use.
const (
	errMsgAtLeastOneField = "少なくとも1つのフィールドを指定してください"
	errMsgIDsNotEmpty     = "並び順のIDリストが空です"
	errMsgInputNotNil     = "更新内容が指定されていません"
)

func validateRequiredName(name string) error {
	if strings.TrimSpace(name) == "" {
		return apperrors.WrapInvalidInput("名前を入力してください")
	}
	if utf8.RuneCountInString(name) > masterNameMaxLength {
		return apperrors.WrapInvalidInput(fmt.Sprintf("名前は%d文字以内で入力してください", masterNameMaxLength))
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

// validateOptionalName は nil 許容の名称バリデーション。
// PATCH 系で nil の場合は更新しない意味なのでスキップ、非 nil の場合のみ検証する。
func validateOptionalName(name *string) error {
	if name == nil {
		return nil
	}
	return validateRequiredName(*name)
}

// validateOwnedMasterFK は request 由来の clinic-scoped master FK の所有権を検証する
// (R24: マスタFKガードの複数イディオムを共有ヘルパーへ統一)。
// find は対象マスタの FindByID を包んだアダプタ。nil の FK はスキップ（optional FK 規約）。
func validateOwnedMasterFK(ctx context.Context, entity string, clinicID uint64, id *uint64,
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

// setNullableUint64Field mirrors internal/service's update_fields.go setNullableUint64Field
// (shared kernel, boundary map §4.2 "維持": fan-in medicalrecord+reservation). clearAssoc
// requests explicitly clearing the association (fields[col] = nil); a non-nil id sets it;
// otherwise the field is left untouched (PATCH semantics — omitted = no change).
func setNullableUint64Field(fields map[string]any, col string, clearAssoc bool, id *uint64) {
	switch {
	case clearAssoc:
		fields[col] = nil
	case id != nil:
		fields[col] = *id
	}
}
