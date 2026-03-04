package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository"
)

// Service contains the business logic layer.
type Service struct {
	repo                  repository.PetRepository
	ownerRepo             repository.OwnerRepository
	medicalRecordRepo     repository.MedicalRecordRepository
	reservationRepo       repository.ReservationRepository
	masterItemRepo        repository.MasterItemRepository
	hospitalizationRepo   repository.HospitalizationRepository
	accountingRepo        repository.AccountingRepository
	examinationRepo       repository.ExaminationRepository
	vaccinationRepo       repository.VaccinationRepository
	trimmingRepo          repository.TrimmingRepository
	clinicRepo            repository.ClinicRepository
	staffRepo             repository.StaffRepository
	inventoryItemRepo     repository.InventoryItemRepository
	db                    interface{ DB() *gorm.DB }
}

// New creates a new Service with the given repositories.
func New(
	repo repository.PetRepository,
	ownerRepo repository.OwnerRepository,
	medicalRecordRepo repository.MedicalRecordRepository,
	reservationRepo repository.ReservationRepository,
	masterItemRepo repository.MasterItemRepository,
	hospitalizationRepo repository.HospitalizationRepository,
	accountingRepo repository.AccountingRepository,
	examinationRepo repository.ExaminationRepository,
	vaccinationRepo repository.VaccinationRepository,
	trimmingRepo repository.TrimmingRepository,
	clinicRepo repository.ClinicRepository,
	staffRepo repository.StaffRepository,
	inventoryItemRepo repository.InventoryItemRepository,
	db interface{ DB() *gorm.DB },
) *Service {
	return &Service{
		repo:                repo,
		ownerRepo:           ownerRepo,
		medicalRecordRepo:   medicalRecordRepo,
		reservationRepo:     reservationRepo,
		masterItemRepo:      masterItemRepo,
		hospitalizationRepo: hospitalizationRepo,
		accountingRepo:      accountingRepo,
		examinationRepo:     examinationRepo,
		vaccinationRepo:     vaccinationRepo,
		trimmingRepo:        trimmingRepo,
		clinicRepo:          clinicRepo,
		staffRepo:           staffRepo,
		inventoryItemRepo:   inventoryItemRepo,
		db:                  db,
	}
}

// GetDB returns the database instance for health checks
func (s *Service) GetDB() (interface{ DB() *gorm.DB }, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	return s.db, nil
}
