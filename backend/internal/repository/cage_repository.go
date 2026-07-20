package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D ⑥ Batch A: cage subpackage を internal/medicalrecord へ roll-up。
// 残存呼び出し側互換のための期限付き facade（削除=BE9-2F）。

type CageRepository = medicalrecord.CageRepository

func NewCageRepository(db *gorm.DB) CageRepository {
	return medicalrecord.NewCageRepository(db)
}
