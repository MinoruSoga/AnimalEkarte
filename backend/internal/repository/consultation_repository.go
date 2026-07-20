package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D ⑥ Batch A: consultation subpackage を internal/medicalrecord へ roll-up。
// 残存呼び出し側互換のための期限付き facade（削除=BE9-2F）。

type ConsultationRepository = medicalrecord.ConsultationRepository

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return medicalrecord.NewConsultationRepository(db)
}
