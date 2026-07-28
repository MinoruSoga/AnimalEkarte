package lstep

// lstep_batch_all_clinics_test.go — BE9-2C L①: lstep_settings_service_test.go から
// LstepBatchService（L⑤で移動予定）を対象とするテストを service 側へ分離残置。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestAllClinicsFiltersBySyncEnabled: AllClinics バッチが is_sync_enabled=false のクリニックをスキップする
func TestAllClinicsFiltersBySyncEnabled(t *testing.T) {
	t.Run("RunNoShowCheckAllClinics skips disabled clinics", func(t *testing.T) {
		processed := make([]uint64, 0)
		clinicRepo := &mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 1}, {ID: 2}}, nil
			},
		}
		resRepo := &batchMockReservationRepo{
			findNoShowCandidatesFn: func(_ context.Context, clinicID uint64) ([]model.Reservation, error) {
				processed = append(processed, clinicID)
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, clinicID uint64) (bool, error) {
				return clinicID == 1, nil // clinic 1 のみ有効
			},
		}
		svc := NewLstepBatchService(
			resRepo, &batchMockTagSyncSvc{}, clinicRepo, &batchMockMedRecordRepo{},
			&batchMockAuditService{}, settingsSvc, nil,
			batchImmediateTransactor{}, &batchNoShowAuditTxLogger{},
		)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []uint64{1}, processed)
	})

	t.Run("RunDormantDetectionAllClinics skips disabled clinics", func(t *testing.T) {
		processed := make([]uint64, 0)
		clinicRepo := &mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 10}, {ID: 20}}, nil
			},
		}
		medRepo := &batchMockMedRecordRepo{
			findDormantCursorFn: func(_ context.Context, clinicID uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				processed = append(processed, clinicID)
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, clinicID uint64) (bool, error) {
				return clinicID == 20, nil // clinic 20 のみ有効
			},
		}
		svc := NewLstepBatchService(
			&batchMockReservationRepo{}, &batchMockTagSyncSvc{}, clinicRepo, medRepo,
			&batchMockAuditService{}, settingsSvc, nil,
			batchImmediateTransactor{}, &batchNoShowAuditTxLogger{},
		)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []uint64{20}, processed)
	})
}
