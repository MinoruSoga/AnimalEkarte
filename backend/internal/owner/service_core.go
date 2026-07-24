package owner

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *ownerService) List(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	owners, total, err := s.repo.FindAll(ctx, clinicIDs, page, limit, search)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list owners", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list owners")
	}
	return owners, total, nil
}

func (s *ownerService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get owner", "error", err)
		return nil, apperrors.Wrap(err, "failed to get owner")
	}
	return owner, nil
}

func (s *ownerService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	owner, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get owner for clinics", "error", err)
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
		slog.ErrorContext(ctx, "failed to create owner with pets", "error", err)
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
		slog.ErrorContext(ctx, "failed to find owner", "error", err)
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

	// 更新フィールドマップ構築（nil フィールドはスキップ）
	fields := buildOwnerUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	owner, err := s.updateOwnerAndFind(
		ctx,
		clinicID,
		id,
		fields,
		"failed to update owner",
		"failed to update owner",
	)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "owner updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))

	return owner, nil
}

func (s *ownerService) ensureOwnerEmailUnique(ctx context.Context, clinicID, currentOwnerID uint64, email string) error {
	if email == "" {
		return nil
	}
	existing, err := s.repo.FindByEmail(ctx, clinicID, email)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check email uniqueness", "error", err)
		return apperrors.Wrap(err, "failed to check email uniqueness")
	}
	if existing != nil && existing.ID != currentOwnerID {
		return apperrors.WrapAlreadyExists("owner", "このメールアドレスはすでに登録されています")
	}
	return nil
}

func (s *ownerService) ensureOwnerPhoneUnique(ctx context.Context, clinicID, currentOwnerID uint64, phone string) error {
	if phone == "" {
		return nil
	}
	existing, err := s.repo.FindByPhone(ctx, clinicID, phone)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check phone uniqueness", "error", err)
		return apperrors.Wrap(err, "failed to check phone uniqueness")
	}
	if existing != nil && existing.ID != currentOwnerID {
		return apperrors.WrapAlreadyExists("owner", "この電話番号はすでに登録されています")
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
	slog.ErrorContext(ctx, logMessage, "error", err, "id", id, "clinic_id", clinicID)
	return apperrors.Wrap(err, wrapMessage)
}

func (s *ownerService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}
	// FK依存チェック: ペットが紐付いている場合は削除を拒否
	petCount, err := s.repo.CountPetsByOwnerID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check pet dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check pet dependencies")
	}
	if petCount > 0 {
		return apperrors.WrapConflict("ペットが紐付いているため削除できません。先にペットを削除してください")
	}

	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete owner", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete owner")
	}
	slog.InfoContext(ctx, "owner deleted",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
