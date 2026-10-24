package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
	Breed         *string
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
	customer, err := s.findCustomerWithOwnerSync(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get health card")
	}

	ownerName := customer.DisplayName
	if ownerName == "" {
		ownerName = customer.RealName
	}

	if customer.OwnerID == nil || customer.Owner == nil {
		return &HealthCardResult{OwnerName: ownerName, Pets: []PetHealthCard{}}, nil
	}
	ownerBasic, err := commonbasic.OwnerOf(customer.Owner)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get health card")
	}
	ownerName = ownerBasic.Name

	vaccinations, err := s.vaccinationRepo.FindByOwner(ctx, clinicID, *customer.OwnerID)
	if err != nil {
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
		petBasic, basicErr := commonbasic.PetOf(p)
		if basicErr != nil {
			return nil, apperrors.Wrap(basicErr, "failed to get health card")
		}
		// EMR-290 / EMR-491 D6: 生存と確認できたペットだけを載せる（死亡と、status 未記録など確定できないものを除く）。
		if state, stateErr := sharedkernel.ClassifyPetLifeState(petBasic.Status, petBasic.DeceasedAt); stateErr != nil || state != sharedkernel.PetLifeStateAlive {
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
		// last_visit は pets.last_visit の廃止後に取得元が無く、導出もしないため常に null にする（LastVisitDate は未設定）。
		pets = append(pets, PetHealthCard{
			PetID:         p.ID,
			PetName:       petBasic.Name,
			Species:       species,
			Breed:         petBasic.Breed,
			LastVisitDate: nil,
			Vaccines:      vaccines,
		})
	}

	return &HealthCardResult{OwnerName: ownerName, Pets: pets}, nil
}
