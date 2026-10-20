package trimming

import (
	"context"
	"log/slog"
	"sort"
	"strings"

	"github.com/lib/pq"

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
	// EMR-457: 任意属性。空文字・空配列は未登録（NULL、曜日は全曜日）。
	ShortName       string
	ImageObjectKey  string
	AllowedWeekdays []int
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
	// nil は既存値を保持する。ShortName/ImageObjectKey の "" と AllowedWeekdays の
	// 空配列は NULL へ戻す（曜日 NULL = 全曜日）。service が正規化してから repo へ渡す。
	ShortName       *string
	ImageObjectKey  *string
	AllowedWeekdays *[]int

	// 画像の置換・解除で不要になった旧 key の削除予約（tombstone）を打つ staff
	// actor。0 は呼び出し元が staff を特定できないことを意味し、削除予約が必要な
	// 更新は既存 lifecycle 契約どおり invalid input で拒否する（fail-closed）。
	ActorStaffID uint64
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

	colTrimmingCourseShortName       = "short_name"
	colTrimmingCourseImageObjectKey  = "image_object_key"
	colTrimmingCourseAllowedWeekdays = "allowed_weekdays"

	minTrimmingCourseWeekday = 0 // Sunday
	maxTrimmingCourseWeekday = 6 // Saturday
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
	if input.ShortName != nil {
		fields[colTrimmingCourseShortName] = nullableCourseText(*input.ShortName)
	}
	if input.ImageObjectKey != nil {
		fields[colTrimmingCourseImageObjectKey] = nullableCourseText(*input.ImageObjectKey)
	}
	if input.AllowedWeekdays != nil {
		if len(*input.AllowedWeekdays) == 0 {
			fields[colTrimmingCourseAllowedWeekdays] = nil
		} else {
			fields[colTrimmingCourseAllowedWeekdays] = toWeekdayArray(*input.AllowedWeekdays)
		}
	}
	return fields
}

// nullableCourseText maps an already-normalized optional text to its column
// value: "" is stored as NULL because the CHECK constraints reject blank text.
func nullableCourseText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func toWeekdayArray(days []int) pq.Int64Array {
	out := make(pq.Int64Array, len(days))
	for i, day := range days {
		out[i] = int64(day)
	}
	return out
}

// normalizeCourseWeekdays validates 0=Sunday..6=Saturday and returns a sorted,
// de-duplicated copy. An empty result means every weekday (stored as NULL).
func normalizeCourseWeekdays(days []int) ([]int, error) {
	seen := make(map[int]bool, len(days))
	out := make([]int, 0, len(days))
	for _, day := range days {
		if day < minTrimmingCourseWeekday || day > maxTrimmingCourseWeekday {
			return nil, apperrors.WrapInvalidInput("allowed_weekdays must be between 0 (Sunday) and 6 (Saturday)")
		}
		if !seen[day] {
			seen[day] = true
			out = append(out, day)
		}
	}
	sort.Ints(out)
	return out, nil
}

// normalizeCourseImageKey accepts only an object key owned by the clinic
// (the trimming image upload contract). URLs, legacy values and other clinics'
// keys are rejected; "" means no image.
func normalizeCourseImageKey(clinicID uint64, value string) (string, error) {
	normalized, err := normalizeTrimmingImageForPersist(clinicID, value)
	if err != nil {
		return "", err
	}
	if normalized == "" {
		return "", nil
	}
	key, err := ownedImageKey(clinicID, normalized)
	if err != nil {
		return "", err
	}
	return key.value, nil
}

type normalizedCourseAttributes struct {
	shortName string
	imageKey  string
	weekdays  []int
}

func normalizeCourseAttributes(clinicID uint64, shortName, imageKey string, weekdays []int) (normalizedCourseAttributes, error) {
	days, err := normalizeCourseWeekdays(weekdays)
	if err != nil {
		return normalizedCourseAttributes{}, err
	}
	key, err := normalizeCourseImageKey(clinicID, imageKey)
	if err != nil {
		return normalizedCourseAttributes{}, err
	}
	return normalizedCourseAttributes{shortName: strings.TrimSpace(shortName), imageKey: key, weekdays: days}, nil
}

// normalizeUpdateCourseAttributes returns a copy of input whose attribute
// pointers hold normalized values; nil pointers (keep) stay nil.
func normalizeUpdateCourseAttributes(clinicID uint64, input *UpdateTrimmingCourseInput) (UpdateTrimmingCourseInput, error) {
	out := *input
	if input.ShortName != nil {
		trimmed := strings.TrimSpace(*input.ShortName)
		out.ShortName = &trimmed
	}
	if input.ImageObjectKey != nil {
		key, err := normalizeCourseImageKey(clinicID, *input.ImageObjectKey)
		if err != nil {
			return UpdateTrimmingCourseInput{}, err
		}
		out.ImageObjectKey = &key
	}
	if input.AllowedWeekdays != nil {
		days, err := normalizeCourseWeekdays(*input.AllowedWeekdays)
		if err != nil {
			return UpdateTrimmingCourseInput{}, err
		}
		out.AllowedWeekdays = &days
	}
	return out, nil
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
	imageLifecycle ImageLifecycleRepository
}

// TrimmingCourseServiceOption configures optional trimming course dependencies.
type TrimmingCourseServiceOption func(*trimmingCourseService)

// WithCourseImageLifecycle wires the shared trimming image key lock/tombstone
// protocol. Without it, writes that set a course image fail closed.
func WithCourseImageLifecycle(lifecycle ImageLifecycleRepository) TrimmingCourseServiceOption {
	return func(s *trimmingCourseService) { s.imageLifecycle = lifecycle }
}

// NewTrimmingCourseService は TrimmingCourseService を生成する
func NewTrimmingCourseService(
	repo TrimmingCourseRepository,
	courseTypeRepo TrimmingCourseTypeRepository,
	transactor Transactor,
	opts ...TrimmingCourseServiceOption,
) TrimmingCourseService {
	s := &trimmingCourseService{
		repo:           repo,
		courseTypeRepo: courseTypeRepo,
		transactor:     transactor,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// lockCourseImageInTx keeps a newly referenced image key alive until commit:
// the same advisory lock as trimming records serializes it with deletion
// reservations, and a tombstoned key must be uploaded again.
func (s *trimmingCourseService) lockCourseImageInTx(ctx context.Context, clinicID uint64, key string) error {
	if key == "" {
		return nil
	}
	if s.imageLifecycle == nil {
		return apperrors.WrapInternalServerError("trimming course image lifecycle is not configured")
	}
	imageKey := ImageKey{clinicID: clinicID, value: key}
	if err := s.imageLifecycle.LockImageKeys(ctx, []ImageKey{imageKey}); err != nil {
		return apperrors.Wrap(err, "failed to lock trimming course image")
	}
	deleted, err := s.imageLifecycle.ImageKeyDeleted(ctx, imageKey)
	if err != nil {
		return apperrors.Wrap(err, "failed to check trimming course image lifecycle")
	}
	if deleted {
		return apperrors.WrapConflict("trimming image has been deleted; upload it again")
	}
	return nil
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
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateNonNegativePrice(input.Price); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate non negative price")
	}
	attrs, err := normalizeCourseAttributes(clinicID, input.ShortName, input.ImageObjectKey, input.AllowedWeekdays)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to validate trimming course attributes")
	}
	sortOrder := input.SortOrder
	course := &model.TrimmingCourse{
		ClinicID:     clinicID,
		Name:         input.Name,
		Price:        input.Price,
		IsActive:     input.IsActive,
		Description:  input.Description,
		CourseTypeID: input.CourseTypeID,
		Duration:     input.Duration,
		SortOrder:    &sortOrder,
	}
	if attrs.shortName != "" {
		course.ShortName = &attrs.shortName
	}
	if attrs.imageKey != "" {
		course.ImageObjectKey = &attrs.imageKey
	}
	if len(attrs.weekdays) > 0 {
		days := toWeekdayArray(attrs.weekdays)
		course.AllowedWeekdays = &days
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
		if err := s.lockCourseImageInTx(txCtx, clinicID, attrs.imageKey); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, course); err != nil {
			return apperrors.Wrap(err, "failed to create trimming course")
		}
		return nil
	}); err != nil {
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
	normalized, err := normalizeUpdateCourseAttributes(clinicID, input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to validate trimming course attributes")
	}
	input = &normalized
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
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
		if err := validateNonNegativePrice(input.Price); err != nil {
			return apperrors.Wrap(err, "failed to validate non negative price")
		}
		fields := buildTrimmingCourseUpdate(input)
		if len(fields) == 0 {
			return apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
		}
		if input.ImageObjectKey != nil {
			if err := s.lockCourseImageInTx(txCtx, clinicID, *input.ImageObjectKey); err != nil {
				return err
			}
		}
		updated, err := s.repo.Update(txCtx, clinicID, id, *input)
		if err != nil {
			return apperrors.Wrap(err, "failed to update trimming course")
		}
		course = updated
		return nil
	}); err != nil {
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
		count, err := s.repo.CountUsageByTrimmingCourseID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check trimming course dependencies")
		}
		if count > 0 {
			return apperrors.WrapConflict("このトリミングコースはトリミング記録で使用中のため削除できません")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete trimming course")
		}
		return nil
	}); err != nil {
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
		return apperrors.Wrap(err, "failed to reorder trimming courses")
	}
	slog.InfoContext(ctx, "trimming courses reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}
