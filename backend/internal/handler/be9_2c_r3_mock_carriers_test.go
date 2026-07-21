package handler

// be9_2c_r3_mock_carriers_test.go — R③でreservationへ移動したappointment_handler_test.goが
// 定義していた共有mockのcarrier複製（残留consumer: appointment_admin_handler_test等。R④で解消）。

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockStaffClinicAssignmentService はテスト用モック。テストで使われるクリニックID（1, 3）すべてに所属を返す。
type mockStaffClinicAssignmentService struct{}

func (m *mockStaffClinicAssignmentService) FindAllByStaffID(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	return []model.StaffClinicAssignment{
		{StaffID: staffID, ClinicID: 1},
		{StaffID: staffID, ClinicID: 3},
	}, nil
}

func (m *mockStaffClinicAssignmentService) FindByClinicID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}

func (m *mockStaffClinicAssignmentService) Create(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}

func (m *mockStaffClinicAssignmentService) Update(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}

func (m *mockStaffClinicAssignmentService) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
