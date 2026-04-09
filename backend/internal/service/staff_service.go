// Package service provides business logic implementations for Staff entity.
package service

import (
	"context"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- StaffService ----

// CreateStaffInput はスタッフ登録時の入力データ（AccountID 決定済みの場合に使用）。
type CreateStaffInput struct {
	ClinicID      uint64
	Name          string
	LicenseNumber string
	OccupationID  *uint64
	SortOrder     int
	AccountID     *uint64

	// LINE予約用フィールド
	StaffType              string
	ReservationDisplayName string
	ReservationVisible     bool
	ReservationComment     string
	ReservationImageURL    string
}

// CreateStaffWithAccountInput はアカウント（email/password）を同時に作成するスタッフ登録用入力DTO。
type CreateStaffWithAccountInput struct {
	ClinicID      uint64
	Name          string
	LicenseNumber string
	OccupationID  *uint64
	SortOrder     int
	Email         string
	Password      string

	// LINE予約用フィールド
	StaffType              string
	ReservationDisplayName string
	ReservationVisible     bool
	ReservationComment     string
	ReservationImageURL    string
}

// UpdateStaffInput はスタッフ部分更新の入力DTO。nil = 未送信フィールド。
type UpdateStaffInput struct {
	Name          *string
	LicenseNumber *string
	OccupationID  *uint64
	SortOrder     *int
	IsActive      *bool

	// LINE予約用フィールド
	StaffType              *string
	ReservationDisplayName *string
	ReservationVisible     *bool
	ReservationComment     *string
	ReservationImageURL    *string
}

type StaffService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	GetByID(ctx context.Context, id uint64) (*model.Staff, error)
	// FindByAccountID はアカウントIDに紐づくスタッフを返す。
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
	// Create はスタッフを作成する（AccountID を呼び出し元で決定済みの場合に使用）。
	Create(ctx context.Context, input *CreateStaffInput) (*model.Staff, error)
	// CreateWithAccount はメールアドレスとパスワードを受け取り、
	// email 重複チェック・bcrypt ハッシュ化・Account 作成・Staff 作成を一括で行う。
	CreateWithAccount(ctx context.Context, input *CreateStaffWithAccountInput) (*model.Staff, error)
	// UpdatePassword はスタッフに紐づくアカウントのパスワードをハッシュ化して更新する。
	UpdatePassword(ctx context.Context, accountID uint64, newPassword string) error
	// SetClinicAssignments はスタッフのクリニック割当をトランザクション内で差し替える。
	// 既存割当を全削除してから新しい clinicIDs を登録する。最初の1件は is_main=true となる。
	SetClinicAssignments(ctx context.Context, staffID uint64, clinicIDs []uint64) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	// GetPermissionGroupIDs はスタッフが所属する権限グループIDリストを返す
	GetPermissionGroupIDs(ctx context.Context, staffID uint64) ([]uint64, error)
	// SetPermissionGroupIDs はスタッフの権限グループを全置換する
	SetPermissionGroupIDs(ctx context.Context, staffID uint64, groupIDs []uint64) error
	// GetExcludedServiceTypeIDs はスタッフの除外サービス種別IDリストを返す
	GetExcludedServiceTypeIDs(ctx context.Context, staffID uint64) ([]uint64, error)
	// SetExcludedServiceTypeIDs はスタッフの除外サービス種別を全置換する
	SetExcludedServiceTypeIDs(ctx context.Context, staffID uint64, typeIDs []uint64) error
}

type staffService struct {
	repo                repository.StaffRepository
	accountRepo         repository.AccountRepository
	assignmentRepo      repository.StaffClinicAssignmentRepository
	reservationRepo     repository.ReservationRepository
	shiftEntryRepo      repository.ShiftEntryRepository
	permissionGroupRepo repository.PermissionGroupRepository
	resStaffRepo        repository.ReservationStaffRepository
}

func NewStaffService(
	repo repository.StaffRepository,
	accountRepo repository.AccountRepository,
	assignmentRepo repository.StaffClinicAssignmentRepository,
	reservationRepo repository.ReservationRepository,
	shiftEntryRepo repository.ShiftEntryRepository,
	permissionGroupRepo repository.PermissionGroupRepository,
	resStaffRepo repository.ReservationStaffRepository,
) StaffService {
	return &staffService{
		repo:                repo,
		accountRepo:         accountRepo,
		assignmentRepo:      assignmentRepo,
		reservationRepo:     reservationRepo,
		shiftEntryRepo:      shiftEntryRepo,
		permissionGroupRepo: permissionGroupRepo,
		resStaffRepo:        resStaffRepo,
	}
}

func (s *staffService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	staff, total, err := s.repo.FindAll(ctx, clinicID, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list staff")
	}
	return staff, total, nil
}

func (s *staffService) GetByID(ctx context.Context, id uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staff")
	}
	return staff, nil
}

func (s *staffService) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByAccountID(ctx, accountID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find staff by account id")
	}
	return staff, nil
}

func (s *staffService) Create(ctx context.Context, input *CreateStaffInput) (*model.Staff, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	input.Name = strings.TrimSpace(input.Name)

	staffType := model.StaffType(input.StaffType)
	if staffType == "" {
		staffType = model.StaffTypeDoctor
	}

	staff := &model.Staff{
		Name:                   input.Name,
		LicenseNumber:          input.LicenseNumber,
		OccupationID:           input.OccupationID,
		SortOrder:              input.SortOrder,
		IsActive:               true,
		AccountID:              input.AccountID,
		StaffType:              staffType,
		ReservationDisplayName: input.ReservationDisplayName,
		ReservationVisible:     input.ReservationVisible,
		ReservationComment:     input.ReservationComment,
		ReservationImageURL:    input.ReservationImageURL,
	}

	if err := s.repo.Create(ctx, staff); err != nil {
		return nil, apperrors.Wrap(err, "failed to create staff")
	}

	slog.InfoContext(ctx, "staff created", slog.Uint64("staff_id", staff.ID))
	return staff, nil
}

// CreateWithAccount はメール・パスワード付きでスタッフを作成する。
// email 重複チェック・bcrypt ハッシュ化・Account 作成・Staff 作成を一括で行う。
func (s *staffService) CreateWithAccount(ctx context.Context, input *CreateStaffWithAccountInput) (*model.Staff, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)

	// email 重複チェック: FindByEmail が NotFound 以外のエラーを返した場合は伝播する
	existing, err := s.accountRepo.FindByEmail(ctx, input.Email)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, apperrors.Wrap(err, "failed to check email uniqueness")
	}
	if existing != nil {
		return nil, apperrors.WrapAlreadyExists("account", input.Email)
	}

	hashed, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		return nil, apperrors.Wrap(hashErr, "failed to hash password")
	}

	account := &model.Account{
		Email:        input.Email,
		PasswordHash: string(hashed),
		IsActive:     true,
	}
	if createErr := s.accountRepo.Create(ctx, account); createErr != nil {
		return nil, apperrors.Wrap(createErr, "failed to create account")
	}

	slog.InfoContext(ctx, "account created for staff", slog.String("email", input.Email))

	staffType := model.StaffType(input.StaffType)
	if staffType == "" {
		staffType = model.StaffTypeDoctor
	}

	staff := &model.Staff{
		Name:                   name,
		LicenseNumber:          input.LicenseNumber,
		OccupationID:           input.OccupationID,
		SortOrder:              input.SortOrder,
		IsActive:               true,
		AccountID:              &account.ID,
		StaffType:              staffType,
		ReservationDisplayName: input.ReservationDisplayName,
		ReservationVisible:     input.ReservationVisible,
		ReservationComment:     input.ReservationComment,
		ReservationImageURL:    input.ReservationImageURL,
	}
	if err := s.repo.Create(ctx, staff); err != nil {
		return nil, apperrors.Wrap(err, "failed to create staff")
	}

	return staff, nil
}

// UpdatePassword はスタッフに紐づくアカウントのパスワードを更新する。
func (s *staffService) UpdatePassword(ctx context.Context, accountID uint64, newPassword string) error {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get account")
	}

	hashed, hashErr := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if hashErr != nil {
		return apperrors.Wrap(hashErr, "failed to hash password")
	}

	account.PasswordHash = string(hashed)
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return apperrors.Wrap(err, "failed to update account password")
	}

	slog.InfoContext(ctx, "password updated", slog.Uint64("account_id", accountID))
	return nil
}

// SetClinicAssignments はスタッフのクリニック割当をトランザクション内で差し替える。
func (s *staffService) SetClinicAssignments(ctx context.Context, staffID uint64, clinicIDs []uint64) error {
	if err := s.assignmentRepo.DeleteByStaffID(ctx, staffID); err != nil {
		return apperrors.Wrap(err, "failed to delete existing clinic assignments")
	}
	for i, clinicID := range clinicIDs {
		assignment := &model.StaffClinicAssignment{
			StaffID:  staffID,
			ClinicID: clinicID,
			IsMain:   i == 0,
		}
		if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
			return apperrors.Wrap(err, "failed to create clinic assignment")
		}
	}
	slog.InfoContext(ctx, "clinic assignments updated", slog.Uint64("staff_id", staffID), slog.Int("count", len(clinicIDs)))
	return nil
}

func (s *staffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error) {
	if input.Name != nil {
		if err := validateRequiredName(*input.Name); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}
	fields := buildStaffUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update staff")
	}
	slog.InfoContext(ctx, "staff updated", slog.Uint64("staff_id", id))
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get updated staff")
	}
	return updated, nil
}

func buildStaffUpdateFields(input *UpdateStaffInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.LicenseNumber != nil {
		fields["license_number"] = *input.LicenseNumber
	}
	if input.OccupationID != nil {
		fields["occupation_id"] = *input.OccupationID
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.StaffType != nil {
		fields["staff_type"] = *input.StaffType
	}
	if input.ReservationDisplayName != nil {
		fields["reservation_display_name"] = *input.ReservationDisplayName
	}
	if input.ReservationVisible != nil {
		fields["reservation_visible"] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields["reservation_comment"] = *input.ReservationComment
	}
	if input.ReservationImageURL != nil {
		fields["reservation_image_url"] = *input.ReservationImageURL
	}
	return fields
}

func (s *staffService) Delete(ctx context.Context, clinicID, id uint64) error {
	reservationExists, err := s.reservationRepo.ExistsByStaffID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if reservationExists {
		return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
	}
	shiftExists, err := s.shiftEntryRepo.ExistsByStaffID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check shift dependency")
	}
	if shiftExists {
		return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete staff")
	}
	slog.InfoContext(ctx, "staff deleted", slog.Uint64("staff_id", id), slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *staffService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder staff")
	}
	return nil
}

// GetPermissionGroupIDs はスタッフが所属する権限グループIDリストを返す
func (s *staffService) GetPermissionGroupIDs(ctx context.Context, staffID uint64) ([]uint64, error) {
	ids, err := s.permissionGroupRepo.GetGroupIDsByStaffID(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get permission group ids")
	}
	return ids, nil
}

// SetPermissionGroupIDs はスタッフの権限グループを全置換する
func (s *staffService) SetPermissionGroupIDs(ctx context.Context, staffID uint64, groupIDs []uint64) error {
	if err := s.permissionGroupRepo.SetStaffGroups(ctx, staffID, groupIDs); err != nil {
		return apperrors.Wrap(err, "failed to set permission group ids")
	}
	return nil
}

// GetExcludedServiceTypeIDs はスタッフの除外サービス種別IDリストを返す
func (s *staffService) GetExcludedServiceTypeIDs(ctx context.Context, staffID uint64) ([]uint64, error) {
	items, err := s.resStaffRepo.FindExcludedServiceTypes(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get excluded service type ids")
	}
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ServiceTypeID)
	}
	return ids, nil
}

// SetExcludedServiceTypeIDs はスタッフの除外サービス種別を全置換する
func (s *staffService) SetExcludedServiceTypeIDs(ctx context.Context, staffID uint64, typeIDs []uint64) error {
	if err := s.resStaffRepo.ReplaceExcludedServiceTypes(ctx, staffID, typeIDs); err != nil {
		return apperrors.Wrap(err, "failed to set excluded service type ids")
	}
	return nil
}
