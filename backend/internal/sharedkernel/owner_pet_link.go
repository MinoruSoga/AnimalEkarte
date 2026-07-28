package sharedkernel

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// OwnerPetLinkVerifier は Owner/Pet の clinic 所有検証 2 メソッドの narrow interface
// （repository.ReservationRepository の AssertOwnerInClinic/FindPetOwnerInClinic 相当・両方 dbOrTx 参加）。
type OwnerPetLinkVerifier interface {
	AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error
	FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error)
}

// ValidateReservationOwnerPetLinks は request 由来 Owner/Pet の clinic 所有と Owner-Pet 整合を検証する（AUD-001）。
// 別 clinic と未存在は区別せず NotFound 系で返す。nil 関連は既存契約どおり許可する。
// 呼び出し側が恒久ドメイン境界を跨ぐ（reservation 系 + medicalrecord の hospitalization 系）ため
// sharedkernel が唯一の実装（BE9-2D ⑤ Batch B で昇格）。
func ValidateReservationOwnerPetLinks(ctx context.Context, repo OwnerPetLinkVerifier, clinicID uint64, ownerID, petID *uint64) error {
	if ownerID != nil {
		if err := repo.AssertOwnerInClinic(ctx, clinicID, *ownerID); err != nil {
			return apperrors.Wrap(err, "failed to verify owner ownership")
		}
	}
	if petID != nil {
		petOwnerID, err := repo.FindPetOwnerInClinic(ctx, clinicID, *petID)
		if err != nil {
			return apperrors.Wrap(err, "failed to verify pet ownership")
		}
		if ownerID != nil && petOwnerID != *ownerID {
			return apperrors.WrapNotFound("pet", fmt.Sprintf("%d", *petID))
		}
	}
	return nil
}
