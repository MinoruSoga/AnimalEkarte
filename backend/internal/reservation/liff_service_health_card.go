package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// HealthCardResult はLIFF健康手帳の集約結果（FE zod healthCardResponseSchema と対応）。
type HealthCardResult struct {
	OwnerName string
	Pets      []PetHealthCard
}

type PetHealthCard struct {
	PetID         uint64
	PetName       string
	Species       string
	Breed         string
	LastVisitDate *time.Time
	Vaccines      []VaccineRecord
}

type VaccineRecord struct {
	VaccineName  string
	VaccinatedAt time.Time
	NextDueAt    *time.Time
}

// GetHealthCard はLINE顧客に紐付くOwnerのペット健康手帳データを集約する。
// Owner未紐付の場合はowner_name（表示名フォールバック）+ pets:[] を返す（リンク前UXを壊さないため）。
func (s *liffService) GetHealthCard(ctx context.Context, clinicID, customerID uint64) (*HealthCardResult, error) {
	customer, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get customer for health card", "error", err)
		return nil, apperrors.Wrap(err, "failed to get health card")
	}

	ownerName := customer.DisplayName
	if ownerName == "" {
		ownerName = customer.RealName
	}

	if customer.OwnerID == nil || customer.Owner == nil {
		return &HealthCardResult{OwnerName: ownerName, Pets: []PetHealthCard{}}, nil
	}
	ownerName = customer.Owner.Name

	vaccinations, err := s.vaccinationRepo.FindByOwner(ctx, clinicID, *customer.OwnerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get vaccinations for health card", "error", err)
		return nil, apperrors.Wrap(err, "failed to get health card")
	}
	vaccinesByPet := make(map[uint64][]model.Vaccination, len(vaccinations))
	for i := range vaccinations {
		v := &vaccinations[i]
		if v.PetID == nil {
			continue
		}
		vaccinesByPet[*v.PetID] = append(vaccinesByPet[*v.PetID], *v)
	}

	pets := make([]PetHealthCard, 0, len(customer.Owner.Pets))
	for i := range customer.Owner.Pets {
		p := &customer.Owner.Pets[i]
		if p.DeceasedAt != nil {
			continue
		}
		species := ""
		if p.AnimalSpecies != nil {
			species = p.AnimalSpecies.Name
		}
		records := vaccinesByPet[p.ID]
		vaccines := make([]VaccineRecord, 0, len(records))
		for j := range records {
			v := &records[j]
			name := ""
			if v.Vaccine != nil {
				name = v.Vaccine.Name
			}
			vaccines = append(vaccines, VaccineRecord{
				VaccineName:  name,
				VaccinatedAt: v.Date,
				NextDueAt:    v.NextDate,
			})
		}
		pets = append(pets, PetHealthCard{
			PetID:         p.ID,
			PetName:       p.Name,
			Species:       species,
			Breed:         p.Breed,
			LastVisitDate: p.LastVisit,
			Vaccines:      vaccines,
		})
	}

	return &HealthCardResult{OwnerName: ownerName, Pets: pets}, nil
}
