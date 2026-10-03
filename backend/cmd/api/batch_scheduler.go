package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/scheduler"
	"github.com/animal-ekarte/backend/internal/support"
)

// schedulerInternalTokenHeader is the application-level privilege header for
// all-clinic scheduled batch routes (DEC-36 / CMD-02).
const schedulerInternalTokenHeader = "X-Scheduler-Token" //nolint:gosec // G101: HTTP header name, not a credential

var errScheduledBatchUnavailable = errors.New("scheduled batch service is unavailable")

type scheduledBatchService interface {
	RunNoShowCheckAllClinicsAt(
		ctx context.Context,
		scheduledAt time.Time,
		runID string,
	) lstep.BatchRunResult
	RunDeliveryTriggerBatchAllClinicsAt(
		ctx context.Context,
		scheduledAt time.Time,
		runID string,
	) lstep.BatchRunResult
	RunDormantDetectionAllClinicsAt(
		ctx context.Context,
		scheduledAt time.Time,
		runID string,
	) lstep.BatchRunResult
}

// planeStateSyncer is the support-side dependency behind JobPlaneSync.
// support.Service satisfies it; nil disables the job fail-closed.
type planeStateSyncer interface {
	SyncPlaneTicketStates(ctx context.Context) support.PlaneSyncResult
}

// planeStateSyncerFor builds the JobPlaneSync dependency only when Plane
// integration is configured. A nil result makes the job fail closed at
// execution time instead of silently pretending to succeed.
func planeStateSyncerFor(db *gorm.DB, tickets support.TicketCreator) planeStateSyncer {
	if db == nil || tickets == nil {
		return nil
	}
	return support.NewService(support.NewRepository(db), tickets)
}

type lstepScheduledJobExecutor struct {
	batch     scheduledBatchService
	planeSync planeStateSyncer
}

func newLstepScheduledJobExecutor(batch scheduledBatchService, planeSync planeStateSyncer) *lstepScheduledJobExecutor {
	return &lstepScheduledJobExecutor{batch: batch, planeSync: planeSync}
}

func registerScheduledJobRoutes(routes gin.IRoutes, batch scheduledBatchService, planeSync planeStateSyncer, internalToken string) {
	// Always register the contract path; middleware fails closed when the shared
	// secret is unset or the header does not match (DEC-36 / CMD-02).
	var protected gin.IRoutes
	switch r := routes.(type) {
	case *gin.Engine:
		protected = r.Group("", requireSchedulerInternalToken(internalToken))
	case *gin.RouterGroup:
		protected = r.Group("", requireSchedulerInternalToken(internalToken))
	default:
		// Production always passes *gin.Engine; refuse unknown IRoutes shapes.
		return
	}
	scheduler.NewHandler(newLstepScheduledJobExecutor(batch, planeSync)).RegisterRoutes(protected)
}

func requireSchedulerInternalToken(expected string) gin.HandlerFunc {
	expectedBytes := []byte(expected)
	return func(c *gin.Context) {
		// Empty expected never authenticates (including empty provided header).
		if len(expectedBytes) == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		provided := []byte(c.GetHeader(schedulerInternalTokenHeader))
		if len(provided) != len(expectedBytes) ||
			subtle.ConstantTimeCompare(provided, expectedBytes) != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

func (e *lstepScheduledJobExecutor) Execute(
	ctx context.Context,
	execution scheduler.Execution,
) (scheduler.Result, error) {
	// FenceToken protects durable-ledger finalization in the Worker coordinator.
	// It is intentionally not presented as an application-side revocation fence;
	// Go work instead obeys ctx and retains each domain's CAS/idempotency checks.
	var result lstep.BatchRunResult
	if execution.Job == scheduler.JobPlaneSync {
		if e == nil || e.planeSync == nil {
			return scheduler.Result{}, errScheduledBatchUnavailable
		}
		syncResult := e.planeSync.SyncPlaneTicketStates(ctx)
		result = lstep.BatchRunResult{
			Processed: syncResult.Processed,
			Succeeded: syncResult.Succeeded,
			Failed:    syncResult.Failed,
		}
	} else {
		if e == nil || e.batch == nil {
			return scheduler.Result{}, errScheduledBatchUnavailable
		}
		switch execution.Job {
		case scheduler.JobNoShow:
			result = e.batch.RunNoShowCheckAllClinicsAt(
				ctx,
				execution.ScheduledAt,
				execution.RunID,
			)
		case scheduler.JobDelivery:
			result = e.batch.RunDeliveryTriggerBatchAllClinicsAt(
				ctx,
				execution.ScheduledAt,
				execution.RunID,
			)
		case scheduler.JobDormant:
			result = e.batch.RunDormantDetectionAllClinicsAt(
				ctx,
				execution.ScheduledAt,
				execution.RunID,
			)
		default:
			return scheduler.Result{}, fmt.Errorf(
				"unsupported scheduled job %q",
				execution.Job,
			)
		}
	}

	if err := result.Validate(); err != nil {
		return scheduler.Result{}, fmt.Errorf("invalid batch result: %w", err)
	}
	response := scheduler.Result{
		Outcome:   scheduledOutcome(result),
		Processed: result.Processed,
		Succeeded: result.Succeeded,
		Failed:    result.Failed,
	}
	if err := response.Validate(); err != nil {
		return scheduler.Result{}, fmt.Errorf("invalid scheduled result: %w", err)
	}
	return response, nil
}

func scheduledOutcome(result lstep.BatchRunResult) scheduler.Outcome {
	switch {
	case result.Failed == 0:
		return scheduler.OutcomeSuccess
	case result.Succeeded == 0:
		return scheduler.OutcomeFailed
	default:
		return scheduler.OutcomePartial
	}
}
