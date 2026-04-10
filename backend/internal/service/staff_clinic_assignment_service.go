package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type StaffClinicAssignmentService interface {
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
	Create(ctx context.Context, assignment *model.StaffClinicAssignment) error
}

type staffClinicAssignmentService struct {
	repo repository.StaffClinicAssignmentRepository
}

func NewStaffClinicAssignmentService(repo repository.StaffClinicAssignmentRepository) StaffClinicAssignmentService {
	return &staffClinicAssignmentService{repo: repo}
}

func (s *staffClinicAssignmentService) FindByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	assignments, err := s.repo.FindByStaffID(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find clinic assignments for staff")
	}
	return assignments, nil
}

func (s *staffClinicAssignmentService) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
	if err := s.repo.Create(ctx, assignment); err != nil {
		return apperrors.Wrap(err, "failed to create staff clinic assignment")
	}
	slog.InfoContext(ctx, "staff clinic assignment created",
		slog.Uint64("staff_id", assignment.StaffID),
		slog.Uint64("clinic_id", assignment.ClinicID))
	return nil
}
