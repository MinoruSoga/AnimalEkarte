// Package scheduler owns the private HTTP boundary used by the durable
// Cloudflare scheduler. It deliberately contains no timer or cron loop.
package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// SchedulerName is the protocol identifier shared with the Cloudflare Worker.
	SchedulerName = "animalekarte-scheduler-v1"
	// RequestTimeout is shorter than both the Worker fetch timeout and its durable lease.
	RequestTimeout = 100 * time.Second

	runActionSuffix = ":run"
	maxRunIDLength  = 256
	// The fixed request has five small fields; 4 KiB leaves ample protocol headroom
	// while preventing an internal/public-routing mistake from creating unbounded reads.
	maxRunRequestBodyBytes = 4 << 10
)

// Job is the closed set of durable jobs exposed at the internal boundary.
// PO-001 segmentation jobs are intentionally absent.
type Job string

const (
	JobNoShow   Job = "no_show"
	JobDelivery Job = "delivery"
	JobDormant  Job = "dormant"
)

// RunRequest is the exact Worker-to-Go request contract.
type RunRequest struct {
	Scheduler     string `json:"scheduler"`
	Job           Job    `json:"job"`
	ScheduledTime int64  `json:"scheduled_time"`
	RunID         string `json:"run_id"`
	FenceToken    int64  `json:"fence_token"`
}

// Execution is the validated representation passed to the application adapter.
type Execution struct {
	Job         Job
	ScheduledAt time.Time
	RunID       string
	FenceToken  int64
}

// Outcome is the terminal job outcome understood by the durable coordinator.
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomePartial Outcome = "partial"
	OutcomeFailed  Outcome = "failed"
)

// Result is the exact Go-to-Worker response contract.
type Result struct {
	Outcome   Outcome `json:"outcome"`
	Processed int     `json:"processed"`
	Succeeded int     `json:"succeeded"`
	Failed    int     `json:"failed"`
}

// Validate enforces the semantic relationship between outcome and counters.
func (r Result) Validate() error {
	if r.Processed < 0 || r.Succeeded < 0 || r.Failed < 0 {
		return errors.New("scheduled result counters must be non-negative")
	}
	if r.Processed != r.Succeeded+r.Failed {
		return errors.New("scheduled result processed must equal succeeded plus failed")
	}
	switch r.Outcome {
	case OutcomeSuccess:
		if r.Failed != 0 {
			return errors.New("successful scheduled result cannot contain failures")
		}
	case OutcomePartial:
		if r.Succeeded == 0 || r.Failed == 0 {
			return errors.New("partial scheduled result requires successes and failures")
		}
	case OutcomeFailed:
		if r.Succeeded != 0 || r.Failed == 0 {
			return errors.New("failed scheduled result requires only failed work")
		}
	default:
		return errors.New("scheduled result outcome is invalid")
	}
	return nil
}

// Executor runs one validated scheduled job.
type Executor interface {
	Execute(ctx context.Context, execution Execution) (Result, error)
}

// Handler is the private scheduled-job HTTP boundary.
type Handler struct {
	executor Executor
}

// NewHandler constructs a scheduled-job handler.
func NewHandler(executor Executor) *Handler {
	return &Handler{executor: executor}
}

// RegisterRoutes registers only the private POST execution contract.
func (h *Handler) RegisterRoutes(routes gin.IRoutes) {
	routes.POST("/_internal/scheduled-jobs/:jobAction", h.run)
}

func (h *Handler) run(c *gin.Context) {
	pathJob, ok := parseJobAction(c.Param("jobAction"))
	if !ok {
		c.JSON(http.StatusNotFound, failedResult())
		return
	}
	if !hasJSONContentType(c.GetHeader("Content-Type")) {
		c.JSON(http.StatusUnsupportedMediaType, failedResult())
		return
	}

	request, err := decodeRunRequest(c)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			c.JSON(http.StatusRequestEntityTooLarge, failedResult())
			return
		}
		c.JSON(http.StatusBadRequest, failedResult())
		return
	}
	if request.validate(pathJob) != nil {
		c.JSON(http.StatusBadRequest, failedResult())
		return
	}
	if h.executor == nil {
		slog.ErrorContext(
			c.Request.Context(),
			"scheduled job executor is not configured",
			"job",
			request.Job,
			"run_id",
			request.RunID,
		)
		c.JSON(http.StatusInternalServerError, failedResult())
		return
	}

	// The Worker lease/fence owns ledger finalization only. Application-side work
	// therefore obeys this deadline and retains domain CAS/idempotency protections.
	ctx, cancel := context.WithTimeout(c.Request.Context(), RequestTimeout)
	defer cancel()
	result, err := h.executor.Execute(ctx, Execution{
		Job:         request.Job,
		ScheduledAt: time.UnixMilli(request.ScheduledTime).UTC(),
		RunID:       request.RunID,
		FenceToken:  request.FenceToken,
	})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"scheduled job execution failed",
			"job",
			request.Job,
			"run_id",
			request.RunID,
			"error",
			err,
		)
		c.JSON(http.StatusInternalServerError, failedResult())
		return
	}
	if validationErr := result.Validate(); validationErr != nil {
		slog.ErrorContext(
			ctx,
			"scheduled job returned an invalid result",
			"job",
			request.Job,
			"run_id",
			request.RunID,
			"error",
			validationErr,
		)
		c.JSON(http.StatusInternalServerError, failedResult())
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseJobAction(action string) (Job, bool) {
	if !strings.HasSuffix(action, runActionSuffix) {
		return "", false
	}
	job := Job(strings.TrimSuffix(action, runActionSuffix))
	return job, job.valid()
}

func (j Job) valid() bool {
	switch j {
	case JobNoShow, JobDelivery, JobDormant:
		return true
	default:
		return false
	}
}

func (r RunRequest) validate(pathJob Job) error {
	if r.Scheduler != SchedulerName {
		return errors.New("scheduler identifier is invalid")
	}
	if !r.Job.valid() || r.Job != pathJob {
		return errors.New("scheduled job does not match path")
	}
	if r.ScheduledTime <= 0 {
		return errors.New("scheduled_time must be positive")
	}
	if strings.TrimSpace(r.RunID) == "" || len(r.RunID) > maxRunIDLength {
		return errors.New("run_id is invalid")
	}
	expectedRunID := fmt.Sprintf(
		"%s:%d:%s",
		SchedulerName,
		r.ScheduledTime,
		r.Job,
	)
	if r.RunID != expectedRunID {
		return errors.New("run_id does not match scheduled job identity")
	}
	if r.FenceToken <= 0 {
		return errors.New("fence_token must be positive")
	}
	return nil
}

func decodeRunRequest(c *gin.Context) (RunRequest, error) {
	var request RunRequest
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxRunRequestBodyBytes,
	)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return RunRequest{}, fmt.Errorf("decode scheduled request: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return RunRequest{}, err
	}
	return request, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("scheduled request contains trailing JSON")
		}
		return fmt.Errorf("decode trailing scheduled request: %w", err)
	}
	return nil
}

func hasJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}

func failedResult() Result {
	return Result{
		Outcome:   OutcomeFailed,
		Processed: 1,
		Failed:    1,
	}
}
