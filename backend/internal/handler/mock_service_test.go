package handler

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetAllPets(ctx context.Context) ([]model.Pet, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Pet), args.Error(1)
}

func (m *MockService) GetPetByID(ctx context.Context, id string) (*model.Pet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockService) CreatePet(ctx context.Context, req *model.CreatePetRequest) (*model.Pet, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockService) UpdatePet(ctx context.Context, id string, req *model.UpdatePetRequest) (*model.Pet, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockService) DeletePet(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Owner Mock Methods
func (m *MockService) GetAllOwners(ctx context.Context) ([]model.Owner, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Owner), args.Error(1)
}

func (m *MockService) GetOwnerByID(ctx context.Context, id string) (*model.Owner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Owner), args.Error(1)
}

func (m *MockService) CreateOwner(ctx context.Context, req *model.CreateOwnerRequest) (*model.Owner, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Owner), args.Error(1)
}

func (m *MockService) UpdateOwner(ctx context.Context, id string, req *model.UpdateOwnerRequest) (*model.Owner, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Owner), args.Error(1)
}

func (m *MockService) DeleteOwner(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Medical Records Mock Methods
func (m *MockService) GetAllMedicalRecords(ctx context.Context) ([]model.MedicalRecord, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MedicalRecord), args.Error(1)
}

func (m *MockService) GetMedicalRecordByID(ctx context.Context, id string) (*model.MedicalRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MedicalRecord), args.Error(1)
}

func (m *MockService) GetMedicalRecordsByPetID(ctx context.Context, petID string) ([]model.MedicalRecord, error) {
	args := m.Called(ctx, petID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MedicalRecord), args.Error(1)
}

func (m *MockService) GetMedicalRecordsByOwnerID(ctx context.Context, ownerID string) ([]model.MedicalRecord, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MedicalRecord), args.Error(1)
}

func (m *MockService) CreateMedicalRecord(ctx context.Context, req *model.CreateMedicalRecordRequest) (*model.MedicalRecord, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MedicalRecord), args.Error(1)
}

func (m *MockService) UpdateMedicalRecord(ctx context.Context, id string, req *model.UpdateMedicalRecordRequest) (*model.MedicalRecord, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MedicalRecord), args.Error(1)
}

func (m *MockService) DeleteMedicalRecord(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetDB Mock Method
func (m *MockService) GetDB() (interface{ DB() *gorm.DB }, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interface{ DB() *gorm.DB }), args.Error(1)
}

// GetMedicalRecordsWithPagination Mock Method
func (m *MockService) GetMedicalRecordsWithPagination(ctx context.Context, page, limit int) (*model.PaginatedMedicalRecords, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PaginatedMedicalRecords), args.Error(1)
}

// Reservation Mock Methods
func (m *MockService) GetAllReservations(ctx context.Context) ([]model.Reservation, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Reservation), args.Error(1)
}

func (m *MockService) GetReservationsByDate(ctx context.Context, date string) ([]model.Reservation, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Reservation), args.Error(1)
}

func (m *MockService) GetReservationByID(ctx context.Context, id string) (*model.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Reservation), args.Error(1)
}

func (m *MockService) GetReservationsByPetID(ctx context.Context, petID string) ([]model.Reservation, error) {
	args := m.Called(ctx, petID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Reservation), args.Error(1)
}

func (m *MockService) GetReservationsByOwnerID(ctx context.Context, ownerID string) ([]model.Reservation, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Reservation), args.Error(1)
}

func (m *MockService) CreateReservation(ctx context.Context, req *model.CreateReservationRequest) (*model.Reservation, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Reservation), args.Error(1)
}

func (m *MockService) UpdateReservation(ctx context.Context, id string, req *model.UpdateReservationRequest) (*model.Reservation, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Reservation), args.Error(1)
}

func (m *MockService) DeleteReservation(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MasterItem Mock Methods
func (m *MockService) GetAllMasterItems(ctx context.Context) ([]model.MasterItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MasterItem), args.Error(1)
}

func (m *MockService) GetMasterItemByID(ctx context.Context, id string) (*model.MasterItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MasterItem), args.Error(1)
}

func (m *MockService) GetMasterItemsByCategory(ctx context.Context, category string) ([]model.MasterItem, error) {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MasterItem), args.Error(1)
}

func (m *MockService) CreateMasterItem(ctx context.Context, req *model.CreateMasterItemRequest) (*model.MasterItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MasterItem), args.Error(1)
}

func (m *MockService) UpdateMasterItem(ctx context.Context, id string, req *model.UpdateMasterItemRequest) (*model.MasterItem, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MasterItem), args.Error(1)
}

func (m *MockService) DeleteMasterItem(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Hospitalization Mock Methods
func (m *MockService) GetAllHospitalizations(ctx context.Context) ([]model.Hospitalization, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Hospitalization), args.Error(1)
}

func (m *MockService) GetHospitalizationByID(ctx context.Context, id string) (*model.Hospitalization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Hospitalization), args.Error(1)
}

func (m *MockService) GetHospitalizationsByPetID(ctx context.Context, petID string) ([]model.Hospitalization, error) {
	args := m.Called(ctx, petID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Hospitalization), args.Error(1)
}

func (m *MockService) GetHospitalizationsByOwnerID(ctx context.Context, ownerID string) ([]model.Hospitalization, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Hospitalization), args.Error(1)
}

func (m *MockService) GetHospitalizationsByStatus(ctx context.Context, status string) ([]model.Hospitalization, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Hospitalization), args.Error(1)
}

func (m *MockService) CreateHospitalization(ctx context.Context, req *model.CreateHospitalizationRequest) (*model.Hospitalization, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Hospitalization), args.Error(1)
}

func (m *MockService) UpdateHospitalization(ctx context.Context, id string, req *model.UpdateHospitalizationRequest) (*model.Hospitalization, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Hospitalization), args.Error(1)
}

func (m *MockService) DeleteHospitalization(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Accounting Mock Methods
func (m *MockService) GetAllAccounting(ctx context.Context) ([]model.Accounting, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Accounting), args.Error(1)
}

func (m *MockService) GetAccountingByID(ctx context.Context, id string) (*model.Accounting, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Accounting), args.Error(1)
}

func (m *MockService) GetAccountingByPetID(ctx context.Context, petID string) ([]model.Accounting, error) {
	args := m.Called(ctx, petID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Accounting), args.Error(1)
}

func (m *MockService) GetAccountingByOwnerID(ctx context.Context, ownerID string) ([]model.Accounting, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Accounting), args.Error(1)
}

func (m *MockService) GetAccountingByStatus(ctx context.Context, status string) ([]model.Accounting, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Accounting), args.Error(1)
}

func (m *MockService) CreateAccounting(ctx context.Context, req *model.CreateAccountingRequest) (*model.Accounting, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Accounting), args.Error(1)
}

func (m *MockService) UpdateAccounting(ctx context.Context, id string, req *model.UpdateAccountingRequest) (*model.Accounting, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Accounting), args.Error(1)
}

func (m *MockService) DeleteAccounting(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Examination Mock Methods
func (m *MockService) GetAllExaminations(ctx context.Context) ([]model.Examination, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Examination), args.Error(1)
}

func (m *MockService) GetExaminationByID(ctx context.Context, id string) (*model.Examination, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Examination), args.Error(1)
}

func (m *MockService) GetExaminationsByPetID(ctx context.Context, petID string) ([]model.Examination, error) {
	args := m.Called(ctx, petID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Examination), args.Error(1)
}

func (m *MockService) GetExaminationsByOwnerID(ctx context.Context, ownerID string) ([]model.Examination, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Examination), args.Error(1)
}

func (m *MockService) GetExaminationsByStatus(ctx context.Context, status string) ([]model.Examination, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Examination), args.Error(1)
}

func (m *MockService) CreateExamination(ctx context.Context, req *model.CreateExaminationRequest) (*model.Examination, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Examination), args.Error(1)
}

func (m *MockService) UpdateExamination(ctx context.Context, id string, req *model.UpdateExaminationRequest) (*model.Examination, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Examination), args.Error(1)
}

func (m *MockService) DeleteExamination(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Vaccination Mock Methods
func (m *MockService) GetAllVaccinations(ctx context.Context) ([]model.Vaccination, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Vaccination), args.Error(1)
}

func (m *MockService) GetVaccinationByID(ctx context.Context, id string) (*model.Vaccination, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Vaccination), args.Error(1)
}

func (m *MockService) GetVaccinationsByPetID(ctx context.Context, petID string) ([]model.Vaccination, error) {
	args := m.Called(ctx, petID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Vaccination), args.Error(1)
}

func (m *MockService) GetVaccinationsByOwnerID(ctx context.Context, ownerID string) ([]model.Vaccination, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Vaccination), args.Error(1)
}

func (m *MockService) CreateVaccination(ctx context.Context, req *model.CreateVaccinationRequest) (*model.Vaccination, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Vaccination), args.Error(1)
}

func (m *MockService) UpdateVaccination(ctx context.Context, id string, req *model.UpdateVaccinationRequest) (*model.Vaccination, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Vaccination), args.Error(1)
}

func (m *MockService) DeleteVaccination(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Trimming Mock Methods
func (m *MockService) GetAllTrimmings(ctx context.Context) ([]model.Trimming, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Trimming), args.Error(1)
}

func (m *MockService) GetTrimmingByID(ctx context.Context, id string) (*model.Trimming, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Trimming), args.Error(1)
}

func (m *MockService) GetTrimmingsByPetID(ctx context.Context, petID string) ([]model.Trimming, error) {
	args := m.Called(ctx, petID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Trimming), args.Error(1)
}

func (m *MockService) GetTrimmingsByOwnerID(ctx context.Context, ownerID string) ([]model.Trimming, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Trimming), args.Error(1)
}

func (m *MockService) GetTrimmingsByStatus(ctx context.Context, status string) ([]model.Trimming, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Trimming), args.Error(1)
}

func (m *MockService) CreateTrimming(ctx context.Context, req *model.CreateTrimmingRequest) (*model.Trimming, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Trimming), args.Error(1)
}

func (m *MockService) UpdateTrimming(ctx context.Context, id string, req *model.UpdateTrimmingRequest) (*model.Trimming, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Trimming), args.Error(1)
}

func (m *MockService) DeleteTrimming(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Clinic Mock Methods
func (m *MockService) GetAllClinics(ctx context.Context) ([]model.Clinic, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Clinic), args.Error(1)
}

func (m *MockService) GetClinicByID(ctx context.Context, id string) (*model.Clinic, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Clinic), args.Error(1)
}

func (m *MockService) CreateClinic(ctx context.Context, req *model.CreateClinicRequest) (*model.Clinic, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Clinic), args.Error(1)
}

func (m *MockService) UpdateClinic(ctx context.Context, id string, req *model.UpdateClinicRequest) (*model.Clinic, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Clinic), args.Error(1)
}

func (m *MockService) DeleteClinic(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Staff Mock Methods
func (m *MockService) GetAllStaff(ctx context.Context) ([]model.Staff, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Staff), args.Error(1)
}

func (m *MockService) GetStaffByID(ctx context.Context, id string) (*model.Staff, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Staff), args.Error(1)
}

func (m *MockService) GetStaffByClinicID(ctx context.Context, clinicID string) ([]model.Staff, error) {
	args := m.Called(ctx, clinicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Staff), args.Error(1)
}

func (m *MockService) CreateStaff(ctx context.Context, req *model.CreateStaffRequest) (*model.Staff, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Staff), args.Error(1)
}

func (m *MockService) UpdateStaff(ctx context.Context, id string, req *model.UpdateStaffRequest) (*model.Staff, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Staff), args.Error(1)
}

func (m *MockService) DeleteStaff(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// InventoryItem Mock Methods
func (m *MockService) GetAllInventoryItems(ctx context.Context) ([]model.InventoryItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.InventoryItem), args.Error(1)
}

func (m *MockService) GetInventoryItemByID(ctx context.Context, id string) (*model.InventoryItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InventoryItem), args.Error(1)
}

func (m *MockService) GetInventoryItemsByCategory(ctx context.Context, category string) ([]model.InventoryItem, error) {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.InventoryItem), args.Error(1)
}

func (m *MockService) GetInventoryItemsByStatus(ctx context.Context, status string) ([]model.InventoryItem, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.InventoryItem), args.Error(1)
}

func (m *MockService) CreateInventoryItem(ctx context.Context, req *model.CreateInventoryItemRequest) (*model.InventoryItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InventoryItem), args.Error(1)
}

func (m *MockService) UpdateInventoryItem(ctx context.Context, id string, req *model.UpdateInventoryItemRequest) (*model.InventoryItem, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.InventoryItem), args.Error(1)
}

func (m *MockService) DeleteInventoryItem(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
