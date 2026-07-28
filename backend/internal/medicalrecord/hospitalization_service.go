package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// DischargeWithBillingInput は退院+会計作成の入力DTO
type DischargeWithBillingInput struct {
	DischargeDate    time.Time
	CreateAccounting bool
	ActorID          *uint64
}

// DischargeWithBillingResult は退院+会計作成のレスポンスDTO
type DischargeWithBillingResult struct {
	HospitalizationID uint64
	AccountingID      *uint64
	Status            string
}

// CreateHospitalizationInput は入院作成の入力DTO
type CreateHospitalizationInput struct {
	OwnerID              uint64
	PetID                uint64
	HospitalizationType  model.HospitalizationType
	StartDate            time.Time
	EndDate              time.Time
	Status               model.HospitalizationStatus
	CageID               *uint64
	DoctorID             *uint64
	Memo                 string
	OwnerRequest         string
	StaffNotes           string
	IsInsurance          bool
	InsuranceCompanyName *string
	InsuranceNumber      *string
}

// UpdateHospitalizationInput は入院更新のサービス入力 DTO
type UpdateHospitalizationInput struct {
	OwnerID              *uint64
	PetID                *uint64
	HospitalizationType  *model.HospitalizationType
	StartDate            *time.Time
	EndDate              *time.Time
	Status               *model.HospitalizationStatus
	CageID               *uint64
	DoctorID             *uint64
	Memo                 *string
	OwnerRequest         *string
	StaffNotes           *string
	IsInsurance          *bool
	InsuranceCompanyName *string
	InsuranceNumber      *string
}

// validatePetNotDeceased は死亡ペットへの入院登録・貼り替えをブロックする（SD-10・臨床安全）。
// FE のペット選択 UI は死亡ペットをクリック不可にするだけで API 直叩きを防げないため、
// BE 側でも fail-closed に検証する。
func validatePetNotDeceased(ctx context.Context, petRepo petFinder, clinicID, petID uint64) error {
	pet, err := petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify pet status")
	}
	if pet.DeceasedAt != nil {
		return apperrors.WrapInvalidInput("死亡したペットは入院登録できません")
	}
	return nil
}

// resolveFinalHospitalizationOwnerPet は PATCH 入力と現在値から最終 Owner/Pet を求める（AUD-004）。
func resolveFinalHospitalizationOwnerPet(existing *model.Hospitalization, input *UpdateHospitalizationInput) (ownerID, petID *uint64) {
	o, p := existing.OwnerID, existing.PetID
	ownerID, petID = &o, &p
	if input.OwnerID != nil {
		ownerID = input.OwnerID
	}
	if input.PetID != nil {
		petID = input.PetID
	}
	return ownerID, petID
}

func buildHospitalizationUpdate(input *UpdateHospitalizationInput) map[string]any {
	fields := make(map[string]any)
	if input.OwnerID != nil {
		fields["owner_id"] = *input.OwnerID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.HospitalizationType != nil {
		fields["hospitalization_type"] = *input.HospitalizationType
	}
	if input.StartDate != nil {
		fields["start_date"] = *input.StartDate
	}
	if input.EndDate != nil {
		fields["end_date"] = *input.EndDate
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.CageID != nil {
		fields["cage_id"] = *input.CageID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.OwnerRequest != nil {
		fields["owner_request"] = *input.OwnerRequest
	}
	if input.StaffNotes != nil {
		fields["staff_notes"] = *input.StaffNotes
	}
	if input.IsInsurance != nil {
		if !*input.IsInsurance {
			// 保険なしに切り替えた場合は保険情報を NULL にする
			fields["insurance_company_name"] = nil
			fields["insurance_number"] = nil
		} else {
			if input.InsuranceCompanyName != nil {
				fields["insurance_company_name"] = *input.InsuranceCompanyName
			}
			if input.InsuranceNumber != nil {
				fields["insurance_number"] = *input.InsuranceNumber
			}
		}
	} else {
		// IsInsurance が nil でも保険フィールド単体の更新は許容する
		if input.InsuranceCompanyName != nil {
			fields["insurance_company_name"] = *input.InsuranceCompanyName
		}
		if input.InsuranceNumber != nil {
			fields["insurance_number"] = *input.InsuranceNumber
		}
	}
	return fields
}

type HospitalizationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error)
}

type hospitalizationService struct {
	hospRepo         HospitalizationRepository
	reservationRepo  sharedkernel.OwnerPetLinkVerifier
	doctorVerifier   hospitalizationDoctorVerifier
	petRepo          petFinder
	cageRepo         cageFinder
	carePlanItemRepo CarePlanItemRepository
	accountingRepo   accountingCreator
	billingItemRepo  billingItemWriter
	transactor       Transactor
	auditTx          AuditTxLogger
}

type hospitalizationDoctorVerifier interface {
	AssertMedicalRecordDoctorInClinic(ctx context.Context, clinicID, doctorID uint64) error
}

// NewHospitalizationService は実消費 repo の個別注入で初期化する（BE9-2D ⑤ Phase1:
// *repository.Repositories 集約と repo-swap tx 機構を treatment ④b と同型で解体。
// DischargeWithBilling は Transactor.WithTx + 各 repo の dbOrTx 参加で原子性を維持する）。
func NewHospitalizationService(
	hospRepo HospitalizationRepository,
	reservationRepo sharedkernel.OwnerPetLinkVerifier,
	petRepo petFinder,
	cageRepo cageFinder,
	carePlanItemRepo CarePlanItemRepository,
	accountingRepo accountingCreator,
	billingItemRepo billingItemWriter,
	transactor Transactor,
) HospitalizationService {
	return NewHospitalizationServiceWithAudit(
		hospRepo,
		reservationRepo,
		petRepo,
		cageRepo,
		carePlanItemRepo,
		accountingRepo,
		billingItemRepo,
		transactor,
		nil,
	)
}

// NewHospitalizationServiceWithAudit は退院会計の tx 内監査依存を追加して初期化する。
// 既存コンストラクタは後方互換のため維持し、会計作成経路では nil 監査依存を fail-closed に扱う。
func NewHospitalizationServiceWithAudit(
	hospRepo HospitalizationRepository,
	reservationRepo sharedkernel.OwnerPetLinkVerifier,
	petRepo petFinder,
	cageRepo cageFinder,
	carePlanItemRepo CarePlanItemRepository,
	accountingRepo accountingCreator,
	billingItemRepo billingItemWriter,
	transactor Transactor,
	auditTx AuditTxLogger,
) HospitalizationService {
	doctorVerifier, _ := reservationRepo.(hospitalizationDoctorVerifier)
	return &hospitalizationService{
		hospRepo:         hospRepo,
		reservationRepo:  reservationRepo,
		doctorVerifier:   doctorVerifier,
		petRepo:          petRepo,
		cageRepo:         cageRepo,
		carePlanItemRepo: carePlanItemRepo,
		accountingRepo:   accountingRepo,
		billingItemRepo:  billingItemRepo,
		transactor:       transactor,
		auditTx:          auditTx,
	}
}

func (s *hospitalizationService) validateDoctor(ctx context.Context, clinicID uint64, doctorID *uint64) error {
	if doctorID == nil {
		return nil
	}
	if *doctorID == 0 {
		return apperrors.WrapInvalidInput("doctor_id must be greater than zero")
	}
	if s.doctorVerifier == nil {
		return apperrors.WrapInternalServerError("hospitalization doctor verifier is required")
	}
	if err := s.doctorVerifier.AssertMedicalRecordDoctorInClinic(ctx, clinicID, *doctorID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization doctor ownership")
	}
	return nil
}

func (s *hospitalizationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	result, total, err := s.hospRepo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list hospitalizations", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list hospitalizations")
	}
	return result, total, nil
}

func (s *hospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	result, err := s.hospRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to get hospitalization")
	}
	return result, nil
}

func (s *hospitalizationService) Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	status := input.Status
	if status == "" {
		status = model.HospitalizationStatusReserved
	}
	// is_insurance == false の場合は保険情報を NULL にする
	var insuranceCompanyName *string
	var insuranceNumber *string
	if input.IsInsurance {
		insuranceCompanyName = input.InsuranceCompanyName
		insuranceNumber = input.InsuranceNumber
	}

	hospitalization := &model.Hospitalization{
		ClinicID:             clinicID,
		OwnerID:              input.OwnerID,
		PetID:                input.PetID,
		HospitalizationType:  input.HospitalizationType,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		Status:               status,
		CageID:               input.CageID,
		DoctorID:             input.DoctorID,
		Memo:                 input.Memo,
		OwnerRequest:         input.OwnerRequest,
		StaffNotes:           input.StaffNotes,
		InsuranceCompanyName: insuranceCompanyName,
		InsuranceNumber:      insuranceNumber,
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("hospitalization write transaction dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Validate every request-derived clinic-scoped FK in the same transaction as persistence.
		ownerID, petID := input.OwnerID, input.PetID
		if err := sharedkernel.ValidateReservationOwnerPetLinks(txCtx, s.reservationRepo, clinicID, &ownerID, &petID); err != nil {
			return err
		}
		if err := validatePetNotDeceased(txCtx, s.petRepo, clinicID, petID); err != nil {
			return err
		}
		if err := validateOwnedMasterFK(txCtx, "cage", clinicID, input.CageID,
			func(actx context.Context, cid, mid uint64) error {
				_, err := s.cageRepo.FindByID(actx, cid, mid)
				return err
			}); err != nil {
			return err
		}
		if err := s.validateDoctor(txCtx, clinicID, input.DoctorID); err != nil {
			return err
		}
		if err := s.hospRepo.Create(txCtx, hospitalization); err != nil {
			slog.ErrorContext(txCtx, "failed to create hospitalization", "error", err)
			return apperrors.Wrap(err, "failed to create hospitalization")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "hospitalization created",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.Uint64("clinic_id", hospitalization.ClinicID))
	return hospitalization, nil
}

func (s *hospitalizationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildHospitalizationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("hospitalization write transaction dependency is required")
	}
	var hosp *model.Hospitalization
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		existing, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to lock hospitalization", "error", err)
			return apperrors.Wrap(err, "failed to lock hospitalization")
		}
		if input.OwnerID != nil || input.PetID != nil {
			finalOwnerID, finalPetID := resolveFinalHospitalizationOwnerPet(existing, input)
			if err := sharedkernel.ValidateReservationOwnerPetLinks(txCtx, s.reservationRepo, clinicID, finalOwnerID, finalPetID); err != nil {
				return err
			}
		}
		if input.PetID != nil {
			if err := validatePetNotDeceased(txCtx, s.petRepo, clinicID, *input.PetID); err != nil {
				return err
			}
		}
		if err := validateOwnedMasterFK(txCtx, "cage", clinicID, input.CageID,
			func(actx context.Context, cid, mid uint64) error {
				_, err := s.cageRepo.FindByID(actx, cid, mid)
				return err
			}); err != nil {
			return err
		}
		if err := s.validateDoctor(txCtx, clinicID, input.DoctorID); err != nil {
			return err
		}
		hosp, err = s.hospRepo.Update(txCtx, clinicID, id, fields)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to update hospitalization", "error", err)
			return apperrors.Wrap(err, "failed to update hospitalization")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "hospitalization updated",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))
	return hosp, nil
}
func (s *hospitalizationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.hospRepo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find hospitalization")
	}
	// FK依存チェック: 入院に紐付く日次記録が存在する場合は削除を拒否
	dailyCount, err := s.hospRepo.CountDailyRecordsByHospitalizationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check daily record dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check daily record dependencies")
	}
	if dailyCount > 0 {
		return apperrors.WrapConflict("日次記録が紐付いているため削除できません。先に日次記録を削除してください")
	}

	// FK依存チェック: 入院に紐付く治療計画が存在する場合は削除を拒否
	planCount, err := s.hospRepo.CountTreatmentPlansByHospitalizationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check treatment plan dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check treatment plan dependencies")
	}
	if planCount > 0 {
		return apperrors.WrapConflict("治療計画が紐付いているため削除できません。先に治療計画を削除してください")
	}

	// FK依存チェック: 入院に紐付くケアプラン項目が存在する場合は削除を拒否
	itemCount, err := s.hospRepo.CountCarePlanItemsByHospitalizationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check care plan item dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check care plan item dependencies")
	}
	if itemCount > 0 {
		return apperrors.WrapConflict("ケアプランが紐付いているため削除できません。先にケアプランを削除してください")
	}

	if err := s.hospRepo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete hospitalization", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete hospitalization")
	}

	slog.InfoContext(ctx, "hospitalization deleted",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}

// DischargeWithBilling は退院処理を行い、オプションで会計レコードを自動生成する。
// care_plan_items を billing_items に変換してトランザクション内で原子的に実行する。
func (s *hospitalizationService) DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error) {
	// 入院レコード取得
	hosp, err := s.hospRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get hospitalization", "error", err)
		return nil, apperrors.Wrap(err, "failed to get hospitalization")
	}
	if hosp.Status == model.HospitalizationStatusDischarged {
		return nil, apperrors.WrapInvalidInput("hospitalization is already discharged")
	}

	result := &DischargeWithBillingResult{
		HospitalizationID: id,
		Status:            string(model.HospitalizationStatusDischarged),
	}

	// BE9-2D ⑤ Phase1: repos.Transaction（tx-bound clone）→ Transactor.WithTx（ctx-txKey）へ変換。
	// 閉包内の read/write は各 repo の dbOrTx が txCtx の ambient tx へ参加する（挙動は旧機構と等価）。
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// 0. Q2-C: FOR UPDATE で直列化し、locked 行の OwnerID/PetID で Q2-A 再検証する。
		// LockByIDForUpdate の行スナップショットにスカラーが含まれるため、検証用 ID が空にならない。
		locked, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock hospitalization for discharge")
		}
		if locked.Status == model.HospitalizationStatusDischarged {
			return apperrors.WrapInvalidInput("hospitalization is already discharged")
		}

		// 1. 汚染行対策: CreateAccounting 有無に関わらず、Update 前に Owner/Pet の clinic 所有を再検証する（AUD-004 Q2-A）。
		if err := s.reservationRepo.AssertOwnerInClinic(txCtx, clinicID, locked.OwnerID); err != nil {
			return apperrors.Wrap(err, "failed to verify hospitalization owner ownership")
		}
		if _, err := s.reservationRepo.FindPetOwnerInClinic(txCtx, clinicID, locked.PetID); err != nil {
			return apperrors.Wrap(err, "failed to verify hospitalization pet ownership")
		}
		if input.CreateAccounting {
			if input.ActorID == nil || *input.ActorID == 0 {
				return apperrors.WrapInvalidInput("staff actor is required to create discharge billing")
			}
			if s.auditTx == nil {
				return apperrors.WrapInternalServerError("hospitalization discharge billing audit dependency is required")
			}
		}

		// 2. 退院ステータスに更新
		dischargedStatus := model.HospitalizationStatusDischarged
		dischargeFields := map[string]any{
			"status":   dischargedStatus,
			"end_date": input.DischargeDate,
		}
		if _, err := s.hospRepo.UpdateIfNotDischarged(txCtx, clinicID, id, dischargeFields); err != nil {
			return apperrors.Wrap(err, "failed to discharge hospitalization")
		}

		if !input.CreateAccounting {
			return nil
		}

		// 2. ケアプラン取得
		carePlanItems, err := s.carePlanItemRepo.FindByHospitalizationID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get care plan items")
		}

		// 3. 会計レコード作成
		billing := &model.Billing{
			ClinicID:          clinicID,
			HospitalizationID: &id,
			PetID:             &locked.PetID,
			OwnerID:           &locked.OwnerID,
			Status:            model.BillingStatusWaiting,
			ScheduledDate:     input.DischargeDate,
		}
		if err := s.accountingRepo.Create(txCtx, clinicID, billing); err != nil {
			return apperrors.Wrap(err, "failed to create billing")
		}

		// 4. ケアプラン → 会計明細に変換
		var subtotalAmount int64
		for i := range carePlanItems {
			item := &carePlanItems[i]
			billingItem := &model.BillingItem{
				BillingID: billing.ID,
				Category: sharedkernel.ResolveItemCategory(sharedkernel.ItemCategoryResolverInput{
					Source:              model.ItemSourceHospitalization,
					CarePlanType:        item.Type,
					IsSurgery:           item.Procedure != nil && item.Procedure.IsSurgery,
					HospitalizationType: locked.HospitalizationType,
				}),
				Name:      item.Name,
				UnitPrice: item.UnitPrice,
				Quantity:  1.0,
				TaxType:   model.TaxTypeExcluded,
				TaxRate:   sharedkernel.DefaultTaxRate,
				Source:    model.ItemSourceHospitalization,
				SortOrder: i,
			}
			if err := s.billingItemRepo.Create(txCtx, billingItem); err != nil {
				return apperrors.Wrap(err, "failed to create billing item")
			}
			subtotalAmount += item.UnitPrice
		}

		// 5. 合計金額更新
		taxAmount := int64(float64(subtotalAmount) * sharedkernel.DefaultTaxRate)
		totalAmount := subtotalAmount + taxAmount
		if len(carePlanItems) > 0 {
			if err := s.billingItemRepo.UpdateBillingTotals(txCtx, clinicID, billing.ID, subtotalAmount, taxAmount, totalAmount); err != nil {
				return apperrors.Wrap(err, "failed to update billing totals")
			}
		}

		// 6. 会計を伴う退院を同じ tx 内で監査する。失敗時は status/会計/明細/合計を一括 rollback する。
		resourceID := id
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorID:    input.ActorID,
			ActorType:  model.AuditActorTypeStaff,
			Action:     model.AuditActionHospitalizationDischargeWithBilling,
			Resource:   model.AuditResourceHospitalization,
			ResourceID: &resourceID,
			NewValue: map[string]any{
				"billing_id":      billing.ID,
				"subtotal_amount": subtotalAmount,
				"tax_amount":      taxAmount,
				"total_amount":    totalAmount,
			},
		}); err != nil {
			return apperrors.Wrap(err, "failed to audit hospitalization discharge billing")
		}

		result.AccountingID = &billing.ID
		return nil
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to discharge hospitalization with billing", "error", err)
		return nil, apperrors.Wrap(err, "failed to discharge hospitalization with billing")
	}

	slog.InfoContext(ctx, "hospitalization discharged",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("create_accounting", input.CreateAccounting))

	return result, nil
}
