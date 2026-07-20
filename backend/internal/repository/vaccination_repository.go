package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// VaccinationRepository is a stable facade alias for medicalrecord.VaccinationRepository
// (BE9-2D roll-up: the implementation moved from internal/repository/vaccination_repository.go
// into internal/medicalrecord). Kept because lstep_tag_sync_service.go,
// lstep_delivery_trigger_service.go and liff_service.go (lstep/liff domains, staying in
// internal/service until their own BE9-2C batches) and the repositories.go aggregator still depend on it.
type VaccinationRepository = medicalrecord.VaccinationRepository

// NewVaccinationRepository constructs the vaccination repository.
func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return medicalrecord.NewVaccinationRepository(db)
}
