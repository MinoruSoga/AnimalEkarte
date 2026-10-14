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
	NextVisit            *time.Time
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

// UpdateHospitalizationInput は入院更新のサービス入力 DTO。
// ClearCageID は cage_id を NULL に戻す（「選択なし」への更新。EMR-316）。CageID との同時指定は不正。
type UpdateHospitalizationInput struct {
	OwnerID              *uint64
	PetID                *uint64
	HospitalizationType  *model.HospitalizationType
	StartDate            *time.Time
	EndDate              *time.Time
	NextVisit            *time.Time
	ClearNextVisit       bool
	Status               *model.HospitalizationStatus
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
	// EMR-353: clear は NULL 化、非 nil は設定、どちらも無ければ変更なし（PATCH 規約）。
	// IsInsurance トグルが早期 return するため、その前に置く（フロントは常にトグルを送る）。
	if input.ClearNextVisit {
		fields["next_visit"] = nil
	} else if input.NextVisit != nil {
		fields["next_visit"] = *input.NextVisit
	}
	if input.Status != nil {
		fields["status"] = *input.Status
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

type HospitalizationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DischargeWithBilling(ctx context.Context, clinicID, id uint64, input DischargeWithBillingInput) (*DischargeWithBillingResult, error)
	SwapCages(ctx context.Context, clinicID, idA, idB uint64, expectedCageIDs [2]uint64) (*model.Hospitalization, *model.Hospitalization, error)
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

// isOccupyingHospitalizationStatus はケージを占有する入院ステータスかを返す
// （CountOccupiedCages と同じ admitted/reserved 判定）。
func isOccupyingHospitalizationStatus(status model.HospitalizationStatus) bool {
	return status == model.HospitalizationStatusAdmitted || status == model.HospitalizationStatusReserved
}

// validateCageAssignmentGuard はケージ割当の連結ガード（EMR-306 Phase1 step3）。
// s.cageRepo が CageMergeRepository 面を持たない場合は 500 fail-closed（Grok review
// LOW-4: fail-open だと cageFinder だけのラッパー差し替えで構成ケージへの割当が
// 素通りするため、Merge/Unmerge と同じく依存不足は内部エラーにする）。
//
// ambient tx 内で validateOwnedMasterFK より前に呼ぶこと: cageRepo.FindByID は
// tx 中に FOR SHARE を取るため、先に LockCageGroupForUpdate (FOR UPDATE) で
// 同一行をロックしないと並行割当同士がデッドロックし得る（reviewer D1）。
// UPDATE 済み行への後続 SHARE は待機しないので、所有権検証は呼び出し位置のまま残す。
// LockCageGroupForUpdate 自体が医院スコープで、存在しない/他医院 id には NotFound を
// 返すため、クロステナント拒否セマンティクスはここでも同等に保証される。
//
// occupies は「この書込みの結果レコードがケージを占有するか」（結果ステータスが
// admitted/reserved）。false（discharged 履歴のバックフィル等）では構成ケージの
// 構造的禁止（422）は維持しつつ、グループ占有検査（409）はスキップする —
// 非占有の履歴行は物理的にケージを占有しないため。
// selfOccupiesCageID は更新経路で「この更新対象の入院が現在占有しているケージ」
// を渡す（create 経路は nil）。占有集計から自身を除外しないと、構成ケージ上の
// 既存入院を代表ケージへ移す唯一妥当なグループ内移動が自己衝突で 409 になる。
func (s *hospitalizationService) validateCageAssignmentGuard(txCtx context.Context, clinicID uint64, cageID *uint64, occupies bool, selfOccupiesCageID *uint64) error {
	if cageID == nil {
		return nil
	}
	mergeRepo, ok := s.cageRepo.(CageMergeRepository)
	if !ok {
		return apperrors.WrapInternalServerError("cage merge repository dependency is required")
	}
	// 代表+構成ケージを id 昇順の1文で FOR UPDATE（占有確認と同一 tx・同一行集合。
	// ロック順は merge/unmerge と同一の昇順でデッドロックを避ける）。
	group, err := mergeRepo.LockCageGroupForUpdate(txCtx, clinicID, *cageID)
	if err != nil {
		return apperrors.Wrap(err, "failed to lock cage group")
	}
	// ORDER BY id ASC なので group[0] は対象とは限らない — ID で走査する（reviewer D3）。
	var target *model.Cage
	for i := range group {
		if group[i].ID == *cageID {
			target = &group[i]
			break
		}
	}
	if target == nil {
		// LockCageGroupForUpdate は対象行の存在を保証するため理論上到達しない防御分岐。
		return apperrors.WrapNotFound("cage", strconv.FormatUint(*cageID, 10))
	}
	// 構成ケージは構造的に割当不可 (422)。代表/グループ判定より先に評価する。
	if target.MergedIntoCageID != nil {
		return apperrors.WrapUnprocessable("連結中の構成ケージには割り当てできません")
	}
	if len(group) <= 1 || !occupies {
		// 未連結ケージには占有チェックを追加しない（従来挙動を維持）。
		// 非占有ステータス（discharged 等）の結果レコードも占有検査をスキップする。
		return nil
	}
	// 代表ケージ: グループ内（代表+構成）の未削除 admitted/reserved 入院を占有とみなす。
	groupIDs := make([]uint64, 0, len(group))
	for i := range group {
		groupIDs = append(groupIDs, group[i].ID)
	}
	occupied, err := mergeRepo.CountOccupiedCages(txCtx, clinicID, groupIDs)
	if err != nil {
		return apperrors.Wrap(err, "failed to count cage group occupancy")
	}
	// 更新対象自身の占有を除外 — 構成ケージ→代表ケージのグループ内移動が
	// 自分自身の入院で 409 になるのを防ぐ（selfOccupiesCageID は占有中のみ非 nil）。
	if selfOccupiesCageID != nil {
		if n := occupied[*selfOccupiesCageID]; n > 1 {
			occupied[*selfOccupiesCageID] = n - 1
		} else {
			delete(occupied, *selfOccupiesCageID)
		}
	}
	if len(occupied) > 0 {
		return apperrors.WrapConflict("連結グループ内のケージが使用中のため割り当てできません")
	}
	return nil
}

func (s *hospitalizationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	result, total, err := s.hospRepo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list hospitalizations")
	}
	return result, total, nil
}

func (s *hospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	result, err := s.hospRepo.FindByID(ctx, clinicID, id)
	if err != nil {
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
	// EMR-316: ケージ未割当（cage_id=NULL）での作成を許可する（BUG-037 の必須化を撤回）。
	// 0 は存在し得ない id なので不正値として拒否する（NULL に丸めない）。
	if input.CageID != nil && *input.CageID == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgCageIDInvalid)
	}
	hospitalization := &model.Hospitalization{
		ClinicID:             clinicID,
		OwnerID:              input.OwnerID,
		PetID:                input.PetID,
		HospitalizationType:  input.HospitalizationType,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		NextVisit:            input.NextVisit,
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
		return s.createHospitalizationInTx(txCtx, clinicID, input, hospitalization)
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
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		existing, err := s.hospRepo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock hospitalization")
		}
		if existing == nil {
			return apperrors.WrapNotFound("hospitalization", strconv.FormatUint(id, 10))
		}
		if input.CageID != nil || input.ClearCageID {
			var cageStale bool
			if input.ExpectedCageIsNull {
				cageStale = existing.CageID != nil
			} else {
				cageStale = existing.CageID == nil || *existing.CageID != *input.ExpectedCageID
			}
			if cageStale {
				return apperrors.WrapConflict(errMsgCageClearStale)
			}
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
		// EMR-306 step3: ケージ割当ガード。ケージ変更または非占有→占有ステータス遷移でのみ
		// 発火する（reviewer D2）。FindByID(FOR SHARE) より先に FOR UPDATE を取るため、
		// validateOwnedMasterFK より前に置く（reviewer D1）。
		// PATCH の最終値を検証する。明示解除と再開が同時の場合、旧ケージは
		// 割当先ではないため、連結状態や占有を検査してはならない。
		finalCageID := existing.CageID
		if input.ClearCageID {
			finalCageID = nil
		} else if input.CageID != nil {
			finalCageID = input.CageID
		}
		cageChanged := finalCageID != nil &&
			(existing.CageID == nil || *finalCageID != *existing.CageID)
		finalStatus := existing.Status
		if input.Status != nil {
			finalStatus = *input.Status
		}
		transitioningIn := !isOccupyingHospitalizationStatus(existing.Status) &&
			isOccupyingHospitalizationStatus(finalStatus)
		if cageChanged || transitioningIn {
			// 結果ステータスが占有なら占有検査を行い、更新対象自身の占有は
			// 集計から除外する（メンバー→代表のグループ内移動を許可するため）。
			var selfOccupiesCageID *uint64
			if isOccupyingHospitalizationStatus(existing.Status) {
				selfOccupiesCageID = existing.CageID
			}
			if err := s.validateCageAssignmentGuard(txCtx, clinicID, finalCageID,
				isOccupyingHospitalizationStatus(finalStatus), selfOccupiesCageID); err != nil {
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
