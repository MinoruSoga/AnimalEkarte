package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D ⑥ Batch A: procedure subpackage を internal/medicalrecord へ roll-up。
// 残存呼び出し側互換のための期限付き facade（削除=BE9-2F）。

type ProcedureRepository = medicalrecord.ProcedureRepository

func NewProcedureRepository(db *gorm.DB) ProcedureRepository {
	return medicalrecord.NewProcedureRepository(db)
}
