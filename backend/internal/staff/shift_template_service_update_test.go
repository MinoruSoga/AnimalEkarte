package staff

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func sampleShiftTemplate(id uint64, name string) *model.ShiftTemplate {
	start, end := "09:00:00", "13:00:00"
	return &model.ShiftTemplate{
		ID:        id,
		Name:      name,
		ShiftType: model.ShiftTypeMorning,
		StartTime: &start,
		EndTime:   &end,
	}
}

func TestShiftTemplateService_Update(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		input   *UpdateShiftTemplateInput
		setupFn func(repo *mockShiftTemplateRepository)
		wantErr bool
	}{
		{
			name:  "正常: 全フィールドnil かつ Breaks も nil → repo.FindByID を呼んでそのまま返す (no-op)",
			id:    1,
			input: &UpdateShiftTemplateInput{},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.findByIDFn = func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					return sampleShiftTemplate(id, "早番"), nil
				}
			},
			wantErr: false,
		},
		{
			name: "正常: Name を更新 → repo.Update が呼ばれる",
			id:   1,
			input: &UpdateShiftTemplateInput{
				Name: strPtr("新しい早番"),
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.updateFn = func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.ShiftTemplate, error) {
					assert.Equal(t, "新しい早番", fields["name"])
					return sampleShiftTemplate(id, "新しい早番"), nil
				}
			},
			wantErr: false,
		},
		{
			name:  "エラー: FindByID が not found → error を返す",
			id:    999,
			input: &UpdateShiftTemplateInput{Name: strPtr("x")},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.ShiftTemplate, error) {
					return nil, apperrors.WrapNotFound("shift_template", "999")
				}
			},
			wantErr: true,
		},
		{
			name: "エラー: repo.Update がエラー → error を返す",
			id:   1,
			input: &UpdateShiftTemplateInput{
				Name: strPtr("新しい早番"),
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.updateFn = func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ShiftTemplate, error) {
					return nil, errors.New("update failed")
				}
			},
			wantErr: true,
		},
		{
			name: "正常: Breaks のみ更新 → repo.UpdateBreaks + FindByID が呼ばれる",
			id:   1,
			input: &UpdateShiftTemplateInput{
				Breaks: &[]ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}},
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.updateBreaksFn = func(_ context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error {
					assert.Equal(t, uint64(1), templateID)
					assert.Len(t, breaks, 1)
					return nil
				}
				repo.findByIDFn = func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					return sampleShiftTemplate(id, "早番"), nil
				}
			},
			wantErr: false,
		},
		{
			name: "エラー: repo.UpdateBreaks がエラー → error を返す",
			id:   1,
			input: &UpdateShiftTemplateInput{
				Breaks: &[]ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}},
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.updateBreaksFn = func(_ context.Context, _ uint64, _ []model.ShiftTemplateBreak) error {
					return errors.New("break save failed")
				}
			},
			wantErr: true,
		},
		{
			name: "エラー: Breaks更新後の FindByID がエラー → error を返す",
			id:   1,
			input: &UpdateShiftTemplateInput{
				Breaks: &[]ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}},
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				callCount := 0
				repo.findByIDFn = func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					callCount++
					if callCount == 1 {
						// 存在チェック用の初回呼び出しは成功させる
						return sampleShiftTemplate(id, "早番"), nil
					}
					return nil, errors.New("reload failed")
				}
			},
			wantErr: true,
		},
		{
			name: "正常: Name と Breaks を同時更新 → 両方が反映される",
			id:   1,
			input: &UpdateShiftTemplateInput{
				Name:   strPtr("複合更新"),
				Breaks: &[]ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}},
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.updateFn = func(_ context.Context, _, id uint64, fields map[string]any) (*model.ShiftTemplate, error) {
					assert.Equal(t, "複合更新", fields["name"])
					return sampleShiftTemplate(id, "複合更新"), nil
				}
				repo.updateBreaksFn = func(_ context.Context, _ uint64, breaks []model.ShiftTemplateBreak) error {
					assert.Len(t, breaks, 1)
					return nil
				}
				repo.findByIDFn = func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					return sampleShiftTemplate(id, "複合更新"), nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftTemplateRepository{}
			tt.setupFn(repo)
			svc := NewShiftTemplateService(repo)

			got, err := svc.Update(context.Background(), 10, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// ---- Tests: Delete ----

func TestShiftTemplateService_Delete(t *testing.T) {
	tests := []struct {
		name             string
		id               uint64
		findByIDErr      error
		countUsageResult int64
		countUsageErr    error
		repoErr          error
		wantErr          bool
		errIs            error
	}{
		{
			name:             "正常: 使用中なし → repo.Delete が呼ばれる",
			id:               1,
			countUsageResult: 0,
			repoErr:          nil,
			wantErr:          false,
		},
		{
			name:        "エラー: FindByID が not found → error を返す",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("shift_template", "999"),
			wantErr:     true,
		},
		{
			name:          "正常: owned child usage count is not consulted",
			id:            1,
			countUsageErr: errors.New("db error"),
			wantErr:       false,
		},
		{
			name:             "正常: owned break children do not block deletion",
			id:               1,
			countUsageResult: 1,
			wantErr:          false,
		},
		{
			name:             "エラー: repo.Delete がエラー → error を返す",
			id:               99,
			countUsageResult: 0,
			repoErr:          errors.New("not found in db"),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftTemplateRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return sampleShiftTemplate(id, "早番"), nil
				},
				countUsageByTemplateID: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.countUsageResult, tt.countUsageErr
				},
				deleteFn: func(_ context.Context, clinicID, id uint64) error {
					return tt.repoErr
				},
			}
			svc := NewShiftTemplateService(repo)

			err := svc.Delete(context.Background(), 10, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errIs != nil {
					assert.True(t, errors.Is(err, tt.errIs), "expected error to wrap %v, got %v", tt.errIs, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShiftTemplateService_Create_RollsBackParentWhenBreakReplacementFails(t *testing.T) {
	parentCreated := false
	repo := &mockShiftTemplateRepository{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			before := parentCreated
			if err := fn(ctx); err != nil {
				parentCreated = before
				return err
			}
			return nil
		},
		createFn: func(_ context.Context, tpl *model.ShiftTemplate) error {
			parentCreated = true
			tpl.ID = 42
			return nil
		},
		updateBreaksFn: func(_ context.Context, _ uint64, _ []model.ShiftTemplateBreak) error {
			return errors.New("replace breaks failed")
		},
	}
	svc := NewShiftTemplateService(repo)

	created, err := svc.Create(context.Background(), 10, &CreateShiftTemplateInput{
		Name:      "早番",
		ShiftType: string(model.ShiftTypeOff),
		Breaks: []ShiftBreakTemplateInput{
			{BreakStart: "12:00:00", BreakEnd: "13:00:00"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, created)
	assert.False(t, parentCreated)
	assert.Equal(t, 1, repo.withTxCalls)
}

func TestShiftTemplateService_Update_ValidatesEffectiveTimesBeforeWrite(t *testing.T) {
	start := "18:00"
	updateCalled := false
	repo := &mockShiftTemplateRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
			return &model.ShiftTemplate{
				ID:        id,
				ClinicID:  clinicID,
				ShiftType: model.ShiftTypeFull,
				StartTime: strPtr("09:00:00"),
				EndTime:   strPtr("17:00:00"),
			}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ShiftTemplate, error) {
			updateCalled = true
			return nil, nil
		},
	}
	svc := NewShiftTemplateService(repo)

	updated, err := svc.Update(context.Background(), 10, 1, &UpdateShiftTemplateInput{StartTime: &start})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, updated)
	assert.False(t, updateCalled)
}

func TestShiftTemplateService_Update_RollsBackParentWhenBreakReplacementFails(t *testing.T) {
	name := "更新後"
	storedName := "更新前"
	repo := &mockShiftTemplateRepository{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			before := storedName
			if err := fn(ctx); err != nil {
				storedName = before
				return err
			}
			return nil
		},
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
			return &model.ShiftTemplate{
				ID:        id,
				ClinicID:  clinicID,
				Name:      storedName,
				ShiftType: model.ShiftTypeOff,
			}, nil
		},
		updateFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.ShiftTemplate, error) {
			storedName = fields[colShiftTemplateName].(string)
			return &model.ShiftTemplate{ID: id, ClinicID: clinicID, Name: storedName, ShiftType: model.ShiftTypeOff}, nil
		},
		updateBreaksFn: func(_ context.Context, _ uint64, _ []model.ShiftTemplateBreak) error {
			return errors.New("replace breaks failed")
		},
	}
	svc := NewShiftTemplateService(repo)
	breaks := []ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}}

	updated, err := svc.Update(context.Background(), 10, 1, &UpdateShiftTemplateInput{
		Name:   &name,
		Breaks: &breaks,
	})

	require.Error(t, err)
	assert.Nil(t, updated)
	assert.Equal(t, "更新前", storedName)
	assert.Equal(t, 1, repo.withTxCalls)
}

func TestShiftTemplateService_Delete_AllowsOwnedBreakChildren(t *testing.T) {
	deleteCalled := false
	repo := &mockShiftTemplateRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
			return &model.ShiftTemplate{
				ID:       id,
				ClinicID: clinicID,
				Breaks: []model.ShiftTemplateBreak{
					{ID: 9, ShiftTemplateID: id},
				},
			}, nil
		},
		countUsageByTemplateID: func(_ context.Context, _, _ uint64) (int64, error) {
			return 1, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	svc := NewShiftTemplateService(repo)

	err := svc.Delete(context.Background(), 10, 1)

	require.NoError(t, err)
	assert.True(t, deleteCalled)
}

// ---- Tests: Reorder ----

func TestShiftTemplateService_Reorder(t *testing.T) {
	tests := []struct {
		name          string
		ids           []uint64
		repoErr       error
		wantErr       bool
		errIs         error
		expectReorder bool
	}{
		{
			name:          "エラー: ids が空 → InvalidInput エラー",
			ids:           []uint64{},
			repoErr:       nil,
			wantErr:       true,
			errIs:         apperrors.ErrInvalidInput,
			expectReorder: false,
		},
		{
			name:          "正常: ids があり → repo.Reorder が呼ばれる",
			ids:           []uint64{3, 1, 2},
			repoErr:       nil,
			wantErr:       false,
			expectReorder: true,
		},
		{
			name:          "エラー: repo.Reorder がエラー → error を返す",
			ids:           []uint64{1, 2},
			repoErr:       errors.New("db error"),
			wantErr:       true,
			expectReorder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reorderCalled := false
			repo := &mockShiftTemplateRepository{
				reorderFn: func(_ context.Context, _ uint64, ids []uint64) error {
					reorderCalled = true
					assert.Equal(t, tt.ids, ids)
					return tt.repoErr
				},
			}
			svc := NewShiftTemplateService(repo)

			err := svc.Reorder(context.Background(), 10, tt.ids)

			assert.Equal(t, tt.expectReorder, reorderCalled, "repo.Reorder called mismatch")
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errIs != nil {
					assert.True(t, errors.Is(err, tt.errIs), "expected error to wrap %v, got %v", tt.errIs, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// boolPtr はテスト用のヘルパー関数
func boolPtr(b bool) *bool { return &b }

// shiftTypePtr はテスト用のヘルパー関数
func shiftTypePtr(s model.ShiftType) *model.ShiftType { return &s }
