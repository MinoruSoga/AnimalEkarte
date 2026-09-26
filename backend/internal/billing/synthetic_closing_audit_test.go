package billing

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// testdbSetupSyntheticAudit は S09 fixture 一式 + audit_logs を AutoMigrate する
// 共有 ekarte_db_test 環境。ekarte_db_test は実 FK（audit_logs.actor_id/clinic_id
// の RESTRICT 制約を含む）を持つため、匿名化と rollback の契約はここでも実制約で
// 検証できる。disposable DB 側（MIGRATE_SQL_INTEGRATION）が append-only trigger の
// 実挙動を補完する。
func testdbSetupSyntheticAudit(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdbSetupSyntheticClosing(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
	resetSyntheticAuditSharedState(t, db)
	return db
}

// resetSyntheticAuditSharedState は共有 ekarte_db_test に残った s09 fixture を
// 製品経路（cleanup token を再計算した既定 policy teardown）で除去する。
// rollback 検証や途中失敗の残留を次回 run が自律回復できるようにするためのもので、
// sentinel 自身は RESTRICT 参照元になるため削除しない（削除すると匿名化済みの
// 監査行が dangling する — 製品動作でも sentinel は teardown 対象外）。
func resetSyntheticAuditSharedState(t *testing.T, db *gorm.DB) {
	t.Helper()
	ctx := context.Background()
	var clinicIDs []uint64
	require.NoError(t, db.Raw(
		"SELECT id FROM clinics WHERE name LIKE ?", syntheticClosingClinicPrefix+"%",
	).Scan(&clinicIDs).Error)
	for _, id := range clinicIDs {
		require.NoError(t, DeleteSyntheticClosingFixture(
			ctx, db, "development", "db", id, SyntheticClosingCleanupToken(id),
		))
	}
	// clinic が残っていない孤立 s09 company だけ直接掃除する（監査行は触らない）。
	require.NoError(t, db.Exec(`DELETE FROM companies WHERE name LIKE ?
		AND id NOT IN (SELECT company_id FROM clinics)`,
		syntheticClosingCompanyPrefix+"%").Error)
}

// createSyntheticAuditFixture は S09 fixture を1セット作り、合成 staff を返す。
func createSyntheticAuditFixture(t *testing.T, db *gorm.DB) (*SyntheticClosingResult, model.Staff) {
	t.Helper()
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db",
		TargetDate: time.Date(2026, 9, 7, 0, 0, 0, 0, jst), PasswordHash: "x",
	})
	require.NoError(t, err)
	var staffRow model.Staff
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).First(&staffRow).Error)
	return got, staffRow
}

// createAuditRow は検証用の audit_logs 行を任意の clinic/actor で作る。
// created_at は保存されるか検証するため呼び出し側が固定値を渡す。
func createAuditRow(t *testing.T, db *gorm.DB, clinicID, actorID *uint64, actorType, action string, createdAt time.Time) *model.AuditLog {
	t.Helper()
	row := &model.AuditLog{
		ClinicID:  clinicID,
		ActorID:   actorID,
		ActorType: actorType,
		Action:    action,
		Resource:  "session",
		CreatedAt: createdAt,
	}
	require.NoError(t, db.Create(row).Error)
	return row
}

// createPlainClinic は S09 / sentinel のどちらでもない通常 clinic + 実 staff を作る。
func createPlainClinic(t *testing.T, db *gorm.DB, name string) (model.Clinic, model.Staff) {
	t.Helper()
	company := &model.Company{Name: name + "-company"}
	require.NoError(t, db.Create(company).Error)
	clinic := &model.Clinic{CompanyID: company.ID, Name: name, IsActive: true}
	require.NoError(t, db.Create(clinic).Error)
	staffRow := &model.Staff{ClinicID: clinic.ID, Name: name + "-staff", IsActive: true, StaffType: model.StaffTypeDoctor}
	require.NoError(t, db.Create(staffRow).Error)
	return *clinic, *staffRow
}

// A1+A2: 監査行を持つ合成 clinic が既定経路（DELETE /api/v1/uat/synthetic-closings/:id
// と同じ DeleteSyntheticClosingFixture）で完遂し、audit 行は削除されず
// actor_id / clinic_id が sentinel へ付け替わる。
func TestDeleteSyntheticClosingFixture_AuditRowsAnonymizedByDefault(t *testing.T) {
	db := testdbSetupSyntheticAudit(t)
	ctx := context.Background()
	got, staffRow := createSyntheticAuditFixture(t, db)

	at := time.Date(2026, 9, 8, 9, 0, 0, 0, time.UTC)
	login := createAuditRow(t, db, &got.ClinicID, &staffRow.ID, model.AuditActorTypeStaff, model.AuditActionAuthLoginSuccess, at)
	logout := createAuditRow(t, db, &got.ClinicID, &staffRow.ID, model.AuditActorTypeStaff, model.AuditActionAuthLogout, at.Add(time.Hour))

	require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))

	// teardown 完遂: clinic・その staffs・fixture 子孫は消える
	require.Error(t, db.WithContext(ctx).First(&model.Clinic{}, got.ClinicID).Error)
	var staffCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Staff{}).Unscoped().Where("clinic_id = ?", got.ClinicID).Count(&staffCount).Error)
	assert.Zero(t, staffCount)
	var billingCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Billing{}).Where("clinic_id = ?", got.ClinicID).Count(&billingCount).Error)
	assert.Zero(t, billingCount)

	var sentinelStaff model.Staff
	require.NoError(t, db.WithContext(ctx).First(&sentinelStaff, "name = ?", syntheticAuditSentinelStaffName).Error)
	var sentinelClinic model.Clinic
	require.NoError(t, db.WithContext(ctx).First(&sentinelClinic, "name = ?", syntheticAuditSentinelClinicName).Error)
	var sentinelCompany model.Company
	require.NoError(t, db.WithContext(ctx).First(&sentinelCompany, "name = ?", syntheticAuditSentinelCompanyName).Error)

	assert.False(t, sentinelStaff.IsActive, "sentinel staff must be disabled")
	assert.Nil(t, sentinelStaff.AccountID, "sentinel staff must carry no login-capable account")
	assert.Equal(t, sentinelClinic.ID, sentinelStaff.ClinicID)
	assert.False(t, sentinelClinic.IsActive, "sentinel clinic must be inactive")
	assert.Equal(t, sentinelCompany.ID, sentinelClinic.CompanyID)
	for _, name := range []string{sentinelStaff.Name, sentinelClinic.Name, sentinelCompany.Name} {
		assert.False(t, strings.HasPrefix(name, syntheticClosingClinicPrefix), "%q must not match the teardown clinic prefix", name)
		assert.False(t, strings.HasPrefix(name, syntheticClosingCompanyPrefix), "%q must not match the teardown company prefix", name)
	}

	// 監査行は物理削除されず、内容を保ったまま sentinel へ付け替わる
	var rows []model.AuditLog
	require.NoError(t, db.WithContext(ctx).Where("id IN ?", []uint64{login.ID, logout.ID}).Order("id").Find(&rows).Error)
	require.Len(t, rows, 2)
	for i, want := range []*model.AuditLog{login, logout} {
		got := rows[i]
		require.NotNil(t, got.ActorID)
		assert.Equal(t, sentinelStaff.ID, *got.ActorID, "actor_id must point at the sentinel staff")
		require.NotNil(t, got.ClinicID)
		assert.Equal(t, sentinelClinic.ID, *got.ClinicID, "clinic_id must point at the sentinel clinic")
		assert.Equal(t, model.AuditActorTypeStaff, got.ActorType, "actor_type stays staff")
		assert.Equal(t, want.Action, got.Action)
		assert.Equal(t, want.Resource, got.Resource)
		assert.True(t, want.CreatedAt.Equal(got.CreatedAt), "created_at must be preserved")
	}
}

// A3: sentinel の lazy find-or-create — audit 行が無い teardown では作らず、
// 作った後は冪等に再利用し、teardown が sentinel を消すこともない。
// 共有 schema では sentinel が先行テスト由来で既に存在し得るため、
// 「作らない」は変化なし、「作る・再利用する」は個数と ID で検証する。
func TestSyntheticClosingAuditSentinel_Lifecycle(t *testing.T) {
	db := testdbSetupSyntheticAudit(t)
	ctx := context.Background()

	countSentinels := func() (companies, clinics, staffs int64) {
		require.NoError(t, db.Model(&model.Company{}).Where("name = ?", syntheticAuditSentinelCompanyName).Count(&companies).Error)
		require.NoError(t, db.Model(&model.Clinic{}).Where("name = ?", syntheticAuditSentinelClinicName).Count(&clinics).Error)
		require.NoError(t, db.Model(&model.Staff{}).Where("name = ?", syntheticAuditSentinelStaffName).Count(&staffs).Error)
		return
	}

	t.Run("audit-free teardown creates no sentinel", func(t *testing.T) {
		beforeCompanies, beforeClinics, beforeStaffs := countSentinels()
		got, _ := createSyntheticAuditFixture(t, db)
		require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))
		companies, clinics, staffs := countSentinels()
		assert.Equal(t, beforeCompanies, companies, "audit-free teardown must not create a sentinel company")
		assert.Equal(t, beforeClinics, clinics, "audit-free teardown must not create a sentinel clinic")
		assert.Equal(t, beforeStaffs, staffs, "audit-free teardown must not create a sentinel staff")
	})

	var firstStaffID, firstClinicID uint64
	t.Run("audit-bearing teardown resolves to a single sentinel set", func(t *testing.T) {
		got, staffRow := createSyntheticAuditFixture(t, db)
		createAuditRow(t, db, &got.ClinicID, &staffRow.ID, model.AuditActorTypeStaff, model.AuditActionAuthLoginSuccess, time.Now())
		require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))
		companies, clinics, staffs := countSentinels()
		assert.Equal(t, int64(1), companies, "exactly one sentinel company per database")
		assert.Equal(t, int64(1), clinics, "exactly one sentinel clinic per database")
		assert.Equal(t, int64(1), staffs, "exactly one sentinel staff per database")
		var sentinelStaff model.Staff
		require.NoError(t, db.First(&sentinelStaff, "name = ?", syntheticAuditSentinelStaffName).Error)
		assert.False(t, sentinelStaff.IsActive)
		assert.Nil(t, sentinelStaff.AccountID)
		firstStaffID = sentinelStaff.ID
		firstClinicID = sentinelStaff.ClinicID
	})

	t.Run("second teardown reuses sentinel and never deletes it", func(t *testing.T) {
		got, staffRow := createSyntheticAuditFixture(t, db)
		createAuditRow(t, db, &got.ClinicID, &staffRow.ID, model.AuditActorTypeStaff, model.AuditActionAuthLogout, time.Now())
		require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))
		companies, clinics, staffs := countSentinels()
		assert.Equal(t, int64(1), companies)
		assert.Equal(t, int64(1), clinics)
		assert.Equal(t, int64(1), staffs)
		var sentinelStaff model.Staff
		require.NoError(t, db.First(&sentinelStaff, "name = ?", syntheticAuditSentinelStaffName).Error)
		assert.Equal(t, firstStaffID, sentinelStaff.ID, "sentinel staff must be reused idempotently")
		assert.Equal(t, firstClinicID, sentinelStaff.ClinicID, "sentinel clinic must be reused idempotently")
	})
}

// A4: 再割当のスコープは厳密 — actor_id は削除対象 clinic の staff が actor の行
// （switch_clinic で他 clinic 属性の行を含む）だけ、clinic_id は削除対象 clinic 上の行
// （system / 実 staff actor を含む）だけ。無関係な行は触らない。
func TestSyntheticClosingAuditReassignment_StrictScope(t *testing.T) {
	db := testdbSetupSyntheticAudit(t)
	ctx := context.Background()
	got, synStaff := createSyntheticAuditFixture(t, db)
	otherClinic, realStaff := createPlainClinic(t, db, "plain-clinic")

	base := time.Date(2026, 9, 9, 8, 0, 0, 0, time.UTC)
	synOnSynthetic := createAuditRow(t, db, &got.ClinicID, &synStaff.ID, model.AuditActorTypeStaff, model.AuditActionAuthLoginSuccess, base)
	synOnOther := createAuditRow(t, db, &otherClinic.ID, &synStaff.ID, model.AuditActorTypeStaff, "switch_clinic", base.Add(time.Minute))
	realOnSynthetic := createAuditRow(t, db, &got.ClinicID, &realStaff.ID, model.AuditActorTypeStaff, model.AuditActionBillingCancel, base.Add(2*time.Minute))
	systemOnSynthetic := createAuditRow(t, db, &got.ClinicID, nil, model.AuditActorTypeSystem, model.AuditActionReservationNoShow, base.Add(3*time.Minute))
	unrelated := createAuditRow(t, db, &otherClinic.ID, &realStaff.ID, model.AuditActorTypeStaff, model.AuditActionAuthLogout, base.Add(4*time.Minute))

	require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))

	var sentinelStaff model.Staff
	require.NoError(t, db.First(&sentinelStaff, "name = ?", syntheticAuditSentinelStaffName).Error)
	var sentinelClinic model.Clinic
	require.NoError(t, db.First(&sentinelClinic, "name = ?", syntheticAuditSentinelClinicName).Error)

	load := func(t *testing.T, id uint64) model.AuditLog {
		t.Helper()
		var row model.AuditLog
		require.NoError(t, db.First(&row, id).Error)
		return row
	}

	// 合成 staff が合成 clinic に残した行: actor・clinic とも sentinel へ
	row := load(t, synOnSynthetic.ID)
	require.NotNil(t, row.ActorID)
	assert.Equal(t, sentinelStaff.ID, *row.ActorID)
	require.NotNil(t, row.ClinicID)
	assert.Equal(t, sentinelClinic.ID, *row.ClinicID)
	assert.Equal(t, model.AuditActorTypeStaff, row.ActorType)
	assert.Equal(t, model.AuditActionAuthLoginSuccess, row.Action)
	assert.True(t, base.Equal(row.CreatedAt))

	// 合成 staff が他 clinic に残した行（switch_clinic 切替先行）:
	// actor は sentinel へ、clinic は切替先のまま
	row = load(t, synOnOther.ID)
	require.NotNil(t, row.ActorID)
	assert.Equal(t, sentinelStaff.ID, *row.ActorID)
	require.NotNil(t, row.ClinicID)
	assert.Equal(t, otherClinic.ID, *row.ClinicID)

	// 実 staff が合成 clinic に残した行: clinic は sentinel へ、actor は実 staff のまま
	row = load(t, realOnSynthetic.ID)
	require.NotNil(t, row.ActorID)
	assert.Equal(t, realStaff.ID, *row.ActorID)
	require.NotNil(t, row.ClinicID)
	assert.Equal(t, sentinelClinic.ID, *row.ClinicID)

	// system 行が合成 clinic 上にある: clinic は sentinel へ、actor は NULL・actor_type は system のまま
	row = load(t, systemOnSynthetic.ID)
	assert.Nil(t, row.ActorID)
	require.NotNil(t, row.ClinicID)
	assert.Equal(t, sentinelClinic.ID, *row.ClinicID)
	assert.Equal(t, model.AuditActorTypeSystem, row.ActorType)

	// 無関係な行（他 clinic × 実 staff）は一切触らない
	row = load(t, unrelated.ID)
	require.NotNil(t, row.ActorID)
	assert.Equal(t, realStaff.ID, *row.ActorID)
	require.NotNil(t, row.ClinicID)
	assert.Equal(t, otherClinic.ID, *row.ClinicID)

	// 無関係な clinic / staff も teardown に消されない
	assert.NoError(t, db.First(&model.Clinic{}, otherClinic.ID).Error)
	assert.NoError(t, db.First(&model.Staff{}, realStaff.ID).Error)
}

// A5: teardown tx 内の失敗は reassignment・sentinel 作成・delete 系列を
// まとめて rollback する。audit UPDATE を一時 trigger で確実に失敗させて検証する。
func TestDeleteSyntheticClosingFixture_AuditPolicyFailureRollsBack(t *testing.T) {
	db := testdbSetupSyntheticAudit(t)
	ctx := context.Background()
	got, staffRow := createSyntheticAuditFixture(t, db)
	at := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	audit := createAuditRow(t, db, &got.ClinicID, &staffRow.ID, model.AuditActorTypeStaff, model.AuditActionAuthLoginSuccess, at)

	// 共有 schema の sentinel 個数を rollback 前後で比較する（先行テスト由来の
	// sentinel が既に存在してよい — 増えないことが契約）。
	var sentinelBefore int64
	require.NoError(t, db.Model(&model.Staff{}).Where("name = ?", syntheticAuditSentinelStaffName).Count(&sentinelBefore).Error)

	require.NoError(t, db.Exec(`
		CREATE OR REPLACE FUNCTION s09_audit_test_block_update()
		RETURNS trigger AS $$
		BEGIN
		    RAISE EXCEPTION 'audit update blocked for rollback test';
		END;
		$$ LANGUAGE plpgsql
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER trg_s09_audit_test_block_update
		BEFORE UPDATE ON audit_logs
		FOR EACH ROW EXECUTE FUNCTION s09_audit_test_block_update()
	`).Error)
	t.Cleanup(func() {
		// rollback 済み fixture は共有 schema 上に残るため、trigger を外してから
		// 既定 policy の teardown で後始末する（順序が重要）。
		_ = db.Exec("DROP TRIGGER IF EXISTS trg_s09_audit_test_block_update ON audit_logs").Error
		_ = db.Exec("DROP FUNCTION IF EXISTS s09_audit_test_block_update()").Error
		assert.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))
	})

	err := DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken)
	require.Error(t, err, "a failing reassignment must abort teardown")

	// delete 系列ごと rollback: fixture は残る
	assert.NoError(t, db.First(&model.Clinic{}, got.ClinicID).Error)
	assert.NoError(t, db.First(&model.Staff{}, staffRow.ID).Error)
	var billingCount int64
	require.NoError(t, db.Model(&model.Billing{}).Where("clinic_id = ?", got.ClinicID).Count(&billingCount).Error)
	assert.Equal(t, int64(5), billingCount)

	// 監査行は元の帰属のまま
	var row model.AuditLog
	require.NoError(t, db.First(&row, audit.ID).Error)
	require.NotNil(t, row.ActorID)
	assert.Equal(t, staffRow.ID, *row.ActorID)
	require.NotNil(t, row.ClinicID)
	assert.Equal(t, got.ClinicID, *row.ClinicID)

	// tx 内で作られた sentinel も rollback で消える — 個数が増えない。
	var sentinelAfter int64
	require.NoError(t, db.Model(&model.Staff{}).Where("name = ?", syntheticAuditSentinelStaffName).Count(&sentinelAfter).Error)
	assert.Equal(t, sentinelBefore, sentinelAfter, "a rolled-back teardown must not leave a new sentinel staff")
}
