package main

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/handler"
	"github.com/animal-ekarte/backend/internal/inventory"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// TestRouteCompositionSmoke_NoPanic registers every route surface in main.go order.
// Trimming remains part of the legacy handler until its route ownership migrates in BE9-2F.
func TestRouteCompositionSmoke_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NotPanics(t, func() {
		r := gin.New()
		legacyHandler := handler.New(
			&config.Config{JWTSecret: "test-secret-for-route-registration"},
			nil,
			nil,
			nil,
		)
		protected := legacyHandler.RegisterRoutes(ctx, r)

		manualarticle.NewHandler(nil, nil, routeSmokeNoopPermission).RegisterRoutes(protected)
		inventory.NewHandler(nil, nil, routeSmokeNoopPermission).RegisterRoutes(protected)
		newMedicalRecordRouteSmokeHandler().RegisterRoutes(protected)

		lstepHandler := newLstepRouteSmokeHandler()
		reservationHandler := newReservationRouteSmokeHandler(lstepHandler.LinkLiffAccount)
		reservationHandler.RegisterRoutes(protected)
		reservationHandler.RegisterLiffRoutes(r)

		newBillingRouteSmokeHandler().RegisterRoutes(protected)
		lstepHandler.RegisterRoutes(protected)
		lstepHandler.RegisterWebhookRoutes(r)
	})
}

func newMedicalRecordRouteSmokeHandler() *medicalrecord.Handler {
	return medicalrecord.NewHandler(
		medicalrecord.NewDiagnosisHandler(nil, nil),
		medicalrecord.NewExamTypeHandler(nil),
		medicalrecord.NewChiefComplaintHandler(nil),
		medicalrecord.NewCheckupHandler(nil, nil),
		medicalrecord.NewCheckupTypeHandler(nil),
		medicalrecord.NewVaccineHandler(nil),
		medicalrecord.NewVaccinationHandler(nil),
		medicalrecord.NewPrescriptionHandler(nil),
		medicalrecord.NewInquiryHandler(nil),
		medicalrecord.NewInquiryTemplateHandler(nil),
		medicalrecord.NewLabImportHandler(nil, nil, nil),
		medicalrecord.NewLabReportHandler(nil),
		medicalrecord.NewVitalHandler(nil, nil),
		medicalrecord.NewClinicalPlanHandler(nil),
		medicalrecord.NewMedicalRecordImageHandler(nil, nil, nil),
		medicalrecord.NewTreatmentHandler(nil, nil),
		medicalrecord.NewHospitalizationHandler(nil),
		medicalrecord.NewHospitalizationPlanHandler(nil),
		medicalrecord.NewDailyRecordHandler(nil),
		medicalrecord.NewCarePlanItemHandler(nil),
		medicalrecord.NewConsultationHandler(nil),
		medicalrecord.NewProcedureHandler(nil),
		medicalrecord.NewMedicineHandler(nil),
		medicalrecord.NewMedicineDoseParamHandler(nil),
		medicalrecord.NewCageHandler(nil),
		medicalrecord.NewTreatmentPlanHandler(nil, nil, nil, nil),
		medicalrecord.NewMedicalRecordHandler(nil),
		medicalrecord.NewMedicalRecordAddendumHandler(nil),
		medicalrecord.NewExaminationHandler(nil),
		routeSmokeNoopPermission,
	)
}

func newReservationRouteSmokeHandler(linkLiffAccount gin.HandlerFunc) *reservation.Handler {
	return reservation.NewHandler(
		reservation.NewReservationTypeHandler(nil, nil, nil, nil),
		reservation.NewReservationTypeGroupHandler(nil),
		reservation.NewReservationTypeLiffHandler(nil),
		reservation.NewReservationStaffHandler(nil),
		reservation.NewReservationScheduleHandler(nil),
		reservation.NewReservationHandler(nil, nil, nil, nil),
		reservation.NewReservationAdminHandler(nil, nil),
		reservation.NewLineReservationSettingHandler(nil),
		reservation.NewLiffHandler(nil, nil),
		routeSmokeNoopHandler,
		func(int) gin.HandlerFunc { return routeSmokeNoopHandler },
		linkLiffAccount,
		routeSmokeNoopPermission,
	)
}

func newBillingRouteSmokeHandler() *billing.Handler {
	hasPermission := func(*gin.Context, string, string) bool { return true }
	return billing.NewHandler(
		billing.NewInsuranceHandler(nil),
		billing.NewCampaignHandler(nil),
		billing.NewPaymentMethodMasterHandler(nil),
		billing.NewEstimateHandler(nil, hasPermission),
		billing.NewBillingConfirmationHandler(nil, routeSmokeNoopPermission),
		billing.NewBillingItemHandler(nil, routeSmokeNoopPermission),
		billing.NewRefundHandler(nil, routeSmokeNoopPermission),
		billing.NewAccountingHandler(nil, nil, hasPermission),
		billing.NewCashRegisterHandler(nil, routeSmokeNoopPermission),
		billing.NewAccountingReportHandler(nil, routeSmokeNoopPermission),
		routeSmokeNoopPermission,
	)
}

func newLstepRouteSmokeHandler() *lstep.Handler {
	return lstep.NewHandler(
		lstep.NewLstepSettingsHandler(nil, routeSmokeNoopPermission),
		lstep.NewLineSendHandler(nil, routeSmokeNoopPermission),
		lstep.NewLineLinkHandler(nil, func(*gin.Context, *model.Owner) {}, routeSmokeNoopPermission),
		lstep.NewLineCustomerHandler(nil, routeSmokeNoopPermission),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		routeSmokeNoopPermission,
		func(...lstep.PermissionRequirement) gin.HandlerFunc { return routeSmokeNoopHandler },
	)
}

func routeSmokeNoopPermission(string, string) gin.HandlerFunc {
	return routeSmokeNoopHandler
}

func routeSmokeNoopHandler(*gin.Context) {}
