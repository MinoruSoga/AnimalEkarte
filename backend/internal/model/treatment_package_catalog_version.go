package model

import "time"

// TreatmentPackageCatalogVersion は医院別 package catalog の永続 CAS 版行
// （EMR-518 §4 / EMR-521）。clinic_id が PK で 1 医院 1 行。membership/order 変更
// （create/delete/reorder）だけが revision を 1 回増分し、text/components/active 編集は
// package.version のみを増分する。row 欠落は read で INSERT せず setup/invariant error。
type TreatmentPackageCatalogVersion struct {
	ClinicID  uint64    `gorm:"column:clinic_id;type:bigint;primaryKey;not null" json:"clinic_id"`
	Revision  int64     `gorm:"column:revision;type:bigint;not null" json:"revision"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null" json:"updated_at"`
}

func (TreatmentPackageCatalogVersion) TableName() string { return "treatment_package_catalog_versions" }
