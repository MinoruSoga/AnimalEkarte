package medicalrecord

// checkup_field_service_test.go — EMR-225 フィールド定義 write サービスのテスト。
//   ユニット部: mock repo + passthrough transactor で検証ロジックを固定
//   （不正 field_type / 選択式の空 options / min>max / options 重複 / nil input / nil transactor）。
//   DB 部: setupCheckupFieldTestDB + testTransactor で実 SQL を通し、A2/A3/A4 の正本を固定
//   （field_type 変更時も checkup_field_results のスナップショット列は不変、
//   soft delete で定義行・結果行が残り一覧から除外、reorder が複合スコープで原子的）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mocks ----

// mockCheckupTypeFieldWriteRepo は write 系メソッドを fn で差し替えるモック。
// 未使用メソッド（FindByCheckupTypeID など）は埋め込み interface に委譲して nil 参照を避ける。
type mockCheckupTypeFieldWriteRepo struct {
	CheckupTypeFieldRepository
	createFieldFn   func(ctx context.Context, field *model.CheckupTypeField) error
	lockFieldByIDFn func(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) (*model.CheckupTypeField, error)
	updateFieldFn   func(ctx context.Context, clinicID, checkupTypeID, fieldID uint64, fields map[string]any) (*model.CheckupTypeField, error)
	deleteFieldFn   func(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) error
	reorderFieldsFn func(ctx context.Context, clinicID, checkupTypeID uint64, ids []uint64) error
}

func (m *mockCheckupTypeFieldWriteRepo) CreateField(ctx context.Context, field *model.CheckupTypeField) error {
	return m.createFieldFn(ctx, field)
}

func (m *mockCheckupTypeFieldWriteRepo) LockFieldByID(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) (*model.CheckupTypeField, error) {
	return m.lockFieldByIDFn(ctx, clinicID, checkupTypeID, fieldID)
}

func (m *mockCheckupTypeFieldWriteRepo) UpdateField(ctx context.Context, clinicID, checkupTypeID, fieldID uint64, fields map[string]any) (*model.CheckupTypeField, error) {
	return m.updateFieldFn(ctx, clinicID, checkupTypeID, fieldID, fields)
}

func (m *mockCheckupTypeFieldWriteRepo) DeleteField(ctx context.Context, clinicID, checkupTypeID, fieldID uint64) error {
	return m.deleteFieldFn(ctx, clinicID, checkupTypeID, fieldID)
}

func (m *mockCheckupTypeFieldWriteRepo) ReorderFields(ctx context.Context, clinicID, checkupTypeID uint64, ids []uint64) error {
	return m.reorderFieldsFn(ctx, clinicID, checkupTypeID, ids)
}

// mockCheckupTypeFinder は親 checkup_type 存在確認の最小 view モック。
type mockCheckupTypeFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
}

func (m *mockCheckupTypeFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func existingCheckupTypeFinder() *mockCheckupTypeFinder {
	return &mockCheckupTypeFinder{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.CheckupType, error) {
			return &model.CheckupType{ID: id}, nil
		},
	}
}

func newCheckupTypeFieldServiceForTest(
	fieldRepo CheckupTypeFieldRepository,
	checkupTypes checkupTypeFinder,
) CheckupTypeFieldService {
	return NewCheckupTypeFieldService(fieldRepo, checkupTypes, passthroughExamTypeTransactor{})
}

// ---- CreateField ----

func TestCheckupTypeFieldService_CreateField(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       *CreateCheckupTypeFieldInput
		setupRepo   func(t *testing.T) *mockCheckupTypeFieldWriteRepo
		setupFinder func(t *testing.T) *mockCheckupTypeFinder
		wantErr     string
		wantErrIs   func(error) bool
	}{
		{
			name: "creates a number field with unit and bounds",
			input: &CreateCheckupTypeFieldInput{
				Name: "体重", FieldType: "number", Unit: "kg",
				MinValue: float64Ptr(1), MaxValue: float64Ptr(80), SortOrder: 1,
			},
			setupRepo: func(t *testing.T) *mockCheckupTypeFieldWriteRepo {
				return &mockCheckupTypeFieldWriteRepo{
					createFieldFn: func(_ context.Context, field *model.CheckupTypeField) error {
						assert.Equal(t, uint64(1), field.ClinicID)
						assert.Equal(t, uint64(7), field.CheckupTypeID)
						assert.Equal(t, "体重", field.Name)
						assert.Equal(t, model.CheckupFieldTypeNumber, field.FieldType)
						assert.Equal(t, "kg", field.Unit)
						assert.Equal(t, 1, field.SortOrder)
						field.ID = 42
						return nil
					},
				}
			},
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
		},
		{
			name: "creates a select field with options",
			input: &CreateCheckupTypeFieldInput{
				Name: "総合評価", FieldType: "single_select",
				Options: []CheckupFieldOptionInput{{Value: "a", Label: "良好"}, {Value: "b", Label: "要注意"}},
			},
			setupRepo: func(t *testing.T) *mockCheckupTypeFieldWriteRepo {
				return &mockCheckupTypeFieldWriteRepo{
					createFieldFn: func(_ context.Context, field *model.CheckupTypeField) error {
						assert.JSONEq(t, `[{"value":"a","label":"良好"},{"value":"b","label":"要注意"}]`, string(field.Options))
						return nil
					},
				}
			},
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
		},
		{
			name:        "rejects nil input",
			input:       nil,
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErrIs:   apperrors.IsInvalidInput,
		},
		{
			name:        "rejects unsupported field_type",
			input:       &CreateCheckupTypeFieldInput{Name: "x", FieldType: "date"},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "field_type",
		},
		{
			name:        "rejects empty name",
			input:       &CreateCheckupTypeFieldInput{Name: "  ", FieldType: "number"},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "",
			wantErrIs:   apperrors.IsInvalidInput,
		},
		{
			name: "rejects select field type with empty options",
			input: &CreateCheckupTypeFieldInput{
				Name: "総合評価", FieldType: "single_select",
			},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "options",
		},
		{
			name: "rejects checklist field type with empty options",
			input: &CreateCheckupTypeFieldInput{
				Name: "処置", FieldType: "checklist", Options: []CheckupFieldOptionInput{},
			},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "options",
		},
		{
			name: "rejects duplicate option value",
			input: &CreateCheckupTypeFieldInput{
				Name: "総合評価", FieldType: "multi_select",
				Options: []CheckupFieldOptionInput{{Value: "a", Label: "A"}, {Value: "a", Label: "B"}},
			},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "一意",
		},
		{
			name: "rejects empty option value",
			input: &CreateCheckupTypeFieldInput{
				Name: "総合評価", FieldType: "single_select",
				Options: []CheckupFieldOptionInput{{Value: " ", Label: "A"}},
			},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "value",
		},
		{
			name: "rejects empty option label",
			input: &CreateCheckupTypeFieldInput{
				Name: "総合評価", FieldType: "single_select",
				Options: []CheckupFieldOptionInput{{Value: "a", Label: ""}},
			},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "label",
		},
		{
			name: "rejects min_value greater than max_value",
			input: &CreateCheckupTypeFieldInput{
				Name: "体重", FieldType: "number", MinValue: float64Ptr(10), MaxValue: float64Ptr(5),
			},
			setupRepo:   func(t *testing.T) *mockCheckupTypeFieldWriteRepo { return &mockCheckupTypeFieldWriteRepo{} },
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder { return existingCheckupTypeFinder() },
			wantErr:     "min_value",
		},
		{
			name:  "returns not found for missing or cross-clinic parent checkup type",
			input: &CreateCheckupTypeFieldInput{Name: "体重", FieldType: "number"},
			setupRepo: func(t *testing.T) *mockCheckupTypeFieldWriteRepo {
				return &mockCheckupTypeFieldWriteRepo{
					createFieldFn: func(_ context.Context, _ *model.CheckupTypeField) error {
						t.Fatal("CreateField must not be reached when the parent lookup fails")
						return nil
					},
				}
			},
			setupFinder: func(t *testing.T) *mockCheckupTypeFinder {
				return &mockCheckupTypeFinder{
					findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
						return nil, apperrors.WrapNotFound("checkup_type", "7")
					},
				}
			},
			wantErrIs: apperrors.IsNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newCheckupTypeFieldServiceForTest(tt.setupRepo(t), tt.setupFinder(t))
			field, err := svc.CreateField(ctx, 1, 7, tt.input)
			if tt.wantErr == "" && tt.wantErrIs == nil {
				require.NoError(t, err)
				require.NotNil(t, field)
				return
			}
			require.Error(t, err)
			if tt.wantErr != "" {
				assert.Contains(t, err.Error(), tt.wantErr)
			}
			if tt.wantErrIs != nil {
				assert.True(t, tt.wantErrIs(err), "unexpected error kind: %v", err)
			}
		})
	}
}

// ---- UpdateField ----

func TestCheckupTypeFieldService_UpdateField(t *testing.T) {
	ctx := context.Background()

	lockedField := func() *model.CheckupTypeField {
		return &model.CheckupTypeField{
			ID: 3, ClinicID: 1, CheckupTypeID: 7,
			Name: "体重", FieldType: model.CheckupFieldTypeNumber,
			MinValue: float64Ptr(1), MaxValue: float64Ptr(80),
		}
	}

	tests := []struct {
		name       string
		input      *UpdateCheckupTypeFieldInput
		existing   *model.CheckupTypeField
		assertCall func(t *testing.T, fields map[string]any)
		wantErr    string
		wantErrIs  func(error) bool
	}{
		{
			name: "updates name and unit",
			input: &UpdateCheckupTypeFieldInput{
				Name: strPtr("体重（朝）"), Unit: strPtr("kg"),
			},
			existing: lockedField(),
			assertCall: func(t *testing.T, fields map[string]any) {
				assert.Equal(t, "体重（朝）", fields["name"])
				assert.Equal(t, "kg", fields["unit"])
				assert.NotContains(t, fields, "field_type")
				assert.NotContains(t, fields, "min_value")
			},
		},
		{
			name:     "allows field_type change to another type",
			input:    &UpdateCheckupTypeFieldInput{FieldType: strPtr("text")},
			existing: lockedField(),
			assertCall: func(t *testing.T, fields map[string]any) {
				assert.Equal(t, model.CheckupFieldTypeText, fields["field_type"])
			},
		},
		{
			name: "clears min/max via clear flags",
			input: &UpdateCheckupTypeFieldInput{
				ClearMinValue: true, ClearMaxValue: true,
			},
			existing: lockedField(),
			assertCall: func(t *testing.T, fields map[string]any) {
				assert.Contains(t, fields, "min_value")
				assert.Nil(t, fields["min_value"])
				assert.Contains(t, fields, "max_value")
				assert.Nil(t, fields["max_value"])
			},
		},
		{
			name:     "rejects unsupported field_type",
			input:    &UpdateCheckupTypeFieldInput{FieldType: strPtr("date")},
			existing: lockedField(),
			wantErr:  "field_type",
		},
		{
			name:     "rejects changing to select type without non-empty options",
			input:    &UpdateCheckupTypeFieldInput{FieldType: strPtr("single_select")},
			existing: lockedField(),
			wantErr:  "options",
		},
		{
			name: "accepts changing to select type when options are provided",
			input: &UpdateCheckupTypeFieldInput{
				FieldType: strPtr("single_select"),
				Options:   &[]CheckupFieldOptionInput{{Value: "a", Label: "良好"}},
			},
			existing: lockedField(),
			assertCall: func(t *testing.T, fields map[string]any) {
				assert.Equal(t, model.CheckupFieldTypeSingleSelect, fields["field_type"])
				assert.JSONEq(t, `[{"value":"a","label":"良好"}]`, string(fields["options"].(datatypes.JSON)))
			},
		},
		{
			name: "rejects clearing options on a select field",
			input: &UpdateCheckupTypeFieldInput{
				Options: &[]CheckupFieldOptionInput{},
			},
			existing: &model.CheckupTypeField{
				ID: 3, ClinicID: 1, CheckupTypeID: 7,
				Name: "総合評価", FieldType: model.CheckupFieldTypeSingleSelect,
				Options: datatypes.JSON([]byte(`[{"value":"a","label":"良好"}]`)),
			},
			wantErr: "options",
		},
		{
			name:     "rejects effective min > max",
			input:    &UpdateCheckupTypeFieldInput{MinValue: float64Ptr(99)},
			existing: lockedField(),
			wantErr:  "min_value",
		},
		{
			name:      "rejects empty update",
			input:     &UpdateCheckupTypeFieldInput{},
			existing:  lockedField(),
			wantErrIs: apperrors.IsInvalidInput,
		},
		{
			name:      "rejects nil input",
			input:     nil,
			existing:  lockedField(),
			wantErrIs: apperrors.IsInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeFieldWriteRepo{
				lockFieldByIDFn: func(_ context.Context, clinicID, checkupTypeID, fieldID uint64) (*model.CheckupTypeField, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), checkupTypeID)
					assert.Equal(t, uint64(3), fieldID)
					return tt.existing, nil
				},
				updateFieldFn: func(_ context.Context, _, _, _ uint64, fields map[string]any) (*model.CheckupTypeField, error) {
					if tt.assertCall != nil {
						tt.assertCall(t, fields)
					}
					return tt.existing, nil
				},
			}
			svc := newCheckupTypeFieldServiceForTest(repo, existingCheckupTypeFinder())
			field, err := svc.UpdateField(ctx, 1, 7, 3, tt.input)
			if tt.wantErr == "" && tt.wantErrIs == nil {
				require.NoError(t, err)
				require.NotNil(t, field)
				return
			}
			require.Error(t, err)
			if tt.wantErr != "" {
				assert.Contains(t, err.Error(), tt.wantErr)
			}
			if tt.wantErrIs != nil {
				assert.True(t, tt.wantErrIs(err), "unexpected error kind: %v", err)
			}
		})
	}
}

// ---- DeleteField ----

func TestCheckupTypeFieldService_DeleteField(t *testing.T) {
	ctx := context.Background()

	t.Run("locks then soft-deletes the field in scope", func(t *testing.T) {
		var locked, deleted bool
		repo := &mockCheckupTypeFieldWriteRepo{
			lockFieldByIDFn: func(_ context.Context, clinicID, checkupTypeID, fieldID uint64) (*model.CheckupTypeField, error) {
				locked = true
				assert.False(t, deleted, "lock must precede delete")
				return &model.CheckupTypeField{ID: fieldID, ClinicID: clinicID, CheckupTypeID: checkupTypeID}, nil
			},
			deleteFieldFn: func(_ context.Context, clinicID, checkupTypeID, fieldID uint64) error {
				deleted = true
				assert.True(t, locked, "delete must happen after lock")
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), checkupTypeID)
				assert.Equal(t, uint64(3), fieldID)
				return nil
			},
		}
		svc := newCheckupTypeFieldServiceForTest(repo, existingCheckupTypeFinder())
		require.NoError(t, svc.DeleteField(ctx, 1, 7, 3))
		assert.True(t, deleted)
	})

	t.Run("propagates not found for missing or cross-clinic field", func(t *testing.T) {
		repo := &mockCheckupTypeFieldWriteRepo{
			lockFieldByIDFn: func(_ context.Context, _, _, _ uint64) (*model.CheckupTypeField, error) {
				return nil, apperrors.WrapNotFound("checkup_type_field", "3")
			},
			deleteFieldFn: func(_ context.Context, _, _, _ uint64) error {
				t.Fatal("DeleteField must not be reached when lock fails")
				return nil
			},
		}
		svc := newCheckupTypeFieldServiceForTest(repo, existingCheckupTypeFinder())
		err := svc.DeleteField(ctx, 1, 7, 3)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// ---- ReorderFields ----

func TestCheckupTypeFieldService_ReorderFields(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		ids       []uint64
		wantErr   string
		wantErrIs func(error) bool
	}{
		{name: "reorders ids", ids: []uint64{3, 1, 2}},
		{name: "rejects empty ids", ids: []uint64{}, wantErrIs: apperrors.IsInvalidInput},
		{name: "rejects nil ids", ids: nil, wantErrIs: apperrors.IsInvalidInput},
		{name: "rejects zero id", ids: []uint64{1, 0}, wantErr: "positive"},
		{name: "rejects duplicate id", ids: []uint64{1, 1, 2}, wantErr: "unique"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupTypeFieldWriteRepo{
				reorderFieldsFn: func(_ context.Context, clinicID, checkupTypeID uint64, ids []uint64) error {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(7), checkupTypeID)
					assert.Equal(t, tt.ids, ids)
					return nil
				},
			}
			svc := newCheckupTypeFieldServiceForTest(repo, existingCheckupTypeFinder())
			err := svc.ReorderFields(ctx, 1, 7, tt.ids)
			if tt.wantErr == "" && tt.wantErrIs == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tt.wantErr != "" {
				assert.Contains(t, err.Error(), tt.wantErr)
			}
			if tt.wantErrIs != nil {
				assert.True(t, tt.wantErrIs(err), "unexpected error kind: %v", err)
			}
		})
	}
}

// nil transactor は composition/wiring 失敗 → 500（MRB-07 と同方針）。
func TestCheckupTypeFieldService_WithTx_NilTransactorIsInternalError(t *testing.T) {
	svc := NewCheckupTypeFieldService(&mockCheckupTypeFieldWriteRepo{}, existingCheckupTypeFinder(), nil)
	_, err := svc.CreateField(context.Background(), 1, 7, &CreateCheckupTypeFieldInput{Name: "x", FieldType: "number"})
	require.Error(t, err)
	assert.False(t, apperrors.IsInvalidInput(err))
	assert.False(t, apperrors.IsNotFound(err))
}

// ---- DB-backed acceptance（A2/A3/A4 の正本） ----

// makeCheckupFieldResultRow はスナップショット列を持つ結果行を直接挿入する。
// service 経由でなく生 INSERT するのは「過去に保存済みの履歴」を再現するため。
func makeCheckupFieldResultRow(t *testing.T, db *gorm.DB, r *model.CheckupFieldResult) *model.CheckupFieldResult {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(r).Error)
	return r
}

// A2: フィールド定義の更新（field_type 変更を含む）が既存 checkup_field_results の
// スナップショット列（field_name/field_type/unit/ref_min/ref_max）を書き換えないこと。
func TestCheckupTypeFieldService_UpdateField_PreservesResultSnapshots(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	svc := NewCheckupTypeFieldService(
		NewCheckupTypeFieldRepository(db),
		NewCheckupTypeRepository(db),
		testTransactor{db: db},
	)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "更新保存飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "更新保存ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CHKSVC-UPD", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診（更新）")

	minV, maxV := 0.0, 4.0
	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石付着度",
		FieldType: model.CheckupFieldTypeNumber, Unit: "段階",
		MinValue: &minV, MaxValue: &maxV, SortOrder: 1,
	})
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)
	result := makeCheckupFieldResultRow(t, db, &model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
		FieldName: field.Name, FieldType: field.FieldType, Unit: field.Unit,
		RefMin: &minV, RefMax: &maxV, ValueNumber: float64Ptr(3),
		Status: model.ExaminationResultStatusNormal, SortOrder: 1,
	})

	newMax := 5.0
	updated, err := svc.UpdateField(ctx, clinicA, ct.ID, field.ID, &UpdateCheckupTypeFieldInput{
		Name:      strPtr("歯石付着度（改訂）"),
		FieldType: strPtr("text"),
		Unit:      strPtr(""),
		MaxValue:  &newMax,
	})
	require.NoError(t, err)
	assert.Equal(t, model.CheckupFieldTypeText, updated.FieldType)

	var historical model.CheckupFieldResult
	require.NoError(t, db.First(&historical, result.ID).Error)
	assert.Equal(t, "歯石付着度", historical.FieldName, "結果のスナップショット名は不変")
	assert.Equal(t, model.CheckupFieldTypeNumber, historical.FieldType, "結果のスナップショット型は不変")
	assert.Equal(t, "段階", historical.Unit)
	require.NotNil(t, historical.RefMin)
	require.NotNil(t, historical.RefMax)
	assert.Equal(t, maxV, *historical.RefMax)
	require.NotNil(t, historical.CheckupTypeFieldID)
	assert.Equal(t, field.ID, *historical.CheckupTypeFieldID, "FK は soft delete でない限り維持される")
}

// A3: soft delete で deleted_at が立ち、一覧から除外されるが定義行と結果行は残る。
func TestCheckupTypeFieldService_DeleteField_SoftDeletePreservesRowAndResults(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	fieldRepo := NewCheckupTypeFieldRepository(db)
	svc := NewCheckupTypeFieldService(fieldRepo, NewCheckupTypeRepository(db), testTransactor{db: db})
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "削除保存飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "削除保存ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CHKSVC-DEL", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診（削除）")
	ctB := makeCheckupTypeMaster(t, db, clinicB, "医院Bの検診")

	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石除去の有無",
		FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1,
	})
	// 別医院の同名フィールド（クロステナント削除拒否の検証対象）。
	fieldB := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicB, CheckupTypeID: ctB.ID, Name: "医院B項目",
		FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1,
	})
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)
	result := makeCheckupFieldResultRow(t, db, &model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
		FieldName: field.Name, FieldType: field.FieldType, ValueBool: boolPtr(true), SortOrder: 1,
	})

	require.NoError(t, svc.DeleteField(ctx, clinicA, ct.ID, field.ID))

	// 一覧（deleted_at IS NULL 経由の GORM soft delete フィルタ）から除外される。
	visible, err := fieldRepo.FindByCheckupTypeID(ctx, clinicA, ct.ID)
	require.NoError(t, err)
	assert.Empty(t, visible, "soft delete 後は一覧から除外される")

	// 定義行は deleted_at 付きで残存する。
	var count int64
	require.NoError(t, db.Model(&model.CheckupTypeField{}).
		Unscoped().
		Where("id = ? AND deleted_at IS NOT NULL", field.ID).
		Count(&count).Error)
	assert.EqualValues(t, 1, count, "定義行は deleted_at 付きで残る")

	// 結果行はスナップショットごと残る。
	var historical model.CheckupFieldResult
	require.NoError(t, db.First(&historical, result.ID).Error)
	assert.Equal(t, "歯石除去の有無", historical.FieldName)
	require.NotNil(t, historical.CheckupTypeFieldID)
	assert.Equal(t, field.ID, *historical.CheckupTypeFieldID)

	// 他医院/別パッケージへの削除は NotFound で拒否される。
	err = svc.DeleteField(ctx, clinicB, ctB.ID, field.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "clinic B が clinic A の field を削除できない")
	var rowB model.CheckupTypeField
	require.NoError(t, db.First(&rowB, fieldB.ID).Error)
	assert.False(t, rowB.DeletedAt.Valid, "別医院のフィールドは無傷")
}

// A4: reorder が sort_order を複合スコープで原子的に振り直す。
func TestCheckupTypeFieldService_ReorderFields_DB(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	fieldRepo := NewCheckupTypeFieldRepository(db)
	svc := NewCheckupTypeFieldService(fieldRepo, NewCheckupTypeRepository(db), testTransactor{db: db})
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診（並替）")
	ctB := makeCheckupTypeMaster(t, db, clinicB, "医院Bの検診（並替）")
	f1 := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "A", FieldType: model.CheckupFieldTypeText, SortOrder: 1})
	f2 := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "B", FieldType: model.CheckupFieldTypeText, SortOrder: 2})
	f3 := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "C", FieldType: model.CheckupFieldTypeText, SortOrder: 3})
	fB := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicB, CheckupTypeID: ctB.ID, Name: "B-側", FieldType: model.CheckupFieldTypeText, SortOrder: 9})

	require.NoError(t, svc.ReorderFields(ctx, clinicA, ct.ID, []uint64{f3.ID, f1.ID, f2.ID}))

	got, err := fieldRepo.FindByCheckupTypeID(ctx, clinicA, ct.ID)
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, []uint64{f3.ID, f1.ID, f2.ID}, []uint64{got[0].ID, got[1].ID, got[2].ID})
	assert.Equal(t, []int{1, 2, 3}, []int{got[0].SortOrder, got[1].SortOrder, got[2].SortOrder})

	// 他医院の id を混ぜると複合スコープで NotFound（部分更新しない）。
	err = svc.ReorderFields(ctx, clinicA, ct.ID, []uint64{f1.ID, fB.ID})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	reloaded, err := fieldRepo.FindByCheckupTypeID(ctx, clinicA, ct.ID)
	require.NoError(t, err)
	assert.Equal(t, []uint64{f3.ID, f1.ID, f2.ID}, []uint64{reloaded[0].ID, reloaded[1].ID, reloaded[2].ID},
		"失敗した reorder は部分適用されない")
}

// 注: strPtr / boolPtr / float64Ptr は既存テストヘルパー
// （medicine_service_test.go / treatment_plan_service_test.go 等、同一パッケージ共有）を再利用する。
