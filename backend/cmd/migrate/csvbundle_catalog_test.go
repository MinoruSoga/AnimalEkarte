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
// 本testは (c) の検証: production の initMissingPackageCatalogs 本体を
// 使い捨て schema（openMigrateTestPool と同 DB）上で exercise する。
// 既存revision保持・不足分のみ追加・冪等を確認する。

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitMissingPackageCatalogs_PreservesExistingAndFillsGaps(t *testing.T) {
	ctx := context.Background()
	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()

	// run毎に一意な schema（並列runが互いのschemaをDROPしないよう固定名は使わない）。
	schema := fmt.Sprintf("pkg_catalog_test_%d", time.Now().UnixNano())
	_, err := pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	require.NoError(t, err)
	defer func() {
		_, _ = pool.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
	}()

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()
	_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schema))
	require.NoError(t, err)

	// clinics / catalog の最小 schema（001 の両表の対象列のみ再現）。
	_, err = conn.Exec(ctx, `CREATE TABLE clinics (
		id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		name text NOT NULL)`)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `CREATE TABLE treatment_package_catalog_versions (
		clinic_id bigint PRIMARY KEY REFERENCES clinics(id) ON DELETE RESTRICT,
		revision bigint NOT NULL CHECK (revision > 0),
		updated_at timestamptz NOT NULL)`)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, `INSERT INTO clinics (name) VALUES ('has catalog'), ('no catalog')`)
	require.NoError(t, err)
	_, err = conn.Exec(ctx, `INSERT INTO treatment_package_catalog_versions
		(clinic_id, revision, updated_at) VALUES (1, 7, clock_timestamp())`)
	require.NoError(t, err)

	tx, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx) //nolint:errcheck
	require.NoError(t, initMissingPackageCatalogs(ctx, tx))
	require.NoError(t, tx.Commit(ctx))

	var rev1, rev2 int64
	require.NoError(t, conn.QueryRow(ctx,
		"SELECT revision FROM treatment_package_catalog_versions WHERE clinic_id = 1").Scan(&rev1))
	require.NoError(t, conn.QueryRow(ctx,
		"SELECT revision FROM treatment_package_catalog_versions WHERE clinic_id = 2").Scan(&rev2))
	// 既存revisionは保持（reset/上書きしない）。
	assert.Equal(t, int64(7), rev1)
	// 不足分のみ revision=1 で追加（架空ではなく永続row）。
	assert.Equal(t, int64(1), rev2)

	// 冪等: 2回目は行数が変わらない。
	tx2, err := conn.Begin(ctx)
	require.NoError(t, err)
	defer tx2.Rollback(ctx) //nolint:errcheck
	require.NoError(t, initMissingPackageCatalogs(ctx, tx2))
	require.NoError(t, tx2.Commit(ctx))
	var count int
	require.NoError(t, conn.QueryRow(ctx,
		"SELECT count(*) FROM treatment_package_catalog_versions").Scan(&count))
	assert.Equal(t, 2, count)
}
