package owner

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *ownerService) List(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	owners, total, err := s.repo.FindAll(ctx, clinicIDs, page, limit, search)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list owners")
	}
	return owners, total, nil
}

func (s *ownerService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get owner")
	}
	return owner, nil
}

func (s *ownerService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	owner, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get owner for clinics")
	}
	return owner, nil
}

// validateOwnerPetsInsuranceOwnership は同時作成する各ペットの InsuranceID
// (nested, request-derived master FK) が呼び出し元 clinic に属することを検証する
// (X-14 U5)。petService.Create/Update と同型の FindByID ガード。nil InsuranceID
// はスキップする。
func (s *ownerService) validateOwnerPetsInsuranceOwnership(ctx context.Context, clinicID uint64, pets []CreatePetForOwnerInput) error {
	for i := range pets {
		pet := &pets[i]
		if pet.InsuranceID == nil {
			continue
		}
		if s.insuranceFinder == nil {
			return apperrors.WrapInternalServerError("owner insurance finder is not configured")
		}
		insurance, err := s.insuranceFinder.FindByID(ctx, clinicID, *pet.InsuranceID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return apperrors.WrapInvalidInput("insurance not found in this clinic")
			}
			return apperrors.Wrap(err, "failed to validate pet insurance")
		}
		if insurance == nil {
			return apperrors.WrapInvalidInput("insurance not found in this clinic")
		}
	}
	return nil
}

func (s *ownerService) CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
	if err := validateCreateOwnerInput(input); err != nil {
		return nil, err
	}

	if err := s.ensureOwnerEmailUnique(ctx, clinicID, 0, input.Email); err != nil {
		return nil, err
	}
	if err := s.ensureOwnerPhoneUnique(ctx, clinicID, 0, input.Phone); err != nil {
		return nil, err
	}

	if err := s.validateOwnerPetsInsuranceOwnership(ctx, clinicID, input.Pets); err != nil {
		return nil, err
	}

	owner := buildOwnerModel(clinicID, input)
	pets := buildOwnerPetModels(input.Pets)

	if err := s.repo.CreateWithPets(ctx, owner, pets); err != nil {
		return nil, apperrors.Wrap(err, "failed to create owner with pets")
	}

	slog.InfoContext(ctx, "owner created with pets",
		slog.Uint64("owner_id", owner.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Int("pets_count", len(pets)))
	if s.tagNotifier != nil {
		if err := s.tagNotifier.SyncOwnerAnimalClassificationTags(ctx, clinicID, owner.ID); err != nil {
			slog.ErrorContext(ctx, "failed to sync animal classification tags after owner create", "error", err, "owner_id", owner.ID, "clinic_id", clinicID)
		}
	}

	return owner, nil
}

func (s *ownerService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	if err := validateUpdateOwnerInput(input); err != nil {
		return nil, err
	}

	if input.Email != nil {
		if err := s.ensureOwnerEmailUnique(ctx, clinicID, id, *input.Email); err != nil {
			return nil, err
		}
	}
	if input.Phone != nil {
		if err := s.ensureOwnerPhoneUnique(ctx, clinicID, id, *input.Phone); err != nil {
			return nil, err
		}
	}

	// SEC-CS-F15: lock owner under write TX and recheck discount_rate against locked snapshot.
	// Handler early check uses pre-TX GetByID; stale equality must not authorize overwrite.
	owner, err := s.repo.UpdateAndFindApplying(ctx, clinicID, id, func(locked *model.Owner) (map[string]any, error) {
		fields := buildOwnerUpdate(input)
		if err := applyOwnerDiscountField(fields, locked, input); err != nil {
			return nil, err
		}
		if len(fields) == 0 {
			return nil, apperrors.WrapInvalidInput("at least one field must be provided")
		}
		return fields, nil
	})
	if err != nil {
		return nil, s.wrapOwnerUpdateError(ctx, clinicID, id, err, "failed to update owner", "failed to update owner")
	}

	slog.InfoContext(ctx, "owner updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))

	return owner, nil
}

func applyOwnerDiscountField(fields map[string]any, locked *model.Owner, input *UpdateOwnerInput) error {
	if input.DiscountRate == nil {
		return nil
	}
	if httpapi.FloatEquals(*input.DiscountRate, locked.DiscountRate) {
		if !input.DiscountEditAllowed {
			delete(fields, colDiscountRate)
		}
		return nil
	}
	if !input.DiscountEditAllowed {
		return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
	}
	return nil
}

func (s *ownerService) ensureOwnerEmailUnique(ctx context.Context, clinicID, currentOwnerID uint64, email string) error {
	if email == "" {
		return nil
	}
	existing, err := s.repo.FindByEmail(ctx, clinicID, email)
	if err != nil {
		return apperrors.Wrap(err, "failed to check email uniqueness")
	}
	if existing != nil && existing.ID != currentOwnerID {
		// BUG-024: 完成文を identifier に渡すと英語テンプレで英日混在になる
		return apperrors.WrapAlreadyExistsMessage("このメールアドレスはすでに登録されています")
	}
	return nil
}

func (s *ownerService) ensureOwnerPhoneUnique(ctx context.Context, clinicID, currentOwnerID uint64, phone string) error {
	if phone == "" {
		return nil
	}
	existing, err := s.repo.FindByPhone(ctx, clinicID, phone)
	if err != nil {
		return apperrors.Wrap(err, "failed to check phone uniqueness")
	}
	if existing != nil && existing.ID != currentOwnerID {
		// BUG-019 / BUG-024: 日本語完成文を message 直渡し（英語テンプレ埋め込み禁止）
		return apperrors.WrapAlreadyExistsMessage("この電話番号はすでに登録されています")
	}
	return nil
}

func (s *ownerService) updateOwnerAndFind(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
	logMessage, wrapMessage string,
) (*model.Owner, error) {
	owner, err := s.repo.UpdateAndFind(ctx, clinicID, id, fields)
	if err != nil {
		return nil, s.wrapOwnerUpdateError(ctx, clinicID, id, err, logMessage, wrapMessage)
	}
	return owner, nil
}

func (s *ownerService) wrapOwnerUpdateError(ctx context.Context, clinicID, id uint64, err error, logMessage, wrapMessage string) error {
	return apperrors.Wrap(err, wrapMessage)
}

func (s *ownerService) Delete(ctx context.Context, clinicID, id uint64) error {
	if !s.deleteGuardEnabled {
		return s.deleteWithoutPetOwnerGuard(ctx, clinicID, id)
	}
	if s.deleteLocker == nil || s.petOwnerCounter == nil || s.deleteTransactor == nil {
		return apperrors.WrapInternalServerError("owner delete guard dependencies are not configured")
	}

	err := s.deleteTransactor.WithTx(ctx, func(txCtx context.Context) error {
		lockedOwner, err := s.deleteLocker.LockForDelete(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to find owner")
		}
		if lockedOwner == nil {
			return apperrors.WrapInternalServerError("owner delete lock returned no owner")
		}

		petCount, err := s.repo.CountPetsByOwnerID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check pet dependencies")
		}
		if petCount > 0 {
			return apperrors.WrapConflict("ペットが紐付いているため削除できません。先にペットを削除してください")
		}

		petOwnerCount, err := s.petOwnerCounter.CountByOwnerID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check pet owner dependencies")
		}
		if petOwnerCount > 0 {
			return apperrors.WrapConflict("副飼主として紐付いているため削除できません。先に紐付けを解除してください")
		}

		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete owner")
		}
		return nil
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "owner deleted",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

// deleteWithoutPetOwnerGuard preserves the existing four-argument constructor
// contract. Production composition uses NewServiceWithPetOwnerDeleteGuard.
func (s *ownerService) deleteWithoutPetOwnerGuard(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}
	// FK依存チェック: ペットが紐付いている場合は削除を拒否
	petCount, err := s.repo.CountPetsByOwnerID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check pet dependencies")
	}
	if petCount > 0 {
		return apperrors.WrapConflict("ペットが紐付いているため削除できません。先にペットを削除してください")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete owner")
	}
	slog.InfoContext(ctx, "owner deleted",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
