package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockMedicalRecordRepository は repository.MedicalRecordRepository のテスト用モック実装（F-3統合正本）。
// fn-field 式: フック未設定時は各メソッドの既定値（多くは nil,nil / ゼロ値）を返す。
// A-5 で禁止した「未設定時に意味のある値を静かに返す」契約は導入しない — 意味のあるデフォルト
// （例: FindByID が draft レコードを返す）が必要な呼出側は、その都度明示的にフックを設定すること。
type mockMedicalRecordRepository struct {
	findAllFn                         func(ctx context.Context, clinicIDs []uint64, filters repository.MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error)
	findByIDFn                        func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	createFn                          func(ctx context.Context, record *model.MedicalRecord) error
	updateFieldsFn                    func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error)
	deleteFn                          func(ctx context.Context, clinicID, id uint64) error
	countByPetIDFn                    func(ctx context.Context, clinicID, petID uint64) (int64, error)
	findFirstVisitDateByPetIDFn       func(ctx context.Context, clinicID, petID uint64) (*time.Time, error)
	countEstimatesByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) (int64, error)
	findOwnerVisitSummaryFn           func(ctx context.Context, clinicID, ownerID uint64) (*repository.OwnerVisitSummary, error)
	countByOwnerIDFn                  func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	countByOwnerIDCallCount           int
	// 以下4フィールドは F-3 統合で追加（旧 mockMedRecordRepoForDelivery / mockMedRecordRepoForLstepVisit
	// が個別に持っていたフック）。未設定時は旧共有モックと同じ nil,nil を返す（挙動不変）。
	findLatestByOwnerFn         func(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error)
	findOwnersByFirstVisitFn    func(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	findOwnersByLastVisitDaysFn func(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error)
	findOwnersByNextVisitRecFn  func(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
}

func (m *mockMedicalRecordRepository) FindAll(ctx context.Context, clinicIDs []uint64, filters repository.MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	if m.findAllFn == nil {
		return []model.MedicalRecord{}, 0, nil
	}
	return m.findAllFn(ctx, clinicIDs, filters, page, limit)
}

func (m *mockMedicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	return m.createFn(ctx, record)
}

func (m *mockMedicalRecordRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any, _ *int) (*model.MedicalRecord, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.MedicalRecord{}, nil
}

func (m *mockMedicalRecordRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordRepository) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	if m.countByPetIDFn != nil {
		return m.countByPetIDFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockMedicalRecordRepository) FindFirstVisitDateByPetID(ctx context.Context, clinicID, petID uint64) (*time.Time, error) {
	if m.findFirstVisitDateByPetIDFn != nil {
		return m.findFirstVisitDateByPetIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) CountEstimatesByMedicalRecordID(ctx context.Context, _, medicalRecordID uint64) (int64, error) {
	if m.countEstimatesByMedicalRecordIDFn != nil {
		return m.countEstimatesByMedicalRecordIDFn(ctx, medicalRecordID)
	}
	return 0, nil
}

func (m *mockMedicalRecordRepository) FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*repository.OwnerVisitSummary, error) {
	if m.findOwnerVisitSummaryFn != nil {
		return m.findOwnerVisitSummaryFn(ctx, clinicID, ownerID)
	}
	return &repository.OwnerVisitSummary{}, nil
}

func (m *mockMedicalRecordRepository) FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error) {
	if m.findLatestByOwnerFn != nil {
		return m.findLatestByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindDormantOwnerEntries(_ context.Context, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindDormantOwnerEntriesCursor(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnersByFirstVisitFn != nil {
		return m.findOwnersByFirstVisitFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error) {
	if m.findOwnersByLastVisitDaysFn != nil {
		return m.findOwnersByLastVisitDaysFn(ctx, clinicID, exactDays, asOf)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) FindOwnersByNextVisitRecommended(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error) {
	if m.findOwnersByNextVisitRecFn != nil {
		return m.findOwnersByNextVisitRecFn(ctx, clinicID, targetDate)
	}
	return nil, nil
}

func (m *mockMedicalRecordRepository) CountByOwnerID(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	m.countByOwnerIDCallCount++
	if m.countByOwnerIDFn != nil {
		return m.countByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}

func (m *mockMedicalRecordRepository) DeleteDraftByAppointmentID(_ context.Context, _, _ uint64) error {
	return nil
}

// LockByIDForUpdate は X-11 finalize-lock テスト用に FindByID と同じ挙動へ委譲する
// （findByIDFn フックで parent status を制御しているテストがそのまま通用する）。
func (m *mockMedicalRecordRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.FindByID(ctx, clinicID, id)
}
