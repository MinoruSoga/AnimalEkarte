package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D ⑦ Batch A: 実装は internal/medicalrecord/medical_record_repository.go へ縦移動済み。
// 残存呼び出し側互換のための期限付き facade（削除=BE9-2F）。

type MedicalRecordRepository = medicalrecord.MedicalRecordRepository

// MedicalRecordListFilters は一覧検索フィルタ DTO（実体は medicalrecord）。
type MedicalRecordListFilters = medicalrecord.MedicalRecordListFilters

func NewMedicalRecordRepository(db *gorm.DB) MedicalRecordRepository {
	return medicalrecord.NewMedicalRecordRepository(db)
}
