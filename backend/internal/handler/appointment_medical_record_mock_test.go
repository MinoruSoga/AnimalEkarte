package handler

// appointment_medical_record_mock_test.go — BE9-2D ⑦ carrier: 移動した
// medical_record_handler_test.go の mockMedicalRecordService の残置コピー
// （appointment_handler_test（reservation 系・残置）が使用。解消 = reservation domain 移行時）。

import (
	"context"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockMedicalRecordService struct {
	listFn                       func(ctx context.Context, clinicIDs []uint64, filters medicalrecord.MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error)
	getByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	getByIDForClinicsFn          func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error)
	countByPetFn                 func(ctx context.Context, clinicID, petID uint64) (int64, error)
	createFn                     func(ctx context.Context, clinicID uint64, input *medicalrecord.CreateMedicalRecordInput) (*model.MedicalRecord, error)
	updateFn                     func(ctx context.Context, clinicID, id uint64, input medicalrecord.UpdateMedicalRecordInput) (*model.MedicalRecord, error)
	deleteFn                     func(ctx context.Context, clinicID, id uint64) error
	updateRecommendationReasonFn func(ctx context.Context, clinicID, id uint64, input medicalrecord.UpdateRecommendationReasonInput) (*model.MedicalRecord, error)
	autoCreateFromReservationFn  func(ctx context.Context, clinicID uint64, reservation *model.Reservation)
}

func (m *mockMedicalRecordService) List(ctx context.Context, clinicIDs []uint64, filters medicalrecord.MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	return m.listFn(ctx, clinicIDs, filters, page, limit)
}

func (m *mockMedicalRecordService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
	if m.getByIDForClinicsFn != nil {
		return m.getByIDForClinicsFn(ctx, clinicIDs, id)
	}
	return nil, nil
}

func (m *mockMedicalRecordService) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	if m.countByPetFn != nil {
		return m.countByPetFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockMedicalRecordService) Create(ctx context.Context, clinicID uint64, input *medicalrecord.CreateMedicalRecordInput) (*model.MedicalRecord, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, input)
	}
	return &model.MedicalRecord{
		ID:                       1,
		ClinicID:                 clinicID,
		RecordNo:                 input.RecordNo,
		Date:                     input.Date,
		OwnerID:                  input.OwnerID,
		PetID:                    input.PetID,
		DoctorID:                 input.DoctorID,
		AppointmentID:            input.AppointmentID,
		VisitType:                &input.VisitType,
		NextVisitRecommendedDate: input.NextVisitRecommendedDate,
		RecommendationReason:     input.RecommendationReason,
		EnteredBy:                input.EnteredBy,
	}, nil
}

func (m *mockMedicalRecordService) Update(ctx context.Context, clinicID, id uint64, input medicalrecord.UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, input)
	}
	return &model.MedicalRecord{}, nil
}

func (m *mockMedicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockMedicalRecordService) CreateSubRecords(_ context.Context, _, _ uint64, _ medicalrecord.CreateSubRecordsInput) {
}

func (m *mockMedicalRecordService) AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation) {
	if m.autoCreateFromReservationFn != nil {
		m.autoCreateFromReservationFn(ctx, clinicID, reservation)
	}
}

func (m *mockMedicalRecordService) DeleteDraftFromReservation(_ context.Context, _, _ uint64) {}

func (m *mockMedicalRecordService) UpdateRecommendationReason(ctx context.Context, clinicID, id uint64, input medicalrecord.UpdateRecommendationReasonInput) (*model.MedicalRecord, error) {
	if m.updateRecommendationReasonFn != nil {
		return m.updateRecommendationReasonFn(ctx, clinicID, id, input)
	}
	return &model.MedicalRecord{}, nil
}
