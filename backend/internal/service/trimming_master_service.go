// Package service provides business logic implementations for TrimmingCourse and TrimmingOption entities.
package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- TrimmingCourseService ----

type TrimmingCourseService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	GetByID(ctx context.Context, id uint64) (*model.TrimmingCourse, error)
	Create(ctx context.Context, course *model.TrimmingCourse) error
	Update(ctx context.Context, course *model.TrimmingCourse) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingCourseService struct {
	repo repository.TrimmingCourseRepository
}

func NewTrimmingCourseService(repo repository.TrimmingCourseRepository) TrimmingCourseService {
	return &trimmingCourseService{repo: repo}
}

func (s *trimmingCourseService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	return s.repo.FindAll(ctx, clinicID)
}
func (s *trimmingCourseService) GetByID(ctx context.Context, id uint64) (*model.TrimmingCourse, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *trimmingCourseService) Create(ctx context.Context, course *model.TrimmingCourse) error {
	return s.repo.Create(ctx, course)
}
func (s *trimmingCourseService) Update(ctx context.Context, course *model.TrimmingCourse) error {
	return s.repo.Update(ctx, course)
}
func (s *trimmingCourseService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *trimmingCourseService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}

// ---- TrimmingOptionService ----

type TrimmingOptionService interface {
	List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	GetByID(ctx context.Context, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, option *model.TrimmingOption) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type trimmingOptionService struct {
	repo repository.TrimmingOptionRepository
}

func NewTrimmingOptionService(repo repository.TrimmingOptionRepository) TrimmingOptionService {
	return &trimmingOptionService{repo: repo}
}

func (s *trimmingOptionService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	return s.repo.FindAll(ctx, clinicID)
}
func (s *trimmingOptionService) GetByID(ctx context.Context, id uint64) (*model.TrimmingOption, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *trimmingOptionService) Create(ctx context.Context, option *model.TrimmingOption) error {
	return s.repo.Create(ctx, option)
}
func (s *trimmingOptionService) Update(ctx context.Context, option *model.TrimmingOption) error {
	return s.repo.Update(ctx, option)
}
func (s *trimmingOptionService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *trimmingOptionService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return s.repo.Reorder(ctx, clinicID, ids)
}
