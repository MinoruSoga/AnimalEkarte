package staff

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockResStaffRepoForStaffPermissions は StaffPermissionService の除外/対応可能サービス種別
// テスト用に repository.ReservationStaffRepository を実装する、対象メソッドのみ関数フィールドで
// 戻り値を制御可能なモック。
type mockResStaffRepoForStaffPermissions struct {
	findAllExcludedFn     func(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error)
	updateExcludedFn      func(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	findAllCapabilitiesFn func(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error)
	updateCapabilitiesFn  func(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
}

func (m *mockResStaffRepoForStaffPermissions) FindAll(_ context.Context, _ uint64) ([]model.Staff, error) {
	return nil, nil
}
func (m *mockResStaffRepoForStaffPermissions) FindByID(_ context.Context, _, _ uint64) (*model.Staff, error) {
	return nil, nil
}
func (m *mockResStaffRepoForStaffPermissions) Create(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}
func (m *mockResStaffRepoForStaffPermissions) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockResStaffRepoForStaffPermissions) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockResStaffRepoForStaffPermissions) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockResStaffRepoForStaffPermissions) UpdateSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *mockResStaffRepoForStaffPermissions) FindAllExcludedReservationTypes(
	ctx context.Context,
	clinicID, staffID uint64,
) ([]model.StaffReservationExclusion, error) {
	if m.findAllExcludedFn != nil {
		return m.findAllExcludedFn(ctx, clinicID, staffID)
	}
	return nil, nil
}
func (m *mockResStaffRepoForStaffPermissions) FindAllExcludedReservationTypesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *mockResStaffRepoForStaffPermissions) UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if m.updateExcludedFn != nil {
		return m.updateExcludedFn(ctx, clinicID, staffID, typeIDs)
	}
	return nil
}
func (m *mockResStaffRepoForStaffPermissions) FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error) {
	if m.findAllCapabilitiesFn != nil {
		return m.findAllCapabilitiesFn(ctx, clinicID, staffID)
	}
	return nil, nil
}
func (m *mockResStaffRepoForStaffPermissions) FindAllReservationCapabilitiesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}
func (m *mockResStaffRepoForStaffPermissions) UpdateReservationCapabilities(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if m.updateCapabilitiesFn != nil {
		return m.updateCapabilitiesFn(ctx, clinicID, staffID, typeIDs)
	}
	return nil
}
func (m *mockResStaffRepoForStaffPermissions) SupportsReservationType(_ context.Context, _, _, _ uint64) (bool, error) {
	return true, nil
}

// newTestStaffServiceForPermissions は StaffPermissionService のテストに必要な最小構成で
// StaffService を組み立てる。権限グループ／予約種別リポジトリ以外は staff_service_test.go 内の
// 共有スタブ（mockStaffRepository 等）を再利用する。
func newTestStaffServiceForPermissions(
	permRepo *mockPermissionGroupRepository,
	resRepo *mockResStaffRepoForStaffPermissions,
) StaffService {
	return NewStaffService(
		&mockStaffRepository{},
		&mockAccountForStaff{},
		&mockAssignmentForStaff{},
		&mockReservationForStaff{},
		&mockShiftEntryForStaff{},
		permRepo,
		resRepo,
		nil,
		nil,
		noopTransactor{},
	)
}

// ---- テスト ----

func TestStaffService_GetPermissionGroupIDs(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint64
		repoErr error
		wantErr bool
	}{
		{
			name: "正常: 権限グループIDリストを返す",
			ids:  []uint64{1, 2, 3},
		},
		{
			name: "正常: 空リスト",
			ids:  []uint64{},
		},
		{
			name:    "エラー: repo エラーを伝播",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permRepo := &mockPermissionGroupRepository{
				findAllGroupIDsByStaffIDFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), staffID)
					return tt.ids, tt.repoErr
				},
			}
			svc := newTestStaffServiceForPermissions(permRepo, &mockResStaffRepoForStaffPermissions{})

			got, err := svc.GetPermissionGroupIDs(context.Background(), 1, 10)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.ids, got)
		})
	}
}

func TestStaffService_SetPermissionGroupIDs(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "正常: 権限グループを全置換する"},
		{name: "エラー: repo エラーを伝播", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotClinicID, gotStaffID uint64
			var gotGroupIDs []uint64
			permRepo := &mockPermissionGroupRepository{
				updateStaffGroupsFn: func(_ context.Context, clinicID, staffID uint64, groupIDs []uint64) error {
					gotClinicID, gotStaffID, gotGroupIDs = clinicID, staffID, groupIDs
					return tt.repoErr
				},
			}
			svc := newTestStaffServiceForPermissions(permRepo, &mockResStaffRepoForStaffPermissions{})

			err := svc.SetPermissionGroupIDs(context.Background(), 1, 10, []uint64{4, 5})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, uint64(1), gotClinicID)
			assert.Equal(t, uint64(10), gotStaffID)
			assert.Equal(t, []uint64{4, 5}, gotGroupIDs)
		})
	}
}

func TestStaffService_GetExcludedReservationTypeIDs(t *testing.T) {
	tests := []struct {
		name    string
		items   []model.StaffReservationExclusion
		repoErr error
		want    []uint64
		wantErr bool
	}{
		{
			name: "正常: 除外サービス種別IDに変換される",
			items: []model.StaffReservationExclusion{
				{StaffID: 1, ReservationTypeID: 10},
				{StaffID: 1, ReservationTypeID: 20},
			},
			want: []uint64{10, 20},
		},
		{
			name:  "正常: 空リスト",
			items: nil,
			want:  []uint64{},
		},
		{
			name:    "エラー: repo エラーを伝播",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resRepo := &mockResStaffRepoForStaffPermissions{
				findAllExcludedFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
					return tt.items, tt.repoErr
				},
			}
			svc := newTestStaffServiceForPermissions(&mockPermissionGroupRepository{}, resRepo)

			got, err := svc.GetExcludedReservationTypeIDs(context.Background(), 1, 1)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStaffService_GetExcludedReservationTypeIDs_UsesAuthenticatedClinicScope(t *testing.T) {
	resRepo := &mockResStaffRepoForStaffPermissions{
		findAllExcludedFn: func(_ context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error) {
			assert.Equal(t, uint64(20), clinicID)
			assert.Equal(t, uint64(7), staffID)
			return []model.StaffReservationExclusion{
				{StaffID: staffID, ReservationTypeID: 200},
			}, nil
		},
	}
	svc := newTestStaffServiceForPermissions(&mockPermissionGroupRepository{}, resRepo)
	scoped, ok := any(svc).(interface {
		GetExcludedReservationTypeIDs(context.Context, uint64, uint64) ([]uint64, error)
	})
	require.True(t, ok, "staff service must require clinicID for exclusion reads")

	ids, err := scoped.GetExcludedReservationTypeIDs(context.Background(), 20, 7)

	require.NoError(t, err)
	assert.Equal(t, []uint64{200}, ids)
}

func TestStaffService_SetExcludedReservationTypeIDs(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "正常: 除外サービス種別を全置換する"},
		{name: "エラー: repo エラーを伝播", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resRepo := &mockResStaffRepoForStaffPermissions{
				updateExcludedFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := newTestStaffServiceForPermissions(&mockPermissionGroupRepository{}, resRepo)

			err := svc.SetExcludedReservationTypeIDs(context.Background(), 1, 10, []uint64{7})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestStaffService_GetCapableReservationTypeIDs(t *testing.T) {
	tests := []struct {
		name    string
		items   []model.StaffReservationCapability
		repoErr error
		want    []uint64
		wantErr bool
	}{
		{
			name: "正常: 対応可能サービス種別IDに変換される",
			items: []model.StaffReservationCapability{
				{StaffID: 1, ReservationTypeID: 30},
			},
			want: []uint64{30},
		},
		{
			name:  "正常: 空リスト",
			items: nil,
			want:  []uint64{},
		},
		{
			name:    "エラー: repo エラーを伝播",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resRepo := &mockResStaffRepoForStaffPermissions{
				findAllCapabilitiesFn: func(_ context.Context, _, _ uint64) ([]model.StaffReservationCapability, error) {
					return tt.items, tt.repoErr
				},
			}
			svc := newTestStaffServiceForPermissions(&mockPermissionGroupRepository{}, resRepo)

			got, err := svc.GetCapableReservationTypeIDs(context.Background(), 1, 1)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStaffService_SetCapableReservationTypeIDs(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{name: "正常: 対応可能サービス種別を全置換する"},
		{name: "エラー: repo エラーを伝播", repoErr: errors.New("db error"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resRepo := &mockResStaffRepoForStaffPermissions{
				updateCapabilitiesFn: func(_ context.Context, _, _ uint64, _ []uint64) error {
					return tt.repoErr
				},
			}
			svc := newTestStaffServiceForPermissions(&mockPermissionGroupRepository{}, resRepo)

			err := svc.SetCapableReservationTypeIDs(context.Background(), 1, 10, []uint64{9})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
