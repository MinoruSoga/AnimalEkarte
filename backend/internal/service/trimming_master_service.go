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

// CreateTrimmingCourseInput はトリミングコース作成の入力DTO
type CreateTrimmingCourseInput struct {
	Name        string
	TargetSize  string
	Price       *int64
	Duration    *int
	IsActive    bool
	Description string
	SortOrder   int
}

type TrimmingCourseService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseInput) (*model.TrimmingCourse, error)
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
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list trimming courses")
	}
	return result, nil
}
func (s *trimmingCourseService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming course")
	}
	return result, nil
}
func (s *trimmingCourseService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	course := &model.TrimmingCourse{
		ClinicID:    clinicID,
		Name:        input.Name,
		Price:       input.Price,
		IsActive:    input.IsActive,
		Description: input.Description,
		Duration:    input.Duration,
		SortOrder:   input.SortOrder,
	}
	if input.TargetSize != "" {
		ts := model.TargetSize(input.TargetSize)
		course.TargetSize = &ts
	}
	if err := s.repo.Create(ctx, course); err != nil {
		return nil, apperrors.Wrap(err, "failed to create trimming course")
	}
	slog.InfoContext(ctx, "trimming course created",
		slog.Uint64("trimming_course_id", course.ID),
		slog.Uint64("clinic_id", clinicID))
	return course, nil
}
func (s *trimmingCourseService) Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
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
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete trimming course")
	}
	slog.InfoContext(ctx, "trimming course deleted", slog.Uint64("trimming_course_id", id))
	return nil
}

func (s *trimmingCourseService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder trimming courses")
	}
	slog.InfoContext(ctx, "trimming courses reordered", slog.Uint64("clinic_id", clinicID))
	return nil
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
		fields["price"] = *input.Price
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

// CreateTrimmingOptionInput はトリミングオプション作成の入力DTO
type CreateTrimmingOptionInput struct {
	Name         string
	Price        *int64
	IsActive     bool
	Description  string
	Duration     *int
	IsCombinable bool
	SortOrder    int
}

type TrimmingOptionService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingOptionInput) (*model.TrimmingOption, error)
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
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list trimming options")
	}
	return result, nil
}
func (s *trimmingOptionService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming option")
	}
	return result, nil
}
func (s *trimmingOptionService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	option := &model.TrimmingOption{
		ClinicID:     clinicID,
		Name:         input.Name,
		Price:        input.Price,
		IsActive:     input.IsActive,
		Description:  input.Description,
		Duration:     input.Duration,
		IsCombinable: input.IsCombinable,
		SortOrder:    input.SortOrder,
	}
	if err := s.repo.Create(ctx, option); err != nil {
		return nil, apperrors.Wrap(err, "failed to create trimming option")
	}
	slog.InfoContext(ctx, "trimming option created",
		slog.Uint64("trimming_option_id", option.ID),
		slog.Uint64("clinic_id", clinicID))
	return option, nil
}
func (s *trimmingOptionService) Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
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
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete trimming option")
	}
	slog.InfoContext(ctx, "trimming option deleted", slog.Uint64("trimming_option_id", id))
	return nil
}

func (s *trimmingOptionService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder trimming options")
	}
	slog.InfoContext(ctx, "trimming options reordered", slog.Uint64("clinic_id", clinicID))
	return nil
}

// UpdateTrimmingOptionInput はトリミングオプション更新のサービス入力 DTO
type UpdateTrimmingOptionInput struct {
	Name         *string
	Price        *int64
	IsActive     *bool
	Description  *string
	Duration     *int
	IsCombinable *bool
	SortOrder    *int
}

func buildTrimmingOptionUpdateFields(input UpdateTrimmingOptionInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Price != nil {
		fields["price"] = *input.Price
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
	if input.IsCombinable != nil {
		fields["is_combinable"] = *input.IsCombinable
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
