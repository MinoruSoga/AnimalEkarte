package medicalrecord

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

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
	// TreatmentPlans are optional nested plans created in the same transaction as the parent
	// (TASK-001-BE atomic create). Empty/nil preserves parent-only create.
	TreatmentPlans []CreateTreatmentPlanInput
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

// validatePetNotDeceased は死亡ペットへの業務書き込みをブロックする（SD-10・臨床安全）。
// 実装正本は sharedkernel.ValidatePetNotDeceased（#261 P0 で昇格）。message は呼び出し元の業務文言。
func validatePetNotDeceased(ctx context.Context, petRepo petFinder, clinicID, petID uint64, message string) error {
	return sharedkernel.ValidatePetNotDeceased(ctx, petRepo, clinicID, petID, message)
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
	hospRepo          HospitalizationRepository
	reservationRepo   sharedkernel.OwnerPetLinkVerifier
	doctorVerifier    hospitalizationDoctorVerifier
	petRepo           petFinder
	cageRepo          cageFinder
	carePlanItemRepo  CarePlanItemRepository
	accountingRepo    accountingCreator
	billingItemRepo   billingItemWriter
	treatmentPlanRepo TreatmentPlanRepository
	transactor        Transactor
	auditTx           AuditTxLogger
}

// HospitalizationServiceOption configures optional dependencies on HospitalizationService.
type HospitalizationServiceOption func(*hospitalizationService)

// WithTreatmentPlanRepository injects the treatment-plan write path used by nested create
// (TASK-001-BE). When nil and TreatmentPlans are present, Create fails closed.
func WithTreatmentPlanRepository(repo TreatmentPlanRepository) HospitalizationServiceOption {
	return func(s *hospitalizationService) {
		s.treatmentPlanRepo = repo
	}
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
	opts ...HospitalizationServiceOption,
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
		opts...,
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
	opts ...HospitalizationServiceOption,
) HospitalizationService {
	doctorVerifier, _ := reservationRepo.(hospitalizationDoctorVerifier)
	svc := &hospitalizationService{
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
	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}
	return svc
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

// defaultHospitalizationStatus picks Create default when client omits status (BUG-031).
// Clinic calendar day uses time.Local (compose TZ=Asia/Tokyo). start_date on/before today → admitted;
// future start_date → reserved. Explicit client status is never overridden.
func defaultHospitalizationStatus(startDate, now time.Time) model.HospitalizationStatus {
	loc := time.Local
	start := startDate.In(loc)
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	n := now.In(loc)
	today := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
	if !startDay.After(today) {
		return model.HospitalizationStatusAdmitted
	}
	return model.HospitalizationStatusReserved
}

func (s *hospitalizationService) Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	status := input.Status
	if status == "" {
		status = defaultHospitalizationStatus(input.StartDate, time.Now())
	}
	// is_insurance == false の場合は保険情報を NULL にする
	var insuranceCompanyName *string
	var insuranceNumber *string
	if input.IsInsurance {
		insuranceCompanyName = input.InsuranceCompanyName
		insuranceNumber = input.InsuranceNumber
	}

	// MRB-06: service-layer date order (request binding already checks when both present).
	if err := validateHospitalizationDateRange(input.StartDate, input.EndDate); err != nil {
		return nil, err
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
		if err := validatePetNotDeceased(txCtx, s.petRepo, clinicID, petID, "死亡したペットは入院登録できません"); err != nil {
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
		// Nested treatment plans share this TX (TASK-001-BE). Use plan repo only —
		// treatmentPlanService.Create opens its own outer transaction and would not join.
		if len(input.TreatmentPlans) > 0 {
			if s.treatmentPlanRepo == nil {
				return apperrors.WrapInternalServerError("hospitalization treatment plan repository is required for nested create")
			}
			hospID := hospitalization.ID
			for i := range input.TreatmentPlans {
				planInput := &input.TreatmentPlans[i]
				if err := validateTreatmentPlanMoney(planInput.UnitPrice, planInput.Quantity, planInput.DiscountRate, planInput.DiscountAmount); err != nil {
					return err
				}
				if planInput.TreatmentContent == "" {
					return apperrors.WrapInvalidInput("treatment_content is required")
				}
				subtotal := computeTreatmentPlanSubtotal(planInput.UnitPrice, planInput.Quantity, planInput.DiscountRate, planInput.DiscountAmount)
				plan := &model.TreatmentPlan{
					ClinicID:          clinicID,
					HospitalizationID: &hospID,
					TreatmentContent:  planInput.TreatmentContent,
					Memo:              planInput.Memo,
					IsInsurance:       planInput.IsInsurance,
					UnitPrice:         planInput.UnitPrice,
					Quantity:          planInput.Quantity,
					DiscountRate:      planInput.DiscountRate,
					DiscountAmount:    planInput.DiscountAmount,
					Subtotal:          subtotal,
					SortOrder:         planInput.SortOrder,
				}
				if err := s.treatmentPlanRepo.Create(txCtx, plan); err != nil {
					slog.ErrorContext(txCtx, "failed to create nested treatment plan", "error", err, "index", i)
					return apperrors.Wrap(err, "failed to create nested treatment plan")
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "hospitalization created",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.Uint64("clinic_id", hospitalization.ClinicID),
		slog.Int("treatment_plan_count", len(input.TreatmentPlans)))
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
		if existing == nil {
			return apperrors.WrapNotFound("hospitalization", strconv.FormatUint(id, 10))
		}
		// MRB-06: merge patch dates with locked row and reject inverted ranges.
		finalStart, finalEnd := existing.StartDate, existing.EndDate
		if input.StartDate != nil {
			finalStart = *input.StartDate
		}
		if input.EndDate != nil {
			finalEnd = *input.EndDate
		}
		if err := validateHospitalizationDateRange(finalStart, finalEnd); err != nil {
			return err
		}
		if input.OwnerID != nil || input.PetID != nil {
			finalOwnerID, finalPetID := resolveFinalHospitalizationOwnerPet(existing, input)
			if err := sharedkernel.ValidateReservationOwnerPetLinks(txCtx, s.reservationRepo, clinicID, finalOwnerID, finalPetID); err != nil {
				return err
			}
		}
		if input.PetID != nil {
			if err := validatePetNotDeceased(txCtx, s.petRepo, clinicID, *input.PetID, "死亡したペットは入院登録できません"); err != nil {
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
		// MRB-05: PATCH-driven discharge bypasses DischargeWithBilling; audit fail-closed.
		if input.Status != nil && *input.Status == model.HospitalizationStatusDischarged &&
			existing.Status != model.HospitalizationStatusDischarged {
			if s.auditTx == nil {
				return apperrors.WrapInternalServerError("hospitalization discharge audit dependency is required")
			}
			resourceID := id
			if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
				ClinicID:   &clinicID,
				ActorType:  auditActorTypeFor(nil),
				Action:     hospitalizationAuditActionDischarge,
				Resource:   model.AuditResourceHospitalization,
				ResourceID: &resourceID,
				OldValue:   hospitalizationAuditValue(existing),
				NewValue:   hospitalizationAuditValue(hosp),
				Metadata:   map[string]any{"via": "update"},
			}); err != nil {
				return apperrors.Wrap(err, "failed to audit hospitalization discharge")
			}
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

// hospitalizationAuditAction* are local action strings so this unit stays within owned paths
// (model/audit_log.go is not owned by U-X02X03X05-MR-HOSPITALIZATION).
const (
	hospitalizationAuditActionDelete    = "hospitalization.delete"
	hospitalizationAuditActionDischarge = "hospitalization.discharge"
)

func hospitalizationAuditValue(h *model.Hospitalization) map[string]any {
	if h == nil {
		return nil
	}
	return map[string]any{
		"id":         h.ID,
		"owner_id":   h.OwnerID,
		"pet_id":     h.PetID,
		"status":     string(h.Status),
		"start_date": h.StartDate,
		"end_date":   h.EndDate,
	}
}

func (s *hospitalizationService) Delete(ctx context.Context, clinicID, id uint64) error {
	existing, err := s.hospRepo.FindByID(ctx, clinicID, id)
	if err != nil {
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

	// MRB-05: soft-delete + fail-closed audit in one transaction.
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("hospitalization write transaction dependency is required")
	}
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("hospitalization delete audit dependency is required")
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.hospRepo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete hospitalization", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete hospitalization")
		}
		resourceID := id
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorType:  auditActorTypeFor(nil),
			Action:     hospitalizationAuditActionDelete,
			Resource:   model.AuditResourceHospitalization,
			ResourceID: &resourceID,
			OldValue:   hospitalizationAuditValue(existing),
		}); err != nil {
			return apperrors.Wrap(err, "failed to audit hospitalization delete")
		}
		return nil
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "hospitalization deleted",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}
