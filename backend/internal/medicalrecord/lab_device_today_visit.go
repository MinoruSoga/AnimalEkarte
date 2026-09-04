package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

type labDeviceTodayVisitFinder struct {
	records MedicalRecordRepository
}

// NewLabDeviceTodayVisitFinder lists today's draft charts for the lab-device wait board.
func NewLabDeviceTodayVisitFinder(records MedicalRecordRepository) LabDeviceTodayVisitFinder {
	return labDeviceTodayVisitFinder{records: records}
}

func (f labDeviceTodayVisitFinder) ListTodayDraft(
	ctx context.Context,
	clinicID uint64,
	date string,
) ([]LabDeviceTodayVisit, error) {
	status := model.MedicalRecordStatusDraft
	records, _, err := f.records.FindAll(ctx, []uint64{clinicID}, MedicalRecordListFilters{
		StartDate: &date,
		EndDate:   &date,
		Status:    &status,
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
			visit.PetName = record.Pet.Name
			visit.PetIsDeceased = record.Pet.Status == model.PetStatusDeceased
			if record.Pet.AnimalSpecies != nil {
				visit.Species = record.Pet.AnimalSpecies.Name
			}
		}
		if record.Owner != nil {
			visit.OwnerName = record.Owner.Name
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
