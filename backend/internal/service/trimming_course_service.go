package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/trimming"
)

// BE9-2E compatibility aliases; remove with the service aggregator in BE9-2F.
type CreateTrimmingCourseInput = trimming.CreateTrimmingCourseInput
type UpdateTrimmingCourseInput = trimming.UpdateTrimmingCourseInput
type TrimmingCourseService = trimming.TrimmingCourseService

func NewTrimmingCourseService(
	repo repository.TrimmingCourseRepository,
	courseTypeRepo repository.TrimmingCourseTypeRepository,
) TrimmingCourseService {
	return trimming.NewTrimmingCourseService(repo, courseTypeRepo)
}
