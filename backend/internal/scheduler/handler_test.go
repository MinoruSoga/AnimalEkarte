package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingExecutor struct {
	execution Execution
	deadline  time.Time
	result    Result
	err       error
	calls     int
}

func (e *recordingExecutor) Execute(ctx context.Context, execution Execution) (Result, error) {
	e.calls++
	e.execution = execution
	e.deadline, _ = ctx.Deadline()
	return e.result, e.err
}

func newSchedulerTestRouter(executor Executor) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(executor).RegisterRoutes(router)
	return router
}

func performScheduledRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func validScheduledRequestBody(job Job) string {
	body, err := json.Marshal(RunRequest{
		Scheduler:     SchedulerName,
		Job:           job,
		ScheduledTime: 1_785_201_200_000,
		RunID:         SchedulerName + ":1785201200000:" + string(job),
		FenceToken:    7,
	})
	if err != nil {
		panic(err)
	}
	return string(body)
}

func TestHandler_Run_ExecutesExactContractWithHundredSecondDeadline(t *testing.T) {
	executor := &recordingExecutor{
		result: Result{Outcome: OutcomeSuccess, Processed: 3, Succeeded: 3, Failed: 0},
	}
	router := newSchedulerTestRouter(executor)

	startedAt := time.Now()
	response := performScheduledRequest(
		t,
		router,
		http.MethodPost,
		"/_internal/scheduled-jobs/no_show:run",
		validScheduledRequestBody(JobNoShow),
	)

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"outcome":"success","processed":3,"succeeded":3,"failed":0}`, response.Body.String())
	require.Equal(t, 1, executor.calls)
	assert.Equal(t, JobNoShow, executor.execution.Job)
	assert.Equal(t, time.UnixMilli(1_785_201_200_000).UTC(), executor.execution.ScheduledAt)
	assert.Equal(t, SchedulerName+":1785201200000:no_show", executor.execution.RunID)
	assert.Equal(t, int64(7), executor.execution.FenceToken)
	require.False(t, executor.deadline.IsZero())
	assert.WithinDuration(t, startedAt.Add(RequestTimeout), executor.deadline, time.Second)
}

func TestHandler_Run_PreservesStrictPartialOutcome(t *testing.T) {
	executor := &recordingExecutor{
		result: Result{Outcome: OutcomePartial, Processed: 3, Succeeded: 2, Failed: 1},
	}
	response := performScheduledRequest(
		t,
		newSchedulerTestRouter(executor),
		http.MethodPost,
		"/_internal/scheduled-jobs/delivery:run",
		validScheduledRequestBody(JobDelivery),
	)

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"outcome":"partial","processed":3,"succeeded":2,"failed":1}`, response.Body.String())
}

func TestHandler_Run_RejectsInvalidBoundaryInputWithoutExecution(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
		code int
	}{
		{
			name: "unknown path job",
			path: "/_internal/scheduled-jobs/ltv:run",
			body: validScheduledRequestBody(JobNoShow),
			code: http.StatusNotFound,
		},
		{
			name: "path must end in exact run action",
			path: "/_internal/scheduled-jobs/no_show:preview",
			body: validScheduledRequestBody(JobNoShow),
			code: http.StatusNotFound,
		},
		{
			name: "scheduler mismatch",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: `{"scheduler":"other","job":"no_show","scheduled_time":1785201200000,"run_id":"run","fence_token":1}`,
			code: http.StatusBadRequest,
		},
		{
			name: "body job mismatch",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: validScheduledRequestBody(JobDormant),
			code: http.StatusBadRequest,
		},
		{
			name: "scheduled time must be positive",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":0,"run_id":"run","fence_token":1}`,
			code: http.StatusBadRequest,
		},
		{
			name: "run id must not be blank",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":1785201200000,"run_id":" ","fence_token":1}`,
			code: http.StatusBadRequest,
		},
		{
			name: "run id must match deterministic schedule identity",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":1785201200000,"run_id":"animalekarte-scheduler-v1:1785201200001:no_show","fence_token":1}`,
			code: http.StatusBadRequest,
		},
		{
			name: "fence token must be positive",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":1785201200000,"run_id":"run","fence_token":0}`,
			code: http.StatusBadRequest,
		},
		{
			name: "unknown field is rejected",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":1785201200000,"run_id":"run","fence_token":1,"extra":true}`,
			code: http.StatusBadRequest,
		},
		{
			name: "trailing JSON is rejected",
			path: "/_internal/scheduled-jobs/no_show:run",
			body: validScheduledRequestBody(JobNoShow) + `{}`,
			code: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &recordingExecutor{
				result: Result{Outcome: OutcomeSuccess},
			}
			response := performScheduledRequest(
				t,
				newSchedulerTestRouter(executor),
				http.MethodPost,
				tt.path,
				tt.body,
			)

			assert.Equal(t, tt.code, response.Code)
			assert.JSONEq(t, `{"outcome":"failed","processed":1,"succeeded":0,"failed":1}`, response.Body.String())
			assert.Zero(t, executor.calls)
		})
	}
}

func TestHandler_Run_RejectsUnsupportedMediaTypeAndOversizedBody(t *testing.T) {
	t.Run("content type must be application json", func(t *testing.T) {
		executor := &recordingExecutor{
			result: Result{Outcome: OutcomeSuccess},
		}
		router := newSchedulerTestRouter(executor)
		request := httptest.NewRequest(
			http.MethodPost,
			"/_internal/scheduled-jobs/no_show:run",
			strings.NewReader(validScheduledRequestBody(JobNoShow)),
		)
		request.Header.Set("Content-Type", "text/plain")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusUnsupportedMediaType, response.Code)
		assert.JSONEq(t, `{"outcome":"failed","processed":1,"succeeded":0,"failed":1}`, response.Body.String())
		assert.Zero(t, executor.calls)
	})

	t.Run("body is capped before decoding", func(t *testing.T) {
		executor := &recordingExecutor{
			result: Result{Outcome: OutcomeSuccess},
		}
		oversizedRunID := strings.Repeat("x", maxRunRequestBodyBytes)
		body := `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":1785201200000,"run_id":"` +
			oversizedRunID +
			`","fence_token":1}`

		response := performScheduledRequest(
			t,
			newSchedulerTestRouter(executor),
			http.MethodPost,
			"/_internal/scheduled-jobs/no_show:run",
			body,
		)

		assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
		assert.JSONEq(t, `{"outcome":"failed","processed":1,"succeeded":0,"failed":1}`, response.Body.String())
		assert.Zero(t, executor.calls)
	})
}

func TestHandler_Run_FailsClosedForExecutorErrorOrInvalidResult(t *testing.T) {
	tests := []struct {
		name     string
		executor *recordingExecutor
	}{
		{
			name: "executor error",
			executor: &recordingExecutor{
				err: errors.New("database unavailable"),
			},
		},
		{
			name: "success cannot contain failed work",
			executor: &recordingExecutor{
				result: Result{Outcome: OutcomeSuccess, Processed: 2, Succeeded: 1, Failed: 1},
			},
		},
		{
			name: "processed must equal succeeded plus failed",
			executor: &recordingExecutor{
				result: Result{Outcome: OutcomePartial, Processed: 3, Succeeded: 1, Failed: 1},
			},
		},
		{
			name: "failed outcome needs a failed item",
			executor: &recordingExecutor{
				result: Result{Outcome: OutcomeFailed, Processed: 0, Succeeded: 0, Failed: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performScheduledRequest(
				t,
				newSchedulerTestRouter(tt.executor),
				http.MethodPost,
				"/_internal/scheduled-jobs/dormant:run",
				validScheduledRequestBody(JobDormant),
			)

			assert.Equal(t, http.StatusInternalServerError, response.Code)
			assert.JSONEq(t, `{"outcome":"failed","processed":1,"succeeded":0,"failed":1}`, response.Body.String())
			assert.NotContains(t, response.Body.String(), "database unavailable")
		})
	}
}

func TestHandler_Run_FailsClosedWhenExecutorIsNotConfigured(t *testing.T) {
	response := performScheduledRequest(
		t,
		newSchedulerTestRouter(nil),
		http.MethodPost,
		"/_internal/scheduled-jobs/no_show:run",
		validScheduledRequestBody(JobNoShow),
	)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"outcome":"failed","processed":1,"succeeded":0,"failed":1}`, response.Body.String())
}
