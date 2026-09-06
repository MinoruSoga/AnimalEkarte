package medicalrecord

import (
	"crypto/rand"
	"fmt"
	"maps"
	"math/big"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
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
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.ClearNextVisitRecommendedDate {
		fields["next_visit_recommended_date"] = nil
	} else if input.NextVisitRecommendedDate != nil {
		fields["next_visit_recommended_date"] = *input.NextVisitRecommendedDate
	}
	if input.VisitType != nil {
		fields["visit_type"] = *input.VisitType
	}
	if input.persistVersion != nil {
		fields["version"] = *input.persistVersion
	}
	if len(input.persistFields) > 0 {
		maps.Copy(fields, input.persistFields)
	}
	return fields
}

func buildMedicalRecordForCreate(clinicID uint64, input *CreateMedicalRecordInput) *model.MedicalRecord {
	visitType := input.VisitType
	record := &model.MedicalRecord{
		ClinicID:                 clinicID,
		RecordNo:                 input.RecordNo,
		Date:                     input.Date,
		OwnerID:                  input.OwnerID,
		PetID:                    input.PetID,
		DoctorID:                 input.DoctorID,
		AppointmentID:            input.AppointmentID,
		VisitType:                &visitType,
		NextVisitRecommendedDate: input.NextVisitRecommendedDate,
		RecommendationReason:     input.RecommendationReason,
		EnteredBy:                input.EnteredBy,
	}
	if input.Status != nil {
		record.Status = *input.Status
	}
	return record
}
