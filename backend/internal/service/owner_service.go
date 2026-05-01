package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- Owner DB column constants ---

const (
	colOwnerName              = "name"
	colOwnerNameKana          = "name_kana"
	colBirthDate              = "birth_date"
	colCompany                = "company"
	colPostalCode             = "postal_code"
	colAddress1               = "address1"
	colAddress2               = "address2"
	colHomePostalCode         = "home_postal_code"
	colHomeAddress1           = "home_address1"
	colHomeAddress2           = "home_address2"
	colPhone                  = "phone"
	colCompanyPhone           = "company_phone"
	colEmail                  = "email"
	colRemarks                = "remarks"
	colIsDangerous            = "is_dangerous"
	colDiscountRate           = "discount_rate"
	colMembershipType         = "membership_type"
	colLstepOptOut            = "lstep_opt_out"
	colLstepOptOutAt          = "lstep_opt_out_at"
	colLstepOptOutReason      = "lstep_opt_out_reason"
	colLineIDConfirmedAt      = "line_id_confirmed_at"
	colDeliveryExcluded       = "delivery_excluded"
	colDeliveryExcludedReason = "delivery_excluded_reason"
	colIsTransferred          = "is_transferred"
	colTransferAt             = "transfer_at"
)

// --- Input DTOs（Service層の公開I/F） ---

// CreatePetForOwnerInput は飼主登録時に同時作成するペットの入力DTO
type CreatePetForOwnerInput struct {
	Name            string
	AnimalSpeciesID uint64
	PetNameKana     string
	Breed           string
	Color           string
	Gender          string
	Status          string
	BirthDate       *time.Time
	Weight          *float64
	NeuteredDate    *time.Time
	AcquisitionType string
	DangerLevel     string
	Food            string
	Environment     string
	InsuranceID     *uint64
	Remarks         string
}

// CreateOwnerInput は飼主作成の入力DTO
// ビジネスルール（DiscountRate 範囲, MembershipType 列挙値）は Service 層で検証する。
type CreateOwnerInput struct {
	OwnerName      string
	OwnerNameKana  string
	BirthDate      *time.Time
	Company        string
	PostalCode     string
	Address1       string
	Address2       string
	HomePostalCode string
	HomeAddress1   string
	HomeAddress2   string
	Phone          string
	CompanyPhone   string
	Email          string
	Remarks        string
	IsDangerous    bool
	DiscountRate   float64
	MembershipType model.MembershipType
	Pets           []CreatePetForOwnerInput
}

// UpdateOwnerInput は飼主更新の入力DTO（全フィールドポインタ型: nil = 未指定, 非nil = 更新対象）
// ビジネスルールは Service 層で検証する。
type UpdateOwnerInput struct {
	OwnerName      *string
	OwnerNameKana  *string
	BirthDate      *time.Time
	Company        *string
	PostalCode     *string
	Address1       *string
	Address2       *string
	HomePostalCode *string
	HomeAddress1   *string
	HomeAddress2   *string
	Phone          *string
	CompanyPhone   *string
	Email          *string
	Remarks        *string
	IsDangerous    *bool
	DiscountRate   *float64
	MembershipType *model.MembershipType
}

// UpdateDeliveryExclusionInput は配信除外フラグ更新の入力DTO（FEAT-381）
type UpdateDeliveryExclusionInput struct {
	Excluded bool
	Reason   *string
}

// UpdateTransferStatusInput は転院フラグ更新の入力DTO（FEAT-381）
type UpdateTransferStatusInput struct {
	IsTransferred bool
}

// buildOwnerUpdate はポインタが非 nil のフィールドのみ map に追加する
func buildOwnerUpdate(input *UpdateOwnerInput) map[string]any {
	fields := make(map[string]any)
	if input.OwnerName != nil {
		fields[colOwnerName] = *input.OwnerName
	}
	if input.OwnerNameKana != nil {
		fields[colOwnerNameKana] = *input.OwnerNameKana
	}
	if input.BirthDate != nil {
		fields[colBirthDate] = *input.BirthDate
	}
	if input.Company != nil {
		fields[colCompany] = *input.Company
	}
	if input.PostalCode != nil {
		fields[colPostalCode] = *input.PostalCode
	}
	if input.Address1 != nil {
		fields[colAddress1] = *input.Address1
	}
	if input.Address2 != nil {
		fields[colAddress2] = *input.Address2
	}
	if input.HomePostalCode != nil {
		fields[colHomePostalCode] = *input.HomePostalCode
	}
	if input.HomeAddress1 != nil {
		fields[colHomeAddress1] = *input.HomeAddress1
	}
	if input.HomeAddress2 != nil {
		fields[colHomeAddress2] = *input.HomeAddress2
	}
	if input.Phone != nil {
		fields[colPhone] = *input.Phone
	}
	if input.CompanyPhone != nil {
		fields[colCompanyPhone] = *input.CompanyPhone
	}
	if input.Email != nil {
		fields[colEmail] = *input.Email
	}
	if input.Remarks != nil {
		fields[colRemarks] = *input.Remarks
	}
	if input.IsDangerous != nil {
		fields[colIsDangerous] = *input.IsDangerous
	}
	if input.DiscountRate != nil {
		fields[colDiscountRate] = *input.DiscountRate
	}
	if input.MembershipType != nil {
		fields[colMembershipType] = *input.MembershipType
	}
	return fields
}

// --- Interface ---

type OwnerService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// LinkLineUserID は飼主の LINE User ID を設定または解除する（BE-005）。nil で解除。
	LinkLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error
	// UpdateDeliveryExclusion は配信除外フラグと理由を更新し、Lステップタグを同期する（FEAT-381）。
	UpdateDeliveryExclusion(ctx context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error)
	// UpdateTransferStatus は転院フラグと転院日時を更新し、Lステップタグを同期する（FEAT-381）。
	UpdateTransferStatus(ctx context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error)
	// ConfirmLineID は LINE ID 紐付け確認日時を現在時刻に設定する（FEAT-381）。
	ConfirmLineID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
}

// --- Implementation ---

type ownerService struct {
	repo       repository.OwnerRepository
	tagSyncSvc LstepTagSyncService
}

func NewOwnerService(repo repository.OwnerRepository, tagSyncSvc LstepTagSyncService) OwnerService {
	return &ownerService{repo: repo, tagSyncSvc: tagSyncSvc}
}

func (s *ownerService) List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	owners, total, err := s.repo.FindAll(ctx, clinicID, page, limit, search)
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

func (s *ownerService) CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
	// 名前バリデーション（スペースのみ・NULL バイト・制御文字チェック）
	if err := validateRequiredName(input.OwnerName); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	input.OwnerName = strings.TrimSpace(input.OwnerName)

	// メールアドレス形式バリデーション（空でない場合）
	if err := validateEmailFormat(input.Email); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate email format")
	}

	// 電話番号形式バリデーション（空でない場合）
	if err := validatePhoneFormat(input.Phone); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate phone format")
	}

	// 郵便番号形式バリデーション（空でない場合）
	if err := validatePostalCodeFormat(input.PostalCode); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate postal code format")
	}

	// ビジネスルールバリデーション
	if err := validateDiscountRate(input.DiscountRate); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate discount rate")
	}
	if err := validateMembershipType(input.MembershipType); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate membership type")
	}
	for i := range input.Pets {
		if err := validatePetGender(input.Pets[i].Gender); err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
		if err := validatePetStatus(input.Pets[i].Status); err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
		if err := validatePetAcquisitionType(input.Pets[i].AcquisitionType); err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
		if err := validatePetDangerLevel(input.Pets[i].DangerLevel); err != nil {
			return nil, apperrors.Wrap(err, fmt.Sprintf("pets[%d]", i))
		}
	}

	// メールアドレス重複チェック（空でない場合のみ）
	if input.Email != "" {
		existing, err := s.repo.FindByEmail(ctx, clinicID, input.Email)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check email uniqueness", "error", err)
			return nil, apperrors.Wrap(err, "failed to check email uniqueness")
		}
		if existing != nil {
			return nil, apperrors.WrapAlreadyExists("owner", "このメールアドレスはすでに登録されています")
		}
	}

	// BUG-064: 電話番号重複チェック（空でない場合のみ）
	if input.Phone != "" {
		existing, err := s.repo.FindByPhone(ctx, clinicID, input.Phone)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check phone uniqueness", "error", err)
			return nil, apperrors.Wrap(err, "failed to check phone uniqueness")
		}
		if existing != nil {
			return nil, apperrors.WrapAlreadyExists("owner", "この電話番号はすでに登録されています")
		}
	}

	// DTO → Model 変換
	membershipType := input.MembershipType
	if membershipType == "" {
		membershipType = model.MembershipTypeNonMember
	}

	owner := &model.Owner{
		ClinicID:       clinicID,
		Name:           input.OwnerName,
		NameKana:       input.OwnerNameKana,
		BirthDate:      input.BirthDate,
		Company:        input.Company,
		PostalCode:     input.PostalCode,
		Address1:       input.Address1,
		Address2:       input.Address2,
		HomePostalCode: input.HomePostalCode,
		HomeAddress1:   input.HomeAddress1,
		HomeAddress2:   input.HomeAddress2,
		Phone:          input.Phone,
		CompanyPhone:   input.CompanyPhone,
		Email:          input.Email,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: membershipType,
	}

	pets := make([]model.Pet, 0, len(input.Pets))
	for i := range input.Pets {
		p := &input.Pets[i]
		pet := model.Pet{
			Name:            p.Name,
			AnimalSpeciesID: p.AnimalSpeciesID,
			NameKana:        p.PetNameKana,
			Breed:           p.Breed,
			Color:           p.Color,
			BirthDate:       p.BirthDate,
			Weight:          p.Weight,
			NeuteredDate:    p.NeuteredDate,
			Food:            p.Food,
			Environment:     p.Environment,
			InsuranceID:     p.InsuranceID,
			Remarks:         p.Remarks,
		}
		if p.Gender != "" {
			pet.Gender = model.PetGender(p.Gender)
		}
		if p.Status != "" {
			pet.Status = model.PetStatus(p.Status)
		}
		if p.AcquisitionType != "" {
			at := model.AcquisitionType(p.AcquisitionType)
			pet.AcquisitionType = &at
		}
		if p.DangerLevel != "" {
			pet.DangerLevel = model.DangerLevel(p.DangerLevel)
		}
		pets = append(pets, pet)
	}

	if err := s.repo.CreateWithPets(ctx, owner, pets); err != nil {
		slog.ErrorContext(ctx, "failed to create owner with pets", "error", err)
		return nil, apperrors.Wrap(err, "failed to create owner with pets")
	}

	slog.InfoContext(ctx, "owner created with pets",
		slog.Uint64("owner_id", owner.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Int("pets_count", len(pets)))

	return owner, nil
}

func (s *ownerService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to find owner", "error", err)
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	// 名前バリデーション（スペースのみ・NULL バイト・制御文字チェック）
	if input.OwnerName != nil {
		if err := validateRequiredName(*input.OwnerName); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate required name")
		}
		trimmed := strings.TrimSpace(*input.OwnerName)
		input.OwnerName = &trimmed
	}

	// メールアドレス形式バリデーション（指定されている場合）
	if input.Email != nil {
		if err := validateEmailFormat(*input.Email); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate email format")
		}
	}

	// 電話番号形式バリデーション（指定されている場合）
	if input.Phone != nil {
		if err := validatePhoneFormat(*input.Phone); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate phone format")
		}
	}

	// 郵便番号形式バリデーション（指定されている場合）
	if input.PostalCode != nil {
		if err := validatePostalCodeFormat(*input.PostalCode); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate postal code format")
		}
	}

	// ビジネスルールバリデーション
	if input.DiscountRate != nil {
		if err := validateDiscountRate(*input.DiscountRate); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate discount rate")
		}
	}
	if input.MembershipType != nil {
		if err := validateMembershipType(*input.MembershipType); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate membership type")
		}
	}

	// メールアドレス変更時の重複チェック（空でない場合のみ）
	if input.Email != nil && *input.Email != "" {
		existing, err := s.repo.FindByEmail(ctx, clinicID, *input.Email)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check email uniqueness", "error", err)
			return nil, apperrors.Wrap(err, "failed to check email uniqueness")
		}
		if existing != nil && existing.ID != id {
			return nil, apperrors.WrapAlreadyExists("owner", "このメールアドレスはすでに登録されています")
		}
	}

	// BUG-064: 電話番号変更時の重複チェック（空でない場合のみ）
	if input.Phone != nil && *input.Phone != "" {
		existing, err := s.repo.FindByPhone(ctx, clinicID, *input.Phone)
		if err != nil {
			slog.ErrorContext(ctx, "failed to check phone uniqueness", "error", err)
			return nil, apperrors.Wrap(err, "failed to check phone uniqueness")
		}
		if existing != nil && existing.ID != id {
			return nil, apperrors.WrapAlreadyExists("owner", "この電話番号はすでに登録されています")
		}
	}

	// 更新フィールドマップ構築（nil フィールドはスキップ）
	fields := buildOwnerUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update owner", "error", err)
		return nil, apperrors.Wrap(err, "failed to update owner")
	}

	slog.InfoContext(ctx, "owner updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))

	// DB の最新状態を返す
	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get updated owner", "error", err)
		return nil, apperrors.Wrap(err, "failed to get updated owner")
	}
	return owner, nil
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

func (s *ownerService) LinkLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}
	if err := s.repo.UpdateLineUserID(ctx, clinicID, id, lineUserID); err != nil {
		slog.ErrorContext(ctx, "failed to link line user id", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to link line user id")
	}
	return nil
}

func (s *ownerService) UpdateDeliveryExclusion(ctx context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	var reason *string
	if input.Excluded && input.Reason != nil {
		trimmed := strings.TrimSpace(*input.Reason)
		if len([]rune(trimmed)) > 100 {
			return nil, apperrors.WrapInvalidInput("delivery_excluded_reason must be 100 characters or less")
		}
		if trimmed != "" {
			reason = &trimmed
		}
	}

	var optOutAt any
	if input.Excluded {
		optOutAt = time.Now()
	}
	fields := map[string]any{
		colDeliveryExcluded:       input.Excluded,
		colDeliveryExcludedReason: reason,
		colLstepOptOut:            input.Excluded,
		colLstepOptOutAt:          optOutAt,
		colLstepOptOutReason:      reason,
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update delivery exclusion", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update delivery exclusion")
	}
	slog.InfoContext(ctx, "delivery exclusion updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("excluded", input.Excluded))
	if s.tagSyncSvc != nil {
		if err := s.tagSyncSvc.SyncExclusionTags(ctx, clinicID, id); err != nil {
			slog.ErrorContext(ctx, "failed to sync exclusion tag after delivery exclusion update", "error", err, "id", id, "clinic_id", clinicID)
		}
	}

	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reload owner after delivery exclusion update", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to reload owner")
	}
	return owner, nil
}

func (s *ownerService) UpdateTransferStatus(ctx context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error) {
	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	var transferAt any
	if input.IsTransferred {
		now := time.Now()
		transferAt = now
	}
	fields := map[string]any{
		colIsTransferred: input.IsTransferred,
		colTransferAt:    transferAt,
	}
	if input.IsTransferred {
		fields[colMembershipType] = model.MembershipTypeTransferred
	} else if owner.MembershipType == model.MembershipTypeTransferred {
		fields[colMembershipType] = model.MembershipTypeNonMember
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update transfer status", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update transfer status")
	}
	slog.InfoContext(ctx, "transfer status updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("is_transferred", input.IsTransferred))
	if s.tagSyncSvc != nil {
		if err := s.tagSyncSvc.SyncExclusionTags(ctx, clinicID, id); err != nil {
			slog.ErrorContext(ctx, "failed to sync exclusion tag after transfer status update", "error", err, "id", id, "clinic_id", clinicID)
		}
	}

	updated, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reload owner after transfer status update", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to reload owner")
	}
	return updated, nil
}

func (s *ownerService) ConfirmLineID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	owner, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LineUserID == nil || strings.TrimSpace(*owner.LineUserID) == "" {
		return nil, apperrors.WrapInvalidInput("LINE User ID is not linked")
	}
	now := time.Now()
	fields := map[string]any{
		colLineIDConfirmedAt: now,
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to confirm line id", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to confirm line id")
	}
	slog.InfoContext(ctx, "line id confirmed",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))
	updated, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reload owner after line id confirmation", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to reload owner")
	}
	return updated, nil
}
