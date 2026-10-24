package sharedkernel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordLocker は LockByIDForUpdate 1メソッドの narrow interface（BE-refactor.md E-5）。
// ctx-txKey 方式のリポジトリフィールドと repo-swap 方式の tx-bound リポジトリの両方がこれを満たす。
type MedicalRecordLocker interface {
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// LockDraftMedicalRecord は子エンティティ writer 共通の親カルテ行ロックを一箇所に集約する
// （BE-refactor.md E-5）。LockByIDForUpdate の行ロックで親カルテの他 writer と直列化し、
// clinic 境界（repo の LockByIDForUpdate が clinic + 論理削除でスコープ）と存在（nil -> NotFound の
// fail-closed、A-5）を保証する。呼び出しは ambient transaction 内で行い、ロック順は呼び出し元が決める。
//
// EMR-495 P1（契約 §0.1）: 旧 status=finalized による拒否は廃止した。medical_records.status 列は
// 存在せず、通常の編集は旧 finalized だけでは拒否されない。呼び出し元ごとの既存保護（版 CAS・監査・
// 目的別規則・検査の確定/履歴 seal）は呼び出し元側に残る。494 の edit fence / 462 医師確認は未参加。
// 関数名は互換のため維持している（Draft の語は歴史的な呼称）。
//
// findErrMsg は LockByIDForUpdate 失敗時の slog.ErrorContext / apperrors.Wrap メッセージ。
// conflictMsg は旧 finalized 拒否の文言で、現在は使われない。既存の約25の呼び出し面を変えないために
// シグネチャを維持している（引数の整理は後続）。
func LockDraftMedicalRecord(ctx context.Context, repo MedicalRecordLocker, clinicID, recordID uint64, findErrMsg, _ string) error {
	parent, err := repo.LockByIDForUpdate(ctx, clinicID, recordID)
	if err != nil {
		slog.ErrorContext(ctx, findErrMsg, "error", err)
		return apperrors.Wrap(err, findErrMsg)
	}
	// fail-closed: parent が nil の場合はカルテ不在として NotFound を返す（BE-refactor.md A-5）。
	if parent == nil {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", recordID))
	}
	return nil
}
