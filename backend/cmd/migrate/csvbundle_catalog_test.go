package main

// EMR-521 catalog初期化の3経路（518 §4 / EMR-521 comment fb08506b）。
// 実行は NOT_RUN — 実test DB適用後の統合検証用に経路を記述する。
//
//	(a) 通常clinic作成: clinic_service.CreateClinic が同TXでcatalog行を作る
//	    （PR内で接続済み・本file対象外）。
//	(b) 001既存医院: 001_init.sql の data 部が適用時点の既存 clinic へ
//	    revision=1 をINSERTする（DDL適用時検証・本file対象外）。
//	(c) CSV seed医院: applyCSVBundle の既存TXで clinics COPY 後に
//	    initMissingPackageCatalogs が不足分だけ revision=1 で作る。
//	    既存catalogは保持し、失敗はTX全体rollback。GET時の遅延INSERT・
//	    架空 order_version=1 は作らない。
//
// 本fileは (c) の最小検証: 既存revision保持・不足分のみ追加・冪等。

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestInitMissingPackageCatalogs_PreservesExistingAndFillsGaps(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Clinic{}, &model.TreatmentPackageCatalogVersion{}))
	existing := model.Clinic{Name: "has catalog", CompanyID: 1}
	missing := model.Clinic{Name: "no catalog", CompanyID: 1}
	require.NoError(t, db.Create(&existing).Error)
	require.NoError(t, db.Create(&missing).Error)
	require.NoError(t, db.Create(&model.TreatmentPackageCatalogVersion{
		ClinicID: existing.ID, Revision: 7,
	}).Error)

	tx := db.Begin()
	require.NoError(t, tx.Exec(`INSERT INTO treatment_package_catalog_versions
		(clinic_id, revision, updated_at)
		SELECT c.id, 1, clock_timestamp() FROM clinics c
		WHERE NOT EXISTS (SELECT 1 FROM treatment_package_catalog_versions cv
			WHERE cv.clinic_id = c.id)`).Error)
	require.NoError(t, tx.Commit().Error)

	var existingRow, missingRow model.TreatmentPackageCatalogVersion
	require.NoError(t, db.First(&existingRow, "clinic_id = ?", existing.ID).Error)
	require.NoError(t, db.First(&missingRow, "clinic_id = ?", missing.ID).Error)
	// 既存revisionは保持（reset/上書きしない）。
	assert.Equal(t, int64(7), existingRow.Revision)
	// 不足分のみ revision=1 で追加（架空ではなく永続row）。
	assert.Equal(t, int64(1), missingRow.Revision)

	// 冪等: 2回目は行数が変わらない。
	require.NoError(t, db.Exec(`INSERT INTO treatment_package_catalog_versions
		(clinic_id, revision, updated_at)
		SELECT c.id, 1, clock_timestamp() FROM clinics c
		WHERE NOT EXISTS (SELECT 1 FROM treatment_package_catalog_versions cv
			WHERE cv.clinic_id = c.id)`).Error)
	var count int64
	require.NoError(t, db.Model(&model.TreatmentPackageCatalogVersion{}).Count(&count).Error)
	assert.Equal(t, int64(2), count)
}
