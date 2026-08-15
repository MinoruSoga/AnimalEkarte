package trimming

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingAuditTxContextKey struct{}

type trimmingAuditRecorder struct {
	entries  []*AuditEntry
	err      error
	sawTxCtx bool
}

func (r *trimmingAuditRecorder) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	r.sawTxCtx = ctx.Value(trimmingAuditTxContextKey{}) == true
	r.entries = append(r.entries, entry)
	return r.err
}

type trimmingAuditRollbackTransactor struct {
	journal    *[]string
	rolledBack bool
}

func (t *trimmingAuditRollbackTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	snapshot := append([]string(nil), (*t.journal)...)
	err := fn(context.WithValue(ctx, trimmingAuditTxContextKey{}, true))
	if err != nil {
		*t.journal = snapshot
		t.rolledBack = true
	}
	return err
}

func newTrimmingAuditTestService(
	reservationRepo TrimmingReservationRepository,
	detailRepo AppointmentTrimmingDetailRepository,
	transactor Transactor,
	auditTx AuditTxLogger,
) TrimmingService {
	return NewTrimmingServiceWithAudit(
		reservationRepo,
		&mockTrimmingReservationTypeRepository{},
		newAcceptingTrimmingStaffRepository(),
		&mockTrimmingUnavailableTimeRepository{},
		detailRepo,
		newActiveTrimmingCourseRepository(),
		newActiveTrimmingOptionRepository(),
		transactor,
		auditTx,
	)
}

func newTrimmingAuditAppointment(id, clinicID uint64) *model.Reservation {
	start := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	return &model.Reservation{
		ID:                id,
		ClinicID:          clinicID,
		ReservationTypeID: 9,
		ReservationType: &model.ReservationType{
			ID:       9,
			ClinicID: clinicID,
			Category: model.ReservationTypeCategoryTrimming,
			IsActive: true,
		},
		StartTime: start,
		EndTime:   end,
		Status:    model.ReservationStatusPending,
		Source:    model.ReservationSourceManual,
	}
}

func cloneTrimmingAuditAppointment(appointment *model.Reservation) *model.Reservation {
	if appointment == nil {
		return nil
	}
	cloned := *appointment
	if appointment.TrimmingDetail != nil {
		detail := *appointment.TrimmingDetail
		detail.Options = append([]model.TrimmingOption(nil), appointment.TrimmingDetail.Options...)
		cloned.TrimmingDetail = &detail
	}
	return &cloned
}

func requireTrimmingAuditContract(
	t *testing.T,
	recorder *trimmingAuditRecorder,
	clinicID, actorID, appointmentID uint64,
	action, mutationPath string,
) *AuditEntry {
	t.Helper()
	require.True(t, recorder.sawTxCtx, "audit must execute inside the business transaction")
	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	require.NotNil(t, entry)
	require.NotNil(t, entry.ClinicID)
	assert.Equal(t, clinicID, *entry.ClinicID)
	require.NotNil(t, entry.ActorID)
	assert.Equal(t, actorID, *entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	assert.Equal(t, action, entry.Action)
	assert.Equal(t, model.AuditResourceTrimming, entry.Resource)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, appointmentID, *entry.ResourceID)
	assert.Equal(t, map[string]any{"mutation_path": mutationPath}, entry.Metadata)
	return entry
}

func TestTrimmingService_AuditContract_CreateAppointment(t *testing.T) {
	const clinicID, actorID, appointmentID = uint64(7), uint64(42), uint64(77)
	appointment := newTrimmingAuditAppointment(appointmentID, clinicID)
	journal := make([]string, 0, 3)
	transactor := &trimmingAuditRollbackTransactor{journal: &journal}
	recorder := &trimmingAuditRecorder{}
	var detailState *model.AppointmentTrimmingDetail

	reservationRepo := &mockTrimmingReservationRepository{
		createFn: func(_ context.Context, created *model.Reservation) error {
			journal = append(journal, "appointment.create")
			created.ID = appointmentID
			appointment = cloneTrimmingAuditAppointment(created)
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return cloneTrimmingAuditAppointment(appointment), nil
		},
	}
	detailRepo := &mockTrimmingDetailRepository{
		createFn: func(_ context.Context, detail *model.AppointmentTrimmingDetail) error {
			journal = append(journal, "detail.create")
			cloned := *detail
			detailState = &cloned
			return nil
		},
		setOptionsFn: func(_ context.Context, _, _ uint64, optionIDs []uint64) error {
			journal = append(journal, "options.replace")
			detailState.Options = []model.TrimmingOption{{ID: optionIDs[0]}, {ID: optionIDs[1]}}
			return nil
		},
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			cloned := *detailState
			cloned.Options = append([]model.TrimmingOption(nil), detailState.Options...)
			return &cloned, nil
		},
	}
	svc := newTrimmingAuditTestService(reservationRepo, detailRepo, transactor, recorder)

	result, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
		ActorID:           ptrUint64(actorID),
		ReservationTypeID: appointment.ReservationTypeID,
		StartTime:         appointment.StartTime,
		EndTime:           appointment.EndTime,
		StyleRequest:      "顔まわりを短く",
		OptionIDs:         []uint64{31, 12},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	entry := requireTrimmingAuditContract(
		t, recorder, clinicID, actorID, appointmentID,
		model.AuditActionTrimmingCreate, trimmingAuditMutationCreateAppointment,
	)
	assert.Nil(t, entry.OldValue)
	newValue, ok := entry.NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, appointmentID, newValue["appointment_id"])
	assert.Equal(t, "顔まわりを短く", newValue["style_request"])
	assert.Equal(t, []uint64{12, 31}, newValue["option_ids"])
	assert.Equal(t, true, newValue["has_detail"])
}

func TestTrimmingService_AuditContract_CreateDetailForExistingAppointment(t *testing.T) {
	const clinicID, actorID, appointmentID = uint64(7), uint64(43), uint64(78)
	appointment := newTrimmingAuditAppointment(appointmentID, clinicID)
	journal := make([]string, 0, 2)
	transactor := &trimmingAuditRollbackTransactor{journal: &journal}
	recorder := &trimmingAuditRecorder{}
	var detailState *model.AppointmentTrimmingDetail
	findDetailCalls := 0

	reservationRepo := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return cloneTrimmingAuditAppointment(appointment), nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return cloneTrimmingAuditAppointment(appointment), nil
		},
	}
	detailRepo := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			findDetailCalls++
			if detailState == nil {
				return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
			}
			cloned := *detailState
			return &cloned, nil
		},
		createFn: func(_ context.Context, detail *model.AppointmentTrimmingDetail) error {
			journal = append(journal, "detail.create")
			cloned := *detail
			detailState = &cloned
			return nil
		},
	}
	svc := newTrimmingAuditTestService(reservationRepo, detailRepo, transactor, recorder)

	result, err := svc.Create(context.Background(), clinicID, &CreateTrimmingInput{
		ActorID:       ptrUint64(actorID),
		AppointmentID: ptrUint64(appointmentID),
		StyleRequest:  "既存予約へ詳細を追加",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, findDetailCalls, 2)
	entry := requireTrimmingAuditContract(
		t, recorder, clinicID, actorID, appointmentID,
		model.AuditActionTrimmingCreate, trimmingAuditMutationCreateDetail,
	)
	oldValue, ok := entry.OldValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, false, oldValue["has_detail"])
	newValue, ok := entry.NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, newValue["has_detail"])
	assert.Equal(t, "既存予約へ詳細を追加", newValue["style_request"])
}

func TestTrimmingService_AuditContract_Update(t *testing.T) {
	const clinicID, actorID, appointmentID = uint64(7), uint64(44), uint64(79)
	appointment := newTrimmingAuditAppointment(appointmentID, clinicID)
	detailState := &model.AppointmentTrimmingDetail{
		ClinicID:      clinicID,
		AppointmentID: appointmentID,
		StyleRequest:  "変更前",
		Options:       []model.TrimmingOption{{ID: 8}},
	}
	journal := make([]string, 0, 2)
	transactor := &trimmingAuditRollbackTransactor{journal: &journal}
	recorder := &trimmingAuditRecorder{}

	reservationRepo := &mockTrimmingReservationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return cloneTrimmingAuditAppointment(appointment), nil
		},
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return cloneTrimmingAuditAppointment(appointment), nil
		},
	}
	detailRepo := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			cloned := *detailState
			cloned.Options = append([]model.TrimmingOption(nil), detailState.Options...)
			return &cloned, nil
		},
		updateFn: func(_ context.Context, detail *model.AppointmentTrimmingDetail) error {
			journal = append(journal, "detail.update")
			cloned := *detail
			detailState = &cloned
			return nil
		},
		setOptionsFn: func(_ context.Context, _, _ uint64, optionIDs []uint64) error {
			journal = append(journal, "options.replace")
			detailState.Options = []model.TrimmingOption{{ID: optionIDs[0]}, {ID: optionIDs[1]}}
			return nil
		},
	}
	svc := newTrimmingAuditTestService(reservationRepo, detailRepo, transactor, recorder)
	styleRequest := "変更後"
	optionIDs := []uint64{21, 3}

	result, err := svc.Update(context.Background(), clinicID, appointmentID, &UpdateTrimmingInput{
		ActorID:      ptrUint64(actorID),
		StyleRequest: &styleRequest,
		OptionIDs:    &optionIDs,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	entry := requireTrimmingAuditContract(
		t, recorder, clinicID, actorID, appointmentID,
		model.AuditActionTrimmingUpdate, trimmingAuditMutationUpdate,
	)
	oldValue, ok := entry.OldValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "変更前", oldValue["style_request"])
	assert.Equal(t, []uint64{8}, oldValue["option_ids"])
	newValue, ok := entry.NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "変更後", newValue["style_request"])
	assert.Equal(t, []uint64{3, 21}, newValue["option_ids"])
}

func TestTrimmingService_AuditContract_Delete(t *testing.T) {
	const clinicID, actorID, appointmentID = uint64(7), uint64(45), uint64(80)
	appointment := newTrimmingAuditAppointment(appointmentID, clinicID)
	detail := &model.AppointmentTrimmingDetail{
		ClinicID:      clinicID,
		AppointmentID: appointmentID,
		StyleRequest:  "削除前",
		Options:       []model.TrimmingOption{{ID: 5}},
	}
	journal := make([]string, 0, 1)
	transactor := &trimmingAuditRollbackTransactor{journal: &journal}
	recorder := &trimmingAuditRecorder{}
	deleted := false
	reservationRepo := &mockTrimmingReservationRepository{
		lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return cloneTrimmingAuditAppointment(appointment), nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			journal = append(journal, "appointment.delete")
			deleted = true
			return nil
		},
	}
	detailRepo := &mockTrimmingDetailRepository{
		findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
			cloned := *detail
			cloned.Options = append([]model.TrimmingOption(nil), detail.Options...)
			return &cloned, nil
		},
	}
	svc := newTrimmingAuditTestService(reservationRepo, detailRepo, transactor, recorder)

	err := svc.Delete(context.Background(), clinicID, appointmentID, ptrUint64(actorID))

	require.NoError(t, err)
	assert.True(t, deleted)
	entry := requireTrimmingAuditContract(
		t, recorder, clinicID, actorID, appointmentID,
		model.AuditActionTrimmingDelete, trimmingAuditMutationDelete,
	)
	oldValue, ok := entry.OldValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "削除前", oldValue["style_request"])
	assert.Equal(t, []uint64{5}, oldValue["option_ids"])
	assert.Nil(t, entry.NewValue)
}

func TestTrimmingService_AuditFailureRollsBackEachWritePath(t *testing.T) {
	sentinel := errors.New("audit persistence failed")

	tests := []struct {
		name string
		run  func(*trimmingAuditRollbackTransactor, *trimmingAuditRecorder, *[]string) error
	}{
		{
			name: "create appointment",
			run: func(tx *trimmingAuditRollbackTransactor, audit *trimmingAuditRecorder, journal *[]string) error {
				appointment := newTrimmingAuditAppointment(81, 7)
				var detail *model.AppointmentTrimmingDetail
				reservationRepo := &mockTrimmingReservationRepository{
					createFn: func(_ context.Context, created *model.Reservation) error {
						*journal = append(*journal, "appointment.create")
						created.ID = appointment.ID
						return nil
					},
					findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
						return cloneTrimmingAuditAppointment(appointment), nil
					},
				}
				detailRepo := &mockTrimmingDetailRepository{
					createFn: func(_ context.Context, created *model.AppointmentTrimmingDetail) error {
						*journal = append(*journal, "detail.create")
						cloned := *created
						detail = &cloned
						return nil
					},
					findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
						return detail, nil
					},
				}
				svc := newTrimmingAuditTestService(reservationRepo, detailRepo, tx, audit)
				_, err := svc.Create(context.Background(), 7, &CreateTrimmingInput{
					ActorID:           ptrUint64(42),
					ReservationTypeID: 9,
					StartTime:         appointment.StartTime,
					EndTime:           appointment.EndTime,
				})
				return err
			},
		},
		{
			name: "create detail for existing appointment",
			run: func(tx *trimmingAuditRollbackTransactor, audit *trimmingAuditRecorder, journal *[]string) error {
				appointment := newTrimmingAuditAppointment(82, 7)
				var detail *model.AppointmentTrimmingDetail
				reservationRepo := &mockTrimmingReservationRepository{
					findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
						return cloneTrimmingAuditAppointment(appointment), nil
					},
					lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
						return cloneTrimmingAuditAppointment(appointment), nil
					},
				}
				detailRepo := &mockTrimmingDetailRepository{
					findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
						if detail == nil {
							return nil, apperrors.WrapNotFound("appointment_trimming_detail", "missing")
						}
						return detail, nil
					},
					createFn: func(_ context.Context, created *model.AppointmentTrimmingDetail) error {
						*journal = append(*journal, "detail.create")
						cloned := *created
						detail = &cloned
						return nil
					},
				}
				svc := newTrimmingAuditTestService(reservationRepo, detailRepo, tx, audit)
				_, err := svc.Create(context.Background(), 7, &CreateTrimmingInput{
					ActorID:       ptrUint64(42),
					AppointmentID: ptrUint64(appointment.ID),
				})
				return err
			},
		},
		{
			name: "update",
			run: func(tx *trimmingAuditRollbackTransactor, audit *trimmingAuditRecorder, journal *[]string) error {
				appointment := newTrimmingAuditAppointment(83, 7)
				detail := &model.AppointmentTrimmingDetail{ClinicID: 7, AppointmentID: appointment.ID}
				reservationRepo := &mockTrimmingReservationRepository{
					findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
						return cloneTrimmingAuditAppointment(appointment), nil
					},
					lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
						return cloneTrimmingAuditAppointment(appointment), nil
					},
				}
				detailRepo := &mockTrimmingDetailRepository{
					findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
						cloned := *detail
						return &cloned, nil
					},
					updateFn: func(_ context.Context, updated *model.AppointmentTrimmingDetail) error {
						*journal = append(*journal, "detail.update")
						cloned := *updated
						detail = &cloned
						return nil
					},
				}
				styleRequest := "rollback"
				svc := newTrimmingAuditTestService(reservationRepo, detailRepo, tx, audit)
				_, err := svc.Update(context.Background(), 7, appointment.ID, &UpdateTrimmingInput{
					ActorID:      ptrUint64(42),
					StyleRequest: &styleRequest,
				})
				return err
			},
		},
		{
			name: "delete",
			run: func(tx *trimmingAuditRollbackTransactor, audit *trimmingAuditRecorder, journal *[]string) error {
				appointment := newTrimmingAuditAppointment(84, 7)
				reservationRepo := &mockTrimmingReservationRepository{
					lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
						return cloneTrimmingAuditAppointment(appointment), nil
					},
					deleteFn: func(_ context.Context, _, _ uint64) error {
						*journal = append(*journal, "appointment.delete")
						return nil
					},
				}
				detailRepo := &mockTrimmingDetailRepository{
					findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.AppointmentTrimmingDetail, error) {
						return &model.AppointmentTrimmingDetail{ClinicID: 7, AppointmentID: appointment.ID}, nil
					},
				}
				svc := newTrimmingAuditTestService(reservationRepo, detailRepo, tx, audit)
				return svc.Delete(context.Background(), 7, appointment.ID, ptrUint64(42))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journal := make([]string, 0, 3)
			tx := &trimmingAuditRollbackTransactor{journal: &journal}
			audit := &trimmingAuditRecorder{err: sentinel}

			err := tt.run(tx, audit, &journal)

			require.Error(t, err)
			assert.ErrorIs(t, err, sentinel)
			assert.True(t, tx.rolledBack)
			assert.Empty(t, journal, "business writes must roll back when durable audit fails")
			assert.True(t, audit.sawTxCtx)
		})
	}
}

func TestTrimmingService_MissingAuditDependencyFailsBeforeEachWritePath(t *testing.T) {
	appointmentID := uint64(91)
	actorID := uint64(42)
	start := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		run  func(TrimmingService) error
	}{
		{
			name: "create appointment",
			run: func(svc TrimmingService) error {
				_, err := svc.Create(context.Background(), 7, &CreateTrimmingInput{
					ActorID:           &actorID,
					ReservationTypeID: 9,
					StartTime:         start,
					EndTime:           start.Add(time.Hour),
				})
				return err
			},
		},
		{
			name: "create detail",
			run: func(svc TrimmingService) error {
				_, err := svc.Create(context.Background(), 7, &CreateTrimmingInput{
					ActorID:       &actorID,
					AppointmentID: &appointmentID,
				})
				return err
			},
		},
		{
			name: "update",
			run: func(svc TrimmingService) error {
				_, err := svc.Update(context.Background(), 7, appointmentID, &UpdateTrimmingInput{
					ActorID: &actorID,
				})
				return err
			},
		},
		{
			name: "delete",
			run: func(svc TrimmingService) error {
				return svc.Delete(context.Background(), 7, appointmentID, &actorID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactionOpened := false
			svc := NewTrimmingService(
				&mockTrimmingReservationRepository{},
				&mockTrimmingReservationTypeRepository{},
				newAcceptingTrimmingStaffRepository(),
				&mockTrimmingUnavailableTimeRepository{},
				&mockTrimmingDetailRepository{},
				newActiveTrimmingCourseRepository(),
				newActiveTrimmingOptionRepository(),
				&mockTransactor{withTxFn: func(context.Context, func(context.Context) error) error {
					transactionOpened = true
					return nil
				}},
			)

			err := tt.run(svc)

			require.Error(t, err)
			assert.False(t, transactionOpened, "business transaction must not open without a durable audit sink")
		})
	}
}

func TestTrimmingService_MissingOrZeroActorFailsBeforeEachWritePath(t *testing.T) {
	appointmentID := uint64(92)
	start := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)

	operations := []struct {
		name string
		run  func(TrimmingService, *uint64) error
	}{
		{
			name: "create appointment",
			run: func(svc TrimmingService, actorID *uint64) error {
				_, err := svc.Create(context.Background(), 7, &CreateTrimmingInput{
					ActorID:           actorID,
					ReservationTypeID: 9,
					StartTime:         start,
					EndTime:           start.Add(time.Hour),
				})
				return err
			},
		},
		{
			name: "create detail",
			run: func(svc TrimmingService, actorID *uint64) error {
				_, err := svc.Create(context.Background(), 7, &CreateTrimmingInput{
					ActorID:       actorID,
					AppointmentID: &appointmentID,
				})
				return err
			},
		},
		{
			name: "update",
			run: func(svc TrimmingService, actorID *uint64) error {
				_, err := svc.Update(context.Background(), 7, appointmentID, &UpdateTrimmingInput{
					ActorID: actorID,
				})
				return err
			},
		},
		{
			name: "delete",
			run: func(svc TrimmingService, actorID *uint64) error {
				return svc.Delete(context.Background(), 7, appointmentID, actorID)
			},
		},
	}
	actors := []struct {
		name string
		id   *uint64
	}{
		{name: "missing", id: nil},
		{name: "zero", id: ptrUint64(0)},
	}

	for _, operation := range operations {
		for _, actor := range actors {
			t.Run(operation.name+"/"+actor.name, func(t *testing.T) {
				transactionOpened := false
				svc := newTrimmingAuditTestService(
					&mockTrimmingReservationRepository{},
					&mockTrimmingDetailRepository{},
					&mockTransactor{withTxFn: func(context.Context, func(context.Context) error) error {
						transactionOpened = true
						return nil
					}},
					&trimmingAuditRecorder{},
				)

				err := operation.run(svc, actor.id)

				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
				assert.False(t, transactionOpened, "invalid actor must fail before opening the business transaction")
			})
		}
	}
}
