package sharedkernel

import (
	"context"
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PetLifeState は共通ペット現行 basic revision から解決した生死の明示状態（EMR-491 D6）。
// string 型にしているのは、未設定のゼロ値 "" が alive にも deceased にも一致しないようにするため。
// 生存確認は PetLifeStateAlive との完全一致だけで行い、それ以外（unknown を含む）は通さない。
type PetLifeState string

const (
	// PetLifeStateAlive は status=alive かつ deceased_at が NULL の場合だけ返る。
	PetLifeStateAlive PetLifeState = "alive"
	// PetLifeStateDeceased は deceased_at が非 NULL、または status=deceased の場合に返る。
	PetLifeStateDeceased PetLifeState = "deceased"
)

// ErrPetLifeStateUnresolved は生死を確定できない場合（binding / subject / 現行 revision の欠落、
// revision の subject 不一致、status が NULL または未知かつ deceased_at が NULL）のセンチネル。
// reader の I/O 失敗（DB エラー等）とは errors.Is で区別でき、どちらも「生存」として扱ってはならない。
// 別センチネルを %w で合成しないこと（apperrors.WrapUnprocessable の注記と同じ理由）。
var ErrPetLifeStateUnresolved = errors.New("pet life state unresolved")

// PetLifeStateReader は clinic スコープでペット 1 件の生死を読む narrow interface（D6 死亡 write ガード）。
// 実装は ambient transaction 内で subject 行を FOR SHARE ロックしてから現行 revision を解決する。
// 返り値は alive / deceased のどちらか、または error（not found / unresolved / reader failure）。
// bool とデフォルト値は返さない。実装は pet package が持ち、各利用側へは constructor 注入する。
type PetLifeStateReader interface {
	ReadPetLifeState(ctx context.Context, clinicID, petID uint64) (PetLifeState, error)
}

// ClassifyPetLifeState は同一 revision の status / deceased_at から D6 真理値表を適用する純粋関数。
//
//   - deceased_at != nil                          -> deceased（status が NULL / 未知でも）
//   - status == deceased                           -> deceased
//   - status == alive かつ deceased_at == nil      -> alive
//   - それ以外（status が NULL / 未知かつ deceased_at == nil）-> ErrPetLifeStateUnresolved（alive にしない）
//
// 日時の捏造・backfill・値の変更は行わない。
func ClassifyPetLifeState(status *model.PetStatus, deceasedAt *time.Time) (PetLifeState, error) {
	if deceasedAt != nil {
		return PetLifeStateDeceased, nil
	}
	if status == nil {
		return "", apperrors.Wrap(ErrPetLifeStateUnresolved, "pet status is not recorded")
	}
	switch *status {
	case model.PetStatusDeceased:
		return PetLifeStateDeceased, nil
	case model.PetStatusAlive:
		return PetLifeStateAlive, nil
	default:
		return "", apperrors.Wrap(ErrPetLifeStateUnresolved, "pet status is not recognized")
	}
}

// ValidatePetNotDeceased は死亡ペットへの業務 write を fail-closed で拒否する。
// 生存と確認できた場合だけ nil を返す。死亡は message（呼び出し元の業務文言）の InvalidInput、
// not found は not found、unresolved / reader 失敗はそのままエラーにする（unknown は決して通さない）。
// reader が nil の場合も内部エラーで拒否する（nil を生存扱いにしない）。
// FE の選択 UI ブロックだけでは API 直叩きを防げないため BE 側でも検証する。
func ValidatePetNotDeceased(ctx context.Context, reader PetLifeStateReader, clinicID, petID uint64, message string) error {
	if reader == nil {
		return apperrors.WrapInternalServerError("pet life state reader is required")
	}
	state, err := reader.ReadPetLifeState(ctx, clinicID, petID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify pet status")
	}
	switch state {
	case PetLifeStateAlive:
		return nil
	case PetLifeStateDeceased:
		return apperrors.WrapInvalidInput(message)
	default:
		return apperrors.Wrap(ErrPetLifeStateUnresolved, "pet life state reader returned no state")
	}
}
