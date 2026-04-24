package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// generateRecordNo は "MR-YYYYMMDD-{clinicID}-{random6}" 形式のカルテ番号を生成する。
// 乱数生成に crypto/rand を使用し、推測困難にする。
func generateRecordNo(date time.Time, clinicID uint64) string {
	datePart := date.Format("20060102")
	randomPart := generateCryptoRandomString(6)
	return fmt.Sprintf("MR-%s-%d-%s", datePart, clinicID, randomPart)
}

// generateCryptoRandomString は crypto/rand を使って指定長の英数字文字列を生成する。
func generateCryptoRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// crypto/rand の失敗は極めて稀だが、安全側に倒して先頭文字を使う
			b[i] = charset[0]
			continue
		}
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

// UpdateMedicalRecordInput はカルテ更新のサービス入力 DTO
type UpdateMedicalRecordInput struct {
	Date          *time.Time
	OwnerID       *uint64
	PetID         *uint64
	DoctorID      *uint64
	AppointmentID *uint64
	Status        *model.MedicalRecordStatus
	Version       *int // 楽観的ロック用: nil の場合はチェックをスキップ
}

func buildMedicalRecordUpdate(input UpdateMedicalRecordInput) map[string]any {
	fields := make(map[string]any)
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.OwnerID != nil {
		fields["owner_id"] = *input.OwnerID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.AppointmentID != nil {
		fields["appointment_id"] = *input.AppointmentID
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	return fields
}

type MedicalRecordService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	Update(ctx context.Context, clinicID, id uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// CreateSubRecords はカルテ作成と同時に inquiry / clinical_plan を best-effort で作成する。
	// 失敗しても呼び出し元のカルテ作成は完了済みのためエラーは握りつぶす（slog.Warn のみ）。
	CreateSubRecords(ctx context.Context, clinicID, recordID uint64, input CreateSubRecordsInput)
	// AutoCreateFromReservation は予約ステータスが「受付済み」に変わったときカルテを best-effort で自動作成する。
	AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation)
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
	Diagnosis2CategoryID *uint64
	Diagnosis2NameID     *uint64
}

type medicalRecordService struct {
	repo             repository.MedicalRecordRepository
	ownerRepo        repository.OwnerRepository
	petRepo          repository.PetRepository
	inquiryRepo      repository.InquiryRepository
	clinicalPlanRepo repository.ClinicalPlanRepository
	lineCustomerRepo repository.LineCustomerRepository
}

func NewMedicalRecordService(
	repo repository.MedicalRecordRepository,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
	inquiryRepo repository.InquiryRepository,
	clinicalPlanRepo repository.ClinicalPlanRepository,
	lineCustomerRepo repository.LineCustomerRepository,
) MedicalRecordService {
	return &medicalRecordService{
		repo:             repo,
		ownerRepo:        ownerRepo,
		petRepo:          petRepo,
		inquiryRepo:      inquiryRepo,
		clinicalPlanRepo: clinicalPlanRepo,
		lineCustomerRepo: lineCustomerRepo,
	}
}

func (s *medicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list medical records", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list medical records")
	}
	return items, total, nil
}

func (s *medicalRecordService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get medical record")
	}
	return result, nil
}

func (s *medicalRecordService) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	count, err := s.repo.CountByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count medical records by pet", "error", err)
		return 0, apperrors.Wrap(err, "failed to count medical records by pet")
	}
	return count, nil
}

func (s *medicalRecordService) Create(ctx context.Context, record *model.MedicalRecord) error {
	// RecordNo が未設定の場合は service 層で自動生成する（handler 層に生成ロジックを置かない）
	if record.RecordNo == "" {
		record.RecordNo = generateRecordNo(record.Date, record.ClinicID)
	}
	if err := s.repo.Create(ctx, record); err != nil {
		slog.ErrorContext(ctx, "failed to create medical record", "error", err, "clinic_id", record.ClinicID)
		return apperrors.Wrap(err, "failed to create medical record")
	}
	slog.InfoContext(ctx, "medical record created",
		slog.Uint64("record_id", record.ID),
		slog.Uint64("clinic_id", record.ClinicID))
	return nil
}

func (s *medicalRecordService) Update(ctx context.Context, clinicID, id uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
	// 楽観的ロックチェックのため現在のレコードを取得
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get medical record")
	}

	// version が指定されている場合は一致確認
	if input.Version != nil && existing.Version != *input.Version {
		return nil, apperrors.WrapConflict("他のユーザーがこのカルテを変更しました。再読み込みしてください")
	}

	// owner_id 変更時: クリニック所属確認
	if input.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, clinicID, *input.OwnerID); err != nil {
			return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
		}
	}

	// pet_id 変更時: クリニック所属確認
	if input.PetID != nil {
		if _, err := s.petRepo.FindByID(ctx, clinicID, *input.PetID); err != nil {
			return nil, apperrors.WrapInvalidInput("pet not found in this clinic")
		}
	}

	fields := buildMedicalRecordUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	// バージョンをインクリメント
	fields["version"] = existing.Version + 1

	record, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to update medical record")
	}
	slog.InfoContext(ctx, "medical record updated",
		slog.Uint64("record_id", id),
		slog.Uint64("clinic_id", clinicID))
	return record, nil
}

func (s *medicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find medical record")
	}
	estimateCount, err := s.repo.CountEstimatesByMedicalRecordID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check estimate dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check estimate dependencies")
	}
	if estimateCount > 0 {
		return apperrors.WrapConflict("この項目は使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete medical record", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete medical record")
	}
	slog.InfoContext(ctx, "medical record deleted",
		slog.Uint64("record_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

// CreateSubRecords はカルテ作成と同時に inquiry / clinical_plan を best-effort で作成する。
// 失敗しても呼び出し元のカルテ作成は完了済みのためエラーは握りつぶし slog.Warn のみ出力する。
func (s *medicalRecordService) CreateSubRecords(ctx context.Context, clinicID, recordID uint64, input CreateSubRecordsInput) {
	// 1. inquiry: フィールドの有無に関わらず常に upsert（空でも OK）
	inquiry := &model.Inquiry{
		MedicalRecordID: recordID,
	}
	if input.ChiefComplaintTypeID != nil {
		inquiry.ChiefComplaintTypeID = input.ChiefComplaintTypeID
	}
	if input.ChiefComplaint != nil {
		inquiry.ChiefComplaint = *input.ChiefComplaint
	}
	if input.Notes != nil {
		inquiry.Notes = *input.Notes
	}
	if _, err := s.inquiryRepo.SaveByMedicalRecordID(ctx, clinicID, inquiry); err != nil {
		slog.WarnContext(ctx, "createSubRecords: failed to upsert inquiry",
			slog.Uint64("medical_record_id", recordID),
			slog.String("error", err.Error()))
	}

	// 2. clinical_plan: 常に GetOrCreate で空レコードを確保し、フィールドがあれば更新
	plan, err := s.clinicalPlanRepo.FindByMedicalRecordID(ctx, clinicID, recordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			slog.WarnContext(ctx, "createSubRecords: failed to find clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
			return
		}
		plan = &model.ClinicalPlan{MedicalRecordID: recordID}
		if err := s.clinicalPlanRepo.Create(ctx, plan); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to create clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
			return
		}
	}
	if input.Plan != nil || input.Assessment != nil || input.Diagnosis1CategoryID != nil || input.Diagnosis1NameID != nil {
		fields := map[string]any{}
		if input.Plan != nil {
			fields["treatment_policy"] = *input.Plan
		}
		if input.Assessment != nil {
			fields["diagnosis_details"] = *input.Assessment
		}
		if input.Diagnosis1CategoryID != nil {
			fields["diagnosis_type_id"] = *input.Diagnosis1CategoryID
		}
		if input.Diagnosis1NameID != nil {
			fields["diagnosis_name_id"] = *input.Diagnosis1NameID
		}
		if input.Diagnosis2CategoryID != nil {
			fields["diagnosis_2_category_id"] = *input.Diagnosis2CategoryID
		}
		if input.Diagnosis2NameID != nil {
			fields["diagnosis_2_name_id"] = *input.Diagnosis2NameID
		}
		if err := s.clinicalPlanRepo.Update(ctx, clinicID, plan.ID, fields); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to update clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
		}
	}
}

// AutoCreateFromReservation は予約ステータスが「受付済み」に変わったときカルテを best-effort で自動作成する。
// 同日同ペットのカルテが既に存在する場合はスキップする（重複防止）。
// LINE予約で owner_id / pet_id が未設定の場合は line_customer から補完を試みる（BUG-386）。
// 失敗してもメイン処理（予約更新）には影響しない。
func (s *medicalRecordService) AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation) {
	// BUG-386: LINE予約で owner_id / pet_id が未設定の場合、line_customer から補完する
	if (reservation.PetID == nil || reservation.OwnerID == nil) &&
		reservation.LineCustomerID != nil &&
		s.lineCustomerRepo != nil {
		customer, err := s.lineCustomerRepo.FindByID(ctx, clinicID, *reservation.LineCustomerID)
		if err == nil && customer != nil && customer.OwnerID != nil {
			reservation.OwnerID = customer.OwnerID
			// ペットが1頭のみ登録されている場合は自動で紐付ける
			if reservation.PetID == nil && customer.Owner != nil && len(customer.Owner.Pets) == 1 {
				reservation.PetID = &customer.Owner.Pets[0].ID
			}
		}
	}

	if reservation.PetID == nil || reservation.OwnerID == nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: skipped — reservation has no pet_id or owner_id",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}

	// 同日同ペットのカルテ存在チェック（重複防止）
	dateStr := reservation.StartTime.Format("2006-01-02")
	_, total, err := s.repo.FindAll(ctx, clinicID, reservation.PetID, nil, &dateStr, &dateStr, 1, 1)
	if err != nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: failed to check existing records",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return
	}
	if total > 0 {
		slog.InfoContext(ctx, "autoCreateFromReservation: skipped — same-day record already exists",
			slog.Uint64("pet_id", *reservation.PetID),
			slog.String("date", dateStr))
		return
	}

	record := &model.MedicalRecord{
		ClinicID:      clinicID,
		Date:          reservation.StartTime,
		OwnerID:       reservation.OwnerID,
		PetID:         reservation.PetID,
		AppointmentID: &reservation.ID,
		Status:        model.MedicalRecordStatusDraft,
	}
	if reservation.DoctorID != nil {
		record.DoctorID = reservation.DoctorID
	}

	if err := s.Create(ctx, record); err != nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: failed to create medical record",
			slog.Uint64("reservation_id", reservation.ID),
			slog.String("error", err.Error()))
		return
	}

	// サブテーブル（inquiry, clinical_plan）を空レコードで作成（best-effort）
	s.CreateSubRecords(ctx, clinicID, record.ID, CreateSubRecordsInput{})
}
