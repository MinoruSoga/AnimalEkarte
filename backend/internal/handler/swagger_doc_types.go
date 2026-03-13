package handler

// このファイルはSwaggerドキュメント生成専用の型定義。
// PaginatedResponse.Data が any のままだと TypeScript 型が unusable になるため、
// エンティティごとに型付きリストレスポンスを定義する。
// 実際の HTTP レスポンスには PaginatedResponse を使い続ける（型は同一）。

import "github.com/animal-ekarte/backend/internal/model"

// OwnerListResponse は飼主一覧のページネーション付きレスポンス
//
// @name OwnerListResponse
type OwnerListResponse struct {
	Data  []model.Owner `json:"data"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

// PetListResponse はペット一覧のページネーション付きレスポンス
//
// @name PetListResponse
type PetListResponse struct {
	Data  []model.Pet `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// VaccinationListResponse は予防接種一覧のページネーション付きレスポンス
//
// @name VaccinationListResponse
type VaccinationListResponse struct {
	Data  []model.Vaccination `json:"data"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// ExamListResponse は検査一覧のページネーション付きレスポンス
//
// @name ExamListResponse
type ExamListResponse struct {
	Data  []model.Exam `json:"data"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

// InventoryListResponse は在庫一覧のページネーション付きレスポンス
//
// @name InventoryListResponse
type InventoryListResponse struct {
	Data  []model.InventoryItem `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// HospitalizationListResponse は入院一覧のページネーション付きレスポンス
//
// @name HospitalizationListResponse
type HospitalizationListResponse struct {
	Data  []model.Hospitalization `json:"data"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
}

// TrimmingListResponse はトリミング一覧のページネーション付きレスポンス
//
// @name TrimmingListResponse
type TrimmingListResponse struct {
	Data  []model.TrimmingRecord `json:"data"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
}

// MedicalRecordListResponse はカルテ一覧のページネーション付きレスポンス
//
// @name MedicalRecordListResponse
type MedicalRecordListResponse struct {
	Data  []model.MedicalRecord `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// BillingListResponse は会計一覧のページネーション付きレスポンス
//
// @name BillingListResponse
type BillingListResponse struct {
	Data  []model.Billing `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}
