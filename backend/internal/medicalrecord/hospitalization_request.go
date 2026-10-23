package medicalrecord

import (
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

type listHospitalizationQuery struct {
	PetID            string
	OwnerID          string
	LifecycleState   string
	PlannedEntryFrom string
	PlannedEntryTo   string
	// RetiredStatus / RetiredStartDate / RetiredEndDate are the pre-stay query names. They are read only
	// to be refused: ignoring them would silently return an unfiltered list.
	RetiredStatus    string
	RetiredStartDate string
	RetiredEndDate   string
}

func newListHospitalizationQuery(values url.Values) listHospitalizationQuery {
	return listHospitalizationQuery{
		PetID:            values.Get("pet_id"),
		OwnerID:          values.Get("owner_id"),
		LifecycleState:   values.Get("lifecycle_state"),
		PlannedEntryFrom: values.Get("planned_entry_from"),
		PlannedEntryTo:   values.Get("planned_entry_to"),
		RetiredStatus:    values.Get("status"),
		RetiredStartDate: values.Get("start_date"),
		RetiredEndDate:   values.Get("end_date"),
	}
}

type listHospitalizationFilters struct {
	PetID            *uint64
	OwnerID          *uint64
	LifecycleState   *string
	PlannedEntryFrom *string
	PlannedEntryTo   *string
}

func (q *listHospitalizationQuery) toServiceFilters() (listHospitalizationFilters, error) {
	for name, value := range map[string]string{"status": q.RetiredStatus, "start_date": q.RetiredStartDate, "end_date": q.RetiredEndDate} {
		if value != "" {
			return listHospitalizationFilters{}, apperrors.WrapInvalidInput(
				name + " は廃止されました。lifecycle_state / planned_entry_from / planned_entry_to を使用してください")
		}
	}
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	from, err := parseOptionalDateQueryFilter(q.PlannedEntryFrom, "planned_entry_from")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	to, err := parseOptionalDateQueryFilter(q.PlannedEntryTo, "planned_entry_to")
	if err != nil {
		return listHospitalizationFilters{}, err
	}
	state := optionalStringQueryFilter(q.LifecycleState)
	if state != nil && !validLifecycleState(*state) {
		return listHospitalizationFilters{}, apperrors.WrapInvalidInput("lifecycle_state is invalid")
	}
	return listHospitalizationFilters{
		PetID:            petID,
		OwnerID:          ownerID,
		LifecycleState:   state,
		PlannedEntryFrom: from,
		PlannedEntryTo:   to,
	}, nil
}

// dischargeWithBillingRequest は退院+会計作成のバインド struct
type dischargeWithBillingRequest struct {
	DischargeDate    time.Time `json:"discharge_date" binding:"required"`
	CreateAccounting bool      `json:"create_accounting"`
}

func (r dischargeWithBillingRequest) toServiceInput(actorID uint64) DischargeWithBillingInput {
	return DischargeWithBillingInput{
		DischargeDate:    r.DischargeDate,
		CreateAccounting: r.CreateAccounting,
		ActorID:          &actorID,
	}
}

// parseHospitalizationNextVisit は next_visit の境界パーサ（EMR-353）。
// 省略(nil)は「変更なし」で nil を返す。非 nil は必須値なので空文字も含め
// YYYY-MM-DD 以外（空白/不正日付/RFC3339 等）はすべて 400 とする。
func parseHospitalizationNextVisit(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	if *raw == "" {
		return nil, apperrors.WrapInvalidInput("next_visit は YYYY-MM-DD 形式で入力してください")
	}
	return httpapi.ParseOptionalDateTimeField(*raw, "next_visit")
}

// createHospitalizationRequest は入院作成のバインド struct
type createHospitalizationRequest struct {
	OwnerID        uint64     `json:"owner_id"               binding:"required"`
	PetID          uint64     `json:"pet_id"                 binding:"required"`
	StayKind       string     `json:"stay_kind"              binding:"required,oneof=inpatient hotel"`
	PlannedEntryAt *time.Time `json:"planned_entry_at"`
	PlannedExitAt  *time.Time `json:"planned_exit_at"`
	NextVisit      *string    `json:"next_visit"`
	// The pre-stay lifecycle names are bound only to be refused: ignoring them would silently drop a
	// value the caller believes was saved (EMR-499 U1; the lifecycle is derived from movement events).
	RetiredType          *string                      `json:"hospitalization_type"`
	RetiredStartDate     *time.Time                   `json:"start_date"`
	RetiredEndDate       *time.Time                   `json:"end_date"`
	RetiredStatus        *string                      `json:"status"`
	CageID               *uint64                      `json:"cage_id"`
	DoctorID             *uint64                      `json:"doctor_id"`
	Memo                 string                       `json:"memo"`
	OwnerRequest         string                       `json:"owner_request"`
	StaffNotes           string                       `json:"staff_notes"`
	IsInsurance          bool                         `json:"is_insurance"`
	InsuranceCompanyName *string                      `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string                      `json:"insurance_number,omitempty"`
	TreatmentPlans       []createTreatmentPlanRequest `json:"treatment_plans"`
}

func (r *createHospitalizationRequest) rejectRetiredFields() error {
	switch {
	case r.RetiredType != nil:
		return errStayLifecycleFieldNotWritable("hospitalization_type")
	case r.RetiredStartDate != nil:
		return errStayLifecycleFieldNotWritable("start_date")
	case r.RetiredEndDate != nil:
		return errStayLifecycleFieldNotWritable("end_date")
	case r.RetiredStatus != nil:
		return errStayLifecycleFieldNotWritable("status")
	}
	return nil
}

func (r *createHospitalizationRequest) toServiceInput() (*CreateHospitalizationInput, error) {
	if err := r.rejectRetiredFields(); err != nil {
		return nil, err
	}
	if err := validatePlannedStayInterval(r.PlannedEntryAt, r.PlannedExitAt); err != nil {
		return nil, err
	}

	// EMR-316: cage_id の省略 / null は「ケージ未割当」として受理する（BUG-037 の必須化を撤回）。
	// 0 は存在し得ない id なので不正値として拒否する。
	if r.CageID != nil && *r.CageID == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgCageIDInvalid)
	}

	plans := make([]CreateTreatmentPlanInput, 0, len(r.TreatmentPlans))
	for i := range r.TreatmentPlans {
		plan := r.TreatmentPlans[i].toServiceInput()
		if plan != nil {
			plans = append(plans, *plan)
		}
	}

	nextVisit, err := parseHospitalizationNextVisit(r.NextVisit)
	if err != nil {
		return nil, err
	}

	return &CreateHospitalizationInput{
		OwnerID:              r.OwnerID,
		PetID:                r.PetID,
		StayKind:             r.StayKind,
		PlannedEntryAt:       r.PlannedEntryAt,
		PlannedExitAt:        r.PlannedExitAt,
		NextVisit:            nextVisit,
		CageID:               r.CageID,
		DoctorID:             r.DoctorID,
		Memo:                 r.Memo,
		OwnerRequest:         r.OwnerRequest,
		StaffNotes:           r.StaffNotes,
		IsInsurance:          r.IsInsurance,
		InsuranceCompanyName: r.InsuranceCompanyName,
		InsuranceNumber:      r.InsuranceNumber,
		TreatmentPlans:       plans,
	}, nil
}

// cageSwapRequest は入院ケージ入れ替えのバインド struct（EMR-306 review: ボード DnD の
// 入れ替えが2本の独立 PATCH では片方だけ commit され得るため、1リクエストで受けて
// 単一トランザクションで交換する）。ExpectedCageIDs は HospitalizationIDs と
// 同じ位置の入院の表示時 cage_id で、古いボードからの誤交換を拒否する。
type cageSwapRequest struct {
	HospitalizationIDs []uint64 `json:"hospitalization_ids" binding:"required,len=2,dive,required,gt=0"`
	ExpectedCageIDs    []uint64 `json:"expected_cage_ids" binding:"required,len=2,dive,required,gt=0"`
}

// updateHospitalizationRequest は入院更新のバインド struct。
// ClearCageID=true で cage_id を NULL に戻す（「選択なし」。EMR-316）。cage_id との同時指定は 400。
type updateHospitalizationRequest struct {
	OwnerID *uint64 `json:"owner_id"`
	PetID   *uint64 `json:"pet_id"`
	// Stay kind, planned interval and the retired lifecycle names are bound only to be refused with a
	// stable 422 (EMR-499 U1): a planned change belongs to extension proposals / acceptance proof, and
	// the lifecycle is derived. Ignoring them would silently drop a value the caller thinks was saved.
	StayKind             *string    `json:"stay_kind"`
	PlannedEntryAt       *time.Time `json:"planned_entry_at"`
	PlannedExitAt        *time.Time `json:"planned_exit_at"`
	HospitalizationType  *string    `json:"hospitalization_type"`
	StartDate            *time.Time `json:"start_date"`
	EndDate              *time.Time `json:"end_date"`
	Status               *string    `json:"status"`
	NextVisit            *string    `json:"next_visit"`
	ClearNextVisit       bool       `json:"clear_next_visit"`
	CageID               *uint64    `json:"cage_id"`
	ClearCageID          bool       `json:"clear_cage_id"`
	ExpectedCageID       *uint64    `json:"expected_cage_id"`
	ExpectedCageIsNull   bool       `json:"expected_cage_is_null"`
	DoctorID             *uint64    `json:"doctor_id"`
	Memo                 *string    `json:"memo"`
	OwnerRequest         *string    `json:"owner_request"`
	StaffNotes           *string    `json:"staff_notes"`
	IsInsurance          *bool      `json:"is_insurance,omitempty"`
	InsuranceCompanyName *string    `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string    `json:"insurance_number,omitempty"`
}

// errStayLifecycleFieldNotWritable refuses a write of a retired hospitalization field (422, stable code).
func errStayLifecycleFieldNotWritable(field string) error {
	return apperrors.WrapUnprocessableWithCode(codeStayLifecycleWriteUnavailable,
		field+" は入院の滞在・状態の導出値になったため、この API では変更できません")
}

func (r *updateHospitalizationRequest) toServiceInput() (UpdateHospitalizationInput, error) {
	switch {
	case r.StayKind != nil:
		return UpdateHospitalizationInput{}, errStayLifecycleFieldNotWritable("stay_kind")
	case r.PlannedEntryAt != nil:
		return UpdateHospitalizationInput{}, errStayLifecycleFieldNotWritable("planned_entry_at")
	case r.PlannedExitAt != nil:
		return UpdateHospitalizationInput{}, errStayLifecycleFieldNotWritable("planned_exit_at")
	case r.HospitalizationType != nil:
		return UpdateHospitalizationInput{}, errStayLifecycleFieldNotWritable("hospitalization_type")
	case r.StartDate != nil:
		return UpdateHospitalizationInput{}, errStayLifecycleFieldNotWritable("start_date")
	case r.EndDate != nil:
		return UpdateHospitalizationInput{}, errStayLifecycleFieldNotWritable("end_date")
	case r.Status != nil:
		return UpdateHospitalizationInput{}, errStayLifecycleFieldNotWritable("status")
	}
	if err := validateCageUpdateGuards(r.CageID, r.ClearCageID, r.ExpectedCageID, r.ExpectedCageIsNull); err != nil {
		return UpdateHospitalizationInput{}, err
	}
	// EMR-353: clear_next_visit と next_visit の同時指定は矛盾なので拒否する。
	if r.ClearNextVisit && r.NextVisit != nil {
		return UpdateHospitalizationInput{}, apperrors.WrapInvalidInput("next_visit と clear_next_visit は同時に指定できません")
	}
	nextVisit, err := parseHospitalizationNextVisit(r.NextVisit)
	if err != nil {
		return UpdateHospitalizationInput{}, err
	}
	input := UpdateHospitalizationInput{
		OwnerID:              r.OwnerID,
		PetID:                r.PetID,
		NextVisit:            nextVisit,
		ClearNextVisit:       r.ClearNextVisit,
		CageID:               r.CageID,
		ClearCageID:          r.ClearCageID,
		ExpectedCageID:       r.ExpectedCageID,
		ExpectedCageIsNull:   r.ExpectedCageIsNull,
		DoctorID:             r.DoctorID,
		Memo:                 r.Memo,
		OwnerRequest:         r.OwnerRequest,
		StaffNotes:           r.StaffNotes,
		IsInsurance:          r.IsInsurance,
		InsuranceCompanyName: r.InsuranceCompanyName,
		InsuranceNumber:      r.InsuranceNumber,
	}
	return input, nil
}
