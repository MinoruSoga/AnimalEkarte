// Package service provides business logic implementations for ReservationType entity.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- Input DTOs ----

// CreateReservationTypeInput はサービス種別作成のための入力データ
type CreateReservationTypeInput struct {
	Name        string
	Color       string
	IsActive    bool
	Description string
	SortOrder   int
	GroupID     *uint64
	Category    string

	// LINE予約用フィールド
	ReservationDisplayName string
	DurationMinutes        *int
	ShortName              string
	ShowShortName          bool
	ReservationVisible     *bool
	ReservationComment     string
	ReservationImageURL    string
	ReservationDayOption   string
	IsInternal             bool
}

// UpdateReservationTypeInput はサービス種別更新のための入力データ（ポインタ型でゼロ値を区別する）
type UpdateReservationTypeInput struct {
	Name         *string
	Color        *string
	IsActive     *bool
	Description  *string
	SortOrder    *int
	GroupID      *uint64
	ClearGroupID bool // true のとき group_id を NULL にクリアする
	Category     *string

	// LINE予約用フィールド
	ReservationDisplayName *string
	DurationMinutes        *int
	ShortName              *string
	ShowShortName          *bool
	ReservationVisible     *bool
	ReservationComment     *string
	ReservationImageURL    *string
	ReservationDayOption   *string
	IsInternal             *bool
}

// ---- DB column constants ----

const (
	colReservationTypeName                = "name"
	colReservationTypeColor               = "color"
	colReservationTypeIsActive            = "is_active"
	colReservationTypeDescription         = "description"
	colReservationTypeSortOrder           = "sort_order"
	colReservationTypeCategory            = "category"
	colReservationTypeReservationDispName = "reservation_display_name"
	colReservationTypeDurationMinutes     = "duration_minutes"
	colReservationTypeShortName           = "short_name"
	colReservationTypeShowShortName       = "show_short_name"
	colReservationTypeReservationVisible  = "reservation_visible"
	colReservationTypeReservationComment  = "reservation_comment"
	colReservationTypeReservationDayOpt   = "reservation_day_option"
	colReservationTypeIsInternal          = "is_internal"
	colReservationTypeReservationImageURL = "reservation_image_url"
	colReservationTypeGroupID             = "group_id"
)

// buildReservationTypeUpdateFields は UpdateReservationTypeInput から nil でないフィールドのみ map に変換する
func buildReservationTypeUpdateFields(input *UpdateReservationTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colReservationTypeName] = *input.Name
	}
	if input.Color != nil {
		fields[colReservationTypeColor] = *input.Color
	}
	if input.IsActive != nil {
		fields[colReservationTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colReservationTypeDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colReservationTypeSortOrder] = *input.SortOrder
	}
	if input.Category != nil {
		fields[colReservationTypeCategory] = model.ReservationTypeCategory(*input.Category)
	}
	if input.ReservationDisplayName != nil {
		fields[colReservationTypeReservationDispName] = *input.ReservationDisplayName
	}
	if input.DurationMinutes != nil {
		fields[colReservationTypeDurationMinutes] = *input.DurationMinutes
	}
	if input.ShortName != nil {
		fields[colReservationTypeShortName] = *input.ShortName
	}
	if input.ShowShortName != nil {
		fields[colReservationTypeShowShortName] = *input.ShowShortName
	}
	if input.ReservationVisible != nil {
		fields[colReservationTypeReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colReservationTypeReservationComment] = *input.ReservationComment
	}
	if input.ReservationImageURL != nil {
		fields[colReservationTypeReservationImageURL] = *input.ReservationImageURL
	}
	if input.ReservationDayOption != nil {
		fields[colReservationTypeReservationDayOpt] = *input.ReservationDayOption
	}
	if input.IsInternal != nil {
		fields[colReservationTypeIsInternal] = *input.IsInternal
	}
	if input.ClearGroupID {
		fields[colReservationTypeGroupID] = nil
	} else if input.GroupID != nil {
		fields[colReservationTypeGroupID] = *input.GroupID
	}
	return fields
}

// ---- ReservationTypeService ----

// CreateUnavailableTimeInput は予約不可時間の作成入力DTO
type CreateUnavailableTimeInput struct {
	UnavailableType string
	DayOfWeek       *int8
	SpecificDate    *time.Time
	StartTime       string
	EndTime         string
}

// ReservationTypeCoreService は予約種別の CRUD・Reorder を担う。
type ReservationTypeCoreService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// ReservationTypeUnavailableTimeService は予約不可時間帯の管理を担う。
type ReservationTypeUnavailableTimeService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	ListUnavailableTimes(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	CreateUnavailableTime(ctx context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error)
	DeleteUnavailableTime(ctx context.Context, clinicID, reservationTypeID, id uint64) error
}

// ReservationTypeOccupationService は診療科目ひもづけを担う。
type ReservationTypeOccupationService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	ListOccupations(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
	UnlinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
}

// ReservationTypeService は3つのサブインターフェースを合成した完全インターフェース。
// handler 層はこのインターフェースを通じてすべての操作を行う。
type ReservationTypeService interface { //nolint:revive // ReservationType is a domain entity name, cannot avoid stutter
	ReservationTypeCoreService
	ReservationTypeUnavailableTimeService
	ReservationTypeOccupationService
}

type reservationTypeService struct {
	repo                repository.ReservationTypeRepository
	reservationRepo     repository.ReservationQueryRepository
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository
	occupationRepo      repository.ReservationTypeOccupationRepository
	baseOccupationRepo  repository.OccupationRepository
}

func NewReservationTypeService(
	repo repository.ReservationTypeRepository,
	reservationRepo repository.ReservationQueryRepository,
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository,
	occupationRepo repository.ReservationTypeOccupationRepository,
	baseOccupationRepo repository.OccupationRepository,
) ReservationTypeService {
	return &reservationTypeService{
		repo:                repo,
		reservationRepo:     reservationRepo,
		unavailableTimeRepo: unavailableTimeRepo,
		occupationRepo:      occupationRepo,
		baseOccupationRepo:  baseOccupationRepo,
	}
}

func (s *reservationTypeService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list service types")
	}
	return items, nil
}

func (s *reservationTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get service type")
	}
	return result, nil
}

func (s *reservationTypeService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	reservationDayOption := model.ReservationDayOption(input.ReservationDayOption)
	if reservationDayOption == "" {
		reservationDayOption = model.DayOptionNone
	}
	durationMinutes := 15
	if input.DurationMinutes != nil {
		durationMinutes = *input.DurationMinutes
	}
	reservationVisible := true
	if input.ReservationVisible != nil {
		reservationVisible = *input.ReservationVisible
	}
	category := model.ReservationTypeCategory(input.Category)
	if category == "" {
		category = model.ReservationTypeCategoryGeneral
	}

	st := &model.ReservationType{
		ClinicID:               clinicID,
		Name:                   input.Name,
		Color:                  input.Color,
		IsActive:               input.IsActive,
		Description:            input.Description,
		SortOrder:              input.SortOrder,
		Category:               category,
		ReservationDisplayName: input.ReservationDisplayName,
		DurationMinutes:        durationMinutes,
		ShortName:              input.ShortName,
		ShowShortName:          input.ShowShortName,
		ReservationVisible:     reservationVisible,
		ReservationComment:     input.ReservationComment,
		ReservationImageURL:    input.ReservationImageURL,
		ReservationDayOption:   reservationDayOption,
		IsInternal:             input.IsInternal,
		GroupID:                input.GroupID,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, apperrors.Wrap(err, "failed to create service type")
	}
	slog.InfoContext(ctx, "service type created", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", st.ID))
	return st, nil
}

func (s *reservationTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildReservationTypeUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update service type")
	}
	slog.InfoContext(ctx, "service type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", id))
	return result, nil
}

func (s *reservationTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	exists, err := s.reservationRepo.ExistsByReservationTypeID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if exists {
		return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete service type")
	}
	slog.InfoContext(ctx, "service type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", id))
	return nil
}

func (s *reservationTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder service types")
	}
	slog.InfoContext(ctx, "reservation_types reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

// ---- 予約不可時間 ----

func (s *reservationTypeService) ListUnavailableTimes(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	// 予約区分の存在確認
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	items, err := s.unavailableTimeRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list unavailable times")
	}
	return items, nil
}

func (s *reservationTypeService) CreateUnavailableTime(ctx context.Context, clinicID, reservationTypeID uint64, input CreateUnavailableTimeInput) (*model.ReservationTypeUnavailableTime, error) {
	// 予約区分の存在確認
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	// 種別バリデーション
	if err := validateUnavailableTimeInput(input); err != nil {
		return nil, err
	}
	// 重複チェック
	existing, err := s.unavailableTimeRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check existing unavailable times")
	}
	if err := checkUnavailableTimeOverlap(existing, input); err != nil {
		return nil, err
	}

	t := &model.ReservationTypeUnavailableTime{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		UnavailableType:   model.UnavailableType(input.UnavailableType),
		DayOfWeek:         input.DayOfWeek,
		SpecificDate:      input.SpecificDate,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
	}
	if err := s.unavailableTimeRepo.Create(ctx, t); err != nil {
		return nil, apperrors.Wrap(err, "failed to create unavailable time")
	}
	slog.InfoContext(ctx, "unavailable time created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID))
	return t, nil
}

func (s *reservationTypeService) DeleteUnavailableTime(ctx context.Context, clinicID, reservationTypeID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return apperrors.Wrap(err, "failed to get reservation type")
	}
	if err := s.unavailableTimeRepo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete unavailable time")
	}
	slog.InfoContext(ctx, "unavailable time deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID),
		slog.Uint64("unavailable_time_id", id))
	return nil
}

// ---- 職種紐付け ----

func (s *reservationTypeService) ListOccupations(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	items, err := s.occupationRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list occupations")
	}
	return items, nil
}

func (s *reservationTypeService) LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	if _, err := s.baseOccupationRepo.FindByID(ctx, clinicID, occupationID); err != nil {
		return nil, apperrors.Wrap(err, "occupation not found")
	}
	o := &model.ReservationTypeOccupation{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		OccupationID:      occupationID,
	}
	if err := s.occupationRepo.Create(ctx, o); err != nil {
		return nil, apperrors.Wrap(err, "failed to link occupation")
	}
	slog.InfoContext(ctx, "occupation linked",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID),
		slog.Uint64("occupation_id", occupationID))
	// Preload して返す
	items, err := s.occupationRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get linked occupation")
	}
	for i := range items {
		if items[i].OccupationID == occupationID {
			return &items[i], nil
		}
	}
	return o, nil
}

func (s *reservationTypeService) UnlinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error {
	if err := s.occupationRepo.Delete(ctx, clinicID, reservationTypeID, occupationID); err != nil {
		return apperrors.Wrap(err, "failed to unlink occupation")
	}
	slog.InfoContext(ctx, "occupation unlinked",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID),
		slog.Uint64("occupation_id", occupationID))
	return nil
}

// ---- バリデーションヘルパー ----

// validateUnavailableTimeInput は CreateUnavailableTimeInput の整合性を検証する
func validateUnavailableTimeInput(input CreateUnavailableTimeInput) error {
	switch input.UnavailableType {
	case string(model.UnavailableTypeWeekly):
		if input.DayOfWeek == nil {
			return apperrors.WrapInvalidInput("曜日の指定は必須です")
		}
		if *input.DayOfWeek < 0 || *input.DayOfWeek > 6 {
			return apperrors.WrapInvalidInput("day_of_week must be between 0 (Sun) and 6 (Sat)")
		}
	case string(model.UnavailableTypeSpecific):
		if input.SpecificDate == nil {
			return apperrors.WrapInvalidInput("specific_date is required for specific type")
		}
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("不正な予約不可時間種別です: %s", input.UnavailableType))
	}
	// StartTime < EndTime（VARCHAR(5) "HH:MM" は辞書順比較で正しく機能する）
	if input.StartTime >= input.EndTime {
		return apperrors.WrapInvalidInput("開始時刻は終了時刻より前に指定してください")
	}
	return nil
}

// checkUnavailableTimeOverlap は既存設定との時間帯重複を検証する
// weekly と specific の混在は許可（LIFF で specific が weekly より優先される）
func checkUnavailableTimeOverlap(existing []model.ReservationTypeUnavailableTime, input CreateUnavailableTimeInput) error {
	for i := range existing {
		if string(existing[i].UnavailableType) != input.UnavailableType {
			continue
		}
		switch existing[i].UnavailableType {
		case model.UnavailableTypeWeekly:
			if input.DayOfWeek == nil || existing[i].DayOfWeek == nil || *existing[i].DayOfWeek != *input.DayOfWeek {
				continue
			}
		case model.UnavailableTypeSpecific:
			if input.SpecificDate == nil || existing[i].SpecificDate == nil {
				continue
			}
			// DATE 型は UTC 午前0時で格納されるため日付文字列で比較
			if existing[i].SpecificDate.UTC().Format("2006-01-02") != input.SpecificDate.UTC().Format("2006-01-02") {
				continue
			}
		}
		// 時間帯が交差するか（start < other.end && end > other.start）
		if input.StartTime < existing[i].EndTime && input.EndTime > existing[i].StartTime {
			return apperrors.WrapConflict("指定した時間帯は既存の予約不可時間と重複しています")
		}
	}
	return nil
}
