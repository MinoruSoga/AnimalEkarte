package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository"
)

// Service contains the business logic layer.
type Service struct {
	repo              repository.PetRepository
	ownerRepo         repository.OwnerRepository
	medicalRecordRepo repository.MedicalRecordRepository
	db                interface{ DB() *gorm.DB }
}

// New creates a new Service with the given repositories.
func New(repo repository.PetRepository, ownerRepo repository.OwnerRepository, medicalRecordRepo repository.MedicalRecordRepository, db interface{ DB() *gorm.DB }) *Service {
	return &Service{
		repo:              repo,
		ownerRepo:         ownerRepo,
		medicalRecordRepo: medicalRecordRepo,
		db:                db,
	}
}

// GetDB returns the database instance for health checks
func (s *Service) GetDB() (interface{ DB() *gorm.DB }, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	return s.db, nil
}
