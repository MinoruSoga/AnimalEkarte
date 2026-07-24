package reservation

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type reservationStaffDeleteRecorder struct {
	calls    int
	clinicID uint64
	staffID  uint64
	err      error
}

func (r *reservationStaffDeleteRecorder) Delete(
	_ context.Context,
	clinicID uint64,
	staffID uint64,
) error {
	r.calls++
	r.clinicID = clinicID
	r.staffID = staffID
	return r.err
}

func TestReservationStaffService_Delete_DelegatesToCanonicalStaffService(t *testing.T) {
	repo := &mockReservationStaffRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Staff, error) {
			t.Fatal("reservation repository ownership lookup must not run during canonical delete")
			return nil, nil
		},
		countUsageByStaffIDFn: func() (int64, error) {
			t.Fatal("reservation repository usage count must not run during canonical delete")
			return 0, nil
		},
		deleteFn: func(context.Context, uint64, uint64) error {
			t.Fatal("reservation repository delete bypass must not run")
			return nil
		},
	}
	deleter := &reservationStaffDeleteRecorder{}
	svc := NewReservationStaffService(repo, &mockTransactor{}, deleter)

	err := svc.Delete(context.Background(), 17, 29)

	require.NoError(t, err)
	assert.Equal(t, 1, deleter.calls)
	assert.Equal(t, uint64(17), deleter.clinicID)
	assert.Equal(t, uint64(29), deleter.staffID)
}

func TestReservationStaffService_Delete_FailsClosedWithoutCanonicalDeleter(t *testing.T) {
	svc := NewReservationStaffService(&mockReservationStaffRepository{}, &mockTransactor{}, nil)

	err := svc.Delete(context.Background(), 17, 29)

	require.Error(t, err)
}

func TestReservationStaffService_Delete_PreservesCanonicalErrorCategory(t *testing.T) {
	deleter := &reservationStaffDeleteRecorder{
		err: apperrors.WrapNotFound("staff", "29"),
	}
	svc := NewReservationStaffService(
		&mockReservationStaffRepository{},
		&mockTransactor{},
		deleter,
	)

	err := svc.Delete(context.Background(), 17, 29)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Equal(t, 1, deleter.calls)
}

func TestReservationStaffDeleteSourceContract(t *testing.T) {
	portType := reflect.TypeOf((*ReservationStaffDeleter)(nil)).Elem()
	require.Equal(t, 1, portType.NumMethod())
	assert.Equal(t, "Delete", portType.Method(0).Name)

	for _, file := range []string{
		"reservation_staff_repository.go",
		"../staff/staff_repository.go",
	} {
		source, err := os.ReadFile(file)
		require.NoError(t, err)
		assert.NotContains(t, string(source), "DeleteForReservation", file)
		assert.NotContains(t, string(source), "CountUsageByStaffID", file)
	}

	serviceSource, err := os.ReadFile("reservation_staff_service.go")
	require.NoError(t, err)
	deleteBody := string(serviceSource)
	require.Contains(t, deleteBody, "s.staffDeleter.Delete(ctx, clinicID, id)")
	assert.False(t, strings.Contains(deleteBody, "s.repo.Delete(ctx, clinicID, id)"))

	compositionSource, err := os.ReadFile("../../cmd/api/composition_reservation.go")
	require.NoError(t, err)
	assert.Contains(
		t,
		string(compositionSource),
		"reservation.NewReservationStaffService(repositories.ReservationStaff, dependencies.Transactor, dependencies.StaffDeleter)",
	)
}
