package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// newLiffSvcForCatalog は liff_service_catalog.go のテスト用に typeLiffRepo /
// trimmingCourseRepo / trimmingOptionRepo / staffRepo を差し替えた liffService を直接構築する
// （newLiffSvc ヘルパーはこれらのフィールドを配線しないため）。
func newLiffSvcForCatalog(course *mockTrimmingCourseRepository, option *mockTrimmingOptionRepository, staff *mockLiffStaffRepository) *liffService {
	return &liffService{
		typeLiffRepo: &mockLiffTypeRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{
					ID:                 id,
					ClinicID:           clinicID,
					IsActive:           true,
					ReservationVisible: true,
				}, nil
			},
		},
		trimmingCourseRepo:  course,
		trimmingOptionRepo:  option,
		staffRepo:           staff,
		unavailableTimeRepo: &mockLiffUnavailableTimeRepository{},
		occupationRepo:      &mockReservationTypeOccupationRepository{},
	}
}

func TestLiffService_GetTrimmingCourses(t *testing.T) {
	ctx := context.Background()

	t.Run("is_active=true のコースのみを返す", func(t *testing.T) {
		svc := newLiffSvcForCatalog(
			&mockTrimmingCourseRepository{
				findAllFn: func(_ context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
					assert.Equal(t, uint64(3), clinicID)
					return []model.TrimmingCourse{
						{ID: 1, Name: "フルコース", IsActive: true},
						{ID: 2, Name: "廃止コース", IsActive: false},
						{ID: 3, Name: "シャンプーのみ", IsActive: true},
					}, nil
				},
			},
			&mockTrimmingOptionRepository{},
			&mockLiffStaffRepository{},
		)

		got, err := svc.GetTrimmingCourses(ctx, 3)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, uint64(1), got[0].ID)
		assert.Equal(t, uint64(3), got[1].ID)
	})

	t.Run("空リストを返す", func(t *testing.T) {
		svc := newLiffSvcForCatalog(
			&mockTrimmingCourseRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.TrimmingCourse, error) {
					return []model.TrimmingCourse{{ID: 1, Name: "廃止コース", IsActive: false}}, nil
				},
			},
			&mockTrimmingOptionRepository{},
			&mockLiffStaffRepository{},
		)

		got, err := svc.GetTrimmingCourses(ctx, 3)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("リポジトリエラーを伝播する", func(t *testing.T) {
		svc := newLiffSvcForCatalog(
			&mockTrimmingCourseRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.TrimmingCourse, error) {
					return nil, errors.New("db error")
				},
			},
			&mockTrimmingOptionRepository{},
			&mockLiffStaffRepository{},
		)

		_, err := svc.GetTrimmingCourses(ctx, 3)
		require.Error(t, err)
	})
}

func TestLiffService_GetTrimmingOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("is_active=true のオプションのみを返す", func(t *testing.T) {
		svc := newLiffSvcForCatalog(
			&mockTrimmingCourseRepository{},
			&mockTrimmingOptionRepository{
				findAllFn: func(_ context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
					assert.Equal(t, uint64(3), clinicID)
					return []model.TrimmingOption{
						{ID: 1, Name: "爪切り", IsActive: true},
						{ID: 2, Name: "廃止オプション", IsActive: false},
					}, nil
				},
			},
			&mockLiffStaffRepository{},
		)

		got, err := svc.GetTrimmingOptions(ctx, 3)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, uint64(1), got[0].ID)
	})

	t.Run("空リストを返す", func(t *testing.T) {
		svc := newLiffSvcForCatalog(
			&mockTrimmingCourseRepository{},
			&mockTrimmingOptionRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.TrimmingOption, error) {
					return []model.TrimmingOption{}, nil
				},
			},
			&mockLiffStaffRepository{},
		)

		got, err := svc.GetTrimmingOptions(ctx, 3)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("リポジトリエラーを伝播する", func(t *testing.T) {
		svc := newLiffSvcForCatalog(
			&mockTrimmingCourseRepository{},
			&mockTrimmingOptionRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.TrimmingOption, error) {
					return nil, errors.New("db error")
				},
			},
			&mockLiffStaffRepository{},
		)

		_, err := svc.GetTrimmingOptions(ctx, 3)
		require.Error(t, err)
	})
}

// TestLiffService_GetStaffs_FindAllError は GetStaffs 自体の FindAll エラーパスを検証する。
// TestLiffService_GetStaffs（liff_service_test.go）は filterVisibleStaffsByTypeID 側の
// エラーパスのみをカバーしており、GetStaffs 自身の s.staffRepo.FindAll エラーは未カバーだった。
func TestLiffService_GetStaffs_FindAllError(t *testing.T) {
	svc := newLiffSvcForCatalog(
		&mockTrimmingCourseRepository{},
		&mockTrimmingOptionRepository{},
		&mockLiffStaffRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return nil, errors.New("db error")
			},
		},
	)

	_, err := svc.GetStaffs(context.Background(), 3, 1)
	require.Error(t, err)
}
