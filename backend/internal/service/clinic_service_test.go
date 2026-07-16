package service

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockClinicRepository は ClinicRepository のテスト用モック実装
type mockClinicRepository struct {
	findAllFn               func(ctx context.Context) ([]model.Clinic, error)
	findByStaffIDFn         func(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	findByIDFn              func(ctx context.Context, id uint64) (*model.Clinic, error)
	getCompanyFn            func(ctx context.Context) (*model.Company, error)
	createFn                func(ctx context.Context, clinic *model.Clinic) error
	updateFn                func(ctx context.Context, id uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, id uint64) error
	countOwnersByClinicIDFn func(ctx context.Context, clinicID uint64) (int64, error)
	countStaffByClinicIDFn  func(ctx context.Context, clinicID uint64) (int64, error)
	countBlockingRefsFn     func(ctx context.Context, clinicID uint64) ([]repository.ClinicDependencyCount, error)
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

func (m *mockClinicRepository) CountBlockingReferencesByClinicID(ctx context.Context, clinicID uint64) ([]repository.ClinicDependencyCount, error) {
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
			svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
			svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
			svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
			svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
		setRulesFn: func(_ context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
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

	svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
				Name:    strPtr("更新後院"),
				Address: strPtr("東京都渋谷区"),
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
			input:       &UpdateClinicInput{Name: strPtr("存在しない院")},
			repoClinic:  nil,
			repoFindErr: apperrors.WrapNotFound("clinic", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:  "returns error on update failure",
			id:    1,
			input: &UpdateClinicInput{Name: strPtr("更新後院")},
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
			svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
	svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
	svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
	svc := NewClinicService(repo, pgRepo, &mockTransactor{})

	result, err := svc.UpdateClinic(context.Background(), 1, &UpdateClinicInput{StandardTaxRate: &invalidRate})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestClinicService_UpdateClinic_RefetchErrorAfterUpdate(t *testing.T) {
	findCalls := 0
	repo := &mockClinicRepository{
		findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			findCalls++
			if findCalls == 1 {
				return &model.Clinic{ID: 1, CompanyID: 5, Name: "旧院名"}, nil
			}
			return nil, errors.New("db error")
		},
		updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	pgRepo := &mockPermissionGroupRepository{}
	svc := NewClinicService(repo, pgRepo, &mockTransactor{})

	result, err := svc.UpdateClinic(context.Background(), 1, &UpdateClinicInput{Name: strPtr("新院名")})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 2, findCalls, "更新後のリフレッシュ取得も呼ばれる")
}

// TestBuildClinicUpdate_AccountingDocumentSettings は #179 follow-up ①（#190）の
// 帳票レイアウト設定が PATCH セマンティクスで更新マップへ反映されることを検証する。
// 指定フィールドは実カラム名付きでマップへ入り、nil フィールドは省略される（既存値保持）。
func TestBuildClinicUpdate_AccountingDocumentSettings(t *testing.T) {
	t.Run("指定した帳票設定が実カラム名付きで更新マップへ入る", func(t *testing.T) {
		fields, err := buildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentShowLogo:                boolPtr(true),
			AccountingDocumentShowRegistrationWarning: boolPtr(false),
			AccountingDocumentShowItemCategory:        boolPtr(false),
			AccountingDocumentFooterNote:              strPtr("ご来院ありがとうございました。"),
		})

		assert.NoError(t, err)
		assert.Equal(t, true, fields["accounting_document_show_logo"])
		assert.Equal(t, false, fields["accounting_document_show_registration_warning"])
		assert.Equal(t, false, fields["accounting_document_show_item_category"])
		assert.Equal(t, "ご来院ありがとうございました。", fields["accounting_document_footer_note"])
	})

	t.Run("空フッター文字列も明示更新として反映される（明示クリア可能）", func(t *testing.T) {
		fields, err := buildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentFooterNote: strPtr(""),
		})

		assert.NoError(t, err)
		got, ok := fields["accounting_document_footer_note"]
		assert.True(t, ok, "空文字でもキーは存在する（フッターを明示的にクリアできる）")
		assert.Equal(t, "", got)
	})

	t.Run("未指定（nil）の帳票設定は更新マップへ入らない（PATCH: 既存値保持）", func(t *testing.T) {
		fields, err := buildClinicUpdate(&UpdateClinicInput{Name: strPtr("更新後院")})

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
		fields, err := buildClinicUpdate(&UpdateClinicInput{
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
		fields, err := buildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentSectionOrder: &order,
		})

		assert.NoError(t, err)
		got, ok := fields["accounting_document_section_order"]
		assert.True(t, ok)
		assert.Equal(t, pq.StringArray(order), got)
	})

	t.Run("#190: 空の順序配列はデフォルト順リセットとして更新マップへ入る", func(t *testing.T) {
		empty := []string{}
		fields, err := buildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentSectionOrder: &empty,
		})

		assert.NoError(t, err)
		got, ok := fields["accounting_document_section_order"]
		assert.True(t, ok, "空配列でもキーは存在する（デフォルト順にリセット）")
		assert.Equal(t, pq.StringArray{}, got)
	})

	t.Run("#190: 未知のセクションキーはエラーになる", func(t *testing.T) {
		invalid := []string{"clinic_header", "unknown_section"}
		_, err := buildClinicUpdate(&UpdateClinicInput{
			AccountingDocumentSectionOrder: &invalid,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "unknown key は WrapInvalidInput を返す")
	})

	t.Run("#190: 重複セクションキーはエラーになる", func(t *testing.T) {
		dup := []string{"clinic_header", "items_table", "clinic_header"}
		_, err := buildClinicUpdate(&UpdateClinicInput{
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
			standardTaxRate:  ptrFloat64(0.10),
			reducedTaxRate:   ptrFloat64(0.08),
			wantErr:          false,
			wantStandardRate: 0.10,
			wantReducedRate:  0.08,
		},
		{
			name:            "standard_tax_rate が1を超える場合はエラー",
			standardTaxRate: ptrFloat64(1.5),
			wantErr:         true,
		},
		{
			name:            "standard_tax_rate が負の場合はエラー",
			standardTaxRate: ptrFloat64(-0.1),
			wantErr:         true,
		},
		{
			name:           "reduced_tax_rate が1を超える場合はエラー",
			reducedTaxRate: ptrFloat64(1.1),
			wantErr:        true,
		},
		{
			name:           "reduced_tax_rate が負の場合はエラー",
			reducedTaxRate: ptrFloat64(-0.01),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := buildClinicUpdate(&UpdateClinicInput{
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
		blockingRefs  []repository.ClinicDependencyCount
		blockingErr   error
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
			blockingRefs: []repository.ClinicDependencyCount{{Label: "予約", Count: 2}},
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
			repoErr:       apperrors.WrapNotFound("clinic", "999"),
			wantErr:       true,
			wantNF:        true,
		},
		// FK cleanup regression tests: soft-deleted PGs must not block clinic hard-delete
		{
			// CountBlockingReferencesByClinicID uses "deleted_at IS NULL" for permission_groups,
			// so soft-deleted PGs return count=0 → service allows deletion → repo cleans them up.
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
			blockingRefs: []repository.ClinicDependencyCount{{Label: "権限グループ", Count: 1}},
			repoErr:      nil,
			wantErr:      true,
			wantConflict: true,
		},
		{
			// The FK 23503 path (repo returning 400 due to soft-deleted PG rows) must not occur.
			// repo.Delete now purges soft-deleted PGs in a transaction, so it never reaches this.
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
				countOwnersByClinicIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.ownerCount, tt.countOwnerErr
				},
				countStaffByClinicIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.staffCount, tt.countStaffErr
				},
				countBlockingRefsFn: func(_ context.Context, _ uint64) ([]repository.ClinicDependencyCount, error) {
					return tt.blockingRefs, tt.blockingErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			pgRepo := &mockPermissionGroupRepository{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo, &mockTransactor{})

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
