package staff

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mockShiftTemplateRepository ----

type mockShiftTemplateRepository struct {
	findAllFn       func(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error)
	findByIDFn      func(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	lockForUpdateFn func(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	createFn        func(ctx context.Context, tpl *model.ShiftTemplate) error
	updateFn        func(ctx context.Context, clinicID, id uint64, cmd UpdateShiftTemplateInput) (*model.ShiftTemplate, error)
	deleteFn        func(ctx context.Context, clinicID, id uint64) error
	updateBreaksFn  func(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error
	reorderFn       func(ctx context.Context, clinicID uint64, ids []uint64) error
	// countUsageByTemplateID is retained only so older Delete table fixtures in
	// shift_template_service_update_test.go still compile; the repository method is gone (AUS-04).
	countUsageByTemplateID func(ctx context.Context, clinicID, id uint64) (int64, error)
	withTxFn               func(ctx context.Context, fn func(context.Context) error) error
	withTxCalls            int
}

func (m *mockShiftTemplateRepository) WithTx(ctx context.Context, fn func(context.Context) error) error {
	m.withTxCalls++
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

func (m *mockShiftTemplateRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.ShiftTemplate{}, nil
}

func (m *mockShiftTemplateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	start, end := "09:00:00", "13:00:00"
	return &model.ShiftTemplate{ID: id, ClinicID: clinicID, ShiftType: model.ShiftTypeMorning, StartTime: &start, EndTime: &end}, nil
}

func (m *mockShiftTemplateRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.ShiftTemplate, error) {
	if m.lockForUpdateFn != nil {
		return m.lockForUpdateFn(ctx, clinicID, id)
	}
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockShiftTemplateRepository) Create(ctx context.Context, tpl *model.ShiftTemplate) error {
	if m.createFn != nil {
		return m.createFn(ctx, tpl)
	}
	return nil
}

func (m *mockShiftTemplateRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateShiftTemplateInput) (*model.ShiftTemplate, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, cmd)
	}
	start, end := "09:00:00", "13:00:00"
	return &model.ShiftTemplate{ID: id, ClinicID: clinicID, ShiftType: model.ShiftTypeMorning, StartTime: &start, EndTime: &end}, nil
}

func (m *mockShiftTemplateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockShiftTemplateRepository) UpdateBreaks(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error {
	if m.updateBreaksFn != nil {
		return m.updateBreaksFn(ctx, templateID, breaks)
	}
	return nil
}

func (m *mockShiftTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- Tests: GetByID ----

func TestShiftTemplateService_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		findByIDErr error
		wantErr     bool
	}{
		{
			name:    "正常: リポジトリが item を返す → 同じ item を返す",
			wantErr: false,
		},
		{
			name:        "エラー: リポジトリがエラー → error を返す",
			findByIDErr: errors.New("not found"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftTemplateRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.ShiftTemplate{ID: id, ClinicID: clinicID, Name: "早番"}, nil
				},
			}
			svc := NewShiftTemplateService(repo)

			got, err := svc.GetByID(context.Background(), 10, 1)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "早番", got.Name)
			}
		})
	}
}

// ---- Tests: buildShiftTemplateUpdate ----

func TestBuildShiftTemplateUpdate(t *testing.T) {
	tests := []struct {
		name  string
		input *UpdateShiftTemplateInput
		want  map[string]any
	}{
		{
			name:  "全フィールド nil → 空マップ",
			input: &UpdateShiftTemplateInput{},
			want:  map[string]any{},
		},
		{
			name: "Name のみ設定",
			input: &UpdateShiftTemplateInput{
				Name: strPtr("新しい早番"),
			},
			want: map[string]any{colShiftTemplateName: "新しい早番"},
		},
		{
			name: "全フィールド設定",
			input: &UpdateShiftTemplateInput{
				Name:      strPtr("遅番"),
				ShiftType: shiftTypePtr(model.ShiftTypeAfternoon),
				StartTime: strPtr("13:00:00"),
				EndTime:   strPtr("18:00"),
				Notes:     strPtr("メモ"),
				SortOrder: shiftTemplateIntPtr(2),
				IsActive:  boolPtr(false),
			},
			want: map[string]any{
				colShiftTemplateName:      "遅番",
				colShiftTemplateShiftType: model.ShiftTypeAfternoon,
				colShiftTemplateStartTime: normalizeTimeString(strPtr("13:00:00")),
				colShiftTemplateEndTime:   normalizeTimeString(strPtr("18:00")),
				colShiftTemplateNotes:     "メモ",
				colShiftTemplateSortOrder: 2,
				colShiftTemplateIsActive:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildShiftTemplateUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func shiftTemplateIntPtr(value int) *int { return &value }

// ---- Tests: List ----

func TestShiftTemplateService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.ShiftTemplate
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "正常: リポジトリが items を返す → 同じ items を返す",
			repoData: []model.ShiftTemplate{
				{ID: 1, ClinicID: 10, Name: "早番", ShiftType: model.ShiftTypeMorning},
				{ID: 2, ClinicID: 10, Name: "遅番", ShiftType: model.ShiftTypeAfternoon},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "エラー: リポジトリがエラー → error を返す",
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftTemplateRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ShiftTemplate, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewShiftTemplateService(repo)

			got, err := svc.List(context.Background(), 10)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

// ---- Tests: Create ----

func TestShiftTemplateService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateShiftTemplateInput
		setupFn func(repo *mockShiftTemplateRepository)
		wantErr bool
		errIs   error
	}{
		{
			name: "正常: 通常シフト(morning)で startTime/endTime あり → repo.Create が呼ばれる",
			input: &CreateShiftTemplateInput{
				Name:      "早番",
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: "08:00",
				EndTime:   "13:00:00",
				IsActive:  boolPtr(true),
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.createFn = func(_ context.Context, tpl *model.ShiftTemplate) error {
					tpl.ID = 1
					return nil
				}
				repo.findByIDFn = func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					return &model.ShiftTemplate{ID: id, Name: "早番", ShiftType: model.ShiftTypeMorning}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "正常: 終日シフト(off)で時刻なし → バリデーションエラーなし",
			input: &CreateShiftTemplateInput{
				Name:      "公休",
				ShiftType: string(model.ShiftTypeOff),
				StartTime: "",
				EndTime:   "",
				IsActive:  boolPtr(true),
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.createFn = func(_ context.Context, tpl *model.ShiftTemplate) error {
					tpl.ID = 2
					return nil
				}
				repo.findByIDFn = func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					return &model.ShiftTemplate{ID: id, Name: "公休", ShiftType: model.ShiftTypeOff}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "エラー: endTime が startTime より前 → InvalidInput エラー",
			input: &CreateShiftTemplateInput{
				Name:      "不正シフト",
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: "13:00:00",
				EndTime:   "08:00",
				IsActive:  boolPtr(true),
			},
			setupFn: func(_ *mockShiftTemplateRepository) {},
			wantErr: true,
			errIs:   apperrors.ErrInvalidInput,
		},
		{
			name: "エラー: repo.Create がエラー → error を返す",
			input: &CreateShiftTemplateInput{
				Name:      "早番",
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: "08:00",
				EndTime:   "13:00:00",
				IsActive:  boolPtr(true),
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.createFn = func(_ context.Context, _ *model.ShiftTemplate) error {
					return errors.New("insert failed")
				}
			},
			wantErr: true,
		},
		{
			name: "正常: 休憩あり → repo.UpdateBreaks が呼ばれる",
			input: &CreateShiftTemplateInput{
				Name:      "早番",
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: "08:00",
				EndTime:   "17:00",
				IsActive:  boolPtr(true),
				Breaks:    []ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}},
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.createFn = func(_ context.Context, tpl *model.ShiftTemplate) error {
					tpl.ID = 3
					return nil
				}
				repo.updateBreaksFn = func(_ context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error {
					assert.Equal(t, uint64(3), templateID)
					assert.Len(t, breaks, 1)
					return nil
				}
				repo.findByIDFn = func(_ context.Context, _, id uint64) (*model.ShiftTemplate, error) {
					return &model.ShiftTemplate{ID: id, Name: "早番"}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "エラー: repo.UpdateBreaks がエラー → error を返す",
			input: &CreateShiftTemplateInput{
				Name:      "早番",
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: "08:00",
				EndTime:   "17:00",
				IsActive:  boolPtr(true),
				Breaks:    []ShiftBreakTemplateInput{{BreakStart: "12:00", BreakEnd: "13:00:00"}},
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.createFn = func(_ context.Context, tpl *model.ShiftTemplate) error {
					tpl.ID = 4
					return nil
				}
				repo.updateBreaksFn = func(_ context.Context, _ uint64, _ []model.ShiftTemplateBreak) error {
					return errors.New("break save failed")
				}
			},
			wantErr: true,
		},
		{
			name: "エラー: 作成後の FindByID がエラー → error を返す",
			input: &CreateShiftTemplateInput{
				Name:      "早番",
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: "08:00",
				EndTime:   "13:00:00",
				IsActive:  boolPtr(true),
			},
			setupFn: func(repo *mockShiftTemplateRepository) {
				repo.createFn = func(_ context.Context, tpl *model.ShiftTemplate) error {
					tpl.ID = 5
					return nil
				}
				repo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.ShiftTemplate, error) {
					return nil, errors.New("not found after create")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftTemplateRepository{}
			tt.setupFn(repo)
			svc := NewShiftTemplateService(repo)

			got, err := svc.Create(context.Background(), 10, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.errIs != nil {
					assert.True(t, errors.Is(err, tt.errIs), "expected error to wrap %v, got %v", tt.errIs, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// ---- Tests: Update ----
