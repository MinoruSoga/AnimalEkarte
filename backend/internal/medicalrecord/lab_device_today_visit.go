package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

type labDeviceTodayVisitFinder struct {
	records MedicalRecordRepository
}

// NewLabDeviceTodayVisitFinder lists today's charts for the lab-device wait board (EMR-495: no status filter).
func NewLabDeviceTodayVisitFinder(records MedicalRecordRepository) LabDeviceTodayVisitFinder {
	return labDeviceTodayVisitFinder{records: records}
}

func (f labDeviceTodayVisitFinder) ListTodayDraft(
	ctx context.Context,
	clinicID uint64,
	date string,
) ([]LabDeviceTodayVisit, error) {
	// EMR-495: status 列の廃止に伴い draft 絞り込みは行わない（当日の全カルテを一覧する）。
	records, _, err := f.records.FindAll(ctx, []uint64{clinicID}, MedicalRecordListFilters{
		StartDate: &date,
		EndDate:   &date,
		Sort:      "pet_name",
		Order:     "asc",
	}, 1, labDeviceTodayVisitLimit)
	if err != nil {
		return nil, err
	}
	visits := make([]LabDeviceTodayVisit, 0, len(records))
	for i := range records {
		record := records[i]
		if record.PetID == nil || *record.PetID == 0 {
			continue
		}
		visit := LabDeviceTodayVisit{
			RecordID: record.ID,
			PetID:    *record.PetID,
		}
		if record.Pet != nil {
			petBasic, petErr := commonbasic.PetOf(record.Pet)
			if petErr != nil {
				return nil, petErr
			}
			visit.PetName = petBasic.Name
			// EMR-491 D6: 生存と確認できないペット（status 未記録を含む）は alive にせず、死亡と同じく選択できない扱いにする。
			state, stateErr := sharedkernel.ClassifyPetLifeState(petBasic.Status, petBasic.DeceasedAt)
			visit.PetIsDeceased = stateErr != nil || state != sharedkernel.PetLifeStateAlive
			if record.Pet.AnimalSpecies != nil {
				visit.Species = record.Pet.AnimalSpecies.Name
			}
		}
		if record.Owner != nil {
			ownerBasic, ownerErr := commonbasic.OwnerOf(record.Owner)
			if ownerErr != nil {
				return nil, ownerErr
			}
			visit.OwnerName = ownerBasic.Name
		}
		if record.Doctor != nil {
			visit.DoctorName = record.Doctor.Name
		}
		visit.VisitType = labDeviceVisitTypeLabel(record.VisitType)
		visits = append(visits, visit)
	}
	return visits, nil
}

func labDeviceVisitTypeLabel(visitType *model.VisitType) string {
	if visitType == nil {
		return ""
	}
	switch *visitType {
	case model.VisitTypeFirst:
		return "初診"
	case model.VisitTypeRevisit:
		return "再診"
	default:
		return string(*visitType)
	}
}
