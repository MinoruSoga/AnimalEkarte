package medicalrecord

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToMedicalRecordResponse_InquirySummaryIncludesNotes(t *testing.T) {
	t.Parallel()

	inquiryID := uint64(9)
	record := &model.MedicalRecord{
		ID:       42,
		ClinicID: 1,
		RecordNo: "MR-042",
		Status:   model.MedicalRecordStatusDraft,
		Version:  1,
		Inquiry: &model.Inquiry{
			ID:             inquiryID,
			ChiefComplaint: "UAT再検証 主訴",
			Notes:          "UAT再検証 治療方針",
		},
	}

	resp := toMedicalRecordResponse(record)
	require.NotNil(t, resp.Inquiry)
	assert.Equal(t, inquiryID, resp.Inquiry.ID)
	assert.Equal(t, "UAT再検証 主訴", resp.Inquiry.ChiefComplaint)
	assert.Equal(t, "UAT再検証 治療方針", resp.Inquiry.Notes)
}

func TestToMedicalRecordResponse_InquirySummaryEmptyNotes(t *testing.T) {
	t.Parallel()

	record := &model.MedicalRecord{
		ID:       1,
		ClinicID: 1,
		RecordNo: "MR-001",
		Status:   model.MedicalRecordStatusFinalized,
		Version:  2,
		Inquiry: &model.Inquiry{
			ID:             1,
			ChiefComplaint: "only chief",
			Notes:          "",
		},
	}

	resp := toMedicalRecordResponse(record)
	require.NotNil(t, resp.Inquiry)
	assert.Equal(t, "", resp.Inquiry.Notes)
	assert.Equal(t, "only chief", resp.Inquiry.ChiefComplaint)
}
