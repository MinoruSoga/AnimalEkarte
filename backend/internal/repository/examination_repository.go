package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D ⑦ Batch A: 実装は internal/medicalrecord/examination_repository.go へ縦移動済み。
// 残存呼び出し側互換のための期限付き facade（削除=BE9-2F）。

type ExaminationRepository = medicalrecord.ExaminationRepository

func NewExaminationRepository(db *gorm.DB) ExaminationRepository {
	return medicalrecord.NewExaminationRepository(db)
}
