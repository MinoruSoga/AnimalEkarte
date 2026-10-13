package support

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestServiceCreate_ConcurrentPlaneClaimResponse(t *testing.T) {
	for _, manual := range []bool{false, true} {
		for _, reloadFails := range []bool{false, true} {
			name := "automatic"
			if manual {
				name = "manual"
			}
			if reloadFails {
				name += "/reload_failure"
			} else {
				name += "/claimed"
			}
			t.Run(name, func(t *testing.T) {
				var logs bytes.Buffer
				previousLogger := slog.Default()
				slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
				defer slog.SetDefault(previousLogger)
				winnerID, winnerURL := "winner", "https://app.plane.so/winner"
				latest := &model.SupportBugReport{ID: 42, ClinicID: 7, Title: "local report", PlaneIssueID: &winnerID, PlaneIssueURL: &winnerURL}
				reads := 0
				repo := &mockRepository{
					createFn:             func(_ context.Context, report *model.SupportBugReport) error { report.ID = 42; return nil },
					setCreationSyncErrFn: func(context.Context, uint64, uint64, string) (bool, error) { return false, nil },
					findByIDForClinicF: func(_ context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
						require.Equal(t, uint64(7), clinicID)
						require.Equal(t, uint64(42), id)
						reads++
						if manual && reads == 1 {
							return &model.SupportBugReport{ID: 42, ClinicID: 7, Title: "local report"}, nil
						}
						if reloadFails {
							return nil, errors.New("reload unavailable")
						}
						return latest, nil
					},
				}
				svc := NewService(repo, &mockTicketCreator{err: &planeUpstreamError{status: 500}}, nil)
				var got *model.SupportBugReport
				var err error
				if manual {
					got, err = svc.EnsurePlaneTicket(context.Background(), 7, 42)
				} else {
					got, err = svc.Create(context.Background(), 7, 1, CreateBugReportInput{Title: "local report"})
				}
				if reloadFails && manual {
					require.Error(t, err)
					assert.True(t, errors.Is(err, apperrors.ErrBadGateway))
					assert.Nil(t, got)
				} else {
					require.NoError(t, err)
					require.NotNil(t, got)
					if !reloadFails {
						require.NotNil(t, got.PlaneIssueID)
						assert.Equal(t, winnerID, *got.PlaneIssueID)
						assert.Equal(t, &winnerURL, got.PlaneIssueURL)
						assert.Nil(t, got.PlaneSyncError)
					}
				}
				if reloadFails {
					assert.Contains(t, logs.String(), "failed to reload competing plane claim")
					assert.Contains(t, logs.String(), "reload unavailable")
				} else {
					assert.NotContains(t, logs.String(), "plane ticket creation failed")
				}
				expectedReads := 1
				if manual {
					expectedReads = 2
				}
				assert.Equal(t, expectedReads, reads)
			})
		}
	}
}

// Genuine creation/persistence failures must not become success after the race fix.
func TestServiceCreate_ManualPlaneFailureRemainsBadGateway(t *testing.T) {
	for _, tc := range []struct {
		name       string
		written    bool
		persistErr error
		wantReads  int
	}{
		{name: "unclaimed creation failure", written: true, wantReads: 1},
		{name: "persistence error", persistErr: errors.New("write unavailable"), wantReads: 1},
		{name: "missing report", persistErr: apperrors.WrapNotFound("support_bug_report", "42"), wantReads: 1},
		{name: "no-op without successful claim", wantReads: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads := 0
			repo := &mockRepository{
				findByIDForClinicF: func(_ context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
					require.Equal(t, uint64(7), clinicID)
					require.Equal(t, uint64(42), id)
					reads++
					return &model.SupportBugReport{ID: id, ClinicID: clinicID}, nil
				},
				setCreationSyncErrFn: func(context.Context, uint64, uint64, string) (bool, error) { return tc.written, tc.persistErr },
			}
			report, err := NewService(repo, &mockTicketCreator{err: &planeUpstreamError{status: 500}}, nil).EnsurePlaneTicket(context.Background(), 7, 42)
			require.Error(t, err)
			assert.ErrorIs(t, err, apperrors.ErrBadGateway)
			assert.Nil(t, report)
			assert.Equal(t, tc.wantReads, reads)
		})
	}
}
