package pet

import (
	"context"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
)

const maxPetOwnerRelationshipRunes = 50

// PetOwnerLinkInput is one requested secondary-owner link.
type PetOwnerLinkInput struct {
	OwnerID      uint64
	Relationship string
}

// ReplacePetOwnersInput is the complete replacement command for a pet's
// secondary owners. Version is optional until the HTTP contract is introduced.
type ReplacePetOwnersInput struct {
	Version   *int
	Links     []PetOwnerLinkInput
	ActorID   *uint64
	ActorType string
	IPAddress string
	UserAgent string
}

// PetOwnerService owns clinic-scoped secondary-owner reads and replacements.
type PetOwnerService interface {
	GetByPetID(ctx context.Context, clinicID, petID uint64) ([]model.PetOwner, error)
	GetSharedPetsByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]SharedPet, error)
	ReplaceForPet(ctx context.Context, clinicID, petID uint64, input *ReplacePetOwnersInput) error
}

type petOwnerService struct {
	pets       PetFinder
	owners     OwnerFinder
	repository PetOwnerRepository
	transactor PetOwnerTransactor
	audit      PetOwnerAuditLogger
}

type petOwnerAuditValue struct {
	OwnerID      uint64 `json:"owner_id"`
	Relationship string `json:"relationship"`
}

// NewPetOwnerService constructs the secondary-owner application service.
func NewPetOwnerService(
	pets PetFinder,
	owners OwnerFinder,
	repository PetOwnerRepository,
	transactor PetOwnerTransactor,
	auditLogger PetOwnerAuditLogger,
) PetOwnerService {
	return &petOwnerService{
		pets:       pets,
		owners:     owners,
		repository: repository,
		transactor: transactor,
		audit:      auditLogger,
	}
}

func (s *petOwnerService) GetByPetID(
	ctx context.Context,
	clinicID, petID uint64,
) ([]model.PetOwner, error) {
	if s.pets == nil || s.repository == nil {
		return nil, apperrors.WrapInternalServerError("pet owner service dependencies are unavailable")
	}
	if _, err := s.pets.FindByID(ctx, clinicID, petID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find pet")
	}
	links, err := s.repository.FindByPetID(ctx, clinicID, petID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find pet owners")
	}
	if links == nil {
		return []model.PetOwner{}, nil
	}
	return links, nil
}

func (s *petOwnerService) GetSharedPetsByOwnerID(
	ctx context.Context,
	clinicID, ownerID uint64,
) ([]SharedPet, error) {
	if s.owners == nil || s.repository == nil {
		return nil, apperrors.WrapInternalServerError("pet owner service dependencies are unavailable")
	}
	if _, err := s.owners.FindByID(ctx, clinicID, ownerID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	sharedPets, err := s.repository.FindSharedPetsByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find shared pets")
	}
	if sharedPets == nil {
		return []SharedPet{}, nil
	}
	return sharedPets, nil
}

func (s *petOwnerService) ReplaceForPet(
	ctx context.Context,
	clinicID, petID uint64,
	input *ReplacePetOwnersInput,
) error {
	if s.pets == nil || s.owners == nil || s.repository == nil || s.transactor == nil || s.audit == nil {
		return apperrors.WrapInternalServerError("pet owner service dependencies are unavailable")
	}

	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		pet, err := s.pets.FindByID(txCtx, clinicID, petID)
		if err != nil {
			return apperrors.Wrap(err, "failed to find pet")
		}
		if pet == nil {
			return apperrors.WrapInternalServerError("pet lookup returned no result")
		}

		links, err := normalizePetOwnerLinks(input, pet.OwnerID)
		if err != nil {
			return err
		}
		for _, link := range links {
			if _, err := s.owners.FindByID(txCtx, clinicID, link.OwnerID); err != nil {
				return apperrors.Wrap(err, "failed to find pet owner")
			}
		}

		oldLinks, err := s.repository.FindByPetID(txCtx, clinicID, petID)
		if err != nil {
			return apperrors.Wrap(err, "failed to find existing pet owners")
		}
		if err := s.repository.ReplaceForPet(txCtx, clinicID, petID, links, input.Version); err != nil {
			return apperrors.Wrap(err, "failed to replace pet owners")
		}

		lockedPet, err := s.pets.FindByID(txCtx, clinicID, petID)
		if err != nil {
			return apperrors.Wrap(err, "failed to verify pet owner replacement")
		}
		if lockedPet == nil {
			return apperrors.WrapInternalServerError("pet verification returned no result")
		}
		if containsPetOwner(links, lockedPet.OwnerID) {
			return apperrors.WrapConflict("主飼主を副飼主として登録することはできません")
		}

		entry := &audit.Entry{
			ClinicID:   &clinicID,
			ActorID:    input.ActorID,
			ActorType:  input.ActorType,
			Action:     model.AuditActionPetOwnerReplace,
			Resource:   "pet",
			ResourceID: &petID,
			OldValue:   petOwnerAuditValues(oldLinks),
			NewValue:   petOwnerAuditValues(links),
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
		}
		if err := s.audit.LogEntryTx(txCtx, entry); err != nil {
			return apperrors.Wrap(err, "failed to audit pet owner replacement")
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to replace pet owners")
	}
	return nil
}

func normalizePetOwnerLinks(
	input *ReplacePetOwnersInput,
	primaryOwnerID uint64,
) ([]model.PetOwner, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("pet owner replacement input is required")
	}
	seen := make(map[uint64]struct{}, len(input.Links))
	links := make([]model.PetOwner, 0, len(input.Links))
	for _, requested := range input.Links {
		if requested.OwnerID == 0 {
			return nil, apperrors.WrapInvalidInput("owner_id is required")
		}
		if requested.OwnerID == primaryOwnerID {
			return nil, apperrors.WrapInvalidInput("主飼主を副飼主として登録することはできません")
		}
		if _, exists := seen[requested.OwnerID]; exists {
			return nil, apperrors.WrapInvalidInput("同じ副飼主を複数登録することはできません")
		}
		seen[requested.OwnerID] = struct{}{}

		relationship := strings.TrimSpace(requested.Relationship)
		length := utf8.RuneCountInString(relationship)
		if length < 1 || length > maxPetOwnerRelationshipRunes {
			return nil, apperrors.WrapInvalidInput("続柄は1文字以上50文字以内で入力してください")
		}
		links = append(links, model.PetOwner{
			OwnerID:      requested.OwnerID,
			Relationship: relationship,
		})
	}
	return links, nil
}

func containsPetOwner(links []model.PetOwner, ownerID uint64) bool {
	for _, link := range links {
		if link.OwnerID == ownerID {
			return true
		}
	}
	return false
}

func petOwnerAuditValues(links []model.PetOwner) []petOwnerAuditValue {
	values := make([]petOwnerAuditValue, len(links))
	for i, link := range links {
		values[i] = petOwnerAuditValue{
			OwnerID:      link.OwnerID,
			Relationship: link.Relationship,
		}
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].OwnerID < values[j].OwnerID
	})
	return values
}
