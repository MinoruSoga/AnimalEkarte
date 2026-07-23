package trimming

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateTrimmingCourseInput はトリミングコース作成の入力DTO
type CreateTrimmingCourseInput struct {
	Name         string
	TargetSize   string
	CourseTypeID *uint64 // #73 種別マスタ FK(nullable)
	Price        *int64
	Duration     *int
	IsActive     bool
	Description  string
	SortOrder    int
}

// UpdateTrimmingCourseInput はトリミングコース更新のサービス入力 DTO
type UpdateTrimmingCourseInput struct {
	Name         *string
	Price        *int64
	IsActive     *bool
	Description  *string
	TargetSize   *string
	CourseTypeID *uint64 // #73 種別マスタ FK(nullable)
	Duration     *int
	SortOrder    *int
}

const (
	colTrimmingCourseName         = "name"
	colTrimmingCoursePrice        = "price"
	colTrimmingCourseIsActive     = "is_active"
	colTrimmingCourseDescription  = "description"
	colTrimmingCourseTargetSize   = "target_size"
	colTrimmingCourseCourseTypeID = "course_type_id"
	colTrimmingCourseDuration     = "duration"
	colTrimmingCourseSortOrder    = "sort_order"
)

func buildTrimmingCourseUpdate(input *UpdateTrimmingCourseInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colTrimmingCourseName] = *input.Name
	}
	if input.Price != nil {
		fields[colTrimmingCoursePrice] = *input.Price
	}
	if input.IsActive != nil {
		fields[colTrimmingCourseIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colTrimmingCourseDescription] = *input.Description
	}
	if input.TargetSize != nil {
		fields[colTrimmingCourseTargetSize] = model.TargetSize(*input.TargetSize)
	}
	if input.CourseTypeID != nil {
		fields[colTrimmingCourseCourseTypeID] = *input.CourseTypeID
	}
	if input.Duration != nil {
		fields[colTrimmingCourseDuration] = *input.Duration
	}
	if input.SortOrder != nil {
		fields[colTrimmingCourseSortOrder] = *input.SortOrder
	}
	return fields
}

// TrimmingCourseService はトリミングコースのビジネスロジックインターフェース
type TrimmingCourseService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseInput) (*model.TrimmingCourse, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingCourseService struct {
	repo           TrimmingCourseRepository
	courseTypeRepo TrimmingCourseTypeRepository
	transactor     Transactor
}

// NewTrimmingCourseService は TrimmingCourseService を生成する
func NewTrimmingCourseService(
	repo TrimmingCourseRepository,
	courseTypeRepo TrimmingCourseTypeRepository,
	transactor Transactor,
) TrimmingCourseService {
	return &trimmingCourseService{
		repo:           repo,
		courseTypeRepo: courseTypeRepo,
		transactor:     transactor,
	}
}

func (s *trimmingCourseService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	result, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list trimming courses", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list trimming courses")
	}
	return result, nil
}

func (s *trimmingCourseService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get trimming course", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get trimming course")
	}
	return result, nil
}

func (s *trimmingCourseService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	course := &model.TrimmingCourse{
		ClinicID:     clinicID,
		Name:         input.Name,
		Price:        input.Price,
		IsActive:     input.IsActive,
		Description:  input.Description,
		CourseTypeID: input.CourseTypeID,
		Duration:     input.Duration,
		SortOrder:    input.SortOrder,
	}
	if input.TargetSize != "" {
		ts := model.TargetSize(input.TargetSize)
		course.TargetSize = &ts
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("trimming course transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if input.CourseTypeID != nil {
			if _, err := s.courseTypeRepo.FindByID(txCtx, clinicID, *input.CourseTypeID); err != nil {
				return apperrors.WrapInvalidInput("course_type_id not found in this clinic")
			}
		}
		if err := s.repo.Create(txCtx, course); err != nil {
			return apperrors.Wrap(err, "failed to create trimming course")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create trimming course", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create trimming course")
	}
	slog.InfoContext(ctx, "trimming course created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("trimming_course_id", course.ID))
	return course, nil
}

func (s *trimmingCourseService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get trimming course", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get trimming course")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("trimming course transaction dependency is required")
	}
	var course *model.TrimmingCourse
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if input.CourseTypeID != nil {
			if _, err := s.courseTypeRepo.FindByID(txCtx, clinicID, *input.CourseTypeID); err != nil {
				return apperrors.WrapInvalidInput("course_type_id not found in this clinic")
			}
		}
		if err := validateOptionalName(input.Name); err != nil {
			return apperrors.Wrap(err, "failed to validate optional name")
		}
		fields := buildTrimmingCourseUpdate(input)
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
		}
		updated, err := s.repo.Update(txCtx, clinicID, id, fields)
		if err != nil {
			return apperrors.Wrap(err, "failed to update trimming course")
		}
		course = updated
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to update trimming course", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update trimming course")
	}
	slog.InfoContext(ctx, "trimming course updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("trimming_course_id", id))
	return course, nil
}

func (s *trimmingCourseService) Delete(ctx context.Context, clinicID, id uint64) error {
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("trimming course transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete trimming course")
		}
		count, err := s.repo.CountUsageByTrimmingCourseID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check trimming course dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("このトリミングコースはトリミング記録で使用中のため削除できません")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to delete trimming course", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete trimming course")
	}
	slog.InfoContext(ctx, "trimming course deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("trimming_course_id", id))
	return nil
}

func (s *trimmingCourseService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder trimming courses", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder trimming courses")
	}
	slog.InfoContext(ctx, "trimming courses reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
