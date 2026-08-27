// Package service provides business logic implementations for Staff entity.
package staff

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
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
	ReservationVisible     *bool
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
	ReservationVisible     *bool
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
	// Password は nil の場合パスワード更新なし。空文字は未送信扱い。
	Password *string

	// LINE予約用フィールド
	StaffType              *string
	ReservationDisplayName *string
	ReservationVisible     *bool
	ReservationComment     *string
	ReservationImageURL    *string

	// AuthorizedClinicIDs and IsSystemAdmin are derived from the authenticated
	// identity, never from the request body. Staff profile and account fields are
	// global across assignments, so non-admin callers must be authorized for
	// every active assignment observed under the mutation transaction.
	AuthorizedClinicIDs []uint64
	IsSystemAdmin       bool
	// ActorStaffID is the authenticated staff performing the update when known.
	// Zero means the actor staff id was not available on the request context.
	ActorStaffID uint64
	// CredentialAudit is derived from the authenticated request context and is
	// required only when Password requests a credential replacement.
	CredentialAudit *CredentialMutationAudit
}

// SetClinicAssignmentsInput はスタッフのクリニック割当を実行者の mutable scope で
// 差分更新するための入力である。ClinicIDs と AuthorizedClinicIDs は変更しない。
type SetClinicAssignmentsInput struct {
	StaffID             uint64
	ClinicIDs           []uint64
	AuthorizedClinicIDs []uint64
	IsSystemAdmin       bool
}

// StaffAssignmentClinicLookup は割当先クリニックを同一 transaction の完了まで
// active のまま保持するために必要な最小依存である。
type StaffAssignmentClinicLookup interface {
	LockActiveByID(ctx context.Context, id uint64) (*model.Clinic, error)
}

// StaffAccountStore は Staff service がアカウント作成・パスワード更新に
// 必要とする最小の read/write port である。実装は受け取った context の
// ambient transaction に参加する。
type StaffAccountStore interface {
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	Create(ctx context.Context, account *model.Account) error
	UpdatePasswordHash(
		ctx context.Context,
		id uint64,
		newHash string,
		updatedAt time.Time,
	) error
	DeletePasswordResetTokens(ctx context.Context, accountID uint64) error
}

// StaffCoreService はスタッフの CRUD・並び替え操作（テスト時に最小モックで済む分割単位）
type StaffCoreService interface {
	List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	GetByID(ctx context.Context, id uint64) (*model.Staff, error)
	GetByIDInClinic(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	// Create はスタッフを作成する（AccountID を呼び出し元で決定済みの場合に使用）。
	Create(ctx context.Context, input *CreateStaffInput) (*model.Staff, error)
	// CreateWithAccount はメールアドレスとパスワードを受け取り、
	// email 重複チェック・bcrypt ハッシュ化・Account 作成・Staff 作成を一括で行う。
	CreateWithAccount(ctx context.Context, input *CreateStaffWithAccountInput) (*model.Staff, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error)
	Delete(ctx context.Context, clinicID, id uint64, isSystemAdmin bool) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// StaffAccountService はアカウント・クリニック割当に関する操作
type StaffAccountService interface {
	// FindByAccountID はアカウントIDに紐づくスタッフを返す。
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
	// SetClinicAssignments はスタッフのクリニック割当をトランザクション内で
	// 実行者の mutable scope に対する差分として更新する。staff・既存割当を決定的な
	// 順序でロックし、解除対象 clinic の予約・シフト依存を同じ transaction で拒否する。
	SetClinicAssignments(ctx context.Context, input *SetClinicAssignmentsInput) error
	// VerifyClinicMembership はスタッフが指定クリニックに所属しているかを確認する。
	// 所属していない場合は ErrNotFound を返す。
	VerifyClinicMembership(ctx context.Context, staffID, clinicID uint64) error
}

// StaffPermissionService は権限グループ・除外予約種別の操作
type StaffPermissionService interface {
	// GetPermissionGroupIDs は認証済みクリニック内でスタッフが所属する権限グループIDを返す
	GetPermissionGroupIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	// SetPermissionGroupIDs はスタッフの権限グループを全置換する
	SetPermissionGroupIDs(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
	// GetExcludedReservationTypeIDs はスタッフの除外サービス種別IDリストを返す
	GetExcludedReservationTypeIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	// SetExcludedReservationTypeIDs はスタッフの除外サービス種別を全置換する
	SetExcludedReservationTypeIDs(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	// GetCapableReservationTypeIDs はスタッフの対応可能サービス種別IDリストを返す
	GetCapableReservationTypeIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	// SetCapableReservationTypeIDs はスタッフの対応可能サービス種別を全置換する
	SetCapableReservationTypeIDs(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
}

// StaffService は StaffCoreService / StaffAccountService / StaffPermissionService を統合したインターフェース。
// validation.go 等、複数サブインターフェースにまたがる処理で使用する。
type StaffService interface {
	StaffCoreService
	StaffAccountService
	StaffPermissionService
}

type staffService struct {
	repo                StaffRepository
	accountRepo         StaffAccountStore
	assignmentRepo      StaffClinicAssignmentRepository
	reservationRepo     StaffAssignmentReservationUsage
	shiftEntryRepo      ShiftEntryRepository
	permissionGroupRepo PermissionGroupRepository
	resStaffRepo        ReservationStaffRepository
	occupationRepo      OccupationRepository
	clinicRepo          StaffAssignmentClinicLookup
	tx                  Transactor
	credentialAudit     CredentialAuditTxLogger
	permissionAudit     PermissionAssignmentAuditTxLogger
}

func NewStaffService(
	repo StaffRepository,
	accountRepo StaffAccountStore,
	assignmentRepo StaffClinicAssignmentRepository,
	reservationRepo StaffAssignmentReservationUsage,
	shiftEntryRepo ShiftEntryRepository,
	permissionGroupRepo PermissionGroupRepository,
	resStaffRepo ReservationStaffRepository,
	occupationRepo OccupationRepository,
	clinicRepo StaffAssignmentClinicLookup,
	tx Transactor,
) StaffService {
	return NewStaffServiceWithCredentialAudit(
		repo,
		accountRepo,
		assignmentRepo,
		reservationRepo,
		shiftEntryRepo,
		permissionGroupRepo,
		resStaffRepo,
		occupationRepo,
		clinicRepo,
		tx,
		nil,
	)
}

// NewStaffServiceWithCredentialAudit constructs staff use cases with a
// transaction-bound audit writer for administrator password replacement.
func NewStaffServiceWithCredentialAudit(
	repo StaffRepository,
	accountRepo StaffAccountStore,
	assignmentRepo StaffClinicAssignmentRepository,
	reservationRepo StaffAssignmentReservationUsage,
	shiftEntryRepo ShiftEntryRepository,
	permissionGroupRepo PermissionGroupRepository,
	resStaffRepo ReservationStaffRepository,
	occupationRepo OccupationRepository,
	clinicRepo StaffAssignmentClinicLookup,
	tx Transactor,
	credentialAudit CredentialAuditTxLogger,
) StaffService {
	return NewStaffServiceWithAudits(
		repo,
		accountRepo,
		assignmentRepo,
		reservationRepo,
		shiftEntryRepo,
		permissionGroupRepo,
		resStaffRepo,
		occupationRepo,
		clinicRepo,
		tx,
		credentialAudit,
		nil,
	)
}

// NewStaffServiceWithAudits constructs staff use cases with both supported
// transaction-bound audit writers.
func NewStaffServiceWithAudits(
	repo StaffRepository,
	accountRepo StaffAccountStore,
	assignmentRepo StaffClinicAssignmentRepository,
	reservationRepo StaffAssignmentReservationUsage,
	shiftEntryRepo ShiftEntryRepository,
	permissionGroupRepo PermissionGroupRepository,
	resStaffRepo ReservationStaffRepository,
	occupationRepo OccupationRepository,
	clinicRepo StaffAssignmentClinicLookup,
	tx Transactor,
	credentialAudit CredentialAuditTxLogger,
	permissionAudit PermissionAssignmentAuditTxLogger,
) StaffService {
	return &staffService{
		repo:                repo,
		accountRepo:         accountRepo,
		assignmentRepo:      assignmentRepo,
		reservationRepo:     reservationRepo,
		shiftEntryRepo:      shiftEntryRepo,
		permissionGroupRepo: permissionGroupRepo,
		resStaffRepo:        resStaffRepo,
		occupationRepo:      occupationRepo,
		clinicRepo:          clinicRepo,
		tx:                  tx,
		credentialAudit:     credentialAudit,
		permissionAudit:     permissionAudit,
	}
}
