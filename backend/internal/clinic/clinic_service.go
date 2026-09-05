package clinic

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// CreateClinicInput はクリニック作成の入力DTO
type CreateClinicInput struct {
	Name               string
	PostalCode         string
	Address            string
	PhoneNumber        string
	FaxNumber          string
	RegistrationNumber string
	DirectorName       string
	Email              string
	Website            string
}

// UpdateClinicInput はクリニック更新の入力DTO（nil = 未指定）
type UpdateClinicInput struct {
	Name                                      *string
	PostalCode                                *string
	Address                                   *string
	PhoneNumber                               *string
	FaxNumber                                 *string
	RegistrationNumber                        *string
	DirectorName                              *string
	Email                                     *string
	Website                                   *string
	LogoURL                                   *string
	IsActive                                  *bool
	StandardTaxRate                           *float64
	ReducedTaxRate                            *float64
	AccountingDocumentShowLogo                *bool
	AccountingDocumentShowRegistrationWarning *bool
	AccountingDocumentShowItemCategory        *bool
	AccountingDocumentFooterNote              *string
	// #190: セクション表示/非表示トグルと表示順 (migration 010)
	AccountingDocumentShowClinicHeader   *bool
	AccountingDocumentShowOwnerPetInfo   *bool
	AccountingDocumentShowItemsTable     *bool
	AccountingDocumentShowPaymentSummary *bool
	AccountingDocumentSectionOrder       *[]string // nil=未変更, non-nil=更新（空配列=デフォルト順にリセット）
}

const (
	colClinicName                                      = "name"
	colClinicPostalCode                                = "postal_code"
	colClinicAddress                                   = "address"
	colClinicPhoneNumber                               = "phone_number"
	colClinicFaxNumber                                 = "fax_number"
	colClinicRegistrationNumber                        = "registration_number"
	colClinicDirectorName                              = "director_name"
	colClinicEmail                                     = "email"
	colClinicWebsite                                   = "website"
	colClinicLogoURL                                   = "logo_url"
	colClinicIsActive                                  = "is_active"
	colClinicStandardTaxRate                           = "standard_tax_rate"
	colClinicReducedTaxRate                            = "reduced_tax_rate"
	colClinicAccountingDocumentShowLogo                = "accounting_document_show_logo"
	colClinicAccountingDocumentShowRegistrationWarning = "accounting_document_show_registration_warning"
	colClinicAccountingDocumentShowItemCategory        = "accounting_document_show_item_category"
	colClinicAccountingDocumentFooterNote              = "accounting_document_footer_note"
	// #190: セクション表示/非表示トグルと表示順 (migration 010)
	colClinicAccountingDocumentShowClinicHeader   = "accounting_document_show_clinic_header"
	colClinicAccountingDocumentShowOwnerPetInfo   = "accounting_document_show_owner_pet_info"
	colClinicAccountingDocumentShowItemsTable     = "accounting_document_show_items_table"
	colClinicAccountingDocumentShowPaymentSummary = "accounting_document_show_payment_summary"
	colClinicAccountingDocumentSectionOrder       = "accounting_document_section_order"
)

// buildClinicUpdate は PATCH 用 map を構築する。
// GORM のゼロ値スキップ問題を回避するために使用する。
// 税率が [0, 1] の範囲外の場合は error を返す。
func BuildClinicUpdate(input *UpdateClinicInput) (map[string]any, error) {
	fields := make(map[string]any)
	applyClinicProfileFields(fields, input)
	if err := applyClinicTaxFields(fields, input); err != nil {
		return nil, err
	}
	applyClinicAccountingDocumentFlags(fields, input)
	if err := applyClinicAccountingDocumentSectionOrder(fields, input); err != nil {
		return nil, err
	}
	return fields, nil
}

// defaultPermissionRule は新規クリニック作成時に "執行"/"一般" グループへ設定する
// リソース×CRUD権限のデフォルト値。
//
// SD-9: 従来 CreateClinic はグループ (permission_groups) だけを作成し、ルール
// (permission_group_rules) を1件も作成していなかった。permission_group_rules に
// 一致行が無いリソースは handler.hasPermission / RequirePermission が deny 方向に
// fallback するため（clinic_handler.go の hasPermission はループ内に一致が無ければ
// 末尾で false を返す）、新規クリニックでは is_system_admin 以外の全スタッフが
// 全リソースへアクセス不能になっていた（疑い=事実）。
//
// 出所: backend/migrations/seeds/002_master/permission_group_rules.csv の
// 執行=奇数ID / 一般=偶数ID / 閲覧専用=group 9 パターン。
// model.AllResources (36) をすべてカバーする。既存デモ seed は明示 rollout 前の
// examination-unconfirm を含めず、同権限は新規クリニックでも default-deny とする。
//
// 設定系フォールバック（cash-register-close / accounting-reports /
// master-payment-method / lstep-csv-import / lstep-analytics / manual-edit /
// lab-import）: 執行=view+edit（create/delete 不可、hospital-settings と同型）、
// 一般=view のみ。
// 例外: closing-settings は /closing-settings/holidays と special-periods が
// create/delete を要求するため、執行は CRUD 全許可（POC-01 契約整合）。
// 例外: master-animal-species は全クリニック共有マスタのため、is_system_admin
// 以外の mutation を禁止する（執行・一般とも view-only）。
type defaultPermissionRule struct {
	resource                                   model.Resource
	execView, execCreate, execEdit, execDelete bool
	genView, genCreate, genEdit, genDelete     bool
}

var defaultPermissionRuleTable = []defaultPermissionRule{
	{model.ResourceReception, true, true, true, true, true, false, false, false},
	{model.ResourceOwners, true, true, true, true, true, true, true, false},
	{model.ResourceReservations, true, true, true, true, true, true, true, false},
	{model.ResourceMedicalRecords, true, true, true, true, true, true, true, false},
	{model.ResourceHospitalization, true, true, true, true, true, true, true, false},
	{model.ResourceTrimming, true, true, true, true, true, true, true, false},
	{model.ResourceExaminations, true, true, true, true, true, true, true, false},
	// Unconfirm changes an immutable clinical workflow boundary. It is never granted by default.
	{model.ResourceExaminationUnconfirm, false, false, false, false, false, false, false, false},
	{model.ResourceAccounting, true, true, true, true, true, false, false, false},
	{model.ResourceVaccinations, true, true, true, true, true, true, true, false},
	{model.ResourceCheckups, true, true, true, true, true, false, false, false},
	{model.ResourceInventory, true, true, true, true, true, false, false, false},
	{model.ResourceEstimates, true, true, true, true, true, false, false, false},
	{model.ResourceShifts, true, true, true, true, true, true, true, false},
	{model.ResourceHospitalSettings, true, false, true, false, true, false, false, false},
	// 共有マスタ: mutation は is_system_admin のみ（権限グループでは view-only）
	{model.ResourceMasterAnimalSpecies, true, false, false, false, true, false, false, false},
	{model.ResourceMasterMedical, true, true, true, true, true, false, false, false},
	{model.ResourceMasterReservationType, true, true, true, true, true, false, false, false},
	{model.ResourceMasterHospitalization, true, true, true, true, true, false, false, false},
	{model.ResourceMasterTrimming, true, true, true, true, true, false, false, false},
	{model.ResourceMasterPermission, true, true, true, true, false, false, false, false},
	{model.ResourceMasterStaff, true, true, true, true, true, false, false, false},
	{model.ResourceMasterInsurance, true, true, true, true, true, false, false, false},
	{model.ResourceMasterMerchandise, true, true, true, true, true, false, false, false},
	{model.ResourceDiscount, true, true, true, true, false, false, false, false},
	{model.ResourceAccountingCancel, true, false, true, false, true, false, false, false},
	{model.ResourceAccountingPostCloseEdit, true, false, true, false, true, false, false, false},
	// 設定系フォールバック（hospital-settings と同型: 執行 view+edit / 一般 view）
	{model.ResourceCashRegisterClose, true, false, true, false, true, false, false, false},
	{model.ResourceAccountingReports, true, false, true, false, true, false, false, false},
	{model.ResourceClosingSettings, true, true, true, true, true, false, false, false},
	{model.ResourcePaymentMethod, true, false, true, false, true, false, false, false},
	{model.ResourceLstepCsvImport, true, false, true, false, true, false, false, false},
	{model.ResourceLstepAnalytics, true, false, true, false, true, false, false, false},
	{model.ResourceManualEdit, true, false, true, false, true, false, false, false},
	{model.ResourceLabImport, true, false, true, false, true, false, false, false},
	// #239 identity-links: fail-closed（通常 staff へ自動付与しない。運用で明示付与）
	{model.ResourceIdentityLinks, false, false, false, false, false, false, false, false},
	// TASK-374 / #211 checkup package import: default-deny（明示付与のみ）
	{model.ResourceCheckupPackageImport, false, false, false, false, false, false, false, false},
}

// buildDefaultPermissionGroupRules は defaultPermissionRuleTable から、指定グループが
// 執行(isExecutive=true)か一般(isExecutive=false)かに応じたルール一覧を組み立てる。
func buildDefaultPermissionGroupRules(isExecutive bool) []model.PermissionGroupRule {
	rules := make([]model.PermissionGroupRule, 0, len(defaultPermissionRuleTable))
	for _, r := range defaultPermissionRuleTable {
		rule := model.PermissionGroupRule{Resource: string(r.resource)}
		if isExecutive {
			rule.CanView, rule.CanCreate, rule.CanEdit, rule.CanDelete = r.execView, r.execCreate, r.execEdit, r.execDelete
		} else {
			rule.CanView, rule.CanCreate, rule.CanEdit, rule.CanDelete = r.genView, r.genCreate, r.genEdit, r.genDelete
		}
		rules = append(rules, rule)
	}
	return rules
}

type ClinicService interface {
	ListClinics(ctx context.Context) ([]model.Clinic, error)
	ListClinicsByIDs(ctx context.Context, ids []uint64) ([]model.Clinic, error)
	ListActiveClinicIDs(ctx context.Context, ids []uint64) ([]uint64, error)
	ListClinicsByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error)
	CreateClinic(ctx context.Context, input *CreateClinicInput) (*model.Clinic, error)
	UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error)
	DeleteClinic(ctx context.Context, id uint64) error
}

type clinicService struct {
	repo                clinicServiceRepository
	permissionGroupRepo PermissionGroupWriter
	transactor          Transactor
}

func NewClinicService(repo clinicServiceRepository, permissionGroupRepo PermissionGroupWriter, transactor Transactor) ClinicService {
	return &clinicService{repo: repo, permissionGroupRepo: permissionGroupRepo, transactor: transactor}
}

func (s *clinicService) ListClinics(ctx context.Context) ([]model.Clinic, error) {
	clinics, err := s.repo.FindAll(ctx)
	if err != nil {
		// login の非 admin 経路は失敗を捨てて続行する。RespondError しないため診断ログを残す。
		slog.ErrorContext(ctx, "failed to list clinics", "error", err)
		return nil, apperrors.Wrap(err, "failed to list clinics")
	}
	return clinics, nil
}

func (s *clinicService) ListClinicsByIDs(ctx context.Context, ids []uint64) ([]model.Clinic, error) {
	clinics, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list clinics by id")
	}
	return clinics, nil
}

func (s *clinicService) ListActiveClinicIDs(ctx context.Context, ids []uint64) ([]uint64, error) {
	active, err := s.repo.FindActiveIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list active clinic ids")
	}
	return active, nil
}

func (s *clinicService) ListClinicsByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error) {
	clinics, err := s.repo.FindByStaffID(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list clinics by staff")
	}
	return clinics, nil
}

func (s *clinicService) GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	clinic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get clinic")
	}
	return clinic, nil
}

func (s *clinicService) CreateClinic(ctx context.Context, input *CreateClinicInput) (*model.Clinic, error) {
	company, err := s.repo.FindCompany(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get company")
	}
	clinic := &model.Clinic{
		CompanyID:          company.ID,
		Name:               input.Name,
		PostalCode:         input.PostalCode,
		Address:            input.Address,
		PhoneNumber:        input.PhoneNumber,
		FaxNumber:          input.FaxNumber,
		RegistrationNumber: input.RegistrationNumber,
		DirectorName:       input.DirectorName,
		Email:              input.Email,
		Website:            input.Website,
		IsActive:           true,
		AccountingDocumentShowRegistrationWarning: true,
		AccountingDocumentShowItemCategory:        true,
		AccountingDocumentShowClinicHeader:        true,
		AccountingDocumentShowOwnerPetInfo:        true,
		AccountingDocumentShowItemsTable:          true,
		AccountingDocumentShowPaymentSummary:      true,
	}

	if err := s.transactor.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, clinic); err != nil {
			return apperrors.Wrap(err, "failed to create clinic")
		}

		// Initialize default permission groups: "執行" (executive) and "一般" (general)
		defaultGroups := []struct {
			name        string
			description string
			sortOrder   int
			isExecutive bool
		}{
			{name: "執行", description: "執行権限", sortOrder: 1, isExecutive: true},
			{name: "一般", description: "一般スタッフ権限", sortOrder: 2, isExecutive: false},
		}

		for _, groupDef := range defaultGroups {
			group := &model.PermissionGroup{
				ClinicID:    clinic.ID,
				Name:        groupDef.name,
				Description: groupDef.description,
				IsActive:    true,
				SortOrder:   groupDef.sortOrder,
			}
			if err := s.permissionGroupRepo.Create(ctx, group); err != nil {
				return apperrors.Wrap(err, "failed to create default permission group")
			}
			// SD-9: グループ作成だけではルール0件のまま残り、is_system_admin 以外の
			// 全スタッフが全リソースへアクセス不能になる（hasPermission の deny-by-default
			// fallback）。作成直後に defaultPermissionRuleTable 由来のルールを流し込む。
			rules := buildDefaultPermissionGroupRules(groupDef.isExecutive)
			if err := s.permissionGroupRepo.UpdateRules(ctx, clinic.ID, group.ID, rules); err != nil {
				return apperrors.Wrap(err, "failed to create default permission group rules")
			}
			slog.InfoContext(ctx, "default permission group created",
				slog.Uint64("clinic_id", clinic.ID),
				slog.String("group_name", group.Name),
				slog.Uint64("group_id", group.ID),
				slog.Int("rule_count", len(rules)))
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create clinic")
	}

	slog.InfoContext(ctx, "clinic created",
		slog.Uint64("clinic_id", clinic.ID))
	return clinic, nil
}

func (s *clinicService) UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	// 存在確認（NotFound を早期返却）
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find clinic for update")
	}
	fields, err := BuildClinicUpdate(input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update clinic clinic")
	}
	if len(fields) == 0 {
		clinic, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get clinic")
		}
		return clinic, nil
	}
	// POC-02 / X-01: update+reload を同一 tx に収め、commit 済み成功を後段 read error で 5xx へ反転させない。
	// reload 失敗時は tx がロールバックするため、write は永続化されない。
	var updated *model.Clinic
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.UpdateClinic(txCtx, id, input); err != nil {
			return apperrors.Wrap(err, "failed to update clinic")
		}
		c, err := s.repo.FindByID(txCtx, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get updated clinic")
		}
		updated = c
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "clinic updated",
		slog.Uint64("clinic_id", id))
	return updated, nil
}

func (s *clinicService) DeleteClinic(ctx context.Context, id uint64) error {
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.ensureClinicCanBeDeleted(txCtx, id); err != nil {
			return err
		}
		if err := s.permissionGroupRepo.DeleteSoftDeletedByClinicID(txCtx, id); err != nil {
			return apperrors.Wrap(err, "failed to clean up soft-deleted permission groups")
		}
		return s.repo.Delete(txCtx, id)
	}); err != nil {
		return apperrors.Wrap(err, "failed to delete clinic")
	}

	slog.InfoContext(ctx, "clinic deleted", slog.Uint64("clinic_id", id))
	return nil
}

func (s *clinicService) ensureClinicCanBeDeleted(ctx context.Context, id uint64) error {
	if _, err := s.repo.LockByIDForUpdate(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to lock clinic for deletion")
	}

	// FK依存チェック: クリニックに関連するオーナーが存在する場合は削除を拒否
	ownerCount, err := s.repo.CountOwnersByClinicID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check owner dependencies")
	}
	if ownerCount > 0 {
		return apperrors.WrapConflict("飼主が紐付いているため削除できません。先に飼主を削除してください")
	}

	// FK依存チェック: クリニックに関連するスタッフが存在する場合は削除を拒否
	staffCount, err := s.repo.CountStaffByClinicID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check staff dependencies")
	}
	if staffCount > 0 {
		return apperrors.WrapConflict("スタッフが紐付いているため削除できません。先にスタッフを削除してください")
	}

	dependencies, err := s.repo.CountBlockingReferencesByClinicID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check clinic dependencies")
	}
	if len(dependencies) > 0 {
		dep := dependencies[0]
		return apperrors.WrapConflict(dep.Label + "が紐付いているため削除できません。関連データを先に整理してください")
	}
	return nil
}
