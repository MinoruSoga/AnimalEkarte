package auth

// permission_group_effective_permissions_test.go — G11-1 (test-effective-permissions-sql-untested)
//
// テスト対象: PermissionGroupRepository.FindAllEffectivePermissionsByStaffID の raw SQL
// (permission_group_repository.go:144-173)。
//
// 保護する不変条件:
//   - 複数グループ所属時のルールは resource 毎に bool_or で合成される
//   - is_active=false のグループのルールは実効権限に混入しない
//   - deleted_at IS NOT NULL (ソフトデリート済み) のグループのルールは混入しない
//   - clinicID 述語により所属外クリニックのグループのルールは混入しない (High-7: マルチクリニック昇格防止)
//
// これらの述語 (pg.is_active = true / pg.deleted_at IS NULL / pg.clinic_id = ?) のいずれかが
// SQL から脱落すると、該当するテストケースが必ず失敗するよう設計されている。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupEffectivePermissionsTestDB は実効権限集約テスト用の DB を整備する。
// permission_group_staff_clinic_isolation_test.go の setupPermissionGroupStaffIsolationTestDB を踏襲し、
// 本テストが追加で必要とする PermissionGroupRule も AutoMigrate + TRUNCATE 対象に含める。
func setupEffectivePermissionsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Staff{}, &model.PermissionGroup{}, &model.StaffPermissionGroup{}, &model.PermissionGroupRule{},
	))
	ensureStaffPermissionGroupsCreatedAt(t, db)
	testdb.Truncate(t, db, "staff_permission_groups", "permission_group_rules", "permission_groups", "staffs")
	return db
}

// makeEffPermGroup は指定クリニック・is_active・deleted_at 状態の権限グループを作成する。
func makeEffPermGroup(t *testing.T, db *gorm.DB, clinicID uint64, name string, isActive bool) *model.PermissionGroup {
	t.Helper()
	g := &model.PermissionGroup{ClinicID: clinicID, Name: name, IsActive: isActive}
	require.NoError(t, db.WithContext(context.Background()).Create(g).Error)
	if !isActive {
		// IsActive は bool ゼロ値のため Create 時に false を明示送出できない(GORM がゼロ値を省略しDB defaultのtrueが入る)。
		// 作成後に明示 UPDATE して false を反映する。
		require.NoError(t, db.Model(&model.PermissionGroup{}).Where("id = ?", g.ID).Update("is_active", false).Error)
		g.IsActive = false
	}
	return g
}

// makeEffPermRule は権限グループに対して resource 単位のルールを作成する。
func makeEffPermRule(t *testing.T, db *gorm.DB, groupID uint64, resource string, canView, canCreate, canEdit, canDelete bool) *model.PermissionGroupRule {
	t.Helper()
	r := &model.PermissionGroupRule{
		GroupID: groupID, Resource: resource,
		CanView: canView, CanCreate: canCreate, CanEdit: canEdit, CanDelete: canDelete,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(r).Error)
	return r
}

// assignStaffToGroup は staff_permission_groups に staff-group の紐付けを直接挿入する
// (UpdateStaffGroups は clinic_id 越境検証を含み、本テストの意図的なクロステナントfixture構築には使えないため)。
func assignStaffToGroup(t *testing.T, db *gorm.DB, staffID, groupID uint64) {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).
		Create(&model.StaffPermissionGroup{StaffID: staffID, GroupID: groupID}).Error)
}

// findRuleByResource はテスト対象メソッドの戻り値から resource で1件検索するテストヘルパー。
func findRuleByResource(t *testing.T, rules []model.PermissionGroupRule, resource string) model.PermissionGroupRule {
	t.Helper()
	for i := range rules {
		if rules[i].Resource == resource {
			return rules[i]
		}
	}
	t.Fatalf("resource %q が結果に含まれていない: %+v", resource, rules)
	return model.PermissionGroupRule{}
}

func TestPermissionGroupRepository_FindAllEffectivePermissions(t *testing.T) {
	const clinicA = uint64(1)
	const clinicB = uint64(2)

	t.Run("複数グループ所属時、resource毎にbool_orで合成される", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		staff := makeDoctor(t, db, clinicA, "bool_or合成テスト用スタッフ")
		groupViewOnly := makeEffPermGroup(t, db, clinicA, "閲覧のみグループ", true)
		groupEditOnly := makeEffPermGroup(t, db, clinicA, "編集のみグループ", true)
		// グループA: can_view=true, can_delete=false / グループB: can_view=false, can_delete=false → 合成後 can_view=true, can_delete=false
		makeEffPermRule(t, db, groupViewOnly.ID, "medical_record", true, false, false, false)
		makeEffPermRule(t, db, groupEditOnly.ID, "medical_record", false, false, true, false)
		assignStaffToGroup(t, db, staff.ID, groupViewOnly.ID)
		assignStaffToGroup(t, db, staff.ID, groupEditOnly.ID)

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)

		r := findRuleByResource(t, rules, "medical_record")
		assert.True(t, r.CanView, "いずれかのグループがcan_view=trueならOR合成でtrueになるべき")
		assert.True(t, r.CanEdit, "いずれかのグループがcan_edit=trueならOR合成でtrueになるべき")
		assert.False(t, r.CanCreate, "両グループともcan_create=falseならfalseのままであるべき")
		assert.False(t, r.CanDelete, "両グループともcan_delete=falseならfalseのままであるべき")
	})

	t.Run("is_active=falseのグループのルールは実効権限から除外される", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		staff := makeDoctor(t, db, clinicA, "is_active除外テスト用スタッフ")
		inactiveGroup := makeEffPermGroup(t, db, clinicA, "無効化グループ", false)
		makeEffPermRule(t, db, inactiveGroup.ID, "billing", true, true, true, true)
		assignStaffToGroup(t, db, staff.ID, inactiveGroup.ID)

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)

		for _, r := range rules {
			assert.NotEqual(t, "billing", r.Resource, "is_active=falseのグループのルールが混入している: %+v", r)
		}
	})

	t.Run("同一resourceでactiveグループとinactiveグループのルールが競合してもinactive側は混入しない", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		staff := makeDoctor(t, db, clinicA, "active_inactive競合テスト用スタッフ")
		activeGroup := makeEffPermGroup(t, db, clinicA, "有効グループ", true)
		inactiveGroup := makeEffPermGroup(t, db, clinicA, "無効グループ", false)
		// 同一resource(billing)に対し、activeはcan_delete=false、inactiveはcan_delete=trueを持つ。
		// is_active述語がGROUP BY前のJOIN ONで正しく効いていれば、inactive側のcan_delete=trueはbool_orに混入しない。
		makeEffPermRule(t, db, activeGroup.ID, "billing", true, false, false, false)
		makeEffPermRule(t, db, inactiveGroup.ID, "billing", true, true, true, true)
		assignStaffToGroup(t, db, staff.ID, activeGroup.ID)
		assignStaffToGroup(t, db, staff.ID, inactiveGroup.ID)

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)

		r := findRuleByResource(t, rules, "billing")
		assert.True(t, r.CanView, "activeグループのcan_view=trueは反映されるべき")
		assert.False(t, r.CanCreate, "inactiveグループのcan_create=trueが混入してはならない")
		assert.False(t, r.CanEdit, "inactiveグループのcan_edit=trueが混入してはならない")
		assert.False(t, r.CanDelete, "inactiveグループのcan_delete=trueが混入してはならない")
	})

	t.Run("deleted_atがNOT NULLのグループのルールは実効権限から除外される", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		staff := makeDoctor(t, db, clinicA, "ソフトデリート除外テスト用スタッフ")
		deletedGroup := makeEffPermGroup(t, db, clinicA, "削除済みグループ", true)
		makeEffPermRule(t, db, deletedGroup.ID, "owner", true, true, true, true)
		assignStaffToGroup(t, db, staff.ID, deletedGroup.ID)
		require.NoError(t, db.Delete(&model.PermissionGroup{}, deletedGroup.ID).Error)

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)

		for _, r := range rules {
			assert.NotEqual(t, "owner", r.Resource, "ソフトデリート済みグループのルールが混入している: %+v", r)
		}
	})

	t.Run("deleted_atがNOT NULLのルールは実効権限から除外される", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		staff := makeDoctor(t, db, clinicA, "ソフトデリートルール除外テスト用スタッフ")
		group := makeEffPermGroup(t, db, clinicA, "有効グループ", true)
		deletedRule := makeEffPermRule(t, db, group.ID, "billing", true, true, true, true)
		assignStaffToGroup(t, db, staff.ID, group.ID)
		require.NoError(t, db.Delete(&model.PermissionGroupRule{}, deletedRule.ID).Error)

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)

		for _, rule := range rules {
			assert.NotEqual(t, "billing", rule.Resource, "ソフトデリート済みルールが混入している: %+v", rule)
		}
	})

	t.Run("High-7: クリニック隔離— clinicID指定時に他クリニックのグループのルールが混入しない", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		// staff は clinicA 所属だが、直接 insert により clinicB のグループにも紐付いている状態を構築する
		// (実運用では UpdateStaffGroups がこれを拒否するが、本テストは SQL 述語自体の防御を検証する)。
		staff := makeDoctor(t, db, clinicA, "クリニック隔離テスト用スタッフ")
		groupA := makeEffPermGroup(t, db, clinicA, "clinic Aグループ", true)
		groupB := makeEffPermGroup(t, db, clinicB, "clinic Bグループ", true)
		makeEffPermRule(t, db, groupA.ID, "vaccination", true, false, false, false)
		makeEffPermRule(t, db, groupB.ID, "vaccination", false, false, false, true)
		assignStaffToGroup(t, db, staff.ID, groupA.ID)
		assignStaffToGroup(t, db, staff.ID, groupB.ID)

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)

		r := findRuleByResource(t, rules, "vaccination")
		assert.True(t, r.CanView, "clinicAのグループのルールは反映されるべき")
		assert.False(t, r.CanDelete, "clinicBのグループのcan_delete=trueが混入してはならない(マルチクリニック昇格防止)")
	})

	t.Run("所属グループなしのスタッフは空スライスを返す", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		staff := makeDoctor(t, db, clinicA, "未所属テスト用スタッフ")

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)
		assert.Empty(t, rules, "所属グループがなければ空スライスを返すべき")
	})

	t.Run("resource毎のGROUP BY — 複数resourceの行数と値が正しく分離される", func(t *testing.T) {
		db := setupEffectivePermissionsTestDB(t)
		repo := NewPermissionGroupRepository(db)
		ctx := context.Background()

		staff := makeDoctor(t, db, clinicA, "GROUP BYテスト用スタッフ")
		groupOne := makeEffPermGroup(t, db, clinicA, "グループ1", true)
		groupTwo := makeEffPermGroup(t, db, clinicA, "グループ2", true)
		// groupOne と groupTwo は同一resource(medical_record)を持つ + groupTwo のみ別resource(billing)を持つ
		makeEffPermRule(t, db, groupOne.ID, "medical_record", true, false, false, false)
		makeEffPermRule(t, db, groupTwo.ID, "medical_record", false, true, false, false)
		makeEffPermRule(t, db, groupTwo.ID, "billing", false, false, false, true)
		assignStaffToGroup(t, db, staff.ID, groupOne.ID)
		assignStaffToGroup(t, db, staff.ID, groupTwo.ID)

		rules, err := repo.FindAllEffectivePermissionsByStaffID(ctx, staff.ID, clinicA)
		require.NoError(t, err)
		require.Len(t, rules, 2, "resource毎に1行ずつ、計2行に集約されるべき")

		medicalRecord := findRuleByResource(t, rules, "medical_record")
		assert.True(t, medicalRecord.CanView)
		assert.True(t, medicalRecord.CanCreate)

		billing := findRuleByResource(t, rules, "billing")
		assert.False(t, billing.CanView)
		assert.True(t, billing.CanDelete)
	})
}
