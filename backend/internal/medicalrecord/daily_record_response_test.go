package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToVitalRecordResponse(t *testing.T) {
	t.Run("populates all fields including staff/daily_record ids", func(t *testing.T) {
		dailyRecordID := uint64(5)
		staffID := uint64(9)
		temperature := 38.5
		heartRate := 100
		respirationRate := 20
		weight := 4.2
		recordedAt := time.Date(2026, 7, 1, 9, 30, 0, 0, time.Local)
		createdAt := time.Date(2026, 7, 1, 9, 31, 0, 0, time.Local)

		vr := &model.VitalRecord{
			ID:              1,
			DailyRecordID:   &dailyRecordID,
			RecordedAt:      recordedAt,
			Temperature:     &temperature,
			HeartRate:       &heartRate,
			RespirationRate: &respirationRate,
			Weight:          &weight,
			Notes:           "stable",
			StaffID:         &staffID,
			CreatedAt:       createdAt,
		}

		resp := toVitalRecordResponse(vr)

		assert.Equal(t, "1", resp.ID)
		assert.Equal(t, "5", resp.DailyRecordID)
		assert.Equal(t, "09:30:00", resp.Time)
		require.NotNil(t, resp.Temperature)
		assert.InDelta(t, temperature, *resp.Temperature, 0.001)
		require.NotNil(t, resp.HeartRate)
		assert.Equal(t, heartRate, *resp.HeartRate)
		require.NotNil(t, resp.RespirationRate)
		assert.Equal(t, respirationRate, *resp.RespirationRate)
		require.NotNil(t, resp.Weight)
		assert.InDelta(t, weight, *resp.Weight, 0.001)
		require.NotNil(t, resp.StaffID)
		assert.Equal(t, "9", *resp.StaffID)
		assert.Equal(t, "stable", resp.Notes)
	})

	t.Run("nil DailyRecordID and StaffID are handled without panic", func(t *testing.T) {
		vr := &model.VitalRecord{
			ID:         2,
			RecordedAt: time.Date(2026, 7, 1, 9, 30, 0, 0, time.Local),
			CreatedAt:  time.Date(2026, 7, 1, 9, 31, 0, 0, time.Local),
		}

		resp := toVitalRecordResponse(vr)

		assert.Equal(t, "2", resp.ID)
		assert.Equal(t, "", resp.DailyRecordID)
		assert.Nil(t, resp.StaffID)
		assert.Nil(t, resp.Temperature)
		assert.Nil(t, resp.HeartRate)
		assert.Nil(t, resp.RespirationRate)
		assert.Nil(t, resp.Weight)
	})
}

func TestToCareLogResponse(t *testing.T) {
	t.Run("populates all fields including staff id", func(t *testing.T) {
		staffID := uint64(7)
		cr := &model.CareLog{
			ID:            1,
			DailyRecordID: 5,
			Time:          "10:15:00",
			Type:          model.CareLogTypeFood,
			Status:        model.CareLogStatusCompleted,
			Value:         "half",
			StaffID:       &staffID,
			Notes:         "notes",
			CreatedAt:     time.Date(2026, 7, 1, 10, 16, 0, 0, time.Local),
		}

		resp := toCareLogResponse(cr)

		assert.Equal(t, "1", resp.ID)
		assert.Equal(t, "5", resp.DailyRecordID)
		assert.Equal(t, "10:15:00", resp.Time)
		assert.Equal(t, "food", resp.Type)
		assert.Equal(t, "completed", resp.Status)
		assert.Equal(t, "half", resp.Value)
		require.NotNil(t, resp.StaffID)
		assert.Equal(t, "7", *resp.StaffID)
		assert.Equal(t, "notes", resp.Notes)
	})

	t.Run("nil StaffID is handled without panic", func(t *testing.T) {
		cr := &model.CareLog{
			ID:            2,
			DailyRecordID: 5,
			Time:          "10:15:00",
			Type:          model.CareLogTypeFood,
			CreatedAt:     time.Date(2026, 7, 1, 10, 16, 0, 0, time.Local),
		}

		resp := toCareLogResponse(cr)

		assert.Equal(t, "2", resp.ID)
		assert.Nil(t, resp.StaffID)
	})
}

func TestToStaffNoteResponse(t *testing.T) {
	t.Run("populates all fields including staff id", func(t *testing.T) {
		staffID := uint64(3)
		sn := &model.StaffNote{
			ID:            1,
			DailyRecordID: 5,
			Time:          "11:00:00",
			Content:       "content",
			StaffID:       &staffID,
			CreatedAt:     time.Date(2026, 7, 1, 11, 1, 0, 0, time.Local),
		}

		resp := toStaffNoteResponse(sn)

		assert.Equal(t, "1", resp.ID)
		assert.Equal(t, "5", resp.DailyRecordID)
		assert.Equal(t, "11:00:00", resp.Time)
		assert.Equal(t, "content", resp.Content)
		require.NotNil(t, resp.StaffID)
		assert.Equal(t, "3", *resp.StaffID)
	})

	t.Run("nil StaffID is handled without panic", func(t *testing.T) {
		sn := &model.StaffNote{
			ID:            2,
			DailyRecordID: 5,
			Time:          "11:00:00",
			Content:       "content",
			CreatedAt:     time.Date(2026, 7, 1, 11, 1, 0, 0, time.Local),
		}

		resp := toStaffNoteResponse(sn)

		assert.Equal(t, "2", resp.ID)
		assert.Nil(t, resp.StaffID)
	})
}

func TestToDailyRecordResponse(t *testing.T) {
	t.Run("aggregates nested vital/care log/staff note slices", func(t *testing.T) {
		dr := &model.DailyRecord{
			ID:                1,
			HospitalizationID: 2,
			Date:              time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local),
			CreatedAt:         time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local),
			UpdatedAt:         time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local),
			VitalRecords: []model.VitalRecord{
				{ID: 1, RecordedAt: time.Date(2026, 7, 1, 9, 0, 0, 0, time.Local), CreatedAt: time.Date(2026, 7, 1, 9, 0, 0, 0, time.Local)},
			},
			CareLogs: []model.CareLog{
				{ID: 1, DailyRecordID: 1, Time: "10:00:00", Type: model.CareLogTypeFood, CreatedAt: time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)},
			},
			StaffNotes: []model.StaffNote{
				{ID: 1, DailyRecordID: 1, Time: "11:00:00", Content: "note", CreatedAt: time.Date(2026, 7, 1, 11, 0, 0, 0, time.Local)},
			},
		}

		resp := toDailyRecordResponse(dr)

		assert.Equal(t, "1", resp.ID)
		assert.Equal(t, "2", resp.HospitalizationID)
		require.Len(t, resp.VitalRecords, 1)
		require.Len(t, resp.CareLogs, 1)
		require.Len(t, resp.StaffNotes, 1)
		assert.Equal(t, "1", resp.VitalRecords[0].ID)
		assert.Equal(t, "1", resp.CareLogs[0].ID)
		assert.Equal(t, "1", resp.StaffNotes[0].ID)
	})

	t.Run("empty nested slices produce empty (not nil) response slices", func(t *testing.T) {
		dr := &model.DailyRecord{
			ID:                2,
			HospitalizationID: 3,
			Date:              time.Date(2026, 7, 2, 0, 0, 0, 0, time.Local),
			CreatedAt:         time.Date(2026, 7, 2, 0, 0, 0, 0, time.Local),
			UpdatedAt:         time.Date(2026, 7, 2, 0, 0, 0, 0, time.Local),
		}

		resp := toDailyRecordResponse(dr)

		assert.NotNil(t, resp.VitalRecords)
		assert.Empty(t, resp.VitalRecords)
		assert.NotNil(t, resp.CareLogs)
		assert.Empty(t, resp.CareLogs)
		assert.NotNil(t, resp.StaffNotes)
		assert.Empty(t, resp.StaffNotes)
	})
}
