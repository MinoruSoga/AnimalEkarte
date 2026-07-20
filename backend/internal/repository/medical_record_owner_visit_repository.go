package repository

import "github.com/animal-ekarte/backend/internal/medicalrecord"

// BE9-2D ⑦ Batch A: 実装（medicalRecordRepository へのメソッド拡張群+DTO 2型）は
// internal/medicalrecord/medical_record_owner_visit_repository.go へ縦移動済み。
// メソッド群は MedicalRecordRepository facade alias 経由で不変。残存呼び出し側互換の DTO alias のみ残す。

type OwnerVisitSummary = medicalrecord.OwnerVisitSummary

type DormantOwnerEntry = medicalrecord.DormantOwnerEntry
