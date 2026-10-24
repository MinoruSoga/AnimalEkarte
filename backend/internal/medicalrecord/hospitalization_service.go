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
	OwnerID uint64
	PetID   uint64
	// StayKind / PlannedEntryAt / PlannedExitAt are stored on stay_cases (EMR-499 U1). They are not
	// lifecycle facts: Create records no movement event, so the effective state starts as unknown.
	StayKind             string
	PlannedEntryAt       *time.Time
	PlannedExitAt        *time.Time
	NextVisit            *time.Time
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

// UpdateHospitalizationInput は入院更新のサービス入力 DTO。
// ClearCageID は cage_id を NULL に戻す（「選択なし」への更新。EMR-316）。CageID との同時指定は不正。
type UpdateHospitalizationInput struct {
	OwnerID              *uint64
	PetID                *uint64
	NextVisit            *time.Time
	ClearNextVisit       bool
	CageID               *uint64
	ClearCageID          bool
	ExpectedCageID       *uint64
	ExpectedCageIsNull   bool
	DoctorID             *uint64
	Memo                 *string
	OwnerRequest         *string
	StaffNotes           *string
	IsInsurance          *bool
	InsuranceCompanyName *string
	InsuranceNumber      *string
}

// validatePetNotDeceased は死亡ペットへの業務書き込みをブロックする（SD-10・臨床安全）。
// 実装正本は sharedkernel.ValidatePetNotDeceased（#261 P0 で昇格・EMR-491 D6 で reader 化）。
// message は呼び出し元の業務文言。reader は ambient transaction 内で呼ぶこと。
func validatePetNotDeceased(ctx context.Context, petLife sharedkernel.PetLifeStateReader, clinicID, petID uint64, message string) error {
	return sharedkernel.ValidatePetNotDeceased(ctx, petLife, clinicID, petID, message)
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
	// EMR-353: clear は NULL 化、非 nil は設定、どちらも無ければ変更なし（PATCH 規約）。
	// IsInsurance トグルが早期 return するため、その前に置く（フロントは常にトグルを送る）。
	if input.ClearNextVisit {
		fields["next_visit"] = nil
	} else if input.NextVisit != nil {
		fields["next_visit"] = *input.NextVisit
	}
	// EMR-316: ClearCageID は NULL 化、非 nil は設定、どちらも無ければ変更なし（PATCH 規約）。
	setNullableUint64Field(fields, colHospitalizationCageID, input.ClearCageID, input.CageID)
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
		applyHospitalizationInsuranceToggle(fields, input)
		return fields
	}
	if input.InsuranceCompanyName != nil {
		fields["insurance_company_name"] = *input.InsuranceCompanyName
	}
	if input.InsuranceNumber != nil {
		fields["insurance_number"] = *input.InsuranceNumber
	}
	return fields
}

func applyHospitalizationInsuranceToggle(fields map[string]any, input *UpdateHospitalizationInput) {
	if !*input.IsInsurance {
		fields["insurance_company_name"] = nil
		fields["insurance_number"] = nil
		return
	}
	if input.InsuranceCompanyName != nil {
		fields["insurance_company_name"] = *input.InsuranceCompanyName
	}
	if input.InsuranceNumber != nil {
		fields["insurance_number"] = *input.InsuranceNumber
	}
}

// HospitalizationWriteResult は入院の書き込み（Create / Update / SwapCages）の結果。
// Stay は書き込みと同一トランザクション内で導出した滞在 projection で、導出に失敗すると書き込みごと
// ロールバックされる（commit 済みの書き込みをエラー応答として返さない）。
// *model.Hospitalization を埋め込むため、行のフィールドはそのまま参照できる。
type HospitalizationWriteResult struct {
	*model.Hospitalization
	Stay HospitalizationStayProjection
}

type HospitalizationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, lifecycleState, plannedEntryFrom, plannedEntryTo *string, page, limit int) ([]model.Hospitalization, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	// StayProjections returns the read-only stay view (type, planned dates, effective lifecycle) per hospitalization ID.
	StayProjections(ctx context.Context, clinicID uint64, rows []model.Hospitalization) (map[uint64]HospitalizationStayProjection, error)
	// Create / Update / SwapCages は書き込み結果に同一 tx 内で導出した滞在 projection を添えて返す（commit 後の再読込は不要）。
	Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*HospitalizationWriteResult, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*HospitalizationWriteResult, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error)
	SwapCages(ctx context.Context, clinicID, idA, idB uint64, expectedCageIDs [2]uint64) (*HospitalizationWriteResult, *HospitalizationWriteResult, error)
}

type hospitalizationService struct {
	hospRepo          HospitalizationRepository
	reservationRepo   sharedkernel.OwnerPetLinkVerifier
	doctorVerifier    hospitalizationDoctorVerifier
	petRepo           petFinder
	petLife           sharedkernel.PetLifeStateReader
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

// WithPetLifeStateReader injects the EMR-491 D6 death reader used by the create / update / cage-swap
// write guards. When it is not injected those writes fail closed (never treated as alive).
func WithPetLifeStateReader(reader sharedkernel.PetLifeStateReader) HospitalizationServiceOption {
	return func(s *hospitalizationService) {
		s.petLife = reader
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

func (s *hospitalizationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, lifecycleState, plannedEntryFrom, plannedEntryTo *string, page, limit int) ([]model.Hospitalization, int64, error) {
	if lifecycleState == nil {
		result, total, err := s.hospRepo.FindAll(ctx, clinicID, petID, ownerID, plannedEntryFrom, plannedEntryTo, page, limit)
		if err != nil {
			return nil, 0, apperrors.Wrap(err, "failed to list hospitalizations")
		}
		return result, total, nil
	}
	return s.listByLifecycleState(ctx, clinicID, petID, ownerID, *lifecycleState, plannedEntryFrom, plannedEntryTo, page, limit)
}

func (s *hospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	result, err := s.hospRepo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get hospitalization")
	}
	return result, nil
}

func (s *hospitalizationService) Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*HospitalizationWriteResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	// is_insurance == false の場合は保険情報を NULL にする
	var insuranceCompanyName *string
	var insuranceNumber *string
	if input.IsInsurance {
		insuranceCompanyName = input.InsuranceCompanyName
		insuranceNumber = input.InsuranceNumber
	}

	if err := validatePlannedStayInterval(input.PlannedEntryAt, input.PlannedExitAt); err != nil {
		return nil, err
	}
	// EMR-316: ケージ未割当（cage_id=NULL）での作成を許可する（BUG-037 の必須化を撤回）。
	// 0 は存在し得ない id なので不正値として拒否する（NULL に丸めない）。
	if input.CageID != nil && *input.CageID == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgCageIDInvalid)
	}
	if !validStayKind(input.StayKind) {
		return nil, apperrors.WrapInvalidInput("stay_kind must be inpatient or hotel")
	}
	hospitalization := &model.Hospitalization{
		ClinicID:             clinicID,
		OwnerID:              input.OwnerID,
		PetID:                input.PetID,
		NextVisit:            input.NextVisit,
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
	// 滞在 projection は作成と同じ tx（txCtx）の中で導出する。失敗すれば作成ごとロールバックされ、
	// commit 済みの作成をエラーとして返して重複再送を誘発することがない。
	var stay HospitalizationStayProjection
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.createHospitalizationInTx(txCtx, clinicID, input, hospitalization); err != nil {
			return err
		}
		var err error
		stay, err = s.writtenStayProjectionInTx(txCtx, clinicID, hospitalization)
		return err
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "hospitalization created",
		slog.Uint64("hospitalization_id", hospitalization.ID),
		slog.Uint64("clinic_id", hospitalization.ClinicID),
		slog.Int("treatment_plan_count", len(input.TreatmentPlans)))
	return &HospitalizationWriteResult{Hospitalization: hospitalization, Stay: stay}, nil
}

func (s *hospitalizationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*HospitalizationWriteResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	// EMR-316: NULL 化と値設定の同時指定は意図が曖昧なので拒否する（clear が黙って勝たない）。
	if err := validateCageUpdateGuards(input.CageID, input.ClearCageID, input.ExpectedCageID, input.ExpectedCageIsNull); err != nil {
		return nil, err
	}
	// EMR-353: clear_next_visit と next_visit の同時指定は direct caller も拒否する。
	if input.ClearNextVisit && input.NextVisit != nil {
		return nil, apperrors.WrapInvalidInput("next_visit と clear_next_visit は同時に指定できません")
	}
	fields := buildHospitalizationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("hospitalization write transaction dependency is required")
	}
	var hosp *model.Hospitalization
	var stay HospitalizationStayProjection
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		existing, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock hospitalization")
		}
		if existing == nil {
			return apperrors.WrapNotFound("hospitalization", strconv.FormatUint(id, 10))
		}
		// 滞在に紐づく入院の患者（pet_id）は stay_cases.pet_id と一致していなければならない。入院行だけの
		// 付け替えは滞在側の予定・実績を別の患者に見せてしまうため、黙って更新せず明示的に拒否する。
		if input.PetID != nil && *input.PetID != existing.PetID && existing.StayCaseID != nil && *existing.StayCaseID != 0 {
			return apperrors.WrapConflict("滞在に紐づく入院の患者は、この API では変更できません")
		}
		if input.CageID != nil || input.ClearCageID {
			var cageStale bool
			if input.ExpectedCageIsNull {
				cageStale = existing.CageID != nil
			} else {
				cageStale = existing.CageID == nil || *existing.CageID != *input.ExpectedCageID
			}
			if cageStale {
				return hospitalizationCageConflict(codeCageAssignmentChanged, errMsgCageClearStale)
			}
		}
		if input.OwnerID != nil || input.PetID != nil {
			finalOwnerID, finalPetID := resolveFinalHospitalizationOwnerPet(existing, input)
			if err := sharedkernel.ValidateReservationOwnerPetLinks(txCtx, s.reservationRepo, clinicID, finalOwnerID, finalPetID); err != nil {
				return err
			}
		}
		if input.PetID != nil {
			if err := validatePetNotDeceased(txCtx, s.petLife, clinicID, *input.PetID, "死亡したペットは入院登録できません"); err != nil {
				return err
			}
		}
		// Only a new non-NULL assignment needs allocation proof. Omission,
		// retention of the locked cage, and explicit clearing keep their meaning.
		if input.CageID != nil && (existing.CageID == nil || *input.CageID != *existing.CageID) {
			group, err := s.lockCageAssignmentGroup(txCtx, clinicID, *input.CageID)
			if err != nil {
				return err
			}
			if err := s.validateCageAssignmentSubjects(txCtx, clinicID, []*model.Hospitalization{existing}); err != nil {
				return err
			}
			if err := s.validateCageAssignmentOccupancy(txCtx, clinicID, group, []uint64{existing.ID}); err != nil {
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
		hosp, err = s.hospRepo.Update(txCtx, clinicID, id, *input)
		if err != nil {
			return apperrors.Wrap(err, "failed to update hospitalization")
		}
		// 滞在 projection は更新と同じ tx 内で導出する。失敗すれば更新ごとロールバックされる。
		stay, err = s.writtenStayProjectionInTx(txCtx, clinicID, hosp)
		return err
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "hospitalization updated",
		slog.Uint64("hospitalization_id", id),
		slog.Uint64("clinic_id", clinicID))
	return &HospitalizationWriteResult{Hospitalization: hosp, Stay: stay}, nil
}

// hospitalizationAuditAction* are local action strings so this unit stays within owned paths
// (model/audit_log.go is not owned by U-X02X03X05-MR-HOSPITALIZATION).
const (
	hospitalizationAuditActionDelete    = "hospitalization.delete"
	hospitalizationAuditActionDischarge = "hospitalization.discharge"
)

// EMR-316: cage_id（NULL 許容 FK）の列名とエラー文言。request / service で共有する。
const (
	colHospitalizationCageID        = "cage_id"
	errMsgCageIDInvalid             = "cage_id must be a positive id"
	errMsgCageIDClearConflict       = "clear_cage_id and cage_id cannot be set together"
	errMsgCageMissingExpected       = "expected_cage_id or expected_cage_is_null is required for cage changes"
	errMsgCageExpectedConflict      = "expected_cage_id and expected_cage_is_null cannot be set together"
	errMsgCageExpectedWithoutChange = "expected_cage_id and expected_cage_is_null require cage_id or clear_cage_id"
	errMsgCageClearStale            = "the cage assignment has changed"
)

func validateCageUpdateGuards(cageID *uint64, clearCage bool, expectedCageID *uint64, expectedNull bool) error {
	if clearCage && cageID != nil {
		return apperrors.WrapInvalidInput(errMsgCageIDClearConflict)
	}
	if cageID != nil && *cageID == 0 {
		return apperrors.WrapInvalidInput(errMsgCageIDInvalid)
	}
	if expectedCageID != nil && *expectedCageID == 0 {
		return apperrors.WrapInvalidInput(errMsgCageIDInvalid)
	}
	if expectedCageID != nil && expectedNull {
		return apperrors.WrapInvalidInput(errMsgCageExpectedConflict)
	}
	hasCageOp := cageID != nil || clearCage
	if hasCageOp && expectedCageID == nil && !expectedNull {
		return apperrors.WrapInvalidInput(errMsgCageMissingExpected)
	}
	if !hasCageOp && (expectedCageID != nil || expectedNull) {
		return apperrors.WrapInvalidInput(errMsgCageExpectedWithoutChange)
	}
	return nil
}

func hospitalizationAuditValue(h *model.Hospitalization) map[string]any {
	if h == nil {
		return nil
	}
	return map[string]any{
		"id":       h.ID,
		"owner_id": h.OwnerID,
		"pet_id":   h.PetID,
		// EMR-499 U1: status/start_date/end_date left the hospitalizations row. The stay case is the
		// stable reference; lifecycle facts are audited by their own movement events.
		"stay_case_id": h.StayCaseID,
		"cage_id":      h.CageID,
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
		return apperrors.Wrap(err, "failed to check daily record dependencies")
	}
	if dailyCount > 0 {
		return apperrors.WrapConflict("日次記録が紐付いているため削除できません。先に日次記録を削除してください")
	}

	// FK依存チェック: 入院に紐付く治療計画が存在する場合は削除を拒否
	planCount, err := s.hospRepo.CountTreatmentPlansByHospitalizationID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check treatment plan dependencies")
	}
	if planCount > 0 {
		return apperrors.WrapConflict("治療計画が紐付いているため削除できません。先に治療計画を削除してください")
	}

	// FK依存チェック: 入院に紐付くケアプラン項目が存在する場合は削除を拒否
	itemCount, err := s.hospRepo.CountCarePlanItemsByHospitalizationID(ctx, clinicID, id)
	if err != nil {
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
