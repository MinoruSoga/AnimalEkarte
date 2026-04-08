// Package service provides business logic implementations for TrimmingCourse and TrimmingOption entities.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- TrimmingCourseService ----

type TrimmingCourseService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingCourseService struct {
	repo repository.TrimmingCourseRepository
}

func NewTrimmingCourseService(repo repository.TrimmingCourseRepository) TrimmingCourseService {
	return &trimmingCourseService{repo: repo}
}

func (s *trimmingCourseService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	return s.repo.FindAll(ctx, clinicID)
}
func (s *trimmingCourseService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}
func (s *trimmingCourseService) Create(ctx context.Context, course *model.TrimmingCourse) error {
	return s.repo.Create(ctx, course)
}
func (s *trimmingCourseService) Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	fields := buildTrimmingCourseUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	course, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update trimming course")
	}
	slog.InfoContext(ctx, "trimming course updated", slog.Uint64("trimming_course_id", id))
	return course, nil
}
func (s *trimmingCourseService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountRecordsByCourseID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check trimming course dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("このトリミングコースはトリミング記録で使用中のため削除できません")
	}
	return s.repo.Delete(ctx, clinicID, id)
}

func (s *trimmingCourseService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}

// UpdateTrimmingCourseInput はトリミングコース更新のサービス入力 DTO
type UpdateTrimmingCourseInput struct {
	Name        *string
	Price       *int64
	IsActive    *bool
	Description *string
	TargetSize  *model.TargetSize
	Duration    *int
	SortOrder   *int
}

func buildTrimmingCourseUpdateFields(input UpdateTrimmingCourseInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Price != nil {
		fields["price"] = input.Price
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.TargetSize != nil {
		fields["target_size"] = *input.TargetSize
	}
	if input.Duration != nil {
		fields["duration"] = *input.Duration
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}

// ---- TrimmingOptionService ----

type TrimmingOptionService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingOptionInput) (*model.TrimmingOption, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingOptionService struct {
	repo repository.TrimmingOptionRepository
}

func NewTrimmingOptionService(repo repository.TrimmingOptionRepository) TrimmingOptionService {
	return &trimmingOptionService{repo: repo}
}

func (s *trimmingOptionService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	return s.repo.FindAll(ctx, clinicID)
}
func (s *trimmingOptionService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}
func (s *trimmingOptionService) Create(ctx context.Context, option *model.TrimmingOption) error {
	return s.repo.Create(ctx, option)
}
func (s *trimmingOptionService) Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
	fields := buildTrimmingOptionUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	option, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update trimming option")
	}
	slog.InfoContext(ctx, "trimming option updated", slog.Uint64("trimming_option_id", id))
	return option, nil
}
func (s *trimmingOptionService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountRecordsByOptionID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check trimming option dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("このトリミングオプションはトリミング記録で使用中のため削除できません")
	}
	return s.repo.Delete(ctx, clinicID, id)
}

func (s *trimmingOptionService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}

// UpdateTrimmingOptionInput はトリミングオプション更新のサービス入力 DTO
type UpdateTrimmingOptionInput struct {
	Name        *string
	Price       *int64
	IsActive    *bool
	Description *string
	Duration    *int
	Combinable  *bool
	SortOrder   *int
}

func buildTrimmingOptionUpdateFields(input UpdateTrimmingOptionInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Price != nil {
		fields["price"] = input.Price
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Duration != nil {
		fields["duration"] = *input.Duration
	}
	if input.Combinable != nil {
		fields["combinable"] = *input.Combinable
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
