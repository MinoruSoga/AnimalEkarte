package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/insurance"
)

// InsuranceRepository is a stable facade alias for the insurance domain package.
type InsuranceRepository = insurance.Repository

// NewInsuranceRepository constructs the insurance repository.
func NewInsuranceRepository(db *gorm.DB) InsuranceRepository {
	return insurance.New(db)
}
