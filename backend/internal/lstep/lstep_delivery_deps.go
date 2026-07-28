package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type deliveryOwnerRepository interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

type deliveryMedicalRecordRepository interface {
	FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error)
	FindOwnersByNextVisitRecommended(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
}

type deliveryVaccinationRepository interface {
	FindOwnersByVaccineDeadline(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
}

type deliveryPetRepository interface {
	FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error)
}

type deliveryTagCacheRepository interface {
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error)
}
