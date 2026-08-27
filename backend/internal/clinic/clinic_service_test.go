package clinic

import (
	"context"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func clinicBoolPtr(value bool) *bool {
	return &value
}

// mockClinicRepository は ClinicRepository のテスト用モック実装
type mockClinicRepository struct {
	findAllFn               func(ctx context.Context) ([]model.Clinic, error)
	findByStaffIDFn         func(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	findByIDFn              func(ctx context.Context, id uint64) (*model.Clinic, error)
	lockForUpdateFn         func(ctx context.Context, id uint64) (*model.Clinic, error)
	getCompanyFn            func(ctx context.Context) (*model.Company, error)
	createFn                func(ctx context.Context, clinic *model.Clinic) error
	updateFn                func(ctx context.Context, id uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, id uint64) error
	countOwnersByClinicIDFn func(ctx context.Context, clinicID uint64) (int64, error)
	countStaffByClinicIDFn  func(ctx context.Context, clinicID uint64) (int64, error)
	countBlockingRefsFn     func(ctx context.Context, clinicID uint64) ([]ClinicDependencyCount, error)
}

// DeleteSoftDeletedByClinicID keeps the shared permission-group mock compatible with
// clinic.PermissionGroupWriter after the write owner gains clinic cleanup.
// Clinic transaction tests use mockClinicPermissionGroupWriter below when they need
// to observe or fail the cleanup call.
func (m *mockPermissionGroupRepository) DeleteSoftDeletedByClinicID(_ context.Context, _ uint64) error {
	return nil
}

type mockClinicPermissionGroupWriter struct {
	*mockPermissionGroupRepository
	deleteSoftDeletedByClinicIDFn func(ctx context.Context, clinicID uint64) error
}

func (m *mockClinicPermissionGroupWriter) DeleteSoftDeletedByClinicID(ctx context.Context, clinicID uint64) error {
	if m.deleteSoftDeletedByClinicIDFn != nil {
		return m.deleteSoftDeletedByClinicIDFn(ctx, clinicID)
	}
	return nil
}

func (m *mockClinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	return m.findAllFn(ctx)
}

func (m *mockClinicRepository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error) {
	if m.findByStaffIDFn == nil {
		return nil, nil
	}
	return m.findByStaffIDFn(ctx, staffID)
}

func (m *mockClinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockClinicRepository) LockActiveByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	return m.FindByID(ctx, id)
}

func (m *mockClinicRepository) LockByIDForUpdate(ctx context.Context, id uint64) (*model.Clinic, error) {
	if m.lockForUpdateFn != nil {
		return m.lockForUpdateFn(ctx, id)
	}
	return &model.Clinic{ID: id, IsActive: true}, nil
}

func (m *mockClinicRepository) FindCompany(ctx context.Context) (*model.Company, error) {
	return m.getCompanyFn(ctx)
}

func (m *mockClinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	return m.createFn(ctx, clinic)
}

func (m *mockClinicRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, id, fields)
}

func (m *mockClinicRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockClinicRepository) CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	if m.countOwnersByClinicIDFn == nil {
		return 0, nil
	}
	return m.countOwnersByClinicIDFn(ctx, clinicID)
}

func (m *mockClinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	if m.countStaffByClinicIDFn == nil {
		return 0, nil
	}
	return m.countStaffByClinicIDFn(ctx, clinicID)
}

func (m *mockClinicRepository) CountBlockingReferencesByClinicID(ctx context.Context, clinicID uint64) ([]ClinicDependencyCount, error) {
	if m.countBlockingRefsFn == nil {
		return nil, nil
	}
	return m.countBlockingRefsFn(ctx, clinicID)
}

func TestClinicService_ListClinics(t *testing.T) {
	tests := []struct {
		name        string
		repoClinics []model.Clinic
		repoErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name: "returns clinic list",
			repoClinics: []model.Clinic{
				{ID: 1, CompanyID: 1, Name: "本院"},
				{ID: 2, CompanyID: 1, Name: "分院"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:        "returns empty list when no clinics exist",
			repoClinics: []model.Clinic{},
			repoErr:     nil,
			wantLen:     0,
			wantErr:     false,
		},
		{
			name:        "propagates repository error",
			repoClinics: nil,
			repoErr:     errors.New("db connection error"),
			wantLen:     0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				findAllFn: func(_ context.Context) ([]model.Clinic, error) {
					return tt.repoClinics, tt.repoErr
				},
			}
			pgRepo := &mockPermissionGroupRepository{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

			clinics, err := svc.ListClinics(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, clinics, tt.wantLen)
			}
		})
	}
}

func TestClinicService_ListClinicsByStaffID(t *testing.T) {
	tests := []struct {
		name        string
		staffID     uint64
		repoClinics []model.Clinic
		repoErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name:    "returns clinics for staff",
			staffID: 1,
			repoClinics: []model.Clinic{
				{ID: 1, CompanyID: 1, Name: "本院"},
				{ID: 2, CompanyID: 1, Name: "分院"},
			},
			wantLen: 2,
		},
		{
			name:        "returns empty list when staff belongs to no clinics",
			staffID:     2,
			repoClinics: []model.Clinic{},
			wantLen:     0,
		},
		{
			name:    "propagates repository error",
			staffID: 3,
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				findByStaffIDFn: func(_ context.Context, staffID uint64) ([]model.Clinic, error) {
					assert.Equal(t, tt.staffID, staffID)
					return tt.repoClinics, tt.repoErr
				},
			}
			pgRepo := &mockPermissionGroupRepository{}
			svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

			clinics, err := svc.ListClinicsByStaffID(context.Background(), tt.staffID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, clinics, tt.wantLen)
			}
		})
	}
}

func TestClinicService_GetClinicByID(t *testing.T) {
	tests := []struct {
		name       string
		id         uint64
		repoClinic *model.Clinic
		repoErr    error
		wantErr    bool
		wantNF     bool
	}{
		{
			name: "returns clinic when found",
			id:   1,
			repoClinic: &model.Clinic{
				ID:        1,
				CompanyID: 1,
				Name:      "本院",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:       "returns not found error when clinic does not exist",
			id:         999,
			repoClinic: nil,
			repoErr:    apperrors.WrapNotFound("clinic", "999"),
			wantErr:    true,
			wantNF:     true,
		},
		{
			name:       "returns error on repository failure",
			id:         1,
			repoClinic: nil,
			repoErr:    errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
					return tt.repoClinic, tt.repoErr
				},
			}
			pgRepo := &mockPermissionGroupRepository{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

			clinic, err := svc.GetClinicByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoClinic, clinic)
			}
		})
	}
}

func TestClinicService_CreateClinic(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateClinicInput
		repoCompany    *model.Company
		repoCompanyErr error
		repoCreateErr  error
		wantErr        bool
		wantCompanyID  uint64
	}{
		{
			name:          "creates clinic successfully with company id set",
			input:         &CreateClinicInput{Name: "新規院"},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			wantErr:       false,
			wantCompanyID: 5,
		},
		{
			name:           "returns error when company retrieval fails",
			input:          &CreateClinicInput{Name: "新規院"},
			repoCompanyErr: apperrors.WrapNotFound("company", "singleton"),
			wantErr:        true,
		},
		{
			name:          "returns error when clinic creation fails",
			input:         &CreateClinicInput{Name: "既存院"},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			repoCreateErr: apperrors.WrapAlreadyExists("clinic", "既存院"),
			wantErr:       true,
		},
		{
			name:          "returns error on repository failure",
			input:         &CreateClinicInput{Name: "エラー院"},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			repoCreateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				getCompanyFn: func(_ context.Context) (*model.Company, error) {
					return tt.repoCompany, tt.repoCompanyErr
				},
				createFn: func(_ context.Context, _ *model.Clinic) error {
					return tt.repoCreateErr
				},
			}
			pgRepo := &mockPermissionGroupRepository{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

			result, err := svc.CreateClinic(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantCompanyID, result.CompanyID)
			}
		})
	}
}

// SD-9: CreateClinic はグループ (permission_groups) を作るだけでルール
// (permission_group_rules) を1件も作らず、is_system_admin 以外の全スタッフが
// 新規クリニックで全リソースへアクセス不能になるバグがあった。
// 修正後は defaultPermissionRuleTable 由来のルールが両グループへ流し込まれることを検証する。
func TestClinicService_CreateClinic_DefaultPermissionGroupRules(t *testing.T) {
	repo := &mockClinicRepository{
		getCompanyFn: func(_ context.Context) (*model.Company, error) {
			return &model.Company{ID: 1, Name: "グループ本社"}, nil
		},
		createFn: func(_ context.Context, clinic *model.Clinic) error {
			clinic.ID = 42
			return nil
		},
	}

	type capturedRules struct {
		groupName string
		rules     []model.PermissionGroupRule
	}
	var created []*model.PermissionGroup
	var captured []capturedRules

	pgRepo := &mockPermissionGroupRepository{
		createFn: func(_ context.Context, group *model.PermissionGroup) error {
			group.ID = uint64(len(created) + 1)
			created = append(created, group)
			return nil
		},
		setRulesFn: func(_ context.Context, clinicID, groupID uint64, rules []model.PermissionGroupRule) error {
			assert.Equal(t, uint64(42), clinicID)
			name := ""
			for _, g := range created {
				if g.ID == groupID {
					name = g.Name
				}
			}
			captured = append(captured, capturedRules{groupName: name, rules: rules})
			return nil
		},
	}

	svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

	result, err := svc.CreateClinic(context.Background(), &CreateClinicInput{Name: "新規院"})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, created, 2, "執行/一般 の2グループが作成されること")
	assert.Len(t, captured, 2, "各グループに UpdateRules が1回ずつ呼ばれること")

	findRule := func(rules []model.PermissionGroupRule, resource model.Resource) *model.PermissionGroupRule {
		for i := range rules {
			if rules[i].Resource == string(resource) {
				return &rules[i]
			}
		}
		return nil
	}

	for _, c := range captured {
		assert.Lenf(t, c.rules, len(model.AllResources),
			"group=%s: 全リソース分のルールが作成されること（ルール0件のデフォルトグループ回帰防止）", c.groupName)

		// 共有マスタ animal-species は is_system_admin 以外 mutation 禁止
		species := findRule(c.rules, model.ResourceMasterAnimalSpecies)
		if assert.NotNilf(t, species, "group=%s: master-animal-species ルールが存在すること", c.groupName) {
			assert.Truef(t, species.CanView, "group=%s: master-animal-species は閲覧可能であること", c.groupName)
			assert.Falsef(t, species.CanCreate, "group=%s: master-animal-species は作成不可（system-admin only）", c.groupName)
			assert.Falsef(t, species.CanEdit, "group=%s: master-animal-species は編集不可（system-admin only）", c.groupName)
			assert.Falsef(t, species.CanDelete, "group=%s: master-animal-species は削除不可（system-admin only）", c.groupName)
		}

		switch c.groupName {
		case "執行":
			owners := findRule(c.rules, model.ResourceOwners)
			if assert.NotNil(t, owners, "執行に owners ルールが存在すること") {
				assert.True(t, owners.CanView, "執行は owners を閲覧可能であること")
				assert.True(t, owners.CanCreate, "執行は owners を作成可能であること")
				assert.True(t, owners.CanEdit, "執行は owners を編集可能であること")
				assert.True(t, owners.CanDelete, "執行は owners を削除可能であること")
			}
			hs := findRule(c.rules, model.ResourceHospitalSettings)
			if assert.NotNil(t, hs, "執行に hospital-settings ルールが存在すること") {
				assert.True(t, hs.CanView)
				assert.True(t, hs.CanEdit)
				assert.False(t, hs.CanCreate, "設定系リソースは執行でも作成不可であること")
				assert.False(t, hs.CanDelete, "設定系リソースは執行でも削除不可であること")
			}
		case "一般":
			owners := findRule(c.rules, model.ResourceOwners)
			if assert.NotNil(t, owners, "一般に owners ルールが存在すること") {
				assert.True(t, owners.CanView)
				assert.True(t, owners.CanCreate)
				assert.True(t, owners.CanEdit)
				assert.False(t, owners.CanDelete, "一般は owners を削除できないこと")
			}
			mp := findRule(c.rules, model.ResourceMasterPermission)
			if assert.NotNil(t, mp, "一般に master-permission ルールが存在すること") {
				assert.False(t, mp.CanView, "一般は権限マスタを閲覧できないこと")
				assert.False(t, mp.CanCreate)
				assert.False(t, mp.CanEdit)
				assert.False(t, mp.CanDelete)
			}
		default:
			t.Fatalf("unexpected group name captured: %q", c.groupName)
		}
	}
}

// TestDefaultPermissionRuleTable_CoversAllResources は defaultPermissionRuleTable が
// model.AllResources (37) を過不足なくカバーし、共有マスタ animal-species が
// 執行・一般とも view-only、examination-unconfirm / checkup-package-import が default-deny であることを固定する。
func TestDefaultPermissionRuleTable_CoversAllResources(t *testing.T) {
	require.Len(t, model.AllResources, 37, "AllResources 件数の契約が変わったら permission rollout を同時に更新すること")
	require.Len(t, defaultPermissionRuleTable, len(model.AllResources),
		"defaultPermissionRuleTable は AllResources と同数であること")

	seen := make(map[model.Resource]int, len(defaultPermissionRuleTable))
	for _, r := range defaultPermissionRuleTable {
		seen[r.resource]++
	}
	for _, res := range model.AllResources {
		assert.Equalf(t, 1, seen[res], "AllResources の %s が defaultPermissionRuleTable に1回だけ現れること", res)
	}
	for res, n := range seen {
		assert.Equalf(t, 1, n, "defaultPermissionRuleTable の %s が重複していないこと", res)
	}

	for _, isExecutive := range []bool{true, false} {
		rules := buildDefaultPermissionGroupRules(isExecutive)
		require.Len(t, rules, len(model.AllResources))
		profile := "一般"
		if isExecutive {
			profile = "執行"
		}
		var species *model.PermissionGroupRule
		for i := range rules {
			if rules[i].Resource == string(model.ResourceMasterAnimalSpecies) {
				species = &rules[i]
				break
			}
		}
		if assert.NotNilf(t, species, "%s に master-animal-species があること", profile) {
			assert.True(t, species.CanView)
			assert.False(t, species.CanCreate)
			assert.False(t, species.CanEdit)
			assert.False(t, species.CanDelete)
		}

		var unconfirm *model.PermissionGroupRule
		for i := range rules {
			if rules[i].Resource == string(model.ResourceExaminationUnconfirm) {
				unconfirm = &rules[i]
				break
			}
		}
		if assert.NotNilf(t, unconfirm, "%s に examination-unconfirm があること", profile) {
			assert.False(t, unconfirm.CanView)
			assert.False(t, unconfirm.CanCreate)
			assert.False(t, unconfirm.CanEdit)
			assert.False(t, unconfirm.CanDelete)
		}

		var pkgImport *model.PermissionGroupRule
		for i := range rules {
			if rules[i].Resource == string(model.ResourceCheckupPackageImport) {
				pkgImport = &rules[i]
				break
			}
		}
		if assert.NotNilf(t, pkgImport, "%s に checkup-package-import があること", profile) {
			assert.False(t, pkgImport.CanView)
			assert.False(t, pkgImport.CanCreate)
			assert.False(t, pkgImport.CanEdit)
			assert.False(t, pkgImport.CanDelete)
		}
	}
}

// demoPermissionSeedGroupProfiles maps 002_master permission_groups IDs.
// 1/3 = 執行, 2/4 = 一般, 9 = 閲覧専用 (view-only).
var demoPermissionSeedGroupProfiles = map[uint64]string{
	1: "executive",
	2: "general",
	3: "executive",
	4: "general",
	9: "view-only",
}

func TestDemoSeedGroupRules_Parity(t *testing.T) {
	require.Len(t, model.AllResources, 37)
	require.Len(t, demoPermissionSeedGroupProfiles, 5, "002_master は 5 権限グループを持つ契約")
	seedResources := make([]model.Resource, 0, len(model.AllResources)-1)
	for _, resource := range model.AllResources {
		if resource != model.ResourceExaminationUnconfirm {
			seedResources = append(seedResources, resource)
		}
	}

	rulesPath := filepath.Join("..", "..", "migrations", "seeds", "002_master", "permission_group_rules.csv")
	f, err := os.Open(rulesPath) //nolint:gosec // fixed seed path relative to backend module root
	require.NoError(t, err, "seed CSV を読めること (cwd は backend/ 想定)")
	defer f.Close() //nolint:errcheck // test cleanup

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	require.Greater(t, len(records), 1, "header + rows")

	header := records[0]
	col := func(name string) int {
		t.Helper()
		for i, h := range header {
			if h == name {
				return i
			}
		}
		t.Fatalf("missing column %q", name)
		return -1
	}
	iGroup := col("group_id")
	iRes := col("resource")
	iView := col("can_view")
	iCreate := col("can_create")
	iEdit := col("can_edit")
	iDelete := col("can_delete")

	type seedRule struct {
		canView, canCreate, canEdit, canDelete bool
	}
	parseBool := func(s string) bool {
		switch s {
		case "t", "true", "TRUE", "1":
			return true
		case "f", "false", "FALSE", "0", "":
			return false
		default:
			t.Fatalf("unexpected bool %q", s)
			return false
		}
	}

	byGroup := make(map[uint64]map[string]seedRule, 9)
	for _, rec := range records[1:] {
		if len(rec) <= iDelete {
			t.Fatalf("short row: %#v", rec)
		}
		gid, err := strconv.ParseUint(rec[iGroup], 10, 64)
		require.NoError(t, err)
		if _, ok := byGroup[gid]; !ok {
			byGroup[gid] = make(map[string]seedRule, len(model.AllResources))
		}
		byGroup[gid][rec[iRes]] = seedRule{
			canView:   parseBool(rec[iView]),
			canCreate: parseBool(rec[iCreate]),
			canEdit:   parseBool(rec[iEdit]),
			canDelete: parseBool(rec[iDelete]),
		}
	}

	require.Len(t, byGroup, 5, "permission_group_rules は 5 グループすべてをカバーすること")

	execExpected := buildDefaultPermissionGroupRules(true)
	genExpected := buildDefaultPermissionGroupRules(false)
	expectedByProfile := map[string][]model.PermissionGroupRule{
		"executive": execExpected,
		"general":   genExpected,
	}

	for gid, profile := range demoPermissionSeedGroupProfiles {
		rules, ok := byGroup[gid]
		require.Truef(t, ok, "group %d のルールが seed に存在すること", gid)
		assert.Lenf(t, rules, len(seedResources),
			"group %d (%s): seed 管理リソース (%d) をカバーすること", gid, profile, len(seedResources))
		assert.NotContains(t, rules, string(model.ResourceExaminationUnconfirm),
			"group %d: examination-unconfirm は明示 rollout 前に seed 付与しない", gid)

		for _, res := range seedResources {
			_, has := rules[string(res)]
			assert.Truef(t, has, "group %d: resource %s が欠落", gid, res)
		}

		species, has := rules[string(model.ResourceMasterAnimalSpecies)]
		if assert.Truef(t, has, "group %d: master-animal-species が存在すること", gid) {
			assert.True(t, species.canView)
			assert.False(t, species.canCreate)
			assert.False(t, species.canEdit)
			assert.False(t, species.canDelete)
		}

		switch profile {
		case "view-only":
			for res, r := range rules {
				assert.Falsef(t, r.canCreate, "group 9 view-only: %s create 不可", res)
				assert.Falsef(t, r.canEdit, "group 9 view-only: %s edit 不可", res)
				assert.Falsef(t, r.canDelete, "group 9 view-only: %s delete 不可", res)
			}
			// master-permission のみ view も false（既存デモ契約）
			mp := rules[string(model.ResourceMasterPermission)]
			assert.False(t, mp.canView, "group 9: master-permission は閲覧不可")
		case "executive", "general":
			want := expectedByProfile[profile]
			require.Len(t, want, len(model.AllResources))
			for _, w := range want {
				if w.Resource == string(model.ResourceExaminationUnconfirm) {
					continue
				}
				got, has := rules[w.Resource]
				if !assert.Truef(t, has, "group %d: %s 欠落", gid, w.Resource) {
					continue
				}
				assert.Equalf(t, w.CanView, got.canView, "group %d %s can_view", gid, w.Resource)
				assert.Equalf(t, w.CanCreate, got.canCreate, "group %d %s can_create", gid, w.Resource)
				assert.Equalf(t, w.CanEdit, got.canEdit, "group %d %s can_edit", gid, w.Resource)
				assert.Equalf(t, w.CanDelete, got.canDelete, "group %d %s can_delete", gid, w.Resource)
			}
		default:
			t.Fatalf("unknown profile %q for group %d", profile, gid)
		}
	}
}

func TestClinicService_UpdateClinic(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		input         *UpdateClinicInput
		repoClinic    *model.Clinic
		repoFindErr   error
		repoUpdateErr error
		wantErr       bool
		wantNF        bool
		wantCompanyID uint64
	}{
		{
			name: "updates clinic successfully and returns fresh record from DB",
			id:   1,
			input: &UpdateClinicInput{
				Name:    clinicStringPtr("更新後院"),
				Address: clinicStringPtr("東京都渋谷区"),
			},
			repoClinic: &model.Clinic{
				ID:        1,
				CompanyID: 5,
				Name:      "更新後院",
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			wantErr:       false,
			wantCompanyID: 5,
		},
		{
			name:        "returns not found error when clinic does not exist",
			id:          999,
			input:       &UpdateClinicInput{Name: clinicStringPtr("存在しない院")},
			repoClinic:  nil,
			repoFindErr: apperrors.WrapNotFound("clinic", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:  "returns error on update failure",
			id:    1,
			input: &UpdateClinicInput{Name: clinicStringPtr("更新後院")},
			repoClinic: &model.Clinic{
				ID:        1,
				CompanyID: 5,
				Name:      "旧院名",
			},
			repoFindErr:   nil,
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
					return tt.repoClinic, tt.repoFindErr
				},
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			pgRepo := &mockPermissionGroupRepository{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

			result, err := svc.UpdateClinic(context.Background(), tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				// 更新後に FindByID でリフレッシュした結果が返ることを確認
				assert.Equal(t, tt.wantCompanyID, result.CompanyID)
				assert.Equal(t, tt.repoClinic.ID, result.ID)
			}
		})
	}
}

func TestClinicService_UpdateClinic_InputNil(t *testing.T) {
	repo := &mockClinicRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			t.Fatal("clinic must not be looked up when input is nil")
			return nil, nil
		},
	}
	pgRepo := &mockPermissionGroupRepository{}
	svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

	result, err := svc.UpdateClinic(context.Background(), 1, nil)

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestClinicService_UpdateClinic_NoFieldsProvided(t *testing.T) {
	existing := &model.Clinic{ID: 1, CompanyID: 5, Name: "既存院"}
	updateCalled := false
	repo := &mockClinicRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			return existing, nil
		},
		updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	pgRepo := &mockPermissionGroupRepository{}
	svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

	result, err := svc.UpdateClinic(context.Background(), 1, &UpdateClinicInput{})

	assert.NoError(t, err)
	assert.Equal(t, existing, result)
	assert.False(t, updateCalled, "更新フィールドが無い場合は repo.Update を呼ばない")
}

func TestClinicService_UpdateClinic_InvalidTaxRate(t *testing.T) {
	invalidRate := 1.5
	repo := &mockClinicRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			return &model.Clinic{ID: 1, CompanyID: 5}, nil
		},
		updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
			t.Fatal("clinic must not be updated when the input fails validation")
			return nil
		},
	}
	pgRepo := &mockPermissionGroupRepository{}
	svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

	result, err := svc.UpdateClinic(context.Background(), 1, &UpdateClinicInput{StandardTaxRate: &invalidRate})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestClinicService_UpdateClinic_RefetchErrorAfterUpdate(t *testing.T) {
	// POC-02 / X-01: update+reload share one WithTx callback. With a real Transactor the
	// failed reload rolls the update back; with the passthrough double we still must not
	// return a success envelope after reload error.
	findCalls := 0
	updateCalls := 0
	withTxCalls := 0
	repo := &mockClinicRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			findCalls++
			if findCalls == 1 {
				return &model.Clinic{ID: 1, CompanyID: 5, Name: "旧院名"}, nil
			}
			return nil, errors.New("db error")
		},
		updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
			updateCalls++
			return nil
		},
	}
	pgRepo := &mockPermissionGroupRepository{}
	tx := &clinicServiceTransactorDouble{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			withTxCalls++
			return fn(ctx)
		},
	}
	svc := NewClinicService(repo, pgRepo, tx)

	result, err := svc.UpdateClinic(context.Background(), 1, &UpdateClinicInput{Name: clinicStringPtr("新院名")})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, withTxCalls, "update+reload must run inside a single WithTx")
	assert.Equal(t, 1, updateCalls)
	assert.Equal(t, 2, findCalls, "pre-check find + in-tx reload")
}

// TestBuildClinicUpdate_AccountingDocumentSettings は #179 follow-up ①（#190）の
// 帳票レイアウト設定が PATCH セマンティクスで更新マップへ反映されることを検証する。
// 指定フィールドは実カラム名付きでマップへ入り、nil フィールドは省略される（既存値保持）。
func TestBuildClinicUpdate_AccountingDocumentSettings(t *testing.T) {
	t.Run("指定した帳票設定が実カラム名付きで更新マップへ入る", func(t *testing.T) {
		fields, err := BuildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentShowLogo:                clinicBoolPtr(true),
			AccountingDocumentShowRegistrationWarning: clinicBoolPtr(false),
			AccountingDocumentShowItemCategory:        clinicBoolPtr(false),
			AccountingDocumentFooterNote:              clinicStringPtr("ご来院ありがとうございました。"),
		})

		assert.NoError(t, err)
		assert.Equal(t, true, fields["accounting_document_show_logo"])
		assert.Equal(t, false, fields["accounting_document_show_registration_warning"])
		assert.Equal(t, false, fields["accounting_document_show_item_category"])
		assert.Equal(t, "ご来院ありがとうございました。", fields["accounting_document_footer_note"])
	})

	t.Run("空フッター文字列も明示更新として反映される（明示クリア可能）", func(t *testing.T) {
		fields, err := BuildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentFooterNote: clinicStringPtr(""),
		})

		assert.NoError(t, err)
		got, ok := fields["accounting_document_footer_note"]
		assert.True(t, ok, "空文字でもキーは存在する（フッターを明示的にクリアできる）")
		assert.Equal(t, "", got)
	})

	t.Run("未指定（nil）の帳票設定は更新マップへ入らない（PATCH: 既存値保持）", func(t *testing.T) {
		fields, err := BuildClinicUpdate(&UpdateClinicInput{Name: clinicStringPtr("更新後院")})

		assert.NoError(t, err)
		for _, col := range []string{
			"accounting_document_show_logo",
			"accounting_document_show_registration_warning",
			"accounting_document_show_item_category",
			"accounting_document_footer_note",
			"accounting_document_show_clinic_header",
			"accounting_document_show_owner_pet_info",
			"accounting_document_show_items_table",
			"accounting_document_show_payment_summary",
			"accounting_document_section_order",
		} {
			_, ok := fields[col]
			assert.Falsef(t, ok, "未指定フィールド %s は更新マップに含めない", col)
		}
	})

	t.Run("#190: セクション表示トグルが実カラム名付きで更新マップへ入る", func(t *testing.T) {
		f := false
		fields, err := BuildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentShowClinicHeader:   &f,
			AccountingDocumentShowOwnerPetInfo:   &f,
			AccountingDocumentShowItemsTable:     &f,
			AccountingDocumentShowPaymentSummary: &f,
		})

		assert.NoError(t, err)
		assert.Equal(t, false, fields["accounting_document_show_clinic_header"])
		assert.Equal(t, false, fields["accounting_document_show_owner_pet_info"])
		assert.Equal(t, false, fields["accounting_document_show_items_table"])
		assert.Equal(t, false, fields["accounting_document_show_payment_summary"])
	})

	t.Run("#190: 有効なセクション順序が更新マップへ入る", func(t *testing.T) {
		order := []string{"payment_summary", "items_table", "clinic_header", "owner_pet_info", "footer_note"}
		fields, err := BuildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentSectionOrder: &order,
		})

		assert.NoError(t, err)
		got, ok := fields["accounting_document_section_order"]
		assert.True(t, ok)
		assert.Equal(t, pq.StringArray(order), got)
	})

	t.Run("#190: 空の順序配列はデフォルト順リセットとして更新マップへ入る", func(t *testing.T) {
		empty := []string{}
		fields, err := BuildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentSectionOrder: &empty,
		})

		assert.NoError(t, err)
		got, ok := fields["accounting_document_section_order"]
		assert.True(t, ok, "空配列でもキーは存在する（デフォルト順にリセット）")
		assert.Equal(t, pq.StringArray{}, got)
	})

	t.Run("#190: 未知のセクションキーはエラーになる", func(t *testing.T) {
		invalid := []string{"clinic_header", "unknown_section"}
		_, err := BuildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentSectionOrder: &invalid,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "unknown key は WrapInvalidInput を返す")
	})

	t.Run("#190: 重複セクションキーはエラーになる", func(t *testing.T) {
		dup := []string{"clinic_header", "items_table", "clinic_header"}
		_, err := BuildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentSectionOrder: &dup,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "重複キーは WrapInvalidInput を返す")
	})
}

func TestBuildClinicUpdate_TaxRateValidation(t *testing.T) {
	tests := []struct {
		name             string
		standardTaxRate  *float64
		reducedTaxRate   *float64
		wantErr          bool
		wantStandardRate float64
		wantReducedRate  float64
	}{
		{
			name:             "有効な税率は更新マップへ入る",
			standardTaxRate:  clinicFloat64Ptr(0.10),
			reducedTaxRate:   clinicFloat64Ptr(0.08),
			wantErr:          false,
			wantStandardRate: 0.10,
			wantReducedRate:  0.08,
		},
		{
			name:            "standard_tax_rate が1を超える場合はエラー",
			standardTaxRate: clinicFloat64Ptr(1.5),
			wantErr:         true,
		},
		{
			name:            "standard_tax_rate が負の場合はエラー",
			standardTaxRate: clinicFloat64Ptr(-0.1),
			wantErr:         true,
		},
		{
			name:           "reduced_tax_rate が1を超える場合はエラー",
			reducedTaxRate: clinicFloat64Ptr(1.1),
			wantErr:        true,
		},
		{
			name:           "reduced_tax_rate が負の場合はエラー",
			reducedTaxRate: clinicFloat64Ptr(-0.01),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := BuildClinicUpdate(&UpdateClinicInput{
				StandardTaxRate: tt.standardTaxRate,
				ReducedTaxRate:  tt.reducedTaxRate,
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
				return
			}
			assert.NoError(t, err)
			if tt.standardTaxRate != nil {
				assert.Equal(t, tt.wantStandardRate, fields["standard_tax_rate"])
			}
			if tt.reducedTaxRate != nil {
				assert.Equal(t, tt.wantReducedRate, fields["reduced_tax_rate"])
			}
		})
	}
}

func TestClinicService_DeleteClinic(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		ownerCount    int64
		staffCount    int64
		countOwnerErr error
		countStaffErr error
		blockingRefs  []ClinicDependencyCount
		blockingErr   error
		lockErr       error
		repoErr       error
		wantErr       bool
		wantNF        bool
		wantConflict  bool
	}{
		{
			name:          "deletes clinic successfully when no dependencies exist",
			id:            1,
			ownerCount:    0,
			staffCount:    0,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       false,
		},
		{
			name:          "returns conflict error when clinic has owners",
			id:            1,
			ownerCount:    5,
			staffCount:    0,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns conflict error when clinic has staff",
			id:            1,
			ownerCount:    0,
			staffCount:    3,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns conflict error when clinic has both owners and staff",
			id:            1,
			ownerCount:    5,
			staffCount:    3,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns error when owner count check fails",
			id:            1,
			ownerCount:    0,
			staffCount:    0,
			countOwnerErr: errors.New("db error"),
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
		},
		{
			name:          "returns error when staff count check fails",
			id:            1,
			ownerCount:    0,
			staffCount:    0,
			countOwnerErr: nil,
			countStaffErr: errors.New("db error"),
			repoErr:       nil,
			wantErr:       true,
		},
		{
			name:         "returns conflict error when clinic has other dependencies",
			id:           1,
			ownerCount:   0,
			staffCount:   0,
			blockingRefs: []ClinicDependencyCount{{Label: "予約", Count: 2}},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:        "returns error when dependency check fails",
			id:          1,
			ownerCount:  0,
			staffCount:  0,
			blockingErr: errors.New("db error"),
			wantErr:     true,
		},
		{
			name:          "returns not found error when clinic does not exist",
			id:            999,
			ownerCount:    0,
			staffCount:    0,
			countOwnerErr: nil,
			countStaffErr: nil,
			lockErr:       apperrors.WrapNotFound("clinic", "999"),
			wantErr:       true,
			wantNF:        true,
		},
		// FK cleanup regression tests: soft-deleted PGs must not block clinic hard-delete
		{
			// CountBlockingReferencesByClinicID uses "deleted_at IS NULL" for permission_groups,
			// so soft-deleted PGs return count=0 → service allows deletion → the permission-group
			// write owner cleans them up in the clinic delete transaction.
			name:         "succeeds when only soft-deleted permission_groups remain",
			id:           1,
			ownerCount:   0,
			staffCount:   0,
			blockingRefs: nil, // soft-deleted PGs are not counted (deleted_at IS NULL filter)
			repoErr:      nil,
			wantErr:      false,
		},
		{
			// Active (non-soft-deleted) PGs must still produce 409 via blockingRefs path.
			name:         "returns conflict when active permission_groups exist",
			id:           1,
			ownerCount:   0,
			staffCount:   0,
			blockingRefs: []ClinicDependencyCount{{Label: "権限グループ", Count: 1}},
			repoErr:      nil,
			wantErr:      true,
			wantConflict: true,
		},
		{
			// The FK 23503 path due only to soft-deleted PG rows must not occur because the
			// permission-group cleanup runs before clinic Delete in the same transaction.
			// If it did, the error would propagate as-is (not swallowed).
			name:         "propagates unexpected repo error without masking",
			id:           1,
			ownerCount:   0,
			staffCount:   0,
			blockingRefs: nil,
			repoErr:      errors.New("unexpected db error"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				lockForUpdateFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
					if tt.lockErr != nil {
						return nil, tt.lockErr
					}
					return &model.Clinic{ID: id, IsActive: true}, nil
				},
				countOwnersByClinicIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.ownerCount, tt.countOwnerErr
				},
				countStaffByClinicIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.staffCount, tt.countStaffErr
				},
				countBlockingRefsFn: func(_ context.Context, _ uint64) ([]ClinicDependencyCount, error) {
					return tt.blockingRefs, tt.blockingErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			pgRepo := &mockPermissionGroupRepository{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{})

			err := svc.DeleteClinic(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type clinicDeleteTxMarkerKey struct{}

func TestClinicService_DeleteClinic_CleanupAndDeleteTransaction(t *testing.T) {
	const clinicID = uint64(41)

	newRepository := func(deleteFn func(context.Context, uint64) error) *mockClinicRepository {
		return &mockClinicRepository{
			lockForUpdateFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
				return &model.Clinic{ID: id}, nil
			},
			deleteFn: deleteFn,
		}
	}

	t.Run("cleanup and clinic delete run in order inside the same transaction context", func(t *testing.T) {
		txMarker := &struct{}{}
		calls := make([]string, 0, 6)
		repo := newRepository(func(ctx context.Context, id uint64) error {
			assert.Equal(t, txMarker, ctx.Value(clinicDeleteTxMarkerKey{}))
			assert.Equal(t, clinicID, id)
			calls = append(calls, "delete-clinic")
			return nil
		})
		repo.lockForUpdateFn = func(ctx context.Context, id uint64) (*model.Clinic, error) {
			assert.Equal(t, txMarker, ctx.Value(clinicDeleteTxMarkerKey{}))
			assert.Equal(t, clinicID, id)
			calls = append(calls, "lock-clinic")
			return &model.Clinic{ID: id}, nil
		}
		repo.countOwnersByClinicIDFn = func(ctx context.Context, id uint64) (int64, error) {
			assert.Equal(t, txMarker, ctx.Value(clinicDeleteTxMarkerKey{}))
			assert.Equal(t, clinicID, id)
			calls = append(calls, "count-owners")
			return 0, nil
		}
		repo.countStaffByClinicIDFn = func(ctx context.Context, id uint64) (int64, error) {
			assert.Equal(t, txMarker, ctx.Value(clinicDeleteTxMarkerKey{}))
			assert.Equal(t, clinicID, id)
			calls = append(calls, "count-staff")
			return 0, nil
		}
		repo.countBlockingRefsFn = func(ctx context.Context, id uint64) ([]ClinicDependencyCount, error) {
			assert.Equal(t, txMarker, ctx.Value(clinicDeleteTxMarkerKey{}))
			assert.Equal(t, clinicID, id)
			calls = append(calls, "count-blocking-references")
			return nil, nil
		}
		pgRepo := &mockClinicPermissionGroupWriter{
			mockPermissionGroupRepository: &mockPermissionGroupRepository{},
			deleteSoftDeletedByClinicIDFn: func(ctx context.Context, id uint64) error {
				assert.Equal(t, txMarker, ctx.Value(clinicDeleteTxMarkerKey{}))
				assert.Equal(t, clinicID, id)
				calls = append(calls, "cleanup-permission-groups")
				return nil
			},
		}
		tx := &clinicServiceTransactorDouble{
			withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				txCtx := context.WithValue(ctx, clinicDeleteTxMarkerKey{}, txMarker)
				return fn(txCtx)
			},
		}

		err := NewClinicService(repo, pgRepo, tx).DeleteClinic(context.Background(), clinicID)

		require.NoError(t, err)
		assert.Equal(t, []string{
			"lock-clinic",
			"count-owners",
			"count-staff",
			"count-blocking-references",
			"cleanup-permission-groups",
			"delete-clinic",
		}, calls)
	})

	t.Run("cleanup failure prevents clinic delete and preserves the cause", func(t *testing.T) {
		cleanupErr := errors.New("permission group cleanup failed")
		deleteCalled := false
		repo := newRepository(func(_ context.Context, _ uint64) error {
			deleteCalled = true
			return nil
		})
		pgRepo := &mockClinicPermissionGroupWriter{
			mockPermissionGroupRepository: &mockPermissionGroupRepository{},
			deleteSoftDeletedByClinicIDFn: func(_ context.Context, id uint64) error {
				assert.Equal(t, clinicID, id)
				return cleanupErr
			},
		}

		err := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{}).
			DeleteClinic(context.Background(), clinicID)

		require.ErrorIs(t, err, cleanupErr)
		assert.False(t, deleteCalled)
	})

	t.Run("clinic delete failure after cleanup preserves the cause", func(t *testing.T) {
		deleteErr := errors.New("clinic delete failed")
		cleanupCalled := false
		repo := newRepository(func(_ context.Context, _ uint64) error {
			return deleteErr
		})
		pgRepo := &mockClinicPermissionGroupWriter{
			mockPermissionGroupRepository: &mockPermissionGroupRepository{},
			deleteSoftDeletedByClinicIDFn: func(_ context.Context, _ uint64) error {
				cleanupCalled = true
				return nil
			},
		}

		err := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{}).
			DeleteClinic(context.Background(), clinicID)

		require.ErrorIs(t, err, deleteErr)
		assert.True(t, cleanupCalled)
	})

	t.Run("transaction start failure runs neither cleanup nor clinic delete", func(t *testing.T) {
		txErr := errors.New("transaction unavailable")
		cleanupCalled := false
		deleteCalled := false
		repo := newRepository(func(_ context.Context, _ uint64) error {
			deleteCalled = true
			return nil
		})
		pgRepo := &mockClinicPermissionGroupWriter{
			mockPermissionGroupRepository: &mockPermissionGroupRepository{},
			deleteSoftDeletedByClinicIDFn: func(_ context.Context, _ uint64) error {
				cleanupCalled = true
				return nil
			},
		}

		err := NewClinicService(repo, pgRepo, &clinicServiceTransactorDouble{withTxErr: txErr}).
			DeleteClinic(context.Background(), clinicID)

		require.ErrorIs(t, err, txErr)
		assert.False(t, cleanupCalled)
		assert.False(t, deleteCalled)
	})

	t.Run("active permission groups retain conflict semantics inside transaction", func(t *testing.T) {
		repo := newRepository(func(_ context.Context, _ uint64) error {
			t.Fatal("clinic delete must not run for an active permission group")
			return nil
		})
		repo.countBlockingRefsFn = func(_ context.Context, _ uint64) ([]ClinicDependencyCount, error) {
			return []ClinicDependencyCount{{Label: "権限グループ", Count: 1}}, nil
		}
		pgRepo := &mockClinicPermissionGroupWriter{
			mockPermissionGroupRepository: &mockPermissionGroupRepository{},
			deleteSoftDeletedByClinicIDFn: func(_ context.Context, _ uint64) error {
				t.Fatal("cleanup must not run for an active permission group")
				return nil
			},
		}
		txCalled := false
		tx := &clinicServiceTransactorDouble{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			txCalled = true
			return fn(ctx)
		}}

		err := NewClinicService(repo, pgRepo, tx).DeleteClinic(context.Background(), clinicID)

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.True(t, txCalled)
	})
}
