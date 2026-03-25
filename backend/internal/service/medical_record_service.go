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

type MedicalRecordService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	GetByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	Update(ctx context.Context, record *model.MedicalRecord) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type medicalRecordService struct {
	repo      repository.MedicalRecordRepository
	ownerRepo repository.OwnerRepository
	petRepo   repository.PetRepository
}

func NewMedicalRecordService(
	repo repository.MedicalRecordRepository,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
) MedicalRecordService {
	return &medicalRecordService{
		repo:      repo,
		ownerRepo: ownerRepo,
		petRepo:   petRepo,
	}
}

func (s *medicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (s *medicalRecordService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *medicalRecordService) GetByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error) {
	return s.repo.FindByRecordNo(ctx, clinicID, recordNo)
}

func (s *medicalRecordService) Create(ctx context.Context, record *model.MedicalRecord) error {
	// RecordNo が未設定の場合は service 層で自動生成する（handler 層に生成ロジックを置かない）
	if record.RecordNo == "" {
		record.RecordNo = generateRecordNo(record.Date, record.ClinicID)
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return err
	}
	slog.InfoContext(ctx, "medical record created",
		slog.Uint64("record_id", record.ID),
		slog.Uint64("clinic_id", record.ClinicID))
	return nil
}

func (s *medicalRecordService) Update(ctx context.Context, record *model.MedicalRecord) error {
	// owner_id 変更時: クリニック所属確認
	if record.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, record.ClinicID, *record.OwnerID); err != nil {
			return apperrors.WrapInvalidInput("owner not found in this clinic")
		}
	}

	// pet_id 変更時: クリニック所属確認
	if record.PetID != nil {
		if _, err := s.petRepo.FindByID(ctx, record.ClinicID, *record.PetID); err != nil {
			return apperrors.WrapInvalidInput("pet not found in this clinic")
		}
	}

	if err := s.repo.Update(ctx, record); err != nil {
		return err
	}
	slog.InfoContext(ctx, "medical record updated",
		slog.Uint64("record_id", record.ID),
		slog.Uint64("clinic_id", record.ClinicID))
	return nil
}

func (s *medicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
