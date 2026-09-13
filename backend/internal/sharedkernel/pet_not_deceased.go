package sharedkernel

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PetByIDFinder は clinic スコープでペット 1 件を読む narrow interface（SD-10 死亡 write ガード）。
// medicalrecord の petFinder / pet 系 repository の FindByID と構造的に互換。
type PetByIDFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

// ValidatePetNotDeceased は死亡ペットへの業務 write を fail-closed で拒否する。
// 死亡契約: status=deceased OR deceased_at != nil（日時の捏造・backfill はしない）。
// FE の選択 UI ブロックだけでは API 直叩きを防げないため BE 側でも検証する。
// message は呼び出し元の業務文言をそのまま返す（既存 hospitalization テストが assert するため）。
func ValidatePetNotDeceased(ctx context.Context, petRepo PetByIDFinder, clinicID, petID uint64, message string) error {
	pet, err := petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify pet status")
	}
	if pet == nil {
		return apperrors.WrapNotFound("pet", "status")
	}
	if pet.Status == model.PetStatusDeceased || pet.DeceasedAt != nil {
		return apperrors.WrapInvalidInput(message)
	}
	return nil
}
