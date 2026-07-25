package service

// This file is test-only by design. BE9-2F removes the production compatibility
// package while retaining a small set of domain tests whose assertions are still
// unique. The aliases below bind those tests directly to their target domains.

import (
	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
)

func NewClinicService(
	repo clinicdomain.ClinicRepository,
	permissionGroupRepo clinicdomain.PermissionGroupWriter,
	transactor clinicdomain.Transactor,
) clinicdomain.ClinicService {
	return clinicdomain.NewClinicService(repo, permissionGroupRepo, transactor)
}

type SetClinicAssignmentsInput = staffdomain.SetClinicAssignmentsInput
type StaffAssignmentClinicLookup = staffdomain.StaffAssignmentClinicLookup
type StaffService = staffdomain.StaffService

func NewStaffService(
	repo staffdomain.StaffRepository,
	accountRepo staffdomain.StaffAccountStore,
	assignmentRepo staffdomain.StaffClinicAssignmentRepository,
	reservationRepo staffdomain.StaffAssignmentReservationUsage,
	shiftEntryRepo staffdomain.ShiftEntryRepository,
	permissionGroupRepo staffdomain.PermissionGroupRepository,
	resStaffRepo staffdomain.ReservationStaffRepository,
	occupationRepo staffdomain.OccupationRepository,
	clinicRepo StaffAssignmentClinicLookup,
	tx staffdomain.Transactor,
) StaffService {
	return staffdomain.NewStaffService(
		repo,
		accountRepo,
		assignmentRepo,
		reservationRepo,
		shiftEntryRepo,
		permissionGroupRepo,
		resStaffRepo,
		occupationRepo,
		clinicRepo,
		tx,
	)
}

type CreateShiftEntryInput = staffdomain.CreateShiftEntryInput

func NewShiftEntryService(
	repo staffdomain.ShiftEntryRepository,
	staffRepo staffdomain.ShiftEntryStaffLocker,
	staffAssignmentRepo staffdomain.ShiftEntryStaffAssignmentLocker,
	tx staffdomain.Transactor,
) staffdomain.ShiftEntryService {
	return staffdomain.NewShiftEntryService(repo, staffRepo, staffAssignmentRepo, tx)
}

func strPtr(value string) *string {
	return &value
}

func ptrFloat64(value float64) *float64 {
	return &value
}

var (
	_ = strPtr
	_ = ptrFloat64
)
