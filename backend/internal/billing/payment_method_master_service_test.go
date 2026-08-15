package billing

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockPaymentMethodMasterRepository は PaymentMethodMasterRepository のテスト用モック実装
type mockPaymentMethodMasterRepository struct {
	findAllFn                     func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	createFn                      func(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error)
	updateFieldsFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	countUsageByPaymentMethodIDFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockPaymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) Create(ctx context.Context, pm *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
	if m.createFn != nil {
		return m.createFn(ctx, pm)
	}
	return pm, nil
}

func (m *mockPaymentMethodMasterRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockPaymentMethodMasterRepository) CountUsageByPaymentMethodID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByPaymentMethodIDFn != nil {
		return m.countUsageByPaymentMethodIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockPaymentMethodMasterRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// ---- buildPaymentMethodUpdate (純粋関数) ----

func TestBuildPaymentMethodUpdate(t *testing.T) {
	ptrStr := func(s string) *string { return &s }
	ptrInt := func(i int) *int { return &i }
	ptrBool := func(b bool) *bool { return &b }

	tests := []struct {
		name  string
		input *UpdatePaymentMethodMasterInput
		want  map[string]any
	}{
		{
			name:  "全フィールド nil は空 map",
			input: &UpdatePaymentMethodMasterInput{},
			want:  map[string]any{},
		},
		{
			name:  "Name のみ",
			input: &UpdatePaymentMethodMasterInput{Name: ptrStr("現金")},
			want:  map[string]any{colPaymentMethodName: "現金"},
		},
		{
			name:  "DisplayOrder のみ",
			input: &UpdatePaymentMethodMasterInput{DisplayOrder: ptrInt(3)},
			want:  map[string]any{colPaymentMethodDisplayOrder: 3},
		},
		{
			name:  "IsActive のみ (false)",
			input: &UpdatePaymentMethodMasterInput{IsActive: ptrBool(false)},
			want:  map[string]any{colPaymentMethodIsActive: false},
		},
		{
			name: "全フィールド指定",
			input: &UpdatePaymentMethodMasterInput{
				Name:         ptrStr("クレジット"),
				DisplayOrder: ptrInt(5),
				IsActive:     ptrBool(true),
			},
			want: map[string]any{
				colPaymentMethodName:         "クレジット",
				colPaymentMethodDisplayOrder: 5,
				colPaymentMethodIsActive:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPaymentMethodUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- テスト ----

func TestPaymentMethodMasterService_List(t *testing.T) {
	methods := []model.PaymentMethodMaster{
		{ID: 1, ClinicID: 1, Name: "現金", DisplayOrder: 1},
		{ID: 2, ClinicID: 1, Name: "クレジット", DisplayOrder: 2},
	}

	tests := []struct {
		name       string
		findAllFn  func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
		wantErr    bool
		wantResult []model.PaymentMethodMaster
	}{
		{
			name: "正常: 全件を返す",
			findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return methods, nil
			},
			wantResult: methods,
		},
		{
			name: "エラー: DB エラーを返す",
			findAllFn: func(_ context.Context, _ uint64) ([]model.PaymentMethodMaster, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{findAllFn: tt.findAllFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			got, err := svc.List(context.Background(), 1)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func TestPaymentMethodMasterService_GetByID(t *testing.T) {
	pm := &model.PaymentMethodMaster{ID: 1, ClinicID: 1, Name: "現金", DisplayOrder: 1}

	tests := []struct {
		name       string
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
		wantErr    bool
		wantResult *model.PaymentMethodMaster
	}{
		{
			name: "正常: 指定IDのレコードを返す",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PaymentMethodMaster, error) {
				return pm, nil
			},
			wantResult: pm,
		},
		{
			name: "エラー: 存在しないID → エラーを返す",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PaymentMethodMaster, error) {
				return nil, apperrors.WrapNotFound("payment_method", "99")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{findByIDFn: tt.findByIDFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			got, err := svc.GetByID(context.Background(), 1, 1)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func TestPaymentMethodMasterService_Create(t *testing.T) {
	created := &model.PaymentMethodMaster{
		ID:           1,
		ClinicID:     1,
		Name:         "クレジット",
		DisplayOrder: 2,
	}

	tests := []struct {
		name       string
		input      *CreatePaymentMethodMasterInput
		createFn   func(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error)
		wantErr    bool
		wantResult *model.PaymentMethodMaster
	}{
		{
			name: "正常: 作成済みレコードを返す",
			input: &CreatePaymentMethodMasterInput{
				Name:         "クレジット",
				DisplayOrder: 2,
			},
			createFn: func(_ context.Context, _ *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
				return created, nil
			},
			wantResult: created,
		},
		{
			name: "エラー: 名前が空 → ErrInvalidInput",
			input: &CreatePaymentMethodMasterInput{
				Name:         "",
				DisplayOrder: 1,
			},
			wantErr: true,
		},
		{
			name: "エラー: DB エラーを返す",
			input: &CreatePaymentMethodMasterInput{
				Name:         "クレジット",
				DisplayOrder: 2,
			},
			createFn: func(_ context.Context, _ *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{createFn: tt.createFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			got, err := svc.Create(context.Background(), 1, tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func TestPaymentMethodMasterService_Delete(t *testing.T) {
	cashKey := "cash"
	ptrStr := func(s string) *string { return &s }

	tests := []struct {
		name                          string
		id                            uint64
		findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
		countUsageByPaymentMethodIDFn func(ctx context.Context, clinicID, id uint64) (int64, error)
		deleteFn                      func(ctx context.Context, clinicID, id uint64) error
		wantErr                       bool
		wantErrIs                     error
		wantErrMsg                    string
	}{
		{
			name: "正常: 未使用のカスタム支払方法を削除",
			id:   1,
			findByIDFn: func(_ context.Context, _, id uint64) (*model.PaymentMethodMaster, error) {
				// system_key nil = カスタム行
				return &model.PaymentMethodMaster{ID: id, Name: "カスタム"}, nil
			},
			countUsageByPaymentMethodIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
		},
		{
			name: "正常: system_key 空文字の行はカスタム扱い・usage=0 なら削除可",
			id:   10,
			findByIDFn: func(_ context.Context, _, id uint64) (*model.PaymentMethodMaster, error) {
				return &model.PaymentMethodMaster{ID: id, Name: "空キー", SystemKey: ptrStr("")}, nil
			},
			countUsageByPaymentMethodIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return nil
			},
		},
		{
			name: "エラー: システム標準行（cash）は usage=0 でも削除不可 → ErrConflict",
			id:   5,
			findByIDFn: func(_ context.Context, _, id uint64) (*model.PaymentMethodMaster, error) {
				return &model.PaymentMethodMaster{ID: id, Name: "現金", SystemKey: &cashKey}, nil
			},
			// countUsage は呼ばれない想定（システムガードが先）
			wantErr:    true,
			wantErrIs:  apperrors.ErrConflict,
			wantErrMsg: "システム標準の支払方法は削除できません",
		},
		{
			name: "エラー: 使用中の支払方法 → ErrConflict",
			id:   2,
			findByIDFn: func(_ context.Context, _, id uint64) (*model.PaymentMethodMaster, error) {
				return &model.PaymentMethodMaster{ID: id}, nil
			},
			countUsageByPaymentMethodIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 5, nil
			},
			wantErr:   true,
			wantErrIs: apperrors.ErrConflict,
		},
		{
			name: "エラー: CountUsageByPaymentMethodID がエラーを返す",
			id:   3,
			findByIDFn: func(_ context.Context, _, id uint64) (*model.PaymentMethodMaster, error) {
				return &model.PaymentMethodMaster{ID: id}, nil
			},
			countUsageByPaymentMethodIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "エラー: Delete がエラーを返す",
			id:   4,
			findByIDFn: func(_ context.Context, _, id uint64) (*model.PaymentMethodMaster, error) {
				return &model.PaymentMethodMaster{ID: id}, nil
			},
			countUsageByPaymentMethodIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{
				findByIDFn:                    tt.findByIDFn,
				countUsageByPaymentMethodIDFn: tt.countUsageByPaymentMethodIDFn,
				deleteFn:                      tt.deleteFn,
			}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			err := svc.Delete(context.Background(), 1, tt.id)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestPaymentMethodMasterService_Update(t *testing.T) {
	cashKey := "cash"
	updated := &model.PaymentMethodMaster{ID: 1, ClinicID: 1, Name: "現金", DisplayOrder: 1}
	systemCash := &model.PaymentMethodMaster{
		ID: 1, ClinicID: 1, Name: "現金", DisplayOrder: 1, SystemKey: &cashKey, IsActive: true,
	}
	customPM := &model.PaymentMethodMaster{
		ID: 2, ClinicID: 1, Name: "カスタム決済", DisplayOrder: 10, IsActive: true,
	}

	ptrStr := func(s string) *string { return &s }
	ptrBool := func(b bool) *bool { return &b }

	// 既定 FindByID: カスタム行（system_key nil）— 既存テスト互換
	defaultFindByID := func(_ context.Context, _, id uint64) (*model.PaymentMethodMaster, error) {
		return &model.PaymentMethodMaster{ID: id, ClinicID: 1, Name: "カスタム"}, nil
	}

	tests := []struct {
		name           string
		input          *UpdatePaymentMethodMasterInput
		findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
		updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error)
		wantErr        bool
		wantErrIs      error
		wantErrMsg     string
		wantResult     *model.PaymentMethodMaster
	}{
		{
			name:  "正常: 名前のみ更新が成功",
			input: &UpdatePaymentMethodMasterInput{Name: ptrStr("現金")},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.PaymentMethodMaster, error) {
				return updated, nil
			},
			wantResult: updated,
		},
		{
			name:  "正常: システム標準行（cash）の名称更新は許可",
			input: &UpdatePaymentMethodMasterInput{Name: ptrStr("現金（店舗）")},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PaymentMethodMaster, error) {
				return systemCash, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
				assert.Equal(t, "現金（店舗）", fields[colPaymentMethodName])
				renamed := *systemCash
				renamed.Name = "現金（店舗）"
				return &renamed, nil
			},
			wantResult: &model.PaymentMethodMaster{
				ID: 1, ClinicID: 1, Name: "現金（店舗）", DisplayOrder: 1, SystemKey: &cashKey, IsActive: true,
			},
		},
		{
			name:  "正常: カスタム行の IsActive=false は許可",
			input: &UpdatePaymentMethodMasterInput{IsActive: ptrBool(false)},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PaymentMethodMaster, error) {
				return customPM, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
				assert.Equal(t, false, fields[colPaymentMethodIsActive])
				deactivated := *customPM
				deactivated.IsActive = false
				return &deactivated, nil
			},
			wantResult: &model.PaymentMethodMaster{
				ID: 2, ClinicID: 1, Name: "カスタム決済", DisplayOrder: 10, IsActive: false,
			},
		},
		{
			name:  "エラー: システム標準行（cash）の IsActive=false → ErrConflict",
			input: &UpdatePaymentMethodMasterInput{IsActive: ptrBool(false)},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PaymentMethodMaster, error) {
				return systemCash, nil
			},
			wantErr:    true,
			wantErrIs:  apperrors.ErrConflict,
			wantErrMsg: "システム標準の支払方法は無効化できません",
		},
		{
			name:    "エラー: input nil はエラー",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "エラー: 空文字名前はエラー",
			input:   &UpdatePaymentMethodMasterInput{Name: ptrStr("")},
			wantErr: true,
		},
		{
			name:    "エラー: フィールドなしはエラー",
			input:   &UpdatePaymentMethodMasterInput{},
			wantErr: true,
		},
		{
			name:  "エラー: 存在しないIDはエラー",
			input: &UpdatePaymentMethodMasterInput{Name: ptrStr("現金")},
			// FindByID 成功後に Update が NotFound を返す経路
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.PaymentMethodMaster, error) {
				return nil, apperrors.WrapNotFound("payment_method", "99")
			},
			wantErr: true,
		},
		{
			name:  "エラー: FindByID が存在しないIDエラーを返す (Update より前段でチェック)",
			input: &UpdatePaymentMethodMasterInput{Name: ptrStr("現金")},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.PaymentMethodMaster, error) {
				return nil, apperrors.WrapNotFound("payment_method", "99")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			findFn := tt.findByIDFn
			if findFn == nil && tt.input != nil {
				// input nil 以外で FindByID 未指定ならカスタム行を返す（nil 受信を避ける）
				findFn = defaultFindByID
			}
			// FindByID 失敗ケースは findByIDFn が明示されているので上書きしない
			if tt.findByIDFn != nil {
				findFn = tt.findByIDFn
			}
			repo := &mockPaymentMethodMasterRepository{findByIDFn: findFn, updateFieldsFn: tt.updateFieldsFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			got, err := svc.Update(context.Background(), 1, 1, tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "want errors.Is(%v), got %v", tt.wantErrIs, err)
				}
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

func TestIsSystemPaymentMethod(t *testing.T) {
	cashKey := "cash"
	empty := ""
	tests := []struct {
		name string
		m    *model.PaymentMethodMaster
		want bool
	}{
		{name: "nil master", m: nil, want: false},
		{name: "system_key nil", m: &model.PaymentMethodMaster{}, want: false},
		{name: "system_key 空文字", m: &model.PaymentMethodMaster{SystemKey: &empty}, want: false},
		{name: "system_key cash", m: &model.PaymentMethodMaster{SystemKey: &cashKey}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isSystemPaymentMethod(tt.m))
		})
	}
}

func TestPaymentMethodMasterService_Reorder(t *testing.T) {
	tests := []struct {
		name      string
		ids       []uint64
		reorderFn func(ctx context.Context, clinicID uint64, ids []uint64) error
		wantErr   bool
	}{
		{
			name: "正常: 並び順を更新する",
			ids:  []uint64{2, 1, 3},
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return nil
			},
		},
		{
			name:    "エラー: IDリストが空",
			ids:     []uint64{},
			wantErr: true,
		},
		{
			name: "エラー: DB エラーを返す",
			ids:  []uint64{1, 2},
			reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockPaymentMethodMasterRepository{reorderFn: tt.reorderFn}
			svc := NewPaymentMethodMasterService(repo)

			// Act
			err := svc.Reorder(context.Background(), 1, tt.ids)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
