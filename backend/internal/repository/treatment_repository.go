package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// BE9-2D sub-batch④b: 実装は internal/medicalrecord/treatment_repository.go へ縦移動済み。
// billing_item_service 等の残存呼び出し側（fan-in>0）互換のための期限付き facade（削除=BE9-2F）。

// TreatmentSortUpdate は並び順一括更新に使う軽量DTO
type TreatmentSortUpdate = medicalrecord.TreatmentSortUpdate

// TreatmentRepository は治療項目の永続化インターフェース
type TreatmentRepository = medicalrecord.TreatmentRepository

// NewTreatmentRepository はTreatmentRepositoryを初期化して返す
func NewTreatmentRepository(db *gorm.DB) TreatmentRepository {
	return medicalrecord.NewTreatmentRepository(db)
}
