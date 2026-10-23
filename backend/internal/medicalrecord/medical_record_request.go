package medicalrecord

import (
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type listMedicalRecordQuery struct {
	PetID           string
	OwnerID         string
	StartDate       string
	EndDate         string
	Search          string
	OwnerName       string
	PetName         string
	ChiefComplaint  string
	DoctorID        string
	AnimalSpeciesID string
	MedicineID      string
	ProcedureID     string
	ConsultationID  string
	InventoryID     string
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
		OwnerName:       values.Get("owner_name"),
		PetName:         values.Get("pet_name"),
		ChiefComplaint:  values.Get("chief_complaint"),
		DoctorID:        values.Get("doctor_id"),
		AnimalSpeciesID: values.Get("animal_species_id"),
		MedicineID:      values.Get("medicine_id"),
		ProcedureID:     values.Get("procedure_id"),
		ConsultationID:  values.Get("consultation_id"),
		InventoryID:     values.Get("inventory_id"),
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
	medicineID, err := parseOptionalUintQueryFilter(q.MedicineID, "medicine_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	procedureID, err := parseOptionalUintQueryFilter(q.ProcedureID, "procedure_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	consultationID, err := parseOptionalUintQueryFilter(q.ConsultationID, "consultation_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	inventoryID, err := parseOptionalUintQueryFilter(q.InventoryID, "inventory_id")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	ownerName, err := validateColumnTextQueryFilter(q.OwnerName, "owner_name")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	petName, err := validateColumnTextQueryFilter(q.PetName, "pet_name")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	chiefComplaint, err := validateColumnTextQueryFilter(q.ChiefComplaint, "chief_complaint")
	if err != nil {
		return listMedicalRecordFilters{}, err
	}
	sortKey, sortOrder := resolveMedicalRecordSort(q.Sort, q.Order)
	return listMedicalRecordFilters{
		PetID:           petID,
		OwnerID:         ownerID,
		StartDate:       startDate,
		EndDate:         endDate,
		DoctorID:        doctorID,
		AnimalSpeciesID: animalSpeciesID,
		MedicineID:      medicineID,
		ProcedureID:     procedureID,
		ConsultationID:  consultationID,
		InventoryID:     inventoryID,
		Search:          q.Search,
		OwnerName:       ownerName,
		PetName:         petName,
		ChiefComplaint:  chiefComplaint,
		Sort:            sortKey,
		Order:           sortOrder,
	}, nil
}

// medicalRecordFinalizeRetiredMessage は旧 finalize/unfinalize（request の status）が未提供であることを示す固定文言。
// 再開の入口として案内する POST /api/v1/reservations/{id}/end-consultation はこの PR では実装されない（B 契約 §2）。
const medicalRecordFinalizeRetiredMessage = "旧 finalize/unfinalize は未提供です。再開は予約の終了操作（POST /api/v1/reservations/{id}/end-consultation）で行ってください。"

// medicalRecordStatusFieldRetiredMessage は更新 request に status が含まれていた場合の固定文言（SPEC-02）。
const medicalRecordStatusFieldRetiredMessage = "カルテの status は廃止されました。更新リクエストに status を含めないでください。診療終了は予約の終了操作（POST /api/v1/reservations/{id}/end-consultation）で記録します。"

// rejectRetiredMedicalRecordStatus は作成 request の status を検査する（更新は SPEC-02 で常に固定エラー）。
// finalized は固定エラー（422 medical_record_status_retired）、draft と未指定は無視（FE 移行期の許容）、
// それ以外の値は従来どおり 400。
func rejectRetiredMedicalRecordStatus(status string) error {
	switch status {
	case "", "draft":
		return nil
	case "finalized":
		return apperrors.WrapMedicalRecordStatusRetired(medicalRecordFinalizeRetiredMessage)
	default:
		return apperrors.WrapInvalidInput("ステータスの値が不正です")
	}
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

	// ClinicalPlan関連フィールド（原子的作成用）
	ChiefComplaint       *string `json:"chief_complaint"`
	ChiefComplaintTypeID *uint64 `json:"chief_complaint_type_id"`
	Plan                 *string `json:"plan"`       // ClinicalPlan.TreatmentPolicyへ
	Assessment           *string `json:"assessment"` // ClinicalPlan.DiagnosisDetailsへ
	Notes                *string `json:"notes"`
	// Diagnoses は診断行（EMR-495 A）。clinical_plan_diagnoses の行へ保存する。未送信 = 診断を変更しない。
	Diagnoses []clinicalPlanDiagnosisRequest `json:"diagnoses"`

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

	visitType := model.VisitTypeRevisit
	if r.VisitType == string(model.VisitTypeFirst) {
		visitType = model.VisitTypeFirst
	}

	// EMR-495: 旧 finalize（status=finalized）は未提供。draft / 未指定は無視する（FE 移行期の許容）。
	if err := rejectRetiredMedicalRecordStatus(r.Status); err != nil {
		return CreateMedicalRecordInput{}, err
	}
	var recommendationReason *string
	if r.RecommendationReason != nil && *r.RecommendationReason != "" {
		recommendationReason = r.RecommendationReason
	}

	return CreateMedicalRecordInput{
		RecordNo:             r.RecordNo,
		Date:                 recordDate,
		OwnerID:              ownerID,
		PetID:                petID,
		DoctorID:             doctorID,
		AppointmentID:        appointmentID,
		VisitType:            visitType,
		RecommendationReason: recommendationReason,
		EnteredBy:            &staffID,
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

func (r *createMedicalRecordRequest) toSubRecordsInput() (CreateSubRecordsInput, error) {
	diagnoses, err := toDiagnosisInputs(r.Diagnoses)
	if err != nil {
		return CreateSubRecordsInput{}, err
	}
	return CreateSubRecordsInput{
		ChiefComplaintTypeID: r.ChiefComplaintTypeID,
		ChiefComplaint:       r.ChiefComplaint,
		Notes:                r.Notes,
		Plan:                 r.Plan,
		Assessment:           r.Assessment,
		Diagnoses:            diagnoses,
	}, nil
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
	return UpdateRecommendationReasonInput(r)
}

// updateMedicalRecordRequest はカルテ更新のバインド struct
type updateMedicalRecordRequest struct {
	Date          *time.Time `json:"date"`
	OwnerID       *uint64    `json:"owner_id"`
	PetID         *uint64    `json:"pet_id"`
	DoctorID      *uint64    `json:"doctor_id"`
	AppointmentID *uint64    `json:"appointment_id"`
	// Status: 更新で変更できる status はもう無いため、キーが存在すれば値に関わらず固定 422（SPEC-02）。
	// 値の enum 検証はしない（バインド時の 400 ではなく固定エラーに統一する）。
	Status    *string `json:"status"`
	Version   *int    `json:"version"`    // 楽観的ロック用
	VisitType *string `json:"visit_type"` // "first" | "revisit"
}

func (r updateMedicalRecordRequest) toServiceInput(actorID uint64) (UpdateMedicalRecordInput, error) {
	// SPEC-02: 更新では変更できる status が存在しない。draft / finalized / その他、status キーが
	// 送られた時点で固定エラー（422 medical_record_status_retired）を返し、何も書かない。
	if r.Status != nil {
		return UpdateMedicalRecordInput{}, apperrors.WrapMedicalRecordStatusRetired(medicalRecordStatusFieldRetiredMessage)
	}

	var visitType *model.VisitType
	if r.VisitType != nil {
		vt, err := httpapi.ValidateEnum(*r.VisitType, model.VisitTypeFirst, model.VisitTypeRevisit)
		if err != nil {
			return UpdateMedicalRecordInput{}, apperrors.WrapInvalidInput("来院区分の値が不正です")
		}
		visitType = &vt
	}

	return UpdateMedicalRecordInput{
		Date:          r.Date,
		OwnerID:       r.OwnerID,
		PetID:         r.PetID,
		DoctorID:      r.DoctorID,
		AppointmentID: r.AppointmentID,
		Version:       r.Version,
		VisitType:     visitType,
		ActorID:       &actorID,
	}, nil
}
