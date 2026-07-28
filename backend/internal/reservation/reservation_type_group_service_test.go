package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ReservationTypeGroup モック ----

type mockReservationTypeGroupRepository struct {
	findAllFn                      func(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error)
	findByIDFn                     func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error)
	countReservationTypesByGroupFn func(ctx context.Context, clinicID, groupID uint64) (int64, error)
	createFn                       func(ctx context.Context, g *model.ReservationTypeGroup) error
	updateFieldsFn                 func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationTypeGroup, error)
	deleteFn                       func(ctx context.Context, clinicID, id uint64) error
	reorderErr                     error
}

func (m *mockReservationTypeGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockReservationTypeGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationTypeGroupRepository) CountUsageByReservationTypeGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	if m.countReservationTypesByGroupFn != nil {
		return m.countReservationTypesByGroupFn(ctx, clinicID, groupID)
	}
	return 0, nil
}

func (m *mockReservationTypeGroupRepository) Create(ctx context.Context, g *model.ReservationTypeGroup) error {
	return m.createFn(ctx, g)
}

func (m *mockReservationTypeGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationTypeGroup, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.ReservationTypeGroup{}, nil
}

func (m *mockReservationTypeGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationTypeGroupRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

// ---- Tests ----

func TestBuildReservationTypeGroupUpdate(t *testing.T) {
	name := "新グループ"
	color := "#FF0000"
	sortOrder := 5
	isActive := true

	tests := []struct {
		name  string
		input *UpdateReservationTypeGroupInput
		want  map[string]any
	}{
		{
			name:  "all fields set",
			input: &UpdateReservationTypeGroupInput{Name: &name, Color: &color, SortOrder: &sortOrder, IsActive: &isActive},
			want: map[string]any{
				colReservationTypeGroupName:      name,
				colReservationTypeGroupColor:     color,
				colReservationTypeGroupSortOrder: sortOrder,
				colReservationTypeGroupIsActive:  isActive,
			},
		},
		{
			name:  "no fields set returns empty map",
			input: &UpdateReservationTypeGroupInput{},
			want:  map[string]any{},
		},
		{
			name:  "only name set",
			input: &UpdateReservationTypeGroupInput{Name: &name},
			want:  map[string]any{colReservationTypeGroupName: name},
		},
		{
			name:  "only color set",
			input: &UpdateReservationTypeGroupInput{Color: &color},
			want:  map[string]any{colReservationTypeGroupColor: color},
		},
		{
			name:  "only sort_order set",
			input: &UpdateReservationTypeGroupInput{SortOrder: &sortOrder},
			want:  map[string]any{colReservationTypeGroupSortOrder: sortOrder},
		},
		{
			name:  "only is_active set",
			input: &UpdateReservationTypeGroupInput{IsActive: &isActive},
			want:  map[string]any{colReservationTypeGroupIsActive: isActive},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildReservationTypeGroupUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReservationTypeGroupService_List(t *testing.T) {
	tests := []struct {
		name       string
		repoGroups []model.ReservationTypeGroup
		repoErr    error
		wantLen    int
		wantErr    bool
	}{
		{
			name: "returns groups",
			repoGroups: []model.ReservationTypeGroup{
				{ID: 1, Name: "午前診察"},
				{ID: 2, Name: "午後診察"},
			},
			wantLen: 2,
		},
		{
			name:       "returns empty list when no groups exist",
			repoGroups: []model.ReservationTypeGroup{},
			wantLen:    0,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeGroupRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.ReservationTypeGroup, error) {
					return tt.repoGroups, tt.repoErr
				},
			}
			svc := NewReservationTypeGroupService(repo)

			groups, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, groups, tt.wantLen)
			}
		})
	}
}

func TestReservationTypeGroupService_GetByID(t *testing.T) {
	existing := &model.ReservationTypeGroup{ID: 1, ClinicID: 1, Name: "午前診察", Color: "#3B82F6"}

	tests := []struct {
		name         string
		id           uint64
		repoGroup    *model.ReservationTypeGroup
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:      "returns group when found",
			id:        1,
			repoGroup: existing,
			wantErr:   false,
		},
		{
			name:         "returns not found error when group does not exist",
			id:           999,
			repoErr:      apperrors.WrapNotFound("reservation_type_group", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeGroupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationTypeGroup, error) {
					return tt.repoGroup, tt.repoErr
				},
			}
			svc := NewReservationTypeGroupService(repo)
			result, err := svc.GetByID(context.Background(), 1, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoGroup, result)
			}
		})
	}
}

func TestReservationTypeGroupService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateReservationTypeGroupInput
		createErr error
		wantErr   bool
		wantColor string
	}{
		{
			name:      "creates group with provided color",
			input:     CreateReservationTypeGroupInput{Name: "午前診察", Color: "#FF0000", SortOrder: 1},
			wantColor: "#FF0000",
			wantErr:   false,
		},
		{
			name:      "uses default color when color is empty",
			input:     CreateReservationTypeGroupInput{Name: "午後診察", Color: ""},
			wantColor: defaultGroupColor,
			wantErr:   false,
		},
		{
			name:    "returns validation error for empty name",
			input:   CreateReservationTypeGroupInput{Name: ""},
			wantErr: true,
		},
		{
			name:      "propagates repository error",
			input:     CreateReservationTypeGroupInput{Name: "グループ名"},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeGroupRepository{
				createFn: func(_ context.Context, g *model.ReservationTypeGroup) error {
					g.ID = 1
					return tt.createErr
				},
			}
			svc := NewReservationTypeGroupService(repo)
			result, err := svc.Create(context.Background(), 1, &tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, tt.wantColor, result.Color)
			}
		})
	}
}

func TestReservationTypeGroupService_Update(t *testing.T) {
	existing := &model.ReservationTypeGroup{ID: 1, ClinicID: 1, Name: "既存グループ", Color: "#3B82F6"}

	tests := []struct {
		name      string
		input     UpdateReservationTypeGroupInput
		updateErr error
		findErr   error
		wantErr   bool
	}{
		{
			name:    "updates group successfully",
			input:   UpdateReservationTypeGroupInput{Name: strPtr("新グループ名")},
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateReservationTypeGroupInput{},
			wantErr: true,
		},
		{
			name:      "propagates update error",
			input:     UpdateReservationTypeGroupInput{Name: strPtr("名前")},
			updateErr: errors.New("update failed"),
			wantErr:   true,
		},
		{
			name:    "returns error when group not found",
			input:   UpdateReservationTypeGroupInput{Name: strPtr("名前")},
			findErr: apperrors.WrapNotFound("reservation_type_group", "1"),
			wantErr: true,
		},
		{
			name:    "returns error when name is blank",
			input:   UpdateReservationTypeGroupInput{Name: strPtr("")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeGroupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationTypeGroup, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return existing, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ReservationTypeGroup, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return existing, nil
				},
			}
			svc := NewReservationTypeGroupService(repo)
			result, err := svc.Update(context.Background(), 1, 1, &tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestReservationTypeGroupService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		categoryCount int64
		countErr      error
		deleteErr     error
		wantErr       bool
		wantConflict  bool
	}{
		{
			name:    "deletes group successfully",
			wantErr: false,
		},
		{
			name:          "returns conflict when group has categories",
			categoryCount: 3,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:     "propagates count error",
			countErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name:      "propagates delete error",
			deleteErr: errors.New("delete failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeGroupRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.ReservationTypeGroup, error) {
					return &model.ReservationTypeGroup{ID: id}, nil
				},
				countReservationTypesByGroupFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.categoryCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewReservationTypeGroupService(repo)
			err := svc.Delete(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReservationTypeGroupService_Update_NilInput(t *testing.T) {
	repo := &mockReservationTypeGroupRepository{}
	svc := NewReservationTypeGroupService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReservationTypeGroupService_Reorder(t *testing.T) {
	tests := []struct {
		name       string
		ids        []uint64
		reorderErr error
		wantErr    bool
	}{
		{
			name:    "reorders groups successfully",
			ids:     []uint64{3, 1, 2},
			wantErr: false,
		},
		{
			name:    "returns error for empty ids",
			ids:     []uint64{},
			wantErr: true,
		},
		{
			name:       "propagates repository error",
			ids:        []uint64{1, 2},
			reorderErr: errors.New("reorder failed"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeGroupRepository{
				reorderErr: tt.reorderErr,
			}
			svc := NewReservationTypeGroupService(repo)
			err := svc.Reorder(context.Background(), 1, tt.ids)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
