package reservation

// reservation_confirm_visit_type_test.go — R2: POST /reservations/:id/confirm-tentative の
// optional visit_type（first/revisit）を handler→service→repo→response で end-to-end に pin する。
// 省略時は tentative 行の既存値（既定 revisit）を維持する契約。

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// makeOnDutyDoctorForJSTDate は confirm 日（JST 暦日）の capacity 要件を満たす
// 出勤医を用意する。doctorless confirm は CountOnDutyDoctors>0 が必要なため。
func makeOnDutyDoctorForJSTDate(t *testing.T, db *gorm.DB, clinicID uint64, name string, start time.Time) {
	t.Helper()
	doctor := makeDoctor(t, db, clinicID, name)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicID, IsMain: true}).Error)
	jst := start.In(config.JST)
	makeShiftEntry(t, db, clinicID, doctor.ID, time.Date(jst.Year(), jst.Month(), jst.Day(), 0, 0, 0, 0, time.UTC))
}

func confirmVisitTypeTentative(t *testing.T, db *gorm.DB, svc *reservationService, clinicID uint64, tag string) *model.Reservation {
	t.Helper()
	_, pet, _ := makeNextVisitIntentFixture(t, db, clinicID, tag)
	intendedDate := time.Now().In(config.JST).AddDate(0, 0, 30)
	tentative, err := svc.SetNextVisitIntent(context.Background(), clinicID, pet.ID, SetNextVisitIntentInput{IntendedDate: intendedDate})
	require.NoError(t, err)
	return tentative
}

func TestConfirmTentative_VisitType(t *testing.T) {
	db := setupNextVisitIntentTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)
	newStart := time.Now().UTC().AddDate(0, 0, 45).Truncate(time.Minute)
	newEnd := newStart.Add(30 * time.Minute)

	t.Run("writes first when explicitly provided", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		tentative := confirmVisitTypeTentative(t, db, svc, clinicA, "CFMVT1")
		makeOnDutyDoctorForJSTDate(t, db, clinicA, "確定VT医1", newStart)
		first := model.VisitTypeFirst

		confirmed, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
			VisitType: &first,
		})
		require.NoError(t, err)
		assert.Equal(t, model.VisitTypeFirst, confirmed.VisitType)

		var stored model.Reservation
		require.NoError(t, db.WithContext(ctx).First(&stored, tentative.ID).Error)
		assert.Equal(t, model.VisitTypeFirst, stored.VisitType)
	})

	t.Run("omitted preserves the existing revisit value", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		tentative := confirmVisitTypeTentative(t, db, svc, clinicA, "CFMVT2")
		makeOnDutyDoctorForJSTDate(t, db, clinicA, "確定VT医2", newStart)

		confirmed, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
		})
		require.NoError(t, err)
		assert.Equal(t, model.VisitTypeRevisit, confirmed.VisitType)
	})

	t.Run("rejects an invalid visit_type at the service boundary", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		tentative := confirmVisitTypeTentative(t, db, svc, clinicA, "CFMVT3")
		invalid := model.VisitType("bogus")

		got, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: newStart,
			EndTime:   newEnd,
			VisitType: &invalid,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, got)
		assert.Equal(t, model.ReservationStatusTentative, reloadReservationIntentStatus(t, db, tentative.ID))
	})

	t.Run("rejects an invalid visit_type for direct repository calls", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		tentative := confirmVisitTypeTentative(t, db, svc, clinicA, "CFMVT4")
		repo := NewReservationRepository(db)
		invalid := model.VisitType("bogus")

		err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			_, confirmErr := repo.ConfirmTentative(txCtx, clinicA, tentative.ID, ConfirmTentativeUpdate{
				StartTime:         newStart,
				EndTime:           newEnd,
				ReservationTypeID: tentative.ReservationTypeID,
				Status:            model.ReservationStatusConfirmed,
				VisitType:         &invalid,
			})
			return confirmErr
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("supports pending with an explicit visit_type", func(t *testing.T) {
		svc := newNextVisitIntentService(t, db, nil)
		tentative := confirmVisitTypeTentative(t, db, svc, clinicA, "CFMVT5")
		// 先行サブテストが newStart に確定済み予約を残すため capacity 衝突しない
		// 同日 +2h の別枠を使う（既存 pending 確定テストと同じ理由）。
		pendingStart := newStart.Add(2 * time.Hour)
		makeOnDutyDoctorForJSTDate(t, db, clinicA, "確定VT医5", pendingStart)
		first := model.VisitTypeFirst
		pending := model.ReservationStatusPending

		confirmed, err := svc.ConfirmTentative(ctx, clinicA, tentative.ID, &ConfirmTentativeInput{
			StartTime: pendingStart,
			EndTime:   pendingStart.Add(30 * time.Minute),
			VisitType: &first,
			Status:    &pending,
		})
		require.NoError(t, err)
		assert.Equal(t, model.ReservationStatusPending, confirmed.Status)
		assert.Equal(t, model.VisitTypeFirst, confirmed.VisitType)
	})
}

func TestConfirmTentativeAppointment_VisitTypeBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	withJSTLocal(t)

	start := time.Date(2026, 12, 1, 10, 0, 0, 0, time.Local)
	end := start.Add(30 * time.Minute)

	t.Run("maps visit_type into the service input", func(t *testing.T) {
		svc := &mockReservationService{
			confirmTentativeFn: func(_ context.Context, _, _ uint64, input *ConfirmTentativeInput) (*model.Reservation, error) {
				require.NotNil(t, input.VisitType)
				assert.Equal(t, model.VisitTypeFirst, *input.VisitType)
				return &model.Reservation{ID: 9, Status: model.ReservationStatusConfirmed}, nil
			},
		}
		router := newConfirmTentativeRouter(svc)
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"start_time": start.Format(time.RFC3339),
			"end_time":   end.Format(time.RFC3339),
			"visit_type": "first",
		})
		req := httptest.NewRequest(http.MethodPost, "/reservations/9/confirm-tentative", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("400 for a visit_type outside the enum", func(t *testing.T) {
		svc := &mockReservationService{}
		router := newConfirmTentativeRouter(svc)
		w := httptest.NewRecorder()
		body, _ := json.Marshal(map[string]any{
			"start_time": start.Format(time.RFC3339),
			"end_time":   end.Format(time.RFC3339),
			"visit_type": "yearly",
		})
		req := httptest.NewRequest(http.MethodPost, "/reservations/9/confirm-tentative", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})
}
