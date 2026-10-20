package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/identitylink"
	"github.com/animal-ekarte/backend/internal/infra"
	appcrypto "github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/inventory"
	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/support"
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
	config            *config.Config
	db                *gorm.DB
	audit             audit.Kernel
	auth              authComposition
	staff             staffComposition
	clinic            clinicComposition
	ownerPet          ownerPetComposition
	reservation       reservationComposition
	billing           billingComposition
	medicalRecord     medicalRecordComposition
	inventory         inventoryRuntime
	trimming          trimmingRuntime
	lstep             *lstep.Application
	supportChat       support.ChatCompleter
	planeTickets      support.TicketCreator
	planeTimeout      time.Duration
	bugReportNotifier support.BugReportNotifier
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
		DB:                  dependencies.DB,
		Cipher:              dependencies.IntegrationCipher,
		SharedFileStorage:   dependencies.SharedStorage,
		Owners:              repositories.ownerPet.Owner,
		OwnerLifecycle:      ownerLifecycleWriterAdapter{inner: repositories.ownerPet.Owner},
		Pets:                repositories.ownerPet.Pet,
		PetLifecycle:        petLifecycleWriterAdapter{inner: repositories.ownerPet.Pet},
		Vaccinations:        repositories.medicalRecord.vaccinations,
		MedicalRecords:      repositories.medicalRecord.medicalRecords,
		Accounting:          repositories.billing.accounting,
		Checkups:            repositories.medicalRecord.checkups,
		BillingItems:        repositories.billing.billingItems,
		ClinicSettings:      repositories.clinic.Settings,
		ReservationSettings: repositories.reservation.LineReservationSettings,
		Reservations:        repositories.reservation.Reservations,
		// EMR-308: next_visit_* タグ同期と tentative リマインダー配信の appointment 参照
		// （candidate 抽出・送信直前再検証）。同一 ReservationStore を共有する。
		Appointments:          repositories.reservation.Reservations,
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
		config:        dependencies.Config,
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
			lstepApplication.TagSync,
		),
		lstep:       lstepApplication,
		supportChat: newSupportChatCompleter(dependencies.Config),
		// Plane 連携は PLANE_API_KEY 未設定の場合 nil（連携無効）を返す。
		// 報告自体は連携の有無に関わらずローカル保存される。
		planeTimeout: time.Duration(dependencies.Config.PlaneTimeoutMS) * time.Millisecond,
		planeTickets: support.NewPlaneTicketCreator(
			dependencies.Config.PlaneBaseURL,
			dependencies.Config.PlaneWebBaseURL,
			dependencies.Config.PlaneAPIKey,
			dependencies.Config.PlaneWorkspaceSlug,
			dependencies.Config.PlaneProjectID,
			dependencies.Config.PlaneProjectIdentifier,
			dependencies.Config.FrontendURL,
			time.Duration(dependencies.Config.PlaneTimeoutMS)*time.Millisecond,
			support.WithPlaneEnvironment(dependencies.Config.PlaneEnvironment),
			support.WithPlaneLabelIDs(
				dependencies.Config.PlaneSourceLabelID,
				dependencies.Config.PlaneLaneLabelID,
			),
		),
		// Discord 通知は DISCORD_BUG_REPORT_WEBHOOK_URL 未設定の場合 nil（通知無効）。
		// 通知の成否に関わらず報告はローカル保存される（best-effort）。
		bugReportNotifier: support.NewDiscordNotifier(
			dependencies.Config.DiscordBugReportWebhookURL,
			time.Duration(dependencies.Config.DiscordWebhookTimeoutMS)*time.Millisecond,
		),
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
			// EMR-308: 予約変更容易点の post-commit next_visit タグ再同期 hook。
			// lstep 側で未設定・同期失敗でも予約操作は失敗しない（best-effort）。
			NextVisitTagSync: lstepApplication.TagSync,

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
	nextVisitTagSync reservation.NextVisitTagSyncer,
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
			nextVisitTagSync,
		),
		courses: trimming.NewTrimmingCourseService(
			repositories.trimmingCourses,
			repositories.trimmingCourseType,
			transactor,
			trimmingCourseImageLifecycleOption(repositories.trimmingDetails),
			trimming.WithCourseImageAudit(trimmingAuditTxAdapter{inner: auditKernel}),
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
	if err := registerBaseRoutes(router, c.lstep.Batch, c.db, c.planeTickets); err != nil {
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
		c.auth.Handler.RequirePermissionAllowingAssignedClinicGrant,
		c.auth.Handler.HasPermission,
	)
	ownerPetHandlers.Owner.RegisterRoutes(protected)
	ownerPetHandlers.Pet.RegisterRoutes(protected)
	c.staff.newHandler(
		c.auth.Handler.RequirePermission,
		c.auth.Handler.HasPermission,
	).RegisterRoutes(protected)
	c.registerClinicRoutes(protected)
	c.registerExistingDomainRoutes(ctx, protected, uploader)

	lstepHandler := c.newLstepHandler()
	reservationHandler := c.newReservationHandler(ctx, lstepHandler)
	reservationHandler.RegisterRoutes(protected)
	reservationHandler.RegisterLiffRoutes(router)
	c.billing.newHandler(
		c.auth.Handler.RequirePermissionAllowingAssignedClinicGrant,
		c.auth.Handler.HasPermission,
	).RegisterRoutes(protected)
	billing.RegisterUATRoutes(router.Group("/api/v1"), &billing.SyntheticClosingHandler{DB: c.db})
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
	ctx context.Context,
	protected *gin.RouterGroup,
	uploader infra.FileUploader,
) {
	manualarticle.NewHandler(
		manualarticle.NewManualArticleService(manualarticle.New(c.db)),
		manualArticleAuditAdapter{logger: c.audit},
		c.auth.Handler.RequirePermissionAllowingAssignedClinicGrant,
	).RegisterRoutes(protected)
	support.NewHandler(
		support.NewService(
			support.NewRepository(c.db),
			c.planeTickets,
			c.bugReportNotifier,
			support.WithAuditLogger(supportAuditAdapter{logger: c.audit}),
			support.WithPlaneTimeout(
				c.planeTimeout,
			),
		),
		uploader,
		supportAuditAdapter{logger: c.audit},
		c.supportChat,
		middleware.RateLimit(
			middleware.NewRateLimitStore(ctx),
			rate.Limit(supportChatRequestsPerMinute)/60.0,
			supportChatBurst,
		),
	).WithManualSearch(
		support.NewManualSearchDeps(
			c.config.SupportHybridSearchEnabled,
			manualarticle.NewCorpusSource(manualarticle.NewManualArticleService(manualarticle.New(c.db))),
			func(ginC *gin.Context) (bool, error) {
				return c.auth.Handler.CheckPermission(ginC, "manual-edit", "view")
			},
			newSupportEmbedder(c.config),
			support.NewManualEmbeddingCache(c.db, c.config.SupportEmbeddingModel),
			c.config.SupportEmbeddingModel,
		),
	).WithLLMBudget(
		support.NewLLMBudgetGranter(c.db),
		c.config.SupportLLMModel,
	).RegisterRoutes(protected)
	identitylink.NewHandler(
		identitylink.NewService(
			identitylink.NewRepository(c.db),
			persistence.NewTransactor(c.db),
			c.audit,
			identitylink.NewDBLinkWriteAuthorizer(c.db),
		),
		c.auth.Handler.RequirePermissionAllowingAssignedClinicGrant,
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
	).WithImageSupport(trimming.ImageDependencies{
		Uploader:             uploader,
		Quota:                medicalrecord.NewPostgresMedicalRecordImageUploadQuotaStore(c.db),
		RequireAnyPermission: adaptTrimmingPermissionAny(c.auth.Handler.RequirePermissionAny),
	}).RegisterRoutes(protected)
	c.medicalRecord.newHandler(
		medicalRecordHTTPDependencies{
			Uploader:          uploader,
			DB:                c.db,
			HasPermission:     c.auth.Handler.HasPermission,
			RequirePermission: c.auth.Handler.RequirePermissionAllowingAssignedClinicGrant,
		},
	).RegisterRoutes(protected)
}

func adaptTrimmingPermissionAny(
	require func(...auth.PermissionRequirement) gin.HandlerFunc,
) trimming.PermissionAnyMiddleware {
	return func(requirements ...trimming.PermissionRequirement) gin.HandlerFunc {
		target := make([]auth.PermissionRequirement, len(requirements))
		for i, requirement := range requirements {
			target[i] = auth.PermissionRequirement{
				Resource: requirement.Resource,
				Action:   requirement.Action,
			}
		}
		return require(target...)
	}
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
			// SEC-CS3-O2: Content-Type 偽装によるグローバル body limit 迂回を塞ぐ
			// （LIFF は JSON のみ。multipart/octet-stream でも同じ上限を強制する）。
			LiffBodyLimit:     middleware.LimitRequestBody(middleware.DefaultJSONBodyMaxBytes),
			LinkLiffAccount:   lstepHandler.LinkLiffAccount,
			RequirePermission: c.auth.Handler.RequirePermissionAllowingAssignedClinicGrant,
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

type supportAuditAdapter struct {
	logger audit.Service
}

func (a supportAuditAdapter) LogEntry(
	ctx context.Context,
	entry support.AuditEntry,
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

// newSupportChatCompleter は既存 OpenAI 互換 completer を返す。
// Sarashina (api.sarashina.ai) の場合は SUPPORT_EXTERNAL_DATA_APPROVED=true の
// 承認済み環境のみ生成し、未承認なら nil（= status unavailable / 検索 UI へ縮退）。
func newSupportChatCompleter(cfg *config.Config) support.ChatCompleter {
	if support.IsSarashinaURL(cfg.SupportLLMBaseURL) && !cfg.SupportExternalDataApproved {
		return nil
	}
	return support.NewChatCompleter(
		cfg.SupportLLMBaseURL,
		cfg.SupportLLMAPIKey,
		cfg.SupportLLMModel,
		time.Duration(cfg.SupportLLMTimeoutMS)*time.Millisecond,
	)
}

// newSupportEmbedder は embeddings クライアントを返す。
// hybrid 無効・未承認・キー未設定なら nil（意味検索は index_missing/縮退）。
// 承認済み Sarashina 以外のエンドポイントも nil（support-index と同一の完全一致ゲート）。
func newSupportEmbedder(cfg *config.Config) support.ManualEmbedder {
	if !cfg.SupportHybridSearchEnabled || !cfg.SupportExternalDataApproved {
		return nil
	}
	if cfg.SupportEmbeddingBaseURL != "https://api.sarashina.ai/v1" {
		return nil
	}
	return support.NewSarashinaEmbedder(
		cfg.SupportEmbeddingBaseURL,
		cfg.SupportEmbeddingAPIKey,
		cfg.SupportEmbeddingModel,
		time.Duration(cfg.SupportEmbeddingTimeoutMS)*time.Millisecond,
	)
}

// trimmingCourseImageLifecycleOption shares the trimming image key lock and
// tombstone protocol with course images (EMR-457). A detail repository without
// the lifecycle capability leaves course image writes fail-closed.
func trimmingCourseImageLifecycleOption(details trimming.AppointmentTrimmingDetailRepository) trimming.TrimmingCourseServiceOption {
	lifecycle, _ := details.(trimming.ImageLifecycleRepository)
	return trimming.WithCourseImageLifecycle(lifecycle)
}
