package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/scheduler"
)

type scheduledBatchCall struct {
	job         scheduler.Job
	scheduledAt time.Time
	runID       string
}

type scheduledBatchStub struct {
	calls   []scheduledBatchCall
	results map[scheduler.Job]lstep.BatchRunResult
}

func (s *scheduledBatchStub) record(
	job scheduler.Job,
	scheduledAt time.Time,
	runID string,
) lstep.BatchRunResult {
	s.calls = append(s.calls, scheduledBatchCall{
		job:         job,
		scheduledAt: scheduledAt,
		runID:       runID,
	})
	return s.results[job]
}

func (s *scheduledBatchStub) RunNoShowCheckAllClinicsAt(
	_ context.Context,
	scheduledAt time.Time,
	runID string,
) lstep.BatchRunResult {
	return s.record(scheduler.JobNoShow, scheduledAt, runID)
}

func (s *scheduledBatchStub) RunDeliveryTriggerBatchAllClinicsAt(
	_ context.Context,
	scheduledAt time.Time,
	runID string,
) lstep.BatchRunResult {
	return s.record(scheduler.JobDelivery, scheduledAt, runID)
}

func (s *scheduledBatchStub) RunDormantDetectionAllClinicsAt(
	_ context.Context,
	scheduledAt time.Time,
	runID string,
) lstep.BatchRunResult {
	return s.record(scheduler.JobDormant, scheduledAt, runID)
}

func TestLstepScheduledJobExecutor_RoutesJobsWithStableScheduleIdentity(t *testing.T) {
	scheduledAt := time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		job     scheduler.Job
		result  lstep.BatchRunResult
		outcome scheduler.Outcome
	}{
		{
			name:    "no show success",
			job:     scheduler.JobNoShow,
			result:  lstep.BatchRunResult{Processed: 2, Succeeded: 2},
			outcome: scheduler.OutcomeSuccess,
		},
		{
			name:    "delivery partial",
			job:     scheduler.JobDelivery,
			result:  lstep.BatchRunResult{Processed: 3, Succeeded: 2, Failed: 1},
			outcome: scheduler.OutcomePartial,
		},
		{
			name:    "dormant failed",
			job:     scheduler.JobDormant,
			result:  lstep.BatchRunResult{Processed: 1, Failed: 1},
			outcome: scheduler.OutcomeFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := &scheduledBatchStub{
				results: map[scheduler.Job]lstep.BatchRunResult{tt.job: tt.result},
			}
			executor := newLstepScheduledJobExecutor(batch)

			result, err := executor.Execute(context.Background(), scheduler.Execution{
				Job:         tt.job,
				ScheduledAt: scheduledAt,
				RunID:       "stable-run-id",
				FenceToken:  9,
			})

			require.NoError(t, err)
			assert.Equal(t, scheduler.Result{
				Outcome:   tt.outcome,
				Processed: tt.result.Processed,
				Succeeded: tt.result.Succeeded,
				Failed:    tt.result.Failed,
			}, result)
			require.Len(t, batch.calls, 1)
			assert.Equal(t, tt.job, batch.calls[0].job)
			assert.Equal(t, scheduledAt, batch.calls[0].scheduledAt)
			assert.Equal(t, "stable-run-id", batch.calls[0].runID)
		})
	}
}

func TestLstepScheduledJobExecutor_RejectsUnsupportedJob(t *testing.T) {
	executor := newLstepScheduledJobExecutor(&scheduledBatchStub{})

	_, err := executor.Execute(context.Background(), scheduler.Execution{
		Job:         scheduler.Job("forbidden"),
		ScheduledAt: time.Now(),
		RunID:       "run",
		FenceToken:  1,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scheduled job")
}

func TestLstepScheduledJobExecutor_FailsClosedForInvalidBatchResult(t *testing.T) {
	batch := &scheduledBatchStub{
		results: map[scheduler.Job]lstep.BatchRunResult{
			scheduler.JobNoShow: {Processed: 2, Succeeded: 1},
		},
	}
	executor := newLstepScheduledJobExecutor(batch)

	_, err := executor.Execute(context.Background(), scheduler.Execution{
		Job:         scheduler.JobNoShow,
		ScheduledAt: time.Now(),
		RunID:       "run",
		FenceToken:  1,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid batch result")
}

func TestRegisterScheduledJobRoutes_ExposesOnlyInternalPOSTContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const token = "test-scheduler-internal-token-32b!!"
	router := gin.New()
	batch := &scheduledBatchStub{
		results: map[scheduler.Job]lstep.BatchRunResult{
			scheduler.JobNoShow: {Processed: 1, Succeeded: 1},
		},
	}
	registerScheduledJobRoutes(router, batch, token)

	body := `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":1785201200000,"run_id":"animalekarte-scheduler-v1:1785201200000:no_show","fence_token":1}`
	request := httptest.NewRequest(
		http.MethodPost,
		"/_internal/scheduled-jobs/no_show:run",
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(schedulerInternalTokenHeader, token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"outcome":"success","processed":1,"succeeded":1,"failed":0}`, response.Body.String())
	require.Len(t, batch.calls, 1)

	unauth := httptest.NewRequest(
		http.MethodPost,
		"/_internal/scheduled-jobs/no_show:run",
		strings.NewReader(body),
	)
	unauth.Header.Set("Content-Type", "application/json")
	unauthResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthResponse, unauth)
	assert.Equal(t, http.StatusUnauthorized, unauthResponse.Code)

	wrong := httptest.NewRequest(
		http.MethodPost,
		"/_internal/scheduled-jobs/no_show:run",
		strings.NewReader(body),
	)
	wrong.Header.Set("Content-Type", "application/json")
	wrong.Header.Set(schedulerInternalTokenHeader, "wrong-scheduler-internal-token-32b!")
	wrongResponse := httptest.NewRecorder()
	router.ServeHTTP(wrongResponse, wrong)
	assert.Equal(t, http.StatusUnauthorized, wrongResponse.Code)
	assert.Len(t, batch.calls, 1)

	getRequest := httptest.NewRequest(
		http.MethodGet,
		"/_internal/scheduled-jobs/no_show:run",
		http.NoBody,
	)
	getRequest.Header.Set(schedulerInternalTokenHeader, token)
	getResponse := httptest.NewRecorder()
	router.ServeHTTP(getResponse, getRequest)
	assert.Equal(t, http.StatusNotFound, getResponse.Code)
}

func TestRegisterBaseRoutes_LocalUploadsAndSchedulerAuthSurface(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("STORAGE_TYPE", "")
	t.Setenv("SCHEDULER_INTERNAL_TOKEN", "test-scheduler-internal-token-32b!!")
	router := gin.New()
	require.NoError(t, registerBaseRoutes(router, nil))
	routes := make(map[string]struct{})
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	assert.Contains(t, routes, http.MethodGet+" /uploads/*filepath")
	assert.Contains(t, routes, http.MethodPost+" /_internal/scheduled-jobs/:jobAction")

	t.Setenv("STORAGE_TYPE", "s3")
	routerS3 := gin.New()
	require.NoError(t, registerBaseRoutes(routerS3, nil))
	routesS3 := make(map[string]struct{})
	for _, route := range routerS3.Routes() {
		routesS3[route.Method+" "+route.Path] = struct{}{}
	}
	assert.NotContains(t, routesS3, http.MethodGet+" /uploads/*filepath")
}

func TestRegisterScheduledJobRoutes_EmptyTokenRejectsAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	batch := &scheduledBatchStub{
		results: map[scheduler.Job]lstep.BatchRunResult{
			scheduler.JobNoShow: {Processed: 1, Succeeded: 1},
		},
	}
	registerScheduledJobRoutes(router, batch, "")
	body := `{"scheduler":"animalekarte-scheduler-v1","job":"no_show","scheduled_time":1785201200000,"run_id":"animalekarte-scheduler-v1:1785201200000:no_show","fence_token":1}`
	request := httptest.NewRequest(
		http.MethodPost,
		"/_internal/scheduled-jobs/no_show:run",
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Empty(t, batch.calls)
}

func TestLstepScheduledJobExecutor_NilBatchFailsClosed(t *testing.T) {
	executor := newLstepScheduledJobExecutor(nil)

	_, err := executor.Execute(context.Background(), scheduler.Execution{
		Job:         scheduler.JobNoShow,
		ScheduledAt: time.Now(),
		RunID:       "run",
		FenceToken:  1,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, errScheduledBatchUnavailable))
}
