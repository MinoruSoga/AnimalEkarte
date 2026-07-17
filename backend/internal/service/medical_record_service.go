package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// UpdateMedicalRecordInput はカルテ更新のサービス入力 DTO
type UpdateMedicalRecordInput struct {
	Date                          *time.Time
	OwnerID                       *uint64
	PetID                         *uint64
	DoctorID                      *uint64
	AppointmentID                 *uint64
	Status                        *model.MedicalRecordStatus
	Version                       *int // 楽観的ロック用: nil の場合はチェックをスキップ
	NextVisitRecommendedDate      *time.Time
	ClearNextVisitRecommendedDate bool // true の場合 DB を NULL にクリアする
	VisitType                     *model.VisitType
	ActorID                       *uint64 // 監査ログ用: 操作スタッフ ID（nil = システム）
}

// CreateMedicalRecordInput はカルテ作成のサービス入力 DTO
type CreateMedicalRecordInput struct {
	RecordNo                 string
	Date                     time.Time
	OwnerID                  *uint64
	PetID                    *uint64
	DoctorID                 *uint64
	AppointmentID            *uint64
	Status                   *model.MedicalRecordStatus
	VisitType                model.VisitType
	NextVisitRecommendedDate *time.Time
	RecommendationReason     *string
	EnteredBy                *uint64
}

// FEAT-381-2 Commit 2: recommendation_reason allowed values (whitelist)
var allowedRecommendationReasons = map[string]struct{}{
	"revisit":    {},
	"checkup":    {},
	"prevention": {},
	"exam":       {},
}

const colRecommendationReason = "recommendation_reason"

// UpdateRecommendationReasonInput は受診推奨理由更新の入力DTO（FEAT-381-2 Commit 2）。
// Reason は revisit / checkup / prevention / exam のいずれか、または "" (未設定 → NULL)。
type UpdateRecommendationReasonInput struct {
	Reason string
}

type MedicalRecordService interface {
	// List は指定した複数医院 (#86 拠点横断) のカルテ一覧を返す。clinicIDs はハンドラ層で所属検証済みであること。
	List(ctx context.Context, clinicIDs []uint64, filters repository.MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	// GetByIDForClinics は複数医院スコープでカルテを1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error)
	CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error)
	Create(ctx context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error)
	Update(ctx context.Context, clinicID, id uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// CreateSubRecords はカルテ作成と同時に inquiry / clinical_plan を best-effort で作成する。
	// 失敗しても呼び出し元のカルテ作成は完了済みのためエラーは握りつぶす（slog.Warn のみ）。
	CreateSubRecords(ctx context.Context, clinicID, recordID uint64, input CreateSubRecordsInput)
	// AutoCreateFromReservation は予約ステータスが確定/受付済みに変わったときカルテを best-effort で自動作成する。
	AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation)
	// DeleteDraftFromReservation は予約キャンセル時に紐づく draft カルテを best-effort で論理削除する (#83 Q10)。
	DeleteDraftFromReservation(ctx context.Context, clinicID, reservationID uint64)
	// UpdateRecommendationReason は受診推奨理由を更新する（FEAT-381-2）。
	// "" は NULL として保存。4値以外は apperrors.WrapInvalidInput を返す。
	UpdateRecommendationReason(ctx context.Context, clinicID, id uint64, input UpdateRecommendationReasonInput) (*model.MedicalRecord, error)
}

// CreateSubRecordsInput はカルテ作成時の inquiry / clinical_plan サブレコード作成 DTO
type CreateSubRecordsInput struct {
	ChiefComplaintTypeID *uint64
	ChiefComplaint       *string
	Notes                *string
	Plan                 *string // → ClinicalPlan.TreatmentPolicy
	Assessment           *string // → ClinicalPlan.DiagnosisDetails
	Diagnosis1CategoryID *uint64
	Diagnosis1NameID     *uint64
	Diagnosis2TypeID     *uint64
	Diagnosis2NameID     *uint64
}

type medicalRecordService struct {
	repo                   repository.MedicalRecordRepository
	ownerRepo              repository.OwnerRepository
	petRepo                repository.PetRepository
	inquiryRepo            repository.InquiryRepository
	clinicalPlanRepo       repository.ClinicalPlanRepository
	chiefComplaintTypeRepo repository.ChiefComplaintTypeRepository
	diagTypeRepo           repository.DiagnosisTypeRepository
	diagNameRepo           repository.DiagnosisNameRepository
	lineCustomerRepo       repository.LineCustomerRepository
	reservationRepo        repository.ReservationRepository
	lstepDeliveryTrigger   LstepDeliveryTriggerService
	auditService           AuditService
	tx                     repository.Transactor
	tagSyncSvc             LstepTagSyncService
}

func NewMedicalRecordService(
	repo repository.MedicalRecordRepository,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
	inquiryRepo repository.InquiryRepository,
	clinicalPlanRepo repository.ClinicalPlanRepository,
	chiefComplaintTypeRepo repository.ChiefComplaintTypeRepository,
	diagTypeRepo repository.DiagnosisTypeRepository,
	diagNameRepo repository.DiagnosisNameRepository,
	lineCustomerRepo repository.LineCustomerRepository,
	reservationRepo repository.ReservationRepository,
	lstepDeliveryTrigger LstepDeliveryTriggerService,
	auditService AuditService,
	tx repository.Transactor,
	tagSyncSvc ...LstepTagSyncService,
) MedicalRecordService {
	var syncSvc LstepTagSyncService
	if len(tagSyncSvc) > 0 {
		syncSvc = tagSyncSvc[0]
	}
	return &medicalRecordService{
		repo:                   repo,
		ownerRepo:              ownerRepo,
		petRepo:                petRepo,
		inquiryRepo:            inquiryRepo,
		clinicalPlanRepo:       clinicalPlanRepo,
		chiefComplaintTypeRepo: chiefComplaintTypeRepo,
		diagTypeRepo:           diagTypeRepo,
		diagNameRepo:           diagNameRepo,
		lineCustomerRepo:       lineCustomerRepo,
		reservationRepo:        reservationRepo,
		lstepDeliveryTrigger:   lstepDeliveryTrigger,
		auditService:           auditService,
		tx:                     tx,
		tagSyncSvc:             syncSvc,
	}
}

// withTx runs fn inside Transactor when configured; nil tx (unit tests) runs fn directly.
func (s *medicalRecordService) withTx(ctx context.Context, fn func(context.Context) error) error {
	if s.tx == nil {
		return fn(ctx)
	}
	return s.tx.WithTx(ctx, fn)
}
