package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestNextVisitExpiryRepair_TagFailuresKeepRetryCandidate(t *testing.T) {
	for _, failure := range []string{"add", "remove"} {
		t.Run(failure, func(t *testing.T) {
			oldTag, newTag := "next_visit_2026-10-11", "next_visit_2026-10-13"
			cache := map[string]bool{oldTag: true}
			remote := map[string]bool{oldTag: true}
			fail := true
			var events []string
			lineID := "mock-only"
			apiErr := errors.New("mock tag failure")
			client := &mockLstepAPIClient{
				getUserTagsFn: func(context.Context, string) ([]string, error) {
					var out []string
					for tag := range remote {
						out = append(out, tag)
					}
					return out, nil
				},
				addTagFn: func(_ context.Context, _, tag string) error {
					events = append(events, "add:"+tag)
					if fail && failure == "add" {
						return apiErr
					}
					remote[tag] = true
					return nil
				},
				removeTagFn: func(_ context.Context, _, tag string) error {
					events = append(events, "remove:"+tag)
					assert.NotEqual(t, newTag, tag)
					if fail && failure == "remove" {
						return apiErr
					}
					delete(remote, tag)
					return nil
				},
			}
			svc := &lstepTagSyncService{
				settingsSvc: &mockLstepSettingsService{},
				ownerRepo: &mockOwnerRepository{findByIDFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
					return &model.Owner{ID: 2, LineUserID: &lineID}, nil
				}},
				buildClientFn: func(context.Context, uint64) (lstepapi.Client, error) { return client, nil },
				nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 6, Status: model.ReservationStatusConfirmed, StartTime: time.Date(2026, 10, 13, 1, 0, 0, 0, time.UTC)}, nil
				}},
				tagCacheRepo: &mockLstepTagCacheRepository{
					findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
						var out []*model.LstepTagCache
						for tag := range cache {
							out = append(out, &model.LstepTagCache{TagName: tag})
						}
						return out, nil
					},
					upsertTagFn: func(_ context.Context, _, _ uint64, tag, _, _ string) error { cache[tag] = true; return nil },
					deleteTagFn: func(_ context.Context, _, _ uint64, tag string) error { delete(cache, tag); return nil },
				},
			}
			require.ErrorIs(t, svc.SyncNextVisitTag(context.Background(), 1, 2), apiErr)
			assert.True(t, cache[oldTag], "failed replacement must keep the expiry candidate")
			assert.Equal(t, "add:"+newTag, events[0], "replacement is added before old tag removal")
			fail = false
			events = nil
			require.NoError(t, svc.SyncNextVisitTag(context.Background(), 1, 2))
			assert.False(t, cache[oldTag])
			assert.True(t, cache[newTag])
			if failure == "remove" {
				assert.NotContains(t, events, "add:"+newTag, "successful prior add is not repeated")
			}
			events = nil
			require.NoError(t, svc.SyncNextVisitTag(context.Background(), 1, 2))
			assert.Empty(t, events, "already correct remote causes no redundant tag mutations")
		})
	}
}
