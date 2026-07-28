package audit

// audit_real_ddl_test.go — X-3 (audit-ip-inet-model-drift): audit_logs を AutoMigrate ではなく
// 001_init.sql の実 DDL から再作成し、model.AuditLog と実スキーマの drift を検出する。
//
// 背景: audit_logs.ip_address は実DDLで `inet NULL` だが、AutoMigrate ベースの
// setupAuditTestDB（audit_repository_test.go、修正前）は model.AuditLog.IPAddress の Go 型
// （旧: string）から `text` カラムを生成するため、本番で発生する `''::inet` の 22P02 エラーを
// テストが検出できなかった（internal/model/schema_drift_test.go の knownSchemaDriftAllowlist["AuditLog.ip_address"]
// 参照）。修正として IPAddress を *string 化し、空文字列は nil に正規化した
// （同 package の service.go にある normalizeIPAddress）。この結果、Go の型システムが
// 「INSERT に空文字列を渡す」という不正状態自体を構築不能にしたため、下記
// TestAuditLogRealDDL_EmptyStringIPAddressRejectedByDB は生SQLで DB 制約そのものを検証する
// 恒久ガードに置き換えている（修正前は model.AuditLog{IPAddress: ""} 経由で 22P02 を実証していた
// — 実際の RED 実行結果は PR 記録を参照）。
//
// 同様に audit_logs.clinic_id は実DDLで `bigint NOT NULL` だが、AutoMigrate は
// model.AuditLog.ClinicID(*uint64) から NULL 許容カラムを生成するため、
// ClinicID=nil の INSERT が AutoMigrate スキーマ上では成功してしまう
// （knownNullabilityDriftAllowlist["AuditLog.clinic_id"] 参照）。ClinicID は意図的に *uint64 の
// ままなので（service.validateAuditLog を DB 制約が最終防衛する設計、model/audit_log.go 参照）、
// repo.Create(ClinicID: nil) は今なお Go レベルで構築可能 — TestAuditLogRealDDL_NilClinicIDFails
// が実DDLレベルでの拒否を検証する。
//
// 共有プール上で DROP+CREATE すると他テストの prepared statement キャッシュを破壊するため、
// isolated test DB を使う。

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// auditLogsFKPattern は clinics/staffs への FK 句を除去する（テスト DB に clinics テーブルは
// 存在せず、staffs は AutoMigrate 済みの他テストに依存するため、checkup_field_cascade_test.go の
// clinicsFKPattern と同じ理由で本テストを自己完結させる）。
var auditLogsFKPattern = regexp.MustCompile(`REFERENCES (clinics|staffs)\(id\)\s+ON DELETE RESTRICT`)

// setupAuditRealDDLTestDB は audit_logs を DROP し、001_init.sql の実 CREATE TABLE 文で
// 再作成する（AutoMigrate 由来の drift を排除する）。
func setupAuditRealDDLTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupIsolatedTestDB(t)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS audit_logs CASCADE").Error)

	sql := readInitMigrationSQL(t)
	ddl := extractCreateTableDDL(t, sql, "audit_logs")
	ddl = auditLogsFKPattern.ReplaceAllString(ddl, "")

	require.NoError(t, db.Exec(ddl).Error, "audit_logs を実DDLで再作成できること")
	return db
}

// TestAuditLogRealDDL_EmptyStringIPAddressRejectedByDB は audit_logs.ip_address(inet NULL) に
// 空文字列を直接 INSERT すると PostgreSQL が 22P02 (invalid_text_representation) を返すことを
// 恒久的に保証する DDL レベルのガードである。
//
// 修正前はこの欠陥を model.AuditLog{IPAddress: ""}（当時 Go 型 string）を repo.Create 経由で
// INSERT することで再現できた（多くの監査書込パス — LogLstepOperationWithMetadata /
// LogMedicalRecordChange / LogVitalChange / LogAddendumCreate / examination_service.LogEntryTx 等 —
// が IPAddress を一度もセットせずゼロ値の "" がそのまま INSERT されていた）。
// 修正後は IPAddress が *string になり、model.AuditLog{IPAddress: ""} は
// 型不一致でコンパイルできない（Go の型システムが不正状態を構築不能にした）。
// そのため本テストは生 SQL で DB 制約そのものを検証し、将来 ip_address の DDL が変わったり、
// 何らかの経路が Go モデルを経由せず空文字列を書き込もうとした場合の回帰を検出する。
func TestAuditLogRealDDL_EmptyStringIPAddressRejectedByDB(t *testing.T) {
	db := setupAuditRealDDLTestDB(t)

	err := db.Exec(
		`INSERT INTO audit_logs (clinic_id, actor_type, action, resource, ip_address) VALUES (?, ?, ?, ?, ?)`,
		1, model.AuditActorTypeSystem, model.AuditActionLstepTagSyncBulk, "lstep", "",
	).Error
	require.Error(t, err, "ip_address(inet) への空文字列 INSERT は 22P02 になるはず（X-3 drift の実証）")
	t.Logf("X-3 evidence — raw SQL insert with empty ip_address failed as expected: %v", err)
}

// TestAuditLogRealDDL_NilIPAddressSucceeds は X-3 修正の GREEN 実証: IPAddress=nil（*string）の
// 監査ログは実DDL（ip_address inet NULL）へ問題なく永続化される。
// 多くの監査書込パスが IPAddress を一度もセットしない（ゼロ値 = nil）ことに対応する正常系。
func TestAuditLogRealDDL_NilIPAddressSucceeds(t *testing.T) {
	db := setupAuditRealDDLTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	clinicID := uint64(1)
	log := &model.AuditLog{
		ClinicID:  &clinicID,
		ActorType: model.AuditActorTypeSystem,
		Action:    model.AuditActionLstepTagSyncBulk,
		Resource:  "lstep",
		IPAddress: nil, // ゼロ値。修正前は "" だった箇所（X-3 GREEN）
	}
	require.NoError(t, repo.Create(ctx, log), "IPAddress=nil は実DDL(ip_address inet NULL)へ問題なく永続化されるはず")

	var got model.AuditLog
	require.NoError(t, db.WithContext(ctx).First(&got, log.ID).Error)
	require.Nil(t, got.IPAddress, "未設定の ip_address は DB から読み戻しても nil であるはず")
}

// TestAuditLogRealDDL_NilClinicIDFails は audit_logs.clinic_id が実DDLで
// `bigint NOT NULL REFERENCES clinics(id)` であるにも関わらず、repo.Create を直接呼ぶと
// （= サービス層 validateAuditLog を経由しない経路）ClinicID=nil の INSERT が
// PostgreSQL の NOT NULL 制約違反 (23502) で失敗することを示す。
// サービス層の validateAuditLog（audit_service.go:94-96）は既に nil ClinicID を拒否しているため
// 本番での到達性は低いが、audit_repository_test.go は AutoMigrate 由来の NULL 許容スキーマ上で
// 「ClinicID=nil でも作成できる」という DDL 非互換の挙動をテストしていた（G12-1 で発見）。
// 本テストは DB 制約そのものを実DDLで検証し、その前提の誤りを是正する。
func TestAuditLogRealDDL_NilClinicIDFails(t *testing.T) {
	db := setupAuditRealDDLTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	ip := "127.0.0.1"
	log := &model.AuditLog{
		ClinicID:  nil,
		ActorType: model.AuditActorTypeSystem,
		Action:    model.AuditActionLstepTagSyncBulk,
		Resource:  "lstep",
		// IPAddress を有効値にして ip_address(inet) 由来のエラーと混同しないようにする
		// （clinic_id NOT NULL 違反のみを単離して検証する）。
		IPAddress: &ip,
	}
	err := repo.Create(ctx, log)
	require.Error(t, err, "ClinicID=nil は実DDL(clinic_id bigint NOT NULL)への INSERT で 23502 になるはず（DDL制約の実証）")
	t.Logf("X-3 evidence — repo.Create with nil ClinicID failed as expected: %v", err)
}
