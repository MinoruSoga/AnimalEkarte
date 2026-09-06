package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/identitylink"
	"github.com/animal-ekarte/backend/internal/infra"
	appcrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/inventory"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/trimming"
)

type runtimeRepositories struct {
	auth          authRepositories
	staff         staffRepositories
	clinic        clinicRepositories
	ownerPet      ownerPetRepositories
	reservation   reservationRepositories
	billing       billingRepositories
	medicalRecord medicalRecordRepositories

	inventory          inventory.Repository
	merchandiseItems   inventory.MerchandiseItemRepository
	trimmingCourses    trimming.TrimmingCourseRepository
	trimmingOptions    trimming.TrimmingOptionRepository
	trimmingCourseType trimming.TrimmingCourseTypeRepository
	trimmingDetails    trimming.AppointmentTrimmingDetailRepository
}

func newRuntimeRepositories(db *gorm.DB) runtimeRepositories {
	staffRepositories := newStaffRepositories(db)
	return runtimeRepositories{
		auth:          newAuthRepositories(db),
		staff:         staffRepositories,
		clinic:        newClinicRepositories(db),
		ownerPet:      newOwnerPetRepositories(db),
		reservation:   newReservationRepositories(db, staffRepositories.Staff, staffRepositories.ShiftEntries, staffRepositories.Occupations),
		billing:       newBillingRepositories(db),
		medicalRecord: newMedicalRecordRepositories(db),

		inventory:          inventory.New(db),
		merchandiseItems:   inventory.NewMerchandiseItemRepository(db),
		trimmingCourses:    trimming.NewTrimmingCourseRepository(db),
		trimmingOptions:    trimming.NewTrimmingOptionRepository(db),
		trimmingCourseType: trimming.NewTrimmingCourseTypeRepository(db),
		trimmingDetails:    trimming.NewAppointmentTrimmingDetailRepository(db),
	}
}

type inventoryRuntime struct {
	inventory        inventory.InventoryService
	merchandiseItems inventory.MerchandiseItemService
}

type trimmingRuntime struct {
	trimming    trimming.Service
	courses     trimming.TrimmingCourseService
	options     trimming.TrimmingOptionService
	courseTypes trimming.TrimmingCourseTypeService
}

type runtimeComposition struct {
	db            *gorm.DB
	audit         audit.Kernel
	auth          authComposition
	staff         staffComposition
	clinic        clinicComposition
	ownerPet      ownerPetComposition
	reservation   reservationComposition
	billing       billingComposition
	medicalRecord medicalRecordComposition
	inventory     inventoryRuntime
	trimming      trimmingRuntime
	lstep         *lstep.Application
}

type runtimeCompositionDependencies struct {
	Config            *config.Config
	DB                *gorm.DB
	IntegrationCipher *appcrypto.AESGCMCipher
	SharedStorage     infra.FileStorage
}

func newRuntimeComposition(
	dependencies runtimeCompositionDependencies,
) runtimeComposition {
	repositories := newRuntimeRepositories(dependencies.DB)
	transactor := persistence.NewTransactor(dependencies.DB)
	auditKernel := audit.NewService(audit.NewRepository(dependencies.DB))
	clinicComposition := newClinicComposition(
		repositories.clinic,
		repositories.auth.PermissionGroups,
		transactor,
		auditKernel,
	)
	staffComposition := newRuntimeStaffComposition(
		repositories,
		transactor,
		auditKernel,
	)
	lstepApplication := newRuntimeLstepApplication(
		dependencies,
		repositories,
		transactor,
		auditKernel,
	)

	return newRuntimeDomainCompositions(
		dependencies,
		repositories,
		transactor,
		auditKernel,
		clinicComposition,
		staffComposition,
		lstepApplication,
	)
}

func newRuntimeStaffComposition(
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	auditLogger audit.TxLogger,
) staffComposition {
	return newStaffComposition(
		repositories.staff,
		staffCompositionDependencies{
			Transactor: transactor,
			Accounts: staffAccountStoreAdapter{
				accounts:    repositories.auth.Accounts,
				resetTokens: repositories.auth.PasswordResetTokens,
			},
			PermissionGroups: repositories.auth.PermissionGroups,
			Reservations:     repositories.reservation.Reservations,
			ReservationStaff: repositories.reservation.ReservationStaff,
			Clinics:          repositories.clinic.Clinics,
			Audit:            auditLogger,
		},
	)
}

func newRuntimeLstepApplication(
	dependencies runtimeCompositionDependencies,
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	auditKernel audit.Kernel,
) *lstep.Application {
	return lstep.NewApplication(&lstep.Dependencies{
		DB:                    dependencies.DB,
		Cipher:                dependencies.IntegrationCipher,
		SharedFileStorage:     dependencies.SharedStorage,
		Owners:                repositories.ownerPet.Owner,
		OwnerLifecycle:        ownerLifecycleWriterAdapter{inner: repositories.ownerPet.Owner},
		Pets:                  repositories.ownerPet.Pet,
		PetLifecycle:          petLifecycleWriterAdapter{inner: repositories.ownerPet.Pet},
		Vaccinations:          repositories.medicalRecord.vaccinations,
		MedicalRecords:        repositories.medicalRecord.medicalRecords,
		Accounting:            repositories.billing.accounting,
		Prescriptions:         repositories.medicalRecord.prescriptions,
		Checkups:              repositories.medicalRecord.checkups,
		BillingItems:          repositories.billing.billingItems,
		ClinicSettings:        repositories.clinic.Settings,
		ReservationSettings:   repositories.reservation.LineReservationSettings,
		Reservations:          repositories.reservation.Reservations,
		Clinics:               repositories.clinic.Clinics,
		Staff:                 repositories.reservation.ReservationStaff,
		Audit:                 auditKernel,
		Transactor:            transactor,
		LifecycleAuditTx:      lstepLifecycleAuditTxAdapter{inner: auditKernel},
		NoShowAuditTx:         lstepNoShowAuditTxAdapter{inner: auditKernel},
		LineLinkAuditTx:       lstepLineLinkAuditTxAdapter{inner: auditKernel},
		AggregationRepository: lstepAggregationRepositoryAdapter{inner: owner.NewLtvRepository(dependencies.DB)},
	})
}

func newRuntimeDomainCompositions(
	dependencies runtimeCompositionDependencies,
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	auditKernel audit.Kernel,
	clinicComposition clinicComposition,
	staffComposition staffComposition,
	lstepApplication *lstep.Application,
) runtimeComposition {
	ownerPetComposition := newRuntimeOwnerPetComposition(
		repositories,
		auditKernel,
		lstepApplication,
	)
	medicalRecordComposition := newRuntimeMedicalRecordComposition(
		dependencies.DB,
		repositories,
		transactor,
		auditKernel,
		lstepApplication,
	)
	billingComposition := newRuntimeBillingComposition(
		repositories,
		transactor,
		auditKernel,
		clinicComposition,
		lstepApplication,
	)
	reservationComposition := newRuntimeReservationComposition(
		dependencies,
		repositories,
		transactor,
		staffComposition,
		medicalRecordComposition,
		lstepApplication,
	)
	authComposition := newRuntimeAuthComposition(
		dependencies.Config,
		repositories,
		transactor,
		auditKernel,
		staffComposition,
		clinicComposition,
	)

	return runtimeComposition{
		db:            dependencies.DB,
		audit:         auditKernel,
		auth:          authComposition,
		staff:         staffComposition,
		clinic:        clinicComposition,
		ownerPet:      ownerPetComposition,
		reservation:   reservationComposition,
		billing:       billingComposition,
		medicalRecord: medicalRecordComposition,
		inventory: newRuntimeInventory(
			repositories,
			transactor,
		),
		trimming: newRuntimeTrimming(
			repositories,
			transactor,
			auditKernel,
		),
		lstep: lstepApplication,
	}
}

func newRuntimeOwnerPetComposition(
	repositories runtimeRepositories,
	auditKernel audit.Kernel,
	lstepApplication *lstep.Application,
) ownerPetComposition {
	return newOwnerPetComposition(
		repositories.ownerPet,
		ownerPetCompositionDependencies{
			Insurance:            repositories.billing.insurance,
			MedicalRecords:       repositories.medicalRecord.medicalRecords,
			OwnerTags:            lstepApplication.TagSync,
			PetTags:              lstepApplication.TagSync,
			ChronicConditionTags: lstepApplication.TagSync,
			Audit:                auditKernel,
		},
	)
}

func newRuntimeMedicalRecordComposition(
	db *gorm.DB,
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	auditKernel audit.Kernel,
	lstepApplication *lstep.Application,
) medicalRecordComposition {
	return newMedicalRecordComposition(
		repositories.medicalRecord,
		medicalRecordCompositionDependencies{
			DB:               db,
			Transactor:       transactor,
			Audit:            auditKernel,
			TagSync:          lstepApplication.TagSync,
			DeliveryTrigger:  lstepApplication.DeliveryTrigger,
			LineCustomers:    lstepApplication.LineCustomers,
			Reservations:     repositories.reservation.Reservations,
			Pets:             repositories.ownerPet.Pet,
			Staff:            repositories.staff.Staff,
			StaffAssignments: repositories.staff.Assignments,
			Inventory:        repositories.inventory,
			Accounting:       repositories.billing.accounting,
			BillingItems:     repositories.billing.billingItems,
		},
	)
}

func newRuntimeBillingComposition(
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	auditKernel audit.Kernel,
	clinicComposition clinicComposition,
	lstepApplication *lstep.Application,
) billingComposition {
	return newBillingComposition(
		repositories.billing,
		billingCompositionDependencies{
			Transactor:       transactor,
			Audit:            auditKernel,
			MedicalRecords:   repositories.medicalRecord.medicalRecords,
			Hospitalizations: repositories.medicalRecord.hospitalizations,
			Reservations:     repositories.reservation.Reservations,
			TagSync:          lstepApplication.TagSync,
			StaffAssignments: repositories.staff.Assignments,
			Treatments:       repositories.medicalRecord.treatments,
			TrimmingCourses:  repositories.trimmingCourses,
			TrimmingOptions:  repositories.trimmingOptions,
			Owners:           repositories.ownerPet.Owner,
			MerchandiseItems: repositories.merchandiseItems,
			ClosingSettings:  clinicComposition.ClosingSettings,
			Clinics:          repositories.clinic.Clinics,
			ClinicHolidays:   repositories.clinic.Holidays,
		},
	)
}

func newRuntimeReservationComposition(
	dependencies runtimeCompositionDependencies,
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	staffComposition staffComposition,
	medicalRecordComposition medicalRecordComposition,
	lstepApplication *lstep.Application,
) reservationComposition {
	return newReservationComposition(
		repositories.reservation,
		reservationServiceDependencies{
			Transactor:      transactor,
			StaffDeleter:    staffComposition.Staff,
			LineCustomers:   lstepApplication.LineCustomers,
			Owners:          repositories.ownerPet.Owner,
			TrimmingCourses: repositories.trimmingCourses,
			TrimmingOptions: repositories.trimmingOptions,
			TrimmingDetails: repositories.trimmingDetails,
			Vaccinations:    repositories.medicalRecord.vaccinations,
			MedicalRecords:  medicalRecordComposition.MedicalRecords,

			NotificationConfig: reservationNotificationConfig(
				dependencies.Config,
			),
			EncryptCredential: reservationCredentialEncryptor(
				dependencies.IntegrationCipher,
			),
			DecryptCredential: reservationCredentialDecryptor(
				dependencies.IntegrationCipher,
			),
			NewLineMessenger: newReservationLineMessenger,
			SendMail:         sendReservationMail,
		},
	)
}

func reservationCredentialEncryptor(
	cipher *appcrypto.AESGCMCipher,
) func(string) (string, error) {
	return func(value string) (string, error) {
		return lstep.EncryptLineCredential(cipher, value)
	}
}

func reservationCredentialDecryptor(
	cipher *appcrypto.AESGCMCipher,
) func(context.Context, string) string {
	return func(ctx context.Context, value string) string {
		return lstep.DecryptLineCredential(ctx, cipher, value)
	}
}

func newReservationLineMessenger(
	channelToken string,
) reservation.LinePusher {
	return lstep.NewLineMessagingService(channelToken)
}

func reservationNotificationConfig(
	cfg *config.Config,
) *reservation.ReservationNotificationConfig {
	return &reservation.ReservationNotificationConfig{
		SMTPHost:    cfg.SMTPHost,
		SMTPPort:    cfg.SMTPPort,
		SMTPUser:    cfg.SMTPUser,
		SMTPPass:    cfg.SMTPPass,
		SMTPFrom:    cfg.SMTPFrom,
		FrontendURL: cfg.FrontendURL,
	}
}

func newRuntimeAuthComposition(
	cfg *config.Config,
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	auditKernel audit.Kernel,
	staffComposition staffComposition,
	clinicComposition clinicComposition,
) authComposition {
	return newAuthComposition(
		repositories.auth,
		authCompositionDependencies{
			Transactor: transactor,
			JWTSecret:  cfg.JWTSecret,
			PasswordResetConfig: auth.PasswordResetConfig{
				SMTPHost: cfg.SMTPHost, SMTPPort: cfg.SMTPPort,
				SMTPUser: cfg.SMTPUser, SMTPPass: cfg.SMTPPass,
				SMTPFrom: cfg.SMTPFrom, FrontendURL: cfg.FrontendURL,
			},
			CookieConfig:     auth.CookieConfigForProduction(cfg.GinMode == "release"),
			Staff:            staffComposition.Staff,
			StaffAssignments: staffComposition.Assignments,
			Clinics:          clinicComposition.Clinic,
			Audit:            auditKernel,
		},
	)
}

func newRuntimeInventory(
	repositories runtimeRepositories,
	transactor persistence.Transactor,
) inventoryRuntime {
	return inventoryRuntime{
		inventory: inventory.NewInventoryService(repositories.inventory),
		// BE-ACT-MERCHANDISE-ATOMIC-DELETE: Delete soft-delete + usage re-check share one tx.
		merchandiseItems: inventory.NewMerchandiseItemService(repositories.merchandiseItems, transactor),
	}
}

func newRuntimeTrimming(
	repositories runtimeRepositories,
	transactor persistence.Transactor,
	auditKernel audit.Kernel,
) trimmingRuntime {
	return trimmingRuntime{
		trimming: trimming.NewServiceWithAudit(
			repositories.reservation.Reservations,
			repositories.reservation.ReservationTypes,
			repositories.reservation.ReservationStaff,
			repositories.reservation.UnavailableTimes,
			repositories.trimmingDetails,
			repositories.trimmingCourses,
			repositories.trimmingOptions,
			transactor,
			trimmingAuditTxAdapter{inner: auditKernel},
		),
		courses: trimming.NewTrimmingCourseService(
			repositories.trimmingCourses,
			repositories.trimmingCourseType,
			transactor,
		),
		options: trimming.NewTrimmingOptionService(
			repositories.trimmingOptions,
			transactor,
		),
		courseTypes: trimming.NewTrimmingCourseTypeService(
			repositories.trimmingCourseType,
			transactor,
		),
	}
}

func (c runtimeComposition) registerRoutes(
	ctx context.Context,
	router *gin.Engine,
	uploader infra.FileUploader,
	isProduction bool,
) error {
	if err := registerBaseRoutes(router, c.lstep.Batch); err != nil {
		return err
	}
	protected, err := c.auth.registerRoutes(
		ctx,
		router.Group("/api/v1"),
		authRouteDependencies{
			Audit:        c.audit,
			IsProduction: isProduction,
			RateLimits:   auth.DefaultAuthRateLimitConfig(),
		},
	)
	if err != nil {
		return err
	}
	// POC-12 / TRM-03 / X-07: unified body size on the authenticated API surface.
	// Global SanitizeNullBytes already caps raw bytes; this re-applies the same
	// ceiling on protected routes so packages without per-handler MaxBytesReader
	// (pet/owner/clinic/trimming/manualarticle) stay bounded.
	protected.Use(middleware.LimitRequestBody(middleware.DefaultJSONBodyMaxBytes))
	c.auth.Handler.RegisterPermissionGroupRoutes(
		protected.Group("/masters"),
	)
	return c.registerDomainRoutes(ctx, router, protected, uploader)
}

func (c runtimeComposition) registerDomainRoutes(
	ctx context.Context,
	router *gin.Engine,
	protected *gin.RouterGroup,
	uploader infra.FileUploader,
) error {
	ownerPetHandlers := c.ownerPet.newHandlers(
		c.lstep,
		c.auth.Handler.RequirePermission,
		c.auth.Handler.HasPermission,
	)
	ownerPetHandlers.Owner.RegisterRoutes(protected)
	ownerPetHandlers.Pet.RegisterRoutes(protected)
	c.staff.newHandler(
		c.auth.Handler.RequirePermission,
		c.auth.Handler.HasPermission,
	).RegisterRoutes(protected)
	c.registerClinicRoutes(protected)
	c.registerExistingDomainRoutes(protected, uploader)

	lstepHandler := c.newLstepHandler()
	reservationHandler := c.newReservationHandler(ctx, lstepHandler)
	reservationHandler.RegisterRoutes(protected)
	reservationHandler.RegisterLiffRoutes(router)
	c.billing.newHandler(
		c.auth.Handler.RequirePermission,
		c.auth.Handler.HasPermission,
	).RegisterRoutes(protected)
	lstepHandler.RegisterRoutes(protected)
	lstepHandler.RegisterWebhookRoutes(
		router,
		middleware.RateLimit(
			middleware.NewRateLimitStore(ctx),
			lineWebhookRequestsPerSecond,
			lineWebhookBurst,
		),
	)
	return nil
}

func (c runtimeComposition) registerClinicRoutes(
	protected *gin.RouterGroup,
) {
	handler := c.clinic.newHandler(c.auth.Handler.RequirePermission)
	handler.RegisterClinicRoutes(protected)
	handler.RegisterClinicHolidayRoutes(protected)
	handler.RegisterCompanyRoutes(protected)
	handler.RegisterClosingSettingsRoutes(protected)
}

func (c runtimeComposition) registerExistingDomainRoutes(
	protected *gin.RouterGroup,
	uploader infra.FileUploader,
) {
	manualarticle.NewHandler(
		manualarticle.NewManualArticleService(manualarticle.New(c.db)),
		manualArticleAuditAdapter{logger: c.audit},
		c.auth.Handler.RequirePermission,
	).RegisterRoutes(protected)
	identitylink.NewHandler(
		identitylink.NewService(
			identitylink.NewRepository(c.db),
			persistence.NewTransactor(c.db),
			c.audit,
		),
		c.auth.Handler.RequirePermission,
	).RegisterRoutes(protected)
	inventory.NewHandler(
		c.inventory.inventory,
		c.inventory.merchandiseItems,
		c.auth.Handler.RequirePermission,
	).RegisterRoutes(protected)
	trimming.NewHandlerWithPermission(
		c.trimming.trimming,
		c.trimming.courses,
		c.trimming.courseTypes,
		c.trimming.options,
		c.auth.Handler.RequirePermission,
	).RegisterRoutes(protected)
	c.medicalRecord.newHandler(
		medicalRecordHTTPDependencies{
			Uploader:          uploader,
			DB:                c.db,
			HasPermission:     c.auth.Handler.HasPermission,
			RequirePermission: c.auth.Handler.RequirePermission,
		},
	).RegisterRoutes(protected)
}

func (c runtimeComposition) newLstepHandler() *lstep.Handler {
	return c.lstep.NewHandler(lstep.HandlerDependencies{
		OwnerLineLinker:      c.ownerPet.OwnerService,
		RequirePermission:    c.auth.Handler.RequirePermission,
		RequireAnyPermission: adaptPermissionAny(c.auth.Handler.RequirePermissionAny),
	})
}

func (c runtimeComposition) newReservationHandler(
	ctx context.Context,
	lstepHandler *lstep.Handler,
) *reservation.Handler {
	liffRateLimitStore := middleware.NewRateLimitStore(ctx)
	return c.reservation.newHandler(
		reservationHandlerDependencies{
			StaffAssignments: c.staff.Assignments,
			LiffAuth: middleware.LiffAuth(
				c.lstep.LineCustomers,
				c.reservation.Repositories.LineReservationSettings,
			),
			LiffRateLimit: func(limit int) gin.HandlerFunc {
				return middleware.LiffRateLimit(liffRateLimitStore, limit)
			},
			LinkLiffAccount:   lstepHandler.LinkLiffAccount,
			RequirePermission: c.auth.Handler.RequirePermission,
		},
	)
}

type manualArticleAuditAdapter struct {
	logger audit.Service
}

func (a manualArticleAuditAdapter) LogEntry(
	ctx context.Context,
	entry manualarticle.AuditEntry,
) error {
	return a.logger.LogEntry(ctx, &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	})
}

type trimmingAuditTxAdapter struct {
	inner audit.TxLogger
}

func (a trimmingAuditTxAdapter) LogEntryTx(
	ctx context.Context,
	entry *trimming.AuditEntry,
) error {
	return a.inner.LogEntryTx(ctx, &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		Metadata:   entry.Metadata,
	})
}
