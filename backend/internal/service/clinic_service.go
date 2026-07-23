// Package service keeps the legacy clinic service surface during BE9 migration.
// The implementation lives in internal/clinic.
package service

import (
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/repository"
)

type CreateClinicInput = clinic.CreateClinicInput
type UpdateClinicInput = clinic.UpdateClinicInput
type ClinicService = clinic.ClinicService
type clinicPermissionGroupWriter = clinic.PermissionGroupWriter

func NewClinicService(
	repo repository.ClinicRepository,
	permissionGroupRepo clinicPermissionGroupWriter,
	transactor repository.Transactor,
) ClinicService {
	return clinic.NewClinicService(repo, permissionGroupRepo, transactor)
}

func buildClinicUpdate(input *UpdateClinicInput) (map[string]any, error) {
	return clinic.BuildClinicUpdate(input)
}
