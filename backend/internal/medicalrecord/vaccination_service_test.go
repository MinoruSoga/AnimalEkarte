package medicalrecord

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestMedicalrecordIsolatedDockerDBHostFallback remaps DB_HOST when bare
// `docker run` cannot resolve compose hostname "db". docker-compose.yml
// publishes Postgres as host port 5434.
func TestMedicalrecordIsolatedDockerDBHostFallback(t *testing.T) {
	ensureMedicalrecordTestDBReachableFromIsolatedDocker(t)
}

func isolatedDockerDBFallback() (host, port string, ok bool) {
	if os.Getenv("TEST_DATABASE_URL") != "" {
		return "", "", false
	}
	if h := os.Getenv("DB_HOST"); h != "" && h != "db" {
		return "", "", false
	}
	if _, err := net.DefaultResolver.LookupHost(context.Background(), "db"); err == nil {
		return "", "", false
	}
	const fallbackHost = "host.docker.internal"
	const fallbackPort = "5434"
	d := net.Dialer{Timeout: 500 * time.Millisecond}
	conn, err := d.DialContext(context.Background(), "tcp", net.JoinHostPort(fallbackHost, fallbackPort))
	if err != nil {
		return "", "", false
	}
	_ = conn.Close()
	return fallbackHost, fallbackPort, true
}

func ensureMedicalrecordTestDBReachableFromIsolatedDocker(t *testing.T) {
	t.Helper()
	host, port, ok := isolatedDockerDBFallback()
	if !ok {
		return
	}
	t.Setenv("DB_HOST", host)
	t.Setenv("DB_PORT", port)
}

// mockVaccinationRepository は VaccinationRepository のテスト用モック実装
type mockVaccinationRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	lockByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	findByOwnerFn  func(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
	createFn       func(ctx context.Context, vaccination *model.Vaccination) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, cmd UpdateVaccinationInput) (*model.Vaccination, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockVaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, search string, page, limit int) ([]model.Vaccination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockVaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Vaccination{ID: id, ClinicID: clinicID, VaccineID: 1}, nil
}

func (m *mockVaccinationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	if m.lockByIDFn != nil {
		return m.lockByIDFn(ctx, clinicID, id)
	}
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Vaccination{ID: id, ClinicID: clinicID, VaccineID: 1}, nil
}

func (m *mockVaccinationRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	return m.createFn(ctx, vaccination)
}

func (m *mockVaccinationRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateVaccinationInput) (*model.Vaccination, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, cmd)
	}
	return &model.Vaccination{ID: id, ClinicID: clinicID}, nil
}

func (m *mockVaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockVaccinationRepository) FindOwnersByVaccineDeadline(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}

type permissiveVaccinationRelationVerifier struct{}

func (*permissiveVaccinationRelationVerifier) AssertOwnerInClinic(context.Context, uint64, uint64) error {
	return nil
}

func (*permissiveVaccinationRelationVerifier) FindPetOwnerInClinic(context.Context, uint64, uint64) (uint64, error) {
	return 1, nil
}

func (*permissiveVaccinationRelationVerifier) AssertMedicalRecordDoctorInClinic(context.Context, uint64, uint64) error {
	return nil
}

type permissiveVaccinationMedicalRecordLocker struct{}

func (*permissiveVaccinationMedicalRecordLocker) LockByIDForUpdate(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
}

func newTestVaccinationService(repo VaccinationRepository, vaccineRepo VaccineRepository, tagSync vaccinationTagSyncer) VaccinationService {
	return NewVaccinationService(
		repo, vaccineRepo, tagSync,
		&permissiveVaccinationRelationVerifier{},
		&permissiveVaccinationMedicalRecordLocker{},
		&mockTransactor{},
	)
}

func TestBuildVaccinationUpdate(t *testing.T) {
	medicalRecordID := uint64(1)
	petID := uint64(2)
	vaccineID := uint64(3)
	date := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	doctorID := uint64(4)
	nextDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	nextScheduleType := model.NextScheduleType("fixed")
	supplemental := "追記"
	lot1, lot2, lot3, lot4 := "L1", "L2", "L3", "L4"
	remarks := "remarks"

	tests := []struct {
		name  string
		input *UpdateVaccinationInput
		want  map[string]any
	}{
		{
			name: "all fields set",
			input: &UpdateVaccinationInput{
				MedicalRecordID: &medicalRecordID, PetID: &petID, VaccineID: &vaccineID, Date: &date,
				DoctorID: &doctorID, NextDate: &nextDate, NextScheduleType: &nextScheduleType,
				Supplemental: &supplemental, Lot1: &lot1, Lot2: &lot2, Lot3: &lot3, Lot4: &lot4, Remarks: &remarks,
			},
			want: map[string]any{
				"medical_record_id":  medicalRecordID,
				"pet_id":             petID,
				"vaccine_id":         vaccineID,
				"date":               date,
				"doctor_id":          doctorID,
				"next_date":          nextDate,
				"next_schedule_type": nextScheduleType,
				"supplemental":       supplemental,
				"lot1":               lot1,
				"lot2":               lot2,
				"lot3":               lot3,
				"lot4":               lot4,
				"remarks":            remarks,
			},
		},
		{
			name:  "no fields set returns empty map",
			input: &UpdateVaccinationInput{},
			want:  map[string]any{},
		},
		{
			name:  "only pet_id set",
			input: &UpdateVaccinationInput{PetID: &petID},
			want:  map[string]any{"pet_id": petID},
		},
		{
			name:  "only doctor_id set",
			input: &UpdateVaccinationInput{DoctorID: &doctorID},
			want:  map[string]any{"doctor_id": doctorID},
		},
		{
			name:  "only next_date set",
			input: &UpdateVaccinationInput{NextDate: &nextDate},
			want:  map[string]any{"next_date": nextDate},
		},
		{
			name:  "only next_schedule_type set",
			input: &UpdateVaccinationInput{NextScheduleType: &nextScheduleType},
			want:  map[string]any{"next_schedule_type": nextScheduleType},
		},
		{
			name:  "only lot fields set",
			input: &UpdateVaccinationInput{Lot1: &lot1, Lot3: &lot3},
			want:  map[string]any{"lot1": lot1, "lot3": lot3},
		},
		{
			name:  "only remarks set",
			input: &UpdateVaccinationInput{Remarks: &remarks},
			want:  map[string]any{"remarks": remarks},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildVaccinationUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVaccinationService_List(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name             string
		clinicID         uint64
		petID            *uint64
		ownerID          *uint64
		page             int
		limit            int
		repoVaccinations []model.Vaccination
		repoTotal        int64
		repoErr          error
		wantLen          int
		wantTotal        int64
		wantErr          bool
	}{
		{
			name:     "returns all vaccinations without filter",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoVaccinations: []model.Vaccination{
				{ID: 1, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
				{ID: 2, MedicalRecordID: ptrUint64(2), VaccineID: 2, Date: now},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    ptrUint64(10),
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoVaccinations: []model.Vaccination{
				{ID: 1, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			petID:    nil,
			ownerID:  ptrUint64(5),
			page:     1,
			limit:    20,
			repoVaccinations: []model.Vaccination{
				{ID: 2, MedicalRecordID: ptrUint64(2), VaccineID: 2, Date: now},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:             "returns empty list when no vaccinations exist",
			clinicID:         1,
			petID:            nil,
			ownerID:          nil,
			page:             1,
			limit:            20,
			repoVaccinations: []model.Vaccination{},
			repoTotal:        0,
			repoErr:          nil,
			wantLen:          0,
			wantTotal:        0,
			wantErr:          false,
		},
		{
			name:             "propagates repository error",
			clinicID:         1,
			petID:            nil,
			ownerID:          nil,
			page:             1,
			limit:            20,
			repoVaccinations: nil,
			repoTotal:        0,
			repoErr:          errors.New("db connection error"),
			wantLen:          0,
			wantTotal:        0,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPetID := (*uint64)(nil)
			capturedOwnerID := (*uint64)(nil)
			repo := &mockVaccinationRepository{
				findAllFn: func(_ context.Context, _ uint64, petID *uint64, ownerID *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
					capturedPetID = petID
					capturedOwnerID = ownerID
					return tt.repoVaccinations, tt.repoTotal, tt.repoErr
				},
			}
			svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

			vaccinations, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, nil, "", tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vaccinations, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, tt.petID, capturedPetID)
				assert.Equal(t, tt.ownerID, capturedOwnerID)
			}
		})
	}
}

func TestVaccinationService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		repoVaccination *model.Vaccination
		repoErr         error
		wantVaccination *model.Vaccination
		wantErr         error
	}{
		{
			name:            "returns vaccination when found",
			clinicID:        1,
			id:              10,
			repoVaccination: &model.Vaccination{ID: 10, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			repoErr:         nil,
			wantVaccination: &model.Vaccination{ID: 10, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			wantErr:         nil,
		},
		{
			name:            "returns not found error when vaccination does not exist",
			clinicID:        1,
			id:              999,
			repoVaccination: nil,
			repoErr:         apperrors.WrapNotFound("vaccination", "999"),
			wantVaccination: nil,
			wantErr:         apperrors.ErrNotFound,
		},
		{
			name:            "returns error on repository failure",
			clinicID:        1,
			id:              10,
			repoVaccination: nil,
			repoErr:         errors.New("db error"),
			wantVaccination: nil,
			wantErr:         errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return tt.repoVaccination, tt.repoErr
				},
			}
			svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

			vaccination, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVaccination, vaccination)
			}
		})
	}
}

func TestVaccinationService_GetByID_NotFound(t *testing.T) {
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, apperrors.WrapNotFound("vaccination", "999")
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, vaccination)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestVaccinationService_Update_LocksRelationsBeforeVaccination(t *testing.T) {
	events := make([]string, 0, 4)
	record := &model.Vaccination{ID: 10, ClinicID: 1, VaccineID: 20}
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			events = append(events, "snapshot")
			return record, nil
		},
		lockByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			events = append(events, "vaccination_lock")
			return record, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
			events = append(events, "update")
			return record, nil
		},
	}
	vaccineRepo := &mockVaccineRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
			events = append(events, "vaccine_lock")
			return &model.Vaccine{ID: id}, nil
		},
	}
	svc := NewVaccinationService(
		repo,
		vaccineRepo,
		nil,
		&permissiveVaccinationRelationVerifier{},
		&permissiveVaccinationMedicalRecordLocker{},
		&mockTransactor{},
	)
	remarks := "updated"

	updated, err := svc.Update(context.Background(), 1, record.ID, &UpdateVaccinationInput{Remarks: &remarks})

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, []string{"snapshot", "vaccine_lock", "vaccination_lock", "update"}, events)
}

func TestVaccinationService_Update_RejectsConcurrentRelationChangeAfterValidation(t *testing.T) {
	updated := false
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: 1, VaccineID: 20}, nil
		},
		lockByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: 1, VaccineID: 21}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
			updated = true
			return &model.Vaccination{}, nil
		},
	}
	svc := NewVaccinationService(
		repo,
		okVaccineRepo(),
		nil,
		&permissiveVaccinationRelationVerifier{},
		&permissiveVaccinationMedicalRecordLocker{},
		&mockTransactor{},
	)
	remarks := "updated"

	got, err := svc.Update(context.Background(), 1, 10, &UpdateVaccinationInput{Remarks: &remarks})

	assert.Nil(t, got)
	assert.True(t, apperrors.IsConflict(err))
	assert.False(t, updated)
}

func TestVaccinationService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		clinicID uint64
		input    *CreateVaccinationInput
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates vaccination successfully",
			clinicID: 1,
			input: &CreateVaccinationInput{
				MedicalRecordID: ptrUint64(1),
				VaccineID:       1,
				Date:            now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when vaccine_id is zero",
			clinicID: 1,
			input:    &CreateVaccinationInput{VaccineID: 0, Date: now},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: &CreateVaccinationInput{
				MedicalRecordID: ptrUint64(1),
				VaccineID:       2,
				Date:            now,
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				createFn: func(_ context.Context, vaccination *model.Vaccination) error {
					vaccination.ID = 10
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
					return &model.Vaccination{ID: id, MedicalRecordID: tt.input.MedicalRecordID, VaccineID: tt.input.VaccineID, Date: tt.input.Date}, nil
				},
			}
			svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

			vaccination, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccination)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccination)
			}
		})
	}
}

func TestVaccinationService_Create_RejectsFutureVaccinationDate(t *testing.T) {
	nowJST := time.Now().In(config.JST)
	today := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 10, 0, 0, 0, config.JST)
	tomorrow := today.AddDate(0, 0, 1)
	yesterday := today.AddDate(0, 0, -1)
	nextDate := today.AddDate(0, 1, 0)
	sameDayNext := today

	tests := []struct {
		name        string
		date        time.Time
		nextDate    *time.Time
		wantErr     bool
		wantInvalid bool
		wantMsg     string
		wantCreate  bool
	}{
		{
			name:        "rejects tomorrow JST",
			date:        tomorrow,
			wantErr:     true,
			wantInvalid: true,
			wantMsg:     "今日以前",
			wantCreate:  false,
		},
		{
			name:       "allows today JST with nil next_date",
			date:       today,
			wantCreate: true,
		},
		{
			name:       "allows future next_date when date is today",
			date:       today,
			nextDate:   &nextDate,
			wantCreate: true,
		},
		{
			name:        "rejects next_date before vaccination date",
			date:        today,
			nextDate:    &yesterday,
			wantErr:     true,
			wantInvalid: true,
			wantMsg:     "次回予定日は接種日より後",
			wantCreate:  false,
		},
		{
			name:        "rejects next_date equal to vaccination date",
			date:        today,
			nextDate:    &sameDayNext,
			wantErr:     true,
			wantInvalid: true,
			wantMsg:     "次回予定日は接種日より後",
			wantCreate:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalled := false
			repo := &mockVaccinationRepository{
				createFn: func(_ context.Context, vaccination *model.Vaccination) error {
					createCalled = true
					vaccination.ID = 10
					return nil
				},
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
					return &model.Vaccination{ID: id, VaccineID: 1, Date: tt.date, NextDate: tt.nextDate}, nil
				},
			}
			svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

			got, err := svc.Create(context.Background(), 1, &CreateVaccinationInput{
				VaccineID: 1,
				Date:      tt.date,
				NextDate:  tt.nextDate,
			})

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
				}
				if tt.wantMsg != "" {
					assert.Contains(t, err.Error(), tt.wantMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
			assert.Equal(t, tt.wantCreate, createCalled)
		})
	}
}

func TestVaccinationService_Update_RejectsFutureVaccinationDate(t *testing.T) {
	nowJST := time.Now().In(config.JST)
	today := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 10, 0, 0, 0, config.JST)
	tomorrow := today.AddDate(0, 0, 1)
	yesterday := today.AddDate(0, 0, -1)
	dayBeforeYesterday := today.AddDate(0, 0, -2)
	nextDate := today.AddDate(0, 1, 0)
	sameDayNext := today
	supplemental := "追記"

	tests := []struct {
		name           string
		input          UpdateVaccinationInput
		storedDate     time.Time
		storedNextDate *time.Time
		wantErr        bool
		wantInvalid    bool
		wantMsg        string
		wantUpdate     bool
	}{
		{
			name: "rejects tomorrow JST",
			input: UpdateVaccinationInput{
				Date: &tomorrow,
			},
			wantErr:     true,
			wantInvalid: true,
			wantMsg:     "今日以前",
			wantUpdate:  false,
		},
		{
			name: "allows today JST with nil next_date",
			input: UpdateVaccinationInput{
				Date: &today,
			},
			wantUpdate: true,
		},
		{
			name: "allows future next_date when date is today",
			input: UpdateVaccinationInput{
				Date:     &today,
				NextDate: &nextDate,
			},
			wantUpdate: true,
		},
		{
			name: "allows next_date after stored date when date is omitted",
			input: UpdateVaccinationInput{
				NextDate: &nextDate,
			},
			storedDate: yesterday,
			wantUpdate: true,
		},
		{
			name: "rejects next_date before stored date when date is omitted",
			input: UpdateVaccinationInput{
				NextDate: &dayBeforeYesterday,
			},
			storedDate:  yesterday,
			wantErr:     true,
			wantInvalid: true,
			wantMsg:     "次回予定日は接種日より後",
			wantUpdate:  false,
		},
		{
			name: "rejects date moved to or after stored next_date",
			input: UpdateVaccinationInput{
				Date: &today,
			},
			storedDate:     dayBeforeYesterday,
			storedNextDate: &today,
			wantErr:        true,
			wantInvalid:    true,
			wantMsg:        "次回予定日は接種日より後",
			wantUpdate:     false,
		},
		{
			name: "rejects next_date equal to vaccination date",
			input: UpdateVaccinationInput{
				Date:     &today,
				NextDate: &sameDayNext,
			},
			wantErr:     true,
			wantInvalid: true,
			wantMsg:     "次回予定日は接種日より後",
			wantUpdate:  false,
		},
		{
			name: "omitting date does not inspect stored date",
			input: UpdateVaccinationInput{
				Supplemental: &supplemental,
			},
			storedDate: tomorrow,
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			stored := tt.storedDate
			if stored.IsZero() {
				stored = today.AddDate(0, 0, -1)
			}
			repo := &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Vaccination, error) {
					return &model.Vaccination{
						ID:        id,
						ClinicID:  clinicID,
						VaccineID: 1,
						Date:      stored,
						NextDate:  tt.storedNextDate,
					}, nil
				},
				updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
					updateCalled = true
					return &model.Vaccination{ID: id, Date: stored}, nil
				},
			}
			svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

			got, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input but got: %v", err)
				}
				if tt.wantMsg != "" {
					assert.Contains(t, err.Error(), tt.wantMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
			assert.Equal(t, tt.wantUpdate, updateCalled)
		})
	}
}

func TestVaccinationService_Create_PostCreateFindByIDErrorRollsBack(t *testing.T) {
	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			vaccination.ID = 10
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.Create(context.Background(), 1, &CreateVaccinationInput{VaccineID: 1, Date: time.Now()})

	assert.Error(t, err)
	assert.Nil(t, vaccination)
}

func TestVaccinationService_Create_PostCreateFindByIDNilRollsBack(t *testing.T) {
	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			vaccination.ID = 11
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, nil
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.Create(context.Background(), 1, &CreateVaccinationInput{VaccineID: 1, Date: time.Now()})

	assert.Error(t, err)
	assert.Nil(t, vaccination)
}

func TestVaccinationService_Create_SyncsVaccineTagBestEffort(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	var syncedClinicID, syncedOwnerID, syncedVaccinationID uint64

	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			vaccination.ID = 30
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID:        id,
				PetID:     &petID,
				VaccineID: 3,
				Date:      time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
				Pet: &model.Pet{
					ID: petID, ClinicID: 1, OwnerID: ownerID,
					Owner: &model.Owner{ID: ownerID, ClinicID: 1},
				},
			}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncVaccineTagFn: func(_ context.Context, clinicID, ownerID, vaccinationID uint64) error {
			syncedClinicID = clinicID
			syncedOwnerID = ownerID
			syncedVaccinationID = vaccinationID
			return errors.New("sync failed")
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), tagSync)

	vaccination, err := svc.Create(context.Background(), 1, &CreateVaccinationInput{
		PetID:     &petID,
		VaccineID: 3,
		Date:      time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
	})

	assert.NoError(t, err)
	assert.NotNil(t, vaccination)
	assert.Equal(t, uint64(1), syncedClinicID)
	assert.Equal(t, ownerID, syncedOwnerID)
	assert.Equal(t, uint64(30), syncedVaccinationID)
}

func TestVaccinationService_Update(t *testing.T) {
	now := time.Now()
	supplemental := "追記情報"
	tests := []struct {
		name    string
		input   UpdateVaccinationInput
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates vaccination successfully",
			input: UpdateVaccinationInput{
				Date:         &now,
				Supplemental: &supplemental,
			},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateVaccinationInput{},
			repoErr: nil,
			wantErr: true,
			wantNF:  false,
		},
		{
			name: "returns not found error when vaccination does not exist",
			input: UpdateVaccinationInput{
				Supplemental: &supplemental,
			},
			repoErr: apperrors.WrapNotFound("vaccination", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateVaccinationInput{
				Supplemental: &supplemental,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Vaccination{ID: 1}, nil
				},
			}
			svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

			vaccination, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccination)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccination)
			}
		})
	}
}

func TestVaccinationService_Update_FindByIDError(t *testing.T) {
	supplemental := "追記"
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, apperrors.WrapNotFound("vaccination", "1")
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.Update(context.Background(), 1, 1, &UpdateVaccinationInput{Supplemental: &supplemental})

	assert.Error(t, err)
	assert.Nil(t, vaccination)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestVaccinationService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes vaccination successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when vaccination does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("vaccination", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
			wantNF:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVaccinationService_Delete_FindByIDError(t *testing.T) {
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, apperrors.WrapNotFound("vaccination", "999")
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), nil)

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestVaccinationService_Update_ResyncsOwnerVaccineTags(t *testing.T) {
	ownerID := uint64(10)
	var syncedOwnerID uint64
	supplemental := "updated"

	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id}, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID:    id,
				PetID: ptrUint64(20),
				Pet: &model.Pet{
					ID: 20, ClinicID: 1, OwnerID: ownerID,
					Owner: &model.Owner{ID: ownerID, ClinicID: 1},
				},
			}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerVaccineTagsFn: func(_ context.Context, _, ownerID uint64) error {
			syncedOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), tagSync)

	vaccination, err := svc.Update(context.Background(), 1, 30, &UpdateVaccinationInput{Supplemental: &supplemental})

	assert.NoError(t, err)
	assert.NotNil(t, vaccination)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestVaccinationService_Delete_ResyncsOwnerVaccineTagsAfterDelete(t *testing.T) {
	ownerID := uint64(10)
	deleted := false
	syncedAfterDelete := false

	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID:    id,
				PetID: ptrUint64(20),
				Pet: &model.Pet{
					ID: 20, ClinicID: 1, OwnerID: ownerID,
					Owner: &model.Owner{ID: ownerID, ClinicID: 1},
				},
			}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleted = true
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerVaccineTagsFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			syncedAfterDelete = deleted
			assert.Equal(t, ownerID, syncedOwnerID)
			return nil
		},
	}
	svc := newTestVaccinationService(repo, okVaccineRepo(), tagSync)

	err := svc.Delete(context.Background(), 1, 30)

	assert.NoError(t, err)
	assert.True(t, syncedAfterDelete)
}

// SEC-DUR-01-MR-T1: 譲渡後もvaccination更新はsnapshot ownerとcurrent pet ownerの差を許容し、clinic外は拒否する。
func TestVaccinationService_Update_AllowsHistoricalOwnerAfterPetTransfer(t *testing.T) {
	const (
		clinicID        = uint64(1)
		vaccinationID   = uint64(5)
		petID           = uint64(10)
		previousOwnerID = uint64(20)
		currentOwnerID  = uint64(21)
		recordID        = uint64(30)
	)

	remarks := "post-transfer remarks"
	updateCalls := 0
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID: id, ClinicID: clinicID, PetID: ptrUint64(petID),
				MedicalRecordID: ptrUint64(recordID), VaccineID: 1,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, cmd UpdateVaccinationInput) (*model.Vaccination, error) {
			updateCalls++
			remarks := ""
			if cmd.Remarks != nil {
				remarks = *cmd.Remarks
			}
			return &model.Vaccination{
				ID: id, ClinicID: clinicID, PetID: ptrUint64(petID),
				MedicalRecordID: ptrUint64(recordID), VaccineID: 1,
				Remarks: remarks,
			}, nil
		},
	}

	t.Run("same_clinic_transfer_succeeds", func(t *testing.T) {
		updateCalls = 0
		// previousOwner must be present as a map value so AssertOwnerInClinic accepts the snapshot owner.
		verifier := &vaccinationRelationVerifierStub{
			petOwners: map[uint64]uint64{petID: currentOwnerID, 999: previousOwnerID},
		}
		locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
			recordID: {
				ID: recordID, ClinicID: clinicID,
				PetID: ptrUint64(petID), OwnerID: ptrUint64(previousOwnerID),
				Status: model.MedicalRecordStatusDraft,
			},
		}}
		svc := NewVaccinationService(repo, okVaccineRepo(), nil, verifier, locker, vaccinationTestTransactor{})

		got, err := svc.Update(context.Background(), clinicID, vaccinationID, &UpdateVaccinationInput{Remarks: &remarks})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 1, updateCalls)
		assert.Equal(t, remarks, got.Remarks)
	})

	t.Run("rejects_foreign_snapshot_owner", func(t *testing.T) {
		updateCalls = 0
		verifier := &vaccinationRelationVerifierStub{
			petOwners: map[uint64]uint64{petID: currentOwnerID},
		}
		locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
			recordID: {
				ID: recordID, ClinicID: clinicID,
				PetID: ptrUint64(petID), OwnerID: ptrUint64(previousOwnerID),
				Status: model.MedicalRecordStatusDraft,
			},
		}}
		svc := NewVaccinationService(repo, okVaccineRepo(), nil, verifier, locker, vaccinationTestTransactor{})

		got, err := svc.Update(context.Background(), clinicID, vaccinationID, &UpdateVaccinationInput{Remarks: &remarks})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, updateCalls)
	})

	t.Run("rejects_foreign_pet", func(t *testing.T) {
		updateCalls = 0
		verifier := &vaccinationRelationVerifierStub{
			petOwners: map[uint64]uint64{999: previousOwnerID},
		}
		locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
			recordID: {
				ID: recordID, ClinicID: clinicID,
				PetID: ptrUint64(petID), OwnerID: ptrUint64(previousOwnerID),
				Status: model.MedicalRecordStatusDraft,
			},
		}}
		svc := NewVaccinationService(repo, okVaccineRepo(), nil, verifier, locker, vaccinationTestTransactor{})

		got, err := svc.Update(context.Background(), clinicID, vaccinationID, &UpdateVaccinationInput{Remarks: &remarks})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, updateCalls)
	})
}
