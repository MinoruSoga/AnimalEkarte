package trimming

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingMasterTxMarkerKey struct{}

func trimmingMasterTxSpy(t *testing.T, calls *int) *mockTransactor {
	t.Helper()
	return &mockTransactor{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			*calls++
			return fn(context.WithValue(ctx, trimmingMasterTxMarkerKey{}, true))
		},
	}
}

func requireTrimmingMasterTxContext(t *testing.T, ctx context.Context) {
	t.Helper()
	assert.Equal(t, true, ctx.Value(trimmingMasterTxMarkerKey{}))
}

func TestTrimmingCourseService_Delete_CountsThenDeletesInTransaction(t *testing.T) {
	var operations []string
	txCalls := 0
	repo := &mockTrimmingCourseRepository{
		countUsageByCourseIDFn: func(ctx context.Context, _, _ uint64) (int64, error) {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "count")
			return 0, nil
		},
		deleteFn: func(ctx context.Context, _, _ uint64) error {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "delete")
			return nil
		},
	}
	svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, trimmingMasterTxSpy(t, &txCalls))

	require.NoError(t, svc.Delete(context.Background(), 1, 10))
	assert.Equal(t, 1, txCalls)
	assert.Equal(t, []string{"count", "delete"}, operations)
}

func TestTrimmingCourseService_Delete_CountConflictSkipsDelete(t *testing.T) {
	var operations []string
	txCalls := 0
	repo := &mockTrimmingCourseRepository{
		countUsageByCourseIDFn: func(ctx context.Context, _, _ uint64) (int64, error) {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "count")
			return 1, nil
		},
		deleteFn: func(ctx context.Context, _, _ uint64) error {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "delete")
			return nil
		},
	}
	svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, trimmingMasterTxSpy(t, &txCalls))

	err := svc.Delete(context.Background(), 1, 10)

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, 1, txCalls)
	assert.Equal(t, []string{"count"}, operations)
}

func TestTrimmingOptionService_Delete_CountConflictSkipsDelete(t *testing.T) {
	var operations []string
	txCalls := 0
	repo := &mockTrimmingOptionRepository{
		countRecordsByOptFn: func(ctx context.Context, _, _ uint64) (int64, error) {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "count")
			return 1, nil
		},
		deleteFn: func(ctx context.Context, _, _ uint64) error {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "delete")
			return nil
		},
	}
	svc := NewTrimmingOptionService(repo, trimmingMasterTxSpy(t, &txCalls))

	err := svc.Delete(context.Background(), 1, 10)

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, 1, txCalls)
	assert.Equal(t, []string{"count"}, operations)
}

func TestTrimmingCourseTypeService_Delete_CountConflictSkipsDelete(t *testing.T) {
	var operations []string
	txCalls := 0
	repo := &mockTrimmingCourseTypeRepository{
		countUsageFn: func(ctx context.Context, _, _ uint64) (int64, error) {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "count")
			return 1, nil
		},
		deleteFn: func(ctx context.Context, _, _ uint64) error {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "delete")
			return nil
		},
	}
	svc := NewTrimmingCourseTypeService(repo, trimmingMasterTxSpy(t, &txCalls))

	err := svc.Delete(context.Background(), 1, 10)

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, 1, txCalls)
	assert.Equal(t, []string{"count"}, operations)
}

func TestTrimmingCourseService_Create_ValidatesCourseTypeAndWritesInTransaction(t *testing.T) {
	const courseTypeID = uint64(20)
	var operations []string
	txCalls := 0
	courseRepo := &mockTrimmingCourseRepository{
		createFn: func(ctx context.Context, course *model.TrimmingCourse) error {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "create")
			course.ID = 30
			return nil
		},
	}
	courseTypeRepo := &mockTrimmingCourseTypeRepository{
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "find-course-type")
			return &model.TrimmingCourseType{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewTrimmingCourseService(courseRepo, courseTypeRepo, trimmingMasterTxSpy(t, &txCalls))

	got, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseInput{
		Name:         "transactional course",
		CourseTypeID: ptrUint64(courseTypeID),
	})

	require.NoError(t, err)
	assert.Equal(t, uint64(30), got.ID)
	assert.Equal(t, 1, txCalls)
	assert.Equal(t, []string{"find-course-type", "create"}, operations)
}

func TestTrimmingCourseService_Update_ValidatesCourseTypeAndWritesInTransaction(t *testing.T) {
	const courseTypeID = uint64(20)
	var operations []string
	txCalls := 0
	courseRepo := &mockTrimmingCourseRepository{
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
			assert.Nil(t, ctx.Value(trimmingMasterTxMarkerKey{}), "existing target lookup keeps its public preflight order")
			operations = append(operations, "find-course")
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID}, nil
		},
		updateFieldsFn: func(ctx context.Context, clinicID, id uint64, _ map[string]any) (*model.TrimmingCourse, error) {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "update")
			return &model.TrimmingCourse{ID: id, ClinicID: clinicID, CourseTypeID: ptrUint64(courseTypeID)}, nil
		},
	}
	courseTypeRepo := &mockTrimmingCourseTypeRepository{
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
			requireTrimmingMasterTxContext(t, ctx)
			operations = append(operations, "find-course-type")
			return &model.TrimmingCourseType{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewTrimmingCourseService(courseRepo, courseTypeRepo, trimmingMasterTxSpy(t, &txCalls))

	got, err := svc.Update(context.Background(), 1, 10, &UpdateTrimmingCourseInput{
		CourseTypeID: ptrUint64(courseTypeID),
	})

	require.NoError(t, err)
	assert.Equal(t, uint64(10), got.ID)
	assert.Equal(t, 1, txCalls)
	assert.Equal(t, []string{"find-course", "find-course-type", "update"}, operations)
}

func TestTrimmingMasterServices_FailClosedWithoutTransactor(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "course",
			run: func() error {
				return NewTrimmingCourseService(&mockTrimmingCourseRepository{}, &mockMinimalCourseTypeRepo{}, nil).
					Delete(context.Background(), 1, 1)
			},
		},
		{
			name: "option",
			run: func() error {
				return NewTrimmingOptionService(&mockTrimmingOptionRepository{}, nil).
					Delete(context.Background(), 1, 1)
			},
		},
		{
			name: "course type",
			run: func() error {
				return NewTrimmingCourseTypeService(&mockTrimmingCourseTypeRepository{}, nil).
					Delete(context.Background(), 1, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			require.Error(t, err)
			assert.ErrorContains(t, err, "transaction dependency is required")
		})
	}
}

func TestTrimmingMasterServices_DoNotWriteWhenTransactionStartFails(t *testing.T) {
	txErr := errors.New("transaction unavailable")
	transactor := &mockTransactor{withTxErr: txErr}
	writeCalls := 0

	courseErr := NewTrimmingCourseService(
		&mockTrimmingCourseRepository{
			deleteFn: func(context.Context, uint64, uint64) error {
				writeCalls++
				return nil
			},
		},
		&mockMinimalCourseTypeRepo{},
		transactor,
	).Delete(context.Background(), 1, 1)
	optionErr := NewTrimmingOptionService(
		&mockTrimmingOptionRepository{
			deleteFn: func(context.Context, uint64, uint64) error {
				writeCalls++
				return nil
			},
		},
		transactor,
	).Delete(context.Background(), 1, 1)
	courseTypeErr := NewTrimmingCourseTypeService(
		&mockTrimmingCourseTypeRepository{
			deleteFn: func(context.Context, uint64, uint64) error {
				writeCalls++
				return nil
			},
		},
		transactor,
	).Delete(context.Background(), 1, 1)

	assert.ErrorIs(t, courseErr, txErr)
	assert.ErrorIs(t, optionErr, txErr)
	assert.ErrorIs(t, courseTypeErr, txErr)
	assert.Zero(t, writeCalls)
}
