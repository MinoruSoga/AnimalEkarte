// Package service keeps the legacy clinic-holiday service surface during BE9 migration.
// The implementation lives in internal/clinic.
package service

import (
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ClinicHolidayService = clinic.ClinicHolidayService

func NewClinicHolidayService(repo repository.ClinicHolidayRepository) ClinicHolidayService {
	return clinic.NewClinicHolidayService(repo)
}
