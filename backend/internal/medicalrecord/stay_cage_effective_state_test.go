package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestStayCageEffectiveState_ConfirmedNormalCases(t *testing.T) {
	entry := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	exit := entry.Add(24 * time.Hour)
	hospitalization, doctor := uint64(7), uint64(8)
	record := int64(9)
	stay := model.StayCase{ID: 11, ClinicID: 3, StayKind: "inpatient", Version: 1, PlannedEntryAt: &entry, PlannedExitAt: &exit, DeletedAt: nil}
	admit := model.StayMovementEvent{ID: 1, ClinicID: 3, StayCaseID: 11, Revision: 1, EventKind: "admit", ActualAt: &entry, HospitalizationID: &hospitalization, DoctorID: &doctor, AdmissionRecordRef: &record, CorrectedEventID: nil}
	discharge := model.StayMovementEvent{ID: 2, ClinicID: 3, StayCaseID: 11, Revision: 2, EventKind: "discharge", ActualAt: &exit, CorrectedEventID: nil}
	for _, tc := range []struct {
		name     string
		events   []model.StayMovementEvent
		want     StayCageOccupancyState
		revision int64
	}{
		{"admit remains occupied despite past planned exit", []model.StayMovementEvent{admit}, StayCageAdmitted, 1},
		{"admit then actual discharge", []model.StayMovementEvent{admit, discharge}, StayCageReleased, 2},
		{"record order is revision not input row order", []model.StayMovementEvent{discharge, admit}, StayCageReleased, 2},
		{"planned cancellation before admission", []model.StayMovementEvent{{ID: 3, ClinicID: 3, StayCaseID: 11, Revision: 1, EventKind: "cancel", ActualAt: nil, CorrectedEventID: nil}}, StayCageReleased, 1},
		{"planned no-show before admission", []model.StayMovementEvent{{ID: 4, ClinicID: 3, StayCaseID: 11, Revision: 1, EventKind: "no_show", ActualAt: nil, CorrectedEventID: nil}}, StayCageReleased, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := projectStayCageEffectiveState(stay, tc.events)
			require.Equal(t, tc.want, got.State)
			require.Equal(t, tc.revision, got.EvidenceVersion)
			require.Equal(t, uint64(3), got.ClinicID)
			require.Equal(t, uint64(11), got.StayCaseID)
		})
	}
}
func TestStayCageEffectiveState_AmbiguousHistoryIsUnknownForThatStay(t *testing.T) {
	entry := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	early := entry.Add(-time.Hour)
	exit := entry.Add(time.Hour)
	hospitalization, doctor, corrected := uint64(7), uint64(8), uint64(1)
	record := int64(9)
	stay := model.StayCase{ID: 11, ClinicID: 3, StayKind: "inpatient", Version: 1, PlannedEntryAt: &entry, PlannedExitAt: &exit, DeletedAt: nil}
	admit := model.StayMovementEvent{ID: 1, ClinicID: 3, StayCaseID: 11, Revision: 1, EventKind: "admit", ActualAt: &entry, HospitalizationID: &hospitalization, DoctorID: &doctor, AdmissionRecordRef: &record, CorrectedEventID: nil}
	discharge := model.StayMovementEvent{ID: 2, ClinicID: 3, StayCaseID: 11, Revision: 2, EventKind: "discharge", ActualAt: &exit, CorrectedEventID: nil}
	for _, mode := range []string{"no-events-not-proven-reserved", "correction-after-admit", "discharge-only", "discharge-before-admit-time", "admit-actual-null", "discharge-actual-null", "cancel-after-admit", "no-show-after-admit", "readmit-after-discharge", "duplicate-revision", "duplicate-id", "foreign-clinic", "foreign-stay", "dangling-correction-reference", "deleted-stay"} {
		t.Run(mode, func(t *testing.T) {
			s := stay
			events := []model.StayMovementEvent{admit, discharge}
			switch mode {
			case "no-events-not-proven-reserved":
				events = nil
			case "correction-after-admit":
				events[1].EventKind = "correction"
				events[1].CorrectedEventID = &corrected
			case "discharge-only":
				events = []model.StayMovementEvent{discharge}
			case "discharge-before-admit-time":
				events[1].ActualAt = &early
			case "admit-actual-null":
				events[0].ActualAt = nil
			case "discharge-actual-null":
				events[1].ActualAt = nil
			case "cancel-after-admit":
				events[1].EventKind = "cancel"
			case "no-show-after-admit":
				events[1].EventKind = "no_show"
			case "readmit-after-discharge":
				again := admit
				again.ID = 3
				again.Revision = 3
				events = append(events, again)
			case "duplicate-revision":
				events[1].Revision = 1
			case "duplicate-id":
				events[1].ID = 1
			case "foreign-clinic":
				events[1].ClinicID = 5
			case "foreign-stay":
				events[1].StayCaseID = 50
			case "dangling-correction-reference":
				events[1].CorrectedEventID = &corrected
			case "deleted-stay":
				s.DeletedAt = &exit
			}
			require.Equal(t, StayCageUnknown, projectStayCageEffectiveState(s, events).State)
			// A malformed stay's history is never reused for another stay.
			require.Equal(t, StayCageAdmitted, projectStayCageEffectiveState(stay, []model.StayMovementEvent{admit}).State)
		})
	}
}
