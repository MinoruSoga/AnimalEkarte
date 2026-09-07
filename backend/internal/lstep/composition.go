package lstep

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
)

// OwnerRepository is the union of owner operations consumed by LSTEP use cases.
type OwnerRepository interface {
	lineLinkOwnerRepo
	tagSyncOwnerRepo
	checkupSyncOwnerRepo
}

// PetRepository is the union of pet operations consumed by LSTEP use cases.
type PetRepository interface {
	tagSyncPetRepo
	checkupSyncPetRepo
	deliveryPetRepository
	lifecyclePetRepository
}

// MedicalRecordRepository is the union of medical-record reads consumed by LSTEP.
type MedicalRecordRepository interface {
	tagSyncMedicalRecordRepo
	deliveryMedicalRecordRepository
	lstepBatchMedicalRecordRepository
}

// VaccinationRepository is the union of vaccination reads consumed by LSTEP.
type VaccinationRepository interface {
	tagSyncVaccinationRepo
	deliveryVaccinationRepository
}

// AuditLogger is the shared audit kernel view consumed by LSTEP.
type AuditLogger interface {
	lstepAuditLogger
	checkupSyncAuditLogger
	lstepBatchAuditService
}

// Transactor preserves the ambient transaction across LSTEP write operations.
type Transactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// Dependencies are cross-domain capabilities required to construct the LSTEP application.
// LSTEP-owned repositories are created from DB inside NewApplication.
type Dependencies struct {
	DB                    *gorm.DB
	Cipher                *crypto.AESGCMCipher
	SharedFileStorage     SharedFileStorage
	Owners                OwnerRepository
	OwnerLifecycle        OwnerLifecycleWriter
	Pets                  PetRepository
	PetLifecycle          PetLifecycleWriter
	Vaccinations          VaccinationRepository
	MedicalRecords        MedicalRecordRepository
	Accounting            tagSyncAccountingRepo
	Prescriptions         tagSyncPrescriptionRepo
	Checkups              tagSyncCheckupRepo
	BillingItems          tagSyncBillingItemRepo
	ClinicSettings        lstepClinicSettingsRepo
	ReservationSettings   lstepLineSettingReader
	Reservations          lstepBatchReservationRepository
	Clinics               lstepBatchClinicRepository
	Staff                 csvImportStaffRepository
	Audit                 AuditLogger
	Transactor            Transactor
	LifecycleAuditTx      LifecycleAuditTxLogger
	NoShowAuditTx         NoShowAuditTxLogger
	LineLinkAuditTx       LineLinkAuditTxLogger
	AggregationRepository OwnerAggregationRepository
}

// Application is the typed LSTEP composition result exposed to real cross-domain consumers.
// HTTP-only services and SharedFile remain private in graph.
type Application struct {
	TagSync         LstepTagSyncService
	DeliveryTrigger LstepDeliveryTriggerService
	Batch           LstepBatchService
	LineCustomers   LineCustomerRepository

	graph applicationGraph
}

// HandlerDependencies are the HTTP-boundary capabilities injected after legacy auth/owner
// services have been constructed.
type HandlerDependencies struct {
	OwnerLineLinker      OwnerLineLinker
	RequirePermission    PermissionMiddleware
	RequireAnyPermission PermissionAnyMiddleware
}

// NewApplication constructs every LSTEP-owned repository and service graph.
func NewApplication(deps *Dependencies) *Application {
	repos := newApplicationRepositories(deps.DB)
	return newApplication(deps, repos)
}

func newApplication(deps *Dependencies, repos *applicationRepositories) *Application {
	core := newApplicationCore(deps, repos)
	delivery := NewLstepDeliveryTriggerService(
		deps.Owners, deps.MedicalRecords, deps.Vaccinations, deps.Pets,
		repos.tagCache, repos.deliveryTriggerLog, core.settings, core.triggerPriority,
	)
	batch := NewLstepBatchService(
		deps.Reservations, core.tagSync, deps.Clinics, deps.MedicalRecords, deps.Audit,
		core.settings, delivery, deps.Transactor, deps.NoShowAuditTx,
	)
	graph := newApplicationGraph(deps, repos, core)
	return &Application{
		TagSync:         core.tagSync,
		DeliveryTrigger: delivery,
		Batch:           batch,
		LineCustomers:   repos.lineCustomers,
		graph:           graph,
	}
}

// NewHandler constructs the package-owned HTTP handler graph.
func (a *Application) NewHandler(deps HandlerDependencies) *Handler {
	return NewHandler(
		NewSettingsHandler(a.graph.settings, deps.RequirePermission),
		NewLineSendHandler(a.graph.lineSend, deps.RequirePermission),
		NewLineLinkHandler(a.graph.lineLink, deps.RequirePermission),
		NewLineCustomerHandler(a.graph.lineCustomer, deps.RequirePermission),
		a.graph.tag, a.graph.tagCodeMapping, a.graph.tagConfig, a.graph.tagSummary,
		a.graph.checkupSync, a.graph.lifecycle, a.graph.deliveryMonitor,
		a.graph.triggerPriority, a.graph.aggregation, a.graph.csvImport,
		a.graph.analytics, a.graph.sharedFile, deps.OwnerLineLinker,
		deps.RequirePermission, deps.RequireAnyPermission,
	)
}

// HandleOwnerDeletion implements the legacy owner handler's narrow deletion-lifecycle port.
// The application keeps the concrete LSTEP lifecycle service private.
func (a *Application) HandleOwnerDeletion(ctx context.Context, clinicID, ownerID uint64) error {
	return a.graph.lifecycle.HandleOwnerDeletion(ctx, clinicID, ownerID)
}

// LinkLiffAccount forwards the reservation LIFF callback without exposing sub-handlers.
func (h *Handler) LinkLiffAccount(c *gin.Context) {
	h.lineLink.LinkLiffAccount(c)
}
