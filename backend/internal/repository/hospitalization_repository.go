package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D ⑤ Batch A: 実装は internal/medicalrecord/hospitalization_repository.go へ縦移動済み。
// 残存呼び出し側互換のための期限付き facade（削除=BE9-2F）。

type HospitalizationRepository = medicalrecord.HospitalizationRepository

func NewHospitalizationRepository(db *gorm.DB) HospitalizationRepository {
	return medicalrecord.NewHospitalizationRepository(db)
}
