package medicalrecord

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
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

func TestToMedicalRecordResponse_IncludesVisitTypeAndChiefComplaintType(t *testing.T) {
	t.Parallel()

	visitType := model.VisitTypeFirst
	typeID := uint64(7)
	record := &model.MedicalRecord{
		ID:        42,
		ClinicID:  1,
		RecordNo:  "MR-042",
		Status:    model.MedicalRecordStatusDraft,
		Version:   1,
		VisitType: &visitType,
		Inquiry: &model.Inquiry{
			ID:                   9,
			ChiefComplaint:       "主訴",
			ChiefComplaintTypeID: &typeID,
		},
	}

	resp := toMedicalRecordResponse(record)
	require.NotNil(t, resp.VisitType)
	assert.Equal(t, "first", *resp.VisitType)
	require.NotNil(t, resp.Inquiry)
	require.NotNil(t, resp.Inquiry.ChiefComplaintTypeID)
	assert.Equal(t, typeID, *resp.Inquiry.ChiefComplaintTypeID)
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

// EMR-356: 問診履歴の体重投影。test-only decode struct で wire key を固定する
// （生成型へは触らない。実装側は additive optional の weight_measurements を返す）。
type weightMeasurementProbe struct {
	ID         json.RawMessage `json:"id"`
	Value      *float64        `json:"value"`
	Unit       string          `json:"unit"`
	RecordedAt time.Time       `json:"recorded_at"`
}

type weightProjectionProbe struct {
	WeightMeasurements []weightMeasurementProbe `json:"weight_measurements"`
}

func marshalResponseForWeightProbe(t *testing.T, record *model.MedicalRecord) ([]byte, weightProjectionProbe) {
	t.Helper()
	raw, err := json.Marshal(toMedicalRecordResponse(record))
	require.NoError(t, err)
	var probe weightProjectionProbe
	require.NoError(t, json.Unmarshal(raw, &probe))
	return raw, probe
}

// そのカルテに紐づく体重測定を id/値/単位/測定日時つきで投影し、複数測定を保持する。
func TestToMedicalRecordResponse_WeightMeasurementsProjection(t *testing.T) {
	t.Parallel()

	recordID := uint64(42)
	petID := uint64(5)
	w1, w2 := 4.2, 4300.0
	m1 := time.Date(2026, 3, 1, 9, 30, 0, 0, time.UTC)
	m2 := time.Date(2026, 3, 1, 15, 0, 0, 0, time.UTC)
	record := &model.MedicalRecord{
		ID:       recordID,
		ClinicID: 1,
		RecordNo: "MR-042",
		Status:   model.MedicalRecordStatusFinalized,
		Version:  3,
		PetID:    &petID,
		Vitals: []model.VitalRecord{
			{ID: 11, ClinicID: 1, PetID: petID, MedicalRecordID: &recordID, RecordedAt: m1, Weight: &w1, WeightUnit: model.BodyWeightUnitKg},
			{ID: 12, ClinicID: 1, PetID: petID, MedicalRecordID: &recordID, RecordedAt: m2, Weight: &w2, WeightUnit: model.BodyWeightUnitG},
		},
	}

	_, probe := marshalResponseForWeightProbe(t, record)
	require.Len(t, probe.WeightMeasurements, 2)

	assert.NotEmpty(t, probe.WeightMeasurements[0].ID)
	require.NotNil(t, probe.WeightMeasurements[0].Value)
	assert.Equal(t, 4.2, *probe.WeightMeasurements[0].Value)
	assert.Equal(t, "Kg", probe.WeightMeasurements[0].Unit)
	assert.True(t, probe.WeightMeasurements[0].RecordedAt.Equal(m1),
		"recorded_at は測定時点を同一 instant で保持する（got %s want %s）",
		probe.WeightMeasurements[0].RecordedAt, m1)

	assert.NotEmpty(t, probe.WeightMeasurements[1].ID)
	require.NotNil(t, probe.WeightMeasurements[1].Value)
	assert.Equal(t, 4300.0, *probe.WeightMeasurements[1].Value)
	assert.Equal(t, "g", probe.WeightMeasurements[1].Unit)
	assert.True(t, probe.WeightMeasurements[1].RecordedAt.Equal(m2))
}

// 投影は「親カルテ + 医院 + ペット」に正確に相関した体重行のみ。体温のみの行、
// 別カルテ/別ペット/別医院行、入院日誌（daily record）由来の行は混ぜない。
func TestToMedicalRecordResponse_WeightMeasurementsOnlyCorrelatedWeights(t *testing.T) {
	t.Parallel()

	recordID := uint64(42)
	otherRecordID := uint64(99)
	petID := uint64(5)
	otherPetID := uint64(8)
	dailyRecordID := uint64(77)
	measuredAt := time.Date(2026, 3, 1, 9, 30, 0, 0, time.UTC)
	w := 4.2
	temperature := 38.5
	record := &model.MedicalRecord{
		ID:       recordID,
		ClinicID: 1,
		RecordNo: "MR-042",
		Status:   model.MedicalRecordStatusDraft,
		Version:  1,
		PetID:    &petID,
		Vitals: []model.VitalRecord{
			{ID: 11, ClinicID: 1, PetID: petID, MedicalRecordID: &recordID, RecordedAt: measuredAt, Weight: &w, WeightUnit: model.BodyWeightUnitKg},
			// 体重を持たない（体温のみ）
			{ID: 12, ClinicID: 1, PetID: petID, MedicalRecordID: &recordID, RecordedAt: measuredAt, Temperature: &temperature, WeightUnit: model.BodyWeightUnitKg},
			// 別カルテ相関
			{ID: 13, ClinicID: 1, PetID: petID, MedicalRecordID: &otherRecordID, RecordedAt: measuredAt, Weight: &w, WeightUnit: model.BodyWeightUnitKg},
			// 別ペット相関
			{ID: 14, ClinicID: 1, PetID: otherPetID, MedicalRecordID: &recordID, RecordedAt: measuredAt, Weight: &w, WeightUnit: model.BodyWeightUnitKg},
			// 別医院相関
			{ID: 15, ClinicID: 2, PetID: petID, MedicalRecordID: &recordID, RecordedAt: measuredAt, Weight: &w, WeightUnit: model.BodyWeightUnitKg},
			// 入院日誌（daily record）由来でカルテ非相関
			{ID: 16, ClinicID: 1, PetID: petID, DailyRecordID: &dailyRecordID, RecordedAt: measuredAt, Weight: &w, WeightUnit: model.BodyWeightUnitKg},
		},
	}

	_, probe := marshalResponseForWeightProbe(t, record)
	require.Len(t, probe.WeightMeasurements, 1)
	require.NotNil(t, probe.WeightMeasurements[0].Value)
	assert.Equal(t, 4.2, *probe.WeightMeasurements[0].Value)
}

// 取得済みの測定なしは [] とし、投影を提供しない旧レスポンスと区別する。
func TestToMedicalRecordResponse_WeightMeasurementsEmptyWhenNoWeights(t *testing.T) {
	t.Parallel()

	record := &model.MedicalRecord{
		ID:       42,
		ClinicID: 1,
		RecordNo: "MR-042",
		Status:   model.MedicalRecordStatusDraft,
		Version:  1,
	}

	raw, err := json.Marshal(toMedicalRecordResponse(record))
	require.NoError(t, err)
	var probe map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &probe))
	_, exists := probe["weight_measurements"]
	assert.True(t, exists, "測定なしと投影未提供を区別する")
	assert.JSONEq(t, "[]", string(probe["weight_measurements"]))
}
