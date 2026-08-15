package medicalrecord

import (
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type listMedicalRecordQuery struct {
	PetID           string
	OwnerID         string
	StartDate       string
	EndDate         string
	Search          string
	Status          string
	DoctorID        string
	AnimalSpeciesID string
	Sort            string
	Order           string
}

func newListMedicalRecordQuery(values url.Values) listMedicalRecordQuery {
	return listMedicalRecordQuery{
		PetID:           values.Get("pet_id"),
		OwnerID:         values.Get("owner_id"),
		StartDate:       values.Get("start_date"),
		EndDate:         values.Get("end_date"),
		Search:          values.Get("search"),
		Status:          values.Get("status"),
		DoctorID:        values.Get("doctor_id"),
		AnimalSpeciesID: values.Get("animal_species_id"),
		Sort:            values.Get("sort"),
		Order:           values.Get("order"),
	}
}

// allowedMedicalRecordSortKeys は B-1 follow-up の列ソート server 化で受理する sort query の許可値。
// medicalRecordSortColumns と1対1対応させ、リポジトリ側の防御的 map lookup と二重に守る。
var allowedMedicalRecordSortKeys = map[string]struct{}{
	"date":       {},
	"owner_name": {},
	"pet_name":   {},
	"status":     {},
}

// resolveMedicalRecordSort は FE の sort/order query を検証済みの (sort, order) に正規化する。
// 許可されていない sort キーは空文字（= repository 側で既定順にフォールバック）にする。
// order は "asc"/"desc" のいずれでもない場合は "desc" にフォールバックする。
func resolveMedicalRecordSort(sort, order string) (sortKey, sortOrder string) {
	if _, ok := allowedMedicalRecordSortKeys[sort]; !ok {
		return "", ""
	}
	if order == "asc" {
		return sort, "asc"
	}
	return sort, "desc"
}

// listMedicalRecordFilters は互換のためのエイリアス。実体は MedicalRecordListFilters。
type listMedicalRecordFilters = MedicalRecordListFilters

func (q *listMedicalRecordQuery) toServiceFilters() (listMedicalRecordFilters, error) {
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	startDate, err := parseOptionalDateQueryFilter(q.StartDate, "start_date")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	endDate, err := parseOptionalDateQueryFilter(q.EndDate, "end_date")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	doctorID, err := parseOptionalUintQueryFilter(q.DoctorID, "doctor_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	animalSpeciesID, err := parseOptionalUintQueryFilter(q.AnimalSpeciesID, "animal_species_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	var status *model.MedicalRecordStatus
	if q.Status != "" {
		parsed, err := validateEnum(q.Status,
			model.MedicalRecordStatusDraft,
			model.MedicalRecordStatusFinalized,
		)
		if err != nil {
			return listMedicalRecordFilters{}, apperrors.WrapInvalidInput("invalid status: " + err.Error())
		}
		status = &parsed
	}
	sortKey, sortOrder := resolveMedicalRecordSort(q.Sort, q.Order)
	return listMedicalRecordFilters{
		PetID:           petID,
		OwnerID:         ownerID,
		StartDate:       startDate,
		EndDate:         endDate,
		Status:          status,
		DoctorID:        doctorID,
		AnimalSpeciesID: animalSpeciesID,
		Search:          q.Search,
		Sort:            sortKey,
		Order:           sortOrder,
	}, nil
}

// createMedicalRecordRequest はカルテ作成のバインド struct
// FE互換: FEから送信されるフィールドと BE期待値を統一
type createMedicalRecordRequest struct {
	// 基本フィールド
	RecordNo      string     `json:"record_no"`                   // optional; 自動生成される
	Date          *time.Time `json:"date"`                        // optional
	VisitDate     *string    `json:"visit_date"`                  // FE送信フィールド（"YYYY-MM-DD"形式）
	VisitType     string     `json:"visit_type"`                  // "first" | "revisit" | "" (省略時は revisit にデフォルト)
	OwnerID       *string    `json:"owner_id" binding:"required"` // FE送信（string）→ uint64に変換
	PetID         *string    `json:"pet_id" binding:"required"`   // FE送信（string）→ uint64に変換
	DoctorID      *string    `json:"doctor_id"`                   // FE送信（string）→ uint64に変換
	AppointmentID *string    `json:"appointment_id"`              // FE送信（string）→ uint64に変換
	Status        string     `json:"status"                       binding:"omitempty,oneof=draft finalized"`

	// BE-006: 次回来院推奨日
	NextVisitRecommendedDate *string `json:"next_visit_recommended_date"` // "YYYY-MM-DD" or null

	// ClinicalPlan関連フィールド（原子的作成用）
	ChiefComplaint       *string `json:"chief_complaint"`
	ChiefComplaintTypeID *uint64 `json:"chief_complaint_type_id"`
	Plan                 *string `json:"plan"`       // ClinicalPlan.TreatmentPolicyへ
	Assessment           *string `json:"assessment"` // ClinicalPlan.DiagnosisDetailsへ
	Notes                *string `json:"notes"`
	Diagnosis1CategoryID *uint64 `json:"diagnosis_1_category_id"`
	Diagnosis1NameID     *uint64 `json:"diagnosis_1_name_id"`
	Diagnosis2TypeID     *uint64 `json:"diagnosis_2_type_id"`
	Diagnosis2NameID     *uint64 `json:"diagnosis_2_name_id"`

	// 受診推奨理由（FEAT-382-2 supplement: 新規作成時受付）
	// 値域: revisit / checkup / prevention / exam / 空 ("")
	RecommendationReason *string `json:"recommendation_reason"`
}

func (r *createMedicalRecordRequest) toServiceInput(staffID uint64) (CreateMedicalRecordInput, error) {
	recordDate, err := r.recordDate()
	if err != nil {
		return CreateMedicalRecordInput{}, err
	}

	if r.OwnerID == nil || *r.OwnerID == "" {
		return CreateMedicalRecordInput{}, apperrors.WrapInvalidInput("owner_id is required")
	}
	if r.PetID == nil || *r.PetID == "" {
		return CreateMedicalRecordInput{}, apperrors.WrapInvalidInput("pet_id is required")
	}

	ownerID, err := parseOptionalMedicalRecordID(r.OwnerID, "owner_id")
	if err != nil {
		return CreateMedicalRecordInput{}, err
	}
	petID, err := parseOptionalMedicalRecordID(r.PetID, "pet_id")
	if err != nil {
		return CreateMedicalRecordInput{}, err
	}
	doctorID, err := parseOptionalMedicalRecordID(r.DoctorID, "doctor_id")
	if err != nil {
		return CreateMedicalRecordInput{}, err
	}
	appointmentID, err := parseOptionalMedicalRecordID(r.AppointmentID, "appointment_id")
	if err != nil {
		return CreateMedicalRecordInput{}, err
	}

	nextVisitDate, err := r.nextVisitDate(recordDate)
	if err != nil {
		return CreateMedicalRecordInput{}, err
	}

	visitType := model.VisitTypeRevisit
	if r.VisitType == string(model.VisitTypeFirst) {
		visitType = model.VisitTypeFirst
	}

	var status *model.MedicalRecordStatus
	if r.Status != "" {
		parsedStatus, err := validateEnum(r.Status,
			model.MedicalRecordStatusDraft,
			model.MedicalRecordStatusFinalized,
		)
		if err != nil {
			return CreateMedicalRecordInput{}, apperrors.WrapInvalidInput("invalid status: " + err.Error())
		}
		status = &parsedStatus
	}
	var recommendationReason *string
	if r.RecommendationReason != nil && *r.RecommendationReason != "" {
		recommendationReason = r.RecommendationReason
	}

	return CreateMedicalRecordInput{
		RecordNo:                 r.RecordNo,
		Date:                     recordDate,
		OwnerID:                  ownerID,
		PetID:                    petID,
		DoctorID:                 doctorID,
		AppointmentID:            appointmentID,
		Status:                   status,
		VisitType:                visitType,
		NextVisitRecommendedDate: nextVisitDate,
		RecommendationReason:     recommendationReason,
		EnteredBy:                &staffID,
	}, nil
}

func (r *createMedicalRecordRequest) recordDate() (time.Time, error) {
	switch {
	case r.Date != nil:
		return *r.Date, nil
	case r.VisitDate != nil && *r.VisitDate != "":
		parsed, err := time.ParseInLocation(time.DateOnly, *r.VisitDate, time.Local)
		if err != nil {
			return time.Time{}, apperrors.WrapInvalidInput("invalid visit_date format (expected YYYY-MM-DD)")
		}
		return parsed, nil
	default:
		return time.Now(), nil
	}
}

func (r *createMedicalRecordRequest) nextVisitDate(recordDate time.Time) (*time.Time, error) {
	if r.NextVisitRecommendedDate == nil || *r.NextVisitRecommendedDate == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation(time.DateOnly, *r.NextVisitRecommendedDate, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid next_visit_recommended_date format (expected YYYY-MM-DD)")
	}
	if !parsed.After(recordDate) {
		return nil, apperrors.WrapInvalidInput("next_visit_recommended_date must be after the record date")
	}
	if parsed.After(recordDate.AddDate(2, 0, 0)) {
		return nil, apperrors.WrapInvalidInput("next_visit_recommended_date must be within 2 years")
	}
	return &parsed, nil
}

func (r *createMedicalRecordRequest) toSubRecordsInput() CreateSubRecordsInput {
	return CreateSubRecordsInput{
		ChiefComplaintTypeID: r.ChiefComplaintTypeID,
		ChiefComplaint:       r.ChiefComplaint,
		Notes:                r.Notes,
		Plan:                 r.Plan,
		Assessment:           r.Assessment,
		Diagnosis1CategoryID: r.Diagnosis1CategoryID,
		Diagnosis1NameID:     r.Diagnosis1NameID,
		Diagnosis2TypeID:     r.Diagnosis2TypeID,
		Diagnosis2NameID:     r.Diagnosis2NameID,
	}
}

func parseOptionalMedicalRecordID(s *string, field string) (*uint64, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	id, err := strconv.ParseUint(*s, 10, 64)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid " + field + " (must be numeric)")
	}
	return &id, nil
}

// patchMedicalRecordRecommendationReasonRequest は受診推奨理由更新リクエスト（FEAT-381-2）。
// Reason は revisit / checkup / prevention / exam のいずれか、または "" (未設定)。
type patchMedicalRecordRecommendationReasonRequest struct {
	Reason string `json:"reason" binding:"max=100"`
}

func (r patchMedicalRecordRecommendationReasonRequest) toServiceInput() UpdateRecommendationReasonInput {
	return UpdateRecommendationReasonInput{Reason: r.Reason}
}

// updateMedicalRecordRequest はカルテ更新のバインド struct
type updateMedicalRecordRequest struct {
	Date                     *time.Time `json:"date"`
	OwnerID                  *uint64    `json:"owner_id"`
	PetID                    *uint64    `json:"pet_id"`
	DoctorID                 *uint64    `json:"doctor_id"`
	AppointmentID            *uint64    `json:"appointment_id"`
	Status                   *string    `json:"status"      binding:"omitempty,oneof=draft finalized"`
	Version                  *int       `json:"version"`                     // 楽観的ロック用
	NextVisitRecommendedDate *string    `json:"next_visit_recommended_date"` // "YYYY-MM-DD" or null
	VisitType                *string    `json:"visit_type"`                  // "first" | "revisit"
}

func (r updateMedicalRecordRequest) toServiceInput(actorID uint64) (UpdateMedicalRecordInput, error) {
	var status *model.MedicalRecordStatus
	if r.Status != nil {
		s, err := validateEnum(*r.Status,
			model.MedicalRecordStatusDraft,
			model.MedicalRecordStatusFinalized,
		)
		if err != nil {
			return UpdateMedicalRecordInput{}, apperrors.WrapInvalidInput("invalid status: " + err.Error())
		}
		status = &s
	}

	var nextVisitDate *time.Time
	var clearNextVisit bool
	if r.NextVisitRecommendedDate != nil {
		if *r.NextVisitRecommendedDate == "" {
			clearNextVisit = true // 空文字 = 明示的クリア → DB を NULL にする
		} else {
			parsed, err := time.ParseInLocation(time.DateOnly, *r.NextVisitRecommendedDate, time.Local)
			if err != nil {
				return UpdateMedicalRecordInput{}, apperrors.WrapInvalidInput("invalid next_visit_recommended_date format (expected YYYY-MM-DD)")
			}
			nextVisitDate = &parsed
		}
	}

	var visitType *model.VisitType
	if r.VisitType != nil {
		vt, err := validateEnum(*r.VisitType, model.VisitTypeFirst, model.VisitTypeRevisit)
		if err != nil {
			return UpdateMedicalRecordInput{}, apperrors.WrapInvalidInput("invalid visit_type: " + err.Error())
		}
		visitType = &vt
	}

	return UpdateMedicalRecordInput{
		Date:                          r.Date,
		OwnerID:                       r.OwnerID,
		PetID:                         r.PetID,
		DoctorID:                      r.DoctorID,
		AppointmentID:                 r.AppointmentID,
		Status:                        status,
		Version:                       r.Version,
		NextVisitRecommendedDate:      nextVisitDate,
		ClearNextVisitRecommendedDate: clearNextVisit,
		VisitType:                     visitType,
		ActorID:                       &actorID,
	}, nil
}
