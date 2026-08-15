package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/scheduler"
)

// schedulerInternalTokenHeader is the application-level privilege header for
// all-clinic scheduled batch routes (DEC-36 / CMD-02).
const schedulerInternalTokenHeader = "X-Scheduler-Token"

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

type lstepScheduledJobExecutor struct {
	batch scheduledBatchService
}

func newLstepScheduledJobExecutor(batch scheduledBatchService) *lstepScheduledJobExecutor {
	return &lstepScheduledJobExecutor{batch: batch}
}

func registerScheduledJobRoutes(routes gin.IRoutes, batch scheduledBatchService, internalToken string) {
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
	scheduler.NewHandler(newLstepScheduledJobExecutor(batch)).RegisterRoutes(protected)
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
	if e == nil || e.batch == nil {
		return scheduler.Result{}, errScheduledBatchUnavailable
	}

	// FenceToken protects durable-ledger finalization in the Worker coordinator.
	// It is intentionally not presented as an application-side revocation fence;
	// Go work instead obeys ctx and retains each domain's CAS/idempotency checks.
	var result lstep.BatchRunResult
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
