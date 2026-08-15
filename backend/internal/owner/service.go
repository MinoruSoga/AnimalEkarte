package owner

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// --- Owner DB column constants ---

const (
	ownerDeliveryReasonMaxLength = 100

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
	colLineIDConfirmedBy      = "line_id_confirmed_by"
	colDeliveryExcluded       = "delivery_excluded"
	colDeliveryExcludedReason = "delivery_excluded_reason"
	colDeliveryCaution        = "delivery_caution"
	colDeliveryCautionReason  = "delivery_caution_reason"
	colIsTransferred          = "is_transferred"
	colTransferAt             = "transfer_at"
	colDMPreference           = "dm_preference"
)

// --- Input DTOs（Service層の公開I/F） ---

// CreatePetForOwnerInput は飼主登録時に同時作成するペットの入力DTO
type CreatePetForOwnerInput struct {
	Name            string
	AnimalSpeciesID uint64
	PetNameKana     string
	Breed           string
	Color           string
	BloodType       string
	MicrochipNumber string
	Gender          string
	Status          string
	BirthDate       *time.Time
	Weight          *float64
	NeuteredDate    *time.Time
	AcquisitionType string
	DangerLevel     string
	DangerReason    *string
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
	// DMPreference は DM 送付希望（#158）。nil=未設定 / true=必要 / false=不要。
	DMPreference *bool
	Pets         []CreatePetForOwnerInput
}

// UpdateOwnerInput は飼主更新の入力DTO（nil = 未指定, 非nil = 更新対象）
// ビジネスルールは Service 層で検証する。
type UpdateOwnerInput struct {
	OwnerName     *string
	OwnerNameKana *string
	// BirthDate は nil=未指定 / &nil=NULLクリア / &&value=更新対象。
	BirthDate      **time.Time
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
	// DMPreference は DM 送付希望（#158）。nil=未指定 / &nil=NULLクリア / &&value=更新対象。
	DMPreference **bool
	// DiscountEditAllowed is set by the HTTP boundary from discount:edit RBAC.
	// Service rechecks discount_rate against the FOR UPDATE locked row (SEC-CS-F15).
	DiscountEditAllowed bool
}

// UpdateDeliveryExclusionInput は配信除外フラグ更新の入力DTO（FEAT-381）
type UpdateDeliveryExclusionInput struct {
	Excluded bool
	Reason   *string
}

// UpdateDeliveryCautionInput は配信注意フラグ更新の入力DTO（FEAT-381-2 Commit 1）
type UpdateDeliveryCautionInput struct {
	Caution bool
	Reason  string
}

// UpdateTransferStatusInput は転院フラグ更新の入力DTO（FEAT-381）
type UpdateTransferStatusInput struct {
	IsTransferred bool
}

// --- Interface ---

// Service is the owner application capability consumed by HTTP and other
// application boundaries.
type Service interface {
	// List は指定した複数医院 (#86 拠点横断) の飼主一覧を返す。clinicIDs はハンドラ層で所属検証済みであること。
	List(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	// GetByIDForClinics は複数医院スコープで飼主を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error)
	CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// LinkLineUserID は飼主の LINE User ID を設定または解除する（BE-005）。nil で解除。
	// 同一クリニック内で重複する lineUserID がある場合は 409 Conflict を返す。
	// actorUserID は監査ログ記録用（nil 可）。
	LinkLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string, actorUserID *uint64) error
	// UpdateDeliveryExclusion は配信除外フラグと理由を更新し、Lステップタグを同期する（FEAT-381）。
	UpdateDeliveryExclusion(ctx context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error)
	// UpdateDeliveryCaution は配信注意フラグと理由を更新し、Lステップタグを同期する（FEAT-381-2）。
	UpdateDeliveryCaution(ctx context.Context, clinicID, id uint64, input UpdateDeliveryCautionInput) (*model.Owner, error)
	// UpdateTransferStatus は転院フラグと転院日時を更新し、Lステップタグを同期する（FEAT-381）。
	UpdateTransferStatus(ctx context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error)
	// ConfirmLineID は LINE ID 紐付け確認日時と確認者を設定する（FEAT-381 + Q22）。
	// actorUserID は確認者として line_id_confirmed_by に記録される（nil 可）。
	ConfirmLineID(ctx context.Context, clinicID, id uint64, actorUserID *uint64) (*model.Owner, error)
}

// InsuranceFinder is the minimal billing-master view needed to validate
// request-derived insurance IDs in nested pet registrations.
type InsuranceFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
}

// TagNotifier is the minimal LSTEP notification capability used by owner
// mutations. Notifications remain best-effort by the existing API contract.
type TagNotifier interface {
	SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error
	SyncExclusionTags(ctx context.Context, clinicID, ownerID uint64) error
}

// AuditLogger is the minimal audit capability used by LINE-owner operations.
type AuditLogger interface {
	LogLstepOperation(
		ctx context.Context,
		clinicID uint64,
		actorID *uint64,
		action, resource string,
		resourceID *uint64,
	) error
}

// PetOwnerCounter is the consumer-side view used to reject deletion of an
// owner who is referenced as a secondary pet owner.
type PetOwnerCounter interface {
	CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

// OwnerDeleteLocker serializes deletion with concurrent owner-dependent
// writers while preserving the owner package's dependency direction.
type OwnerDeleteLocker interface {
	LockForDelete(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

// OwnerDeleteTransactor owns the owner deletion transaction boundary.
type OwnerDeleteTransactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// --- Implementation ---

type ownerService struct {
	repo               ServiceRepository
	insuranceFinder    InsuranceFinder
	tagNotifier        TagNotifier
	auditLogger        AuditLogger
	deleteLocker       OwnerDeleteLocker
	petOwnerCounter    PetOwnerCounter
	deleteTransactor   OwnerDeleteTransactor
	deleteGuardEnabled bool
}

// NewService constructs the owner use case.
func NewService(
	repo ServiceRepository,
	insuranceFinder InsuranceFinder,
	tagNotifier TagNotifier,
	auditLogger AuditLogger,
) Service {
	return &ownerService{
		repo:               repo,
		insuranceFinder:    insuranceFinder,
		tagNotifier:        tagNotifier,
		auditLogger:        auditLogger,
		deleteGuardEnabled: false,
	}
}

// NewServiceWithPetOwnerDeleteGuard constructs the production owner use case
// with the transaction and dependency ports required for guarded deletion.
func NewServiceWithPetOwnerDeleteGuard(
	repo ServiceRepository,
	insuranceFinder InsuranceFinder,
	tagNotifier TagNotifier,
	auditLogger AuditLogger,
	deleteLocker OwnerDeleteLocker,
	petOwnerCounter PetOwnerCounter,
	deleteTransactor OwnerDeleteTransactor,
) Service {
	return &ownerService{
		repo:               repo,
		insuranceFinder:    insuranceFinder,
		tagNotifier:        tagNotifier,
		auditLogger:        auditLogger,
		deleteLocker:       deleteLocker,
		petOwnerCounter:    petOwnerCounter,
		deleteTransactor:   deleteTransactor,
		deleteGuardEnabled: true,
	}
}
