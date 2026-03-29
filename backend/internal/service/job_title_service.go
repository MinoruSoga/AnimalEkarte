// Package service provides business logic implementations for JobTitle entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- JobTitleService ----

// UpdateJobTitleInput holds the fields that can be updated via PATCH.
// All fields are pointers: nil means "not provided / skip".
type UpdateJobTitleInput struct {
	Name        *string
	Description *string
	SortOrder   *int
	IsActive    *bool
}

type JobTitleService interface {
	List(ctx context.Context, clinicID uint64) ([]model.JobTitle, error)
	GetByID(ctx context.Context, id uint64) (*model.JobTitle, error)
	Create(ctx context.Context, jobTitle *model.JobTitle) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateJobTitleInput) (*model.JobTitle, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type jobTitleService struct {
	repo repository.JobTitleRepository
}

func NewJobTitleService(repo repository.JobTitleRepository) JobTitleService {
	return &jobTitleService{repo: repo}
}

func (s *jobTitleService) List(ctx context.Context, clinicID uint64) ([]model.JobTitle, error) {
	return s.repo.FindAll(ctx, clinicID)
}

func (s *jobTitleService) GetByID(ctx context.Context, id uint64) (*model.JobTitle, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *jobTitleService) Create(ctx context.Context, jobTitle *model.JobTitle) error {
	if err := s.repo.Create(ctx, jobTitle); err != nil {
		return apperrors.Wrap(err, "failed to create job title")
	}
	slog.InfoContext(ctx, "job title created",
		slog.Uint64("job_title_id", jobTitle.ID),
		slog.Uint64("clinic_id", jobTitle.ClinicID))
	return nil
}

func (s *jobTitleService) Update(ctx context.Context, clinicID, id uint64, input *UpdateJobTitleInput) (*model.JobTitle, error) {
	fields := buildJobTitleUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update job title")
	}
	slog.InfoContext(ctx, "job title updated",
		slog.Uint64("job_title_id", id),
		slog.Uint64("clinic_id", clinicID))
	return s.repo.FindByID(ctx, id)
}

func (s *jobTitleService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *jobTitleService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}

func buildJobTitleUpdateFields(input *UpdateJobTitleInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}
