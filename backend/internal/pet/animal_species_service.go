// Package pet provides business logic implementations for AnimalSpecies entity.
package pet

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// 列名定数
const (
	colAnimalSpeciesName      = "name"
	colAnimalSpeciesIsActive  = "is_active"
	colAnimalSpeciesSortOrder = "sort_order"
)

// Audit action strings for global animal_species writes (model constants not required).
const (
	auditActionAnimalSpeciesCreate  = "animal_species.create"
	auditActionAnimalSpeciesUpdate  = "animal_species.update"
	auditActionAnimalSpeciesDelete  = "animal_species.delete"
	auditActionAnimalSpeciesReorder = "animal_species.reorder"
	auditResourceAnimalSpecies      = "animal_species"
)

// buildAnimalSpeciesUpdate はポインタが非 nil のフィールドのみ map に追加する
func buildAnimalSpeciesUpdate(input *UpdateAnimalSpeciesInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colAnimalSpeciesName] = *input.Name
	}
	if input.IsActive != nil {
		fields[colAnimalSpeciesIsActive] = *input.IsActive
	}
	if input.SortOrder != nil {
		fields[colAnimalSpeciesSortOrder] = *input.SortOrder
	}
	return fields
}

// ---- Input DTOs ----

// AnimalSpeciesMutationMeta carries actor/clinic context for fail-closed audit.
// ClinicID is the executing clinic (audit.Entry requires non-zero clinic_id even for global masters).
type AnimalSpeciesMutationMeta struct {
	ClinicID  uint64
	ActorID   *uint64
	ActorType string
	IPAddress string
	UserAgent string
}

// CreateAnimalSpeciesInput は動物種類作成の入力DTO
type CreateAnimalSpeciesInput struct {
	Name      string
	IsActive  bool
	SortOrder int
}

// UpdateAnimalSpeciesInput は動物種類更新の入力DTO（nil = 未指定 = 更新しない）
type UpdateAnimalSpeciesInput struct {
	Name      *string
	IsActive  *bool
	SortOrder *int
}

// AnimalSpeciesService はペット種類マスタのビジネスロジック層
type AnimalSpeciesService interface {
	List(ctx context.Context) ([]model.AnimalSpecies, error)
	GetByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error)
	Create(ctx context.Context, input *CreateAnimalSpeciesInput, meta AnimalSpeciesMutationMeta) (*model.AnimalSpecies, error)
	Update(ctx context.Context, id uint64, input *UpdateAnimalSpeciesInput, meta AnimalSpeciesMutationMeta) (*model.AnimalSpecies, error)
	Delete(ctx context.Context, id uint64, meta AnimalSpeciesMutationMeta) error
	Reorder(ctx context.Context, ids []uint64, meta AnimalSpeciesMutationMeta) error
}

type animalSpeciesService struct {
	repo    AnimalSpeciesRepository
	petRepo AnimalSpeciesUsageCounter
	audit   AnimalSpeciesAuditLogger
	tx      AnimalSpeciesTransactor
}

// NewAnimalSpeciesService はAnimalSpeciesServiceを初期化して返す。
// audit/tx が未配線の場合でも system_admin 境界は handler 側で閉じる（DEC-31）。
// 監査は NewAnimalSpeciesServiceWithAudit で注入する。
func NewAnimalSpeciesService(repo AnimalSpeciesRepository, petRepo AnimalSpeciesUsageCounter) AnimalSpeciesService {
	return &animalSpeciesService{repo: repo, petRepo: petRepo}
}

// NewAnimalSpeciesServiceWithAudit wires fail-closed audit on the same transaction as writes.
func NewAnimalSpeciesServiceWithAudit(
	repo AnimalSpeciesRepository,
	petRepo AnimalSpeciesUsageCounter,
	auditLogger AnimalSpeciesAuditLogger,
	tx AnimalSpeciesTransactor,
) AnimalSpeciesService {
	return &animalSpeciesService{repo: repo, petRepo: petRepo, audit: auditLogger, tx: tx}
}

func (s *animalSpeciesService) List(ctx context.Context) ([]model.AnimalSpecies, error) {
	items, err := s.repo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list animal species", "error", err)
		return nil, apperrors.Wrap(err, "failed to list animal species")
	}
	return items, nil
}

func (s *animalSpeciesService) GetByID(ctx context.Context, id uint64) (*model.AnimalSpecies, error) {
	result, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get animal species", "error", err, "id", id)
		return nil, apperrors.Wrap(err, "failed to get animal species")
	}
	return result, nil
}

func (s *animalSpeciesService) Create(ctx context.Context, input *CreateAnimalSpeciesInput, meta AnimalSpeciesMutationMeta) (*model.AnimalSpecies, error) {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	species := &model.AnimalSpecies{
		Name:      input.Name,
		IsActive:  input.IsActive,
		SortOrder: input.SortOrder,
	}
	err := s.withOptionalAuditTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, species); err != nil {
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				input.Name,
				apperrors.ConstraintAnimalSpeciesName,
				apperrors.CodeAnimalSpeciesNameConflict,
			); conflict != nil {
				return conflict
			}
			slog.ErrorContext(txCtx, "failed to create animal species", "error", err)
			return apperrors.Wrap(err, "failed to create animal species")
		}
		return s.writeSpeciesAudit(txCtx, meta, auditActionAnimalSpeciesCreate, &species.ID, nil, species)
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "animal species created",
		slog.Uint64("animal_species_id", species.ID))
	return species, nil
}

func (s *animalSpeciesService) Update(ctx context.Context, id uint64, input *UpdateAnimalSpeciesInput, meta AnimalSpeciesMutationMeta) (*model.AnimalSpecies, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildAnimalSpeciesUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}

	var result *model.AnimalSpecies
	err := s.withOptionalAuditTx(ctx, func(txCtx context.Context) error {
		old, err := s.repo.FindByID(txCtx, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get animal species", "error", err)
			return apperrors.Wrap(err, "failed to get animal species")
		}
		updated, err := s.repo.Update(txCtx, id, fields)
		if err != nil {
			nameForConflict := ""
			if input.Name != nil {
				nameForConflict = *input.Name
			}
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				nameForConflict,
				apperrors.ConstraintAnimalSpeciesName,
				apperrors.CodeAnimalSpeciesNameConflict,
			); conflict != nil {
				return conflict
			}
			slog.ErrorContext(txCtx, "failed to update animal species", "error", err, "id", id)
			return apperrors.Wrap(err, "failed to update animal species")
		}
		result = updated
		return s.writeSpeciesAudit(txCtx, meta, auditActionAnimalSpeciesUpdate, &id, old, updated)
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "animal species updated",
		slog.Uint64("animal_species_id", id))
	return result, nil
}

func (s *animalSpeciesService) Delete(ctx context.Context, id uint64, meta AnimalSpeciesMutationMeta) error {
	err := s.withOptionalAuditTx(ctx, func(txCtx context.Context) error {
		old, err := s.repo.FindByID(txCtx, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to find animal species")
		}
		count, err := s.petRepo.CountUsageByAnimalSpeciesID(txCtx, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check animal species dependencies", "error", err, "id", id)
			return apperrors.Wrap(err, "failed to check animal species dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("この動物種はペット情報で使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete animal species", "error", err, "id", id)
			return apperrors.Wrap(err, "failed to delete animal species")
		}
		return s.writeSpeciesAudit(txCtx, meta, auditActionAnimalSpeciesDelete, &id, old, nil)
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "animal species deleted", slog.Uint64("animal_species_id", id))
	return nil
}

func (s *animalSpeciesService) Reorder(ctx context.Context, ids []uint64, meta AnimalSpeciesMutationMeta) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	err := s.withOptionalAuditTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Reorder(txCtx, ids); err != nil {
			slog.ErrorContext(txCtx, "failed to reorder animal species", "error", err)
			return apperrors.Wrap(err, "failed to reorder animal species")
		}
		return s.writeSpeciesAudit(txCtx, meta, auditActionAnimalSpeciesReorder, nil, nil, map[string]any{"ids": ids})
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "animal species reordered", slog.Int("count", len(ids)))
	return nil
}

func (s *animalSpeciesService) withOptionalAuditTx(ctx context.Context, fn func(context.Context) error) error {
	if s.audit != nil {
		if s.tx == nil {
			return apperrors.WrapInternalServerError("animal species audit requires a transaction owner")
		}
		if err := s.tx.WithTx(ctx, fn); err != nil {
			return apperrors.Wrap(err, "failed animal species mutation with audit")
		}
		return nil
	}
	return fn(ctx)
}

func (s *animalSpeciesService) writeSpeciesAudit(
	ctx context.Context,
	meta AnimalSpeciesMutationMeta,
	action string,
	resourceID *uint64,
	oldValue any,
	newValue any,
) error {
	if s.audit == nil {
		return nil
	}
	if meta.ClinicID == 0 {
		return apperrors.WrapInvalidInput("clinic context is required to audit animal species changes")
	}
	clinicID := meta.ClinicID
	actorType := meta.ActorType
	if actorType == "" {
		actorType = "staff"
	}
	entry := &audit.Entry{
		ClinicID:   &clinicID,
		ActorID:    meta.ActorID,
		ActorType:  actorType,
		Action:     action,
		Resource:   auditResourceAnimalSpecies,
		ResourceID: resourceID,
		OldValue:   oldValue,
		NewValue:   newValue,
		IPAddress:  meta.IPAddress,
		UserAgent:  meta.UserAgent,
	}
	if err := s.audit.LogEntryTx(ctx, entry); err != nil {
		return apperrors.Wrap(err, "failed to audit animal species mutation")
	}
	return nil
}
