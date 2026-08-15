package owner

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type lineWebhookEventType string

const (
	lineWebhookEventFollow   lineWebhookEventType = "follow"
	lineWebhookEventUnfollow lineWebhookEventType = "unfollow"
)

func TestOwnerRepository_LineWebhookCASUpdates_RollBackWithAmbientTransaction(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	base := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)
	rollbackErr := errors.New("force LINE webhook CAS rollback")

	tests := []struct {
		name             string
		eventType        lineWebhookEventType
		initialFollowed  *time.Time
		initialBlocked   *time.Time
		staleEventAt     time.Time
		acceptedEventAt  time.Time
		wantTxFollowed   *time.Time
		wantTxBlocked    *time.Time
		wantBaseFollowed *time.Time
		wantBaseBlocked  *time.Time
	}{
		{
			name:             "follow update preserves clinic line ID and ordering predicates then rolls back",
			eventType:        lineWebhookEventFollow,
			initialFollowed:  timePointer(base),
			initialBlocked:   timePointer(base.Add(time.Minute)),
			staleEventAt:     base.Add(time.Minute),
			acceptedEventAt:  base.Add(2 * time.Minute),
			wantTxFollowed:   timePointer(base.Add(2 * time.Minute)),
			wantTxBlocked:    nil,
			wantBaseFollowed: timePointer(base),
			wantBaseBlocked:  timePointer(base.Add(time.Minute)),
		},
		{
			name:             "unfollow update accepts the follow timestamp boundary then rolls back",
			eventType:        lineWebhookEventUnfollow,
			initialFollowed:  timePointer(base.Add(2 * time.Minute)),
			initialBlocked:   timePointer(base),
			staleEventAt:     base.Add(time.Minute),
			acceptedEventAt:  base.Add(2 * time.Minute),
			wantTxFollowed:   timePointer(base.Add(2 * time.Minute)),
			wantTxBlocked:    timePointer(base.Add(2 * time.Minute)),
			wantBaseFollowed: timePointer(base.Add(2 * time.Minute)),
			wantBaseBlocked:  timePointer(base),
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const clinicID = uint64(1)
			lineUserID := fmt.Sprintf("U-webhook-tx-%d", i)
			record := makeTestOwner(t, db, clinicID, fmt.Sprintf("LINE webhook tx owner %d", i))
			require.NoError(t, repo.UpdateLineUserID(
				context.Background(),
				clinicID,
				record.ID,
				&lineUserID,
			))
			require.NoError(t, db.Model(&model.Owner{}).
				Where("id = ?", record.ID).
				Updates(map[string]any{
					"line_followed_at": tt.initialFollowed,
					"line_blocked_at":  tt.initialBlocked,
				}).Error)

			err := persistence.NewTransactor(db).WithTx(
				context.Background(),
				func(txCtx context.Context) error {
					updated, updateErr := updateLineWebhookEventForTest(
						txCtx,
						repo,
						tt.eventType,
						clinicID+1,
						record.ID,
						lineUserID,
						tt.acceptedEventAt.Add(time.Minute),
					)
					require.NoError(t, updateErr)
					assert.False(t, updated, "foreign clinic must not update the owner")

					updated, updateErr = updateLineWebhookEventForTest(
						txCtx,
						repo,
						tt.eventType,
						clinicID,
						record.ID,
						"U-stale-link",
						tt.acceptedEventAt.Add(time.Minute),
					)
					require.NoError(t, updateErr)
					assert.False(t, updated, "stale LINE User ID must fail the CAS")

					updated, updateErr = updateLineWebhookEventForTest(
						txCtx,
						repo,
						tt.eventType,
						clinicID,
						record.ID,
						lineUserID,
						tt.staleEventAt,
					)
					require.NoError(t, updateErr)
					assert.False(t, updated, "stale event ordering must be a no-op")

					updated, updateErr = updateLineWebhookEventForTest(
						txCtx,
						repo,
						tt.eventType,
						clinicID,
						record.ID,
						lineUserID,
						tt.acceptedEventAt,
					)
					require.NoError(t, updateErr)
					require.True(t, updated)

					got := findLineWebhookOwnerForTest(t, txCtx, db, record.ID)
					assertOptionalTimeEqual(t, tt.wantTxFollowed, got.LineFollowedAt)
					assertOptionalTimeEqual(t, tt.wantTxBlocked, got.LineBlockedAt)
					return rollbackErr
				},
			)
			require.ErrorIs(t, err, rollbackErr)

			persisted := findLineWebhookOwnerForTest(
				t,
				context.Background(),
				db,
				record.ID,
			)
			assertOptionalTimeEqual(t, tt.wantBaseFollowed, persisted.LineFollowedAt)
			assertOptionalTimeEqual(t, tt.wantBaseBlocked, persisted.LineBlockedAt)
		})
	}
}

func updateLineWebhookEventForTest(
	ctx context.Context,
	repo Repository,
	eventType lineWebhookEventType,
	clinicID, ownerID uint64,
	lineUserID string,
	eventAt time.Time,
) (bool, error) {
	switch eventType {
	case lineWebhookEventFollow:
		return repo.UpdateLineFollowedAt(ctx, clinicID, ownerID, lineUserID, eventAt)
	case lineWebhookEventUnfollow:
		return repo.UpdateLineBlockedAt(ctx, clinicID, ownerID, lineUserID, eventAt)
	default:
		return false, fmt.Errorf("unsupported LINE webhook event type %q", eventType)
	}
}

func findLineWebhookOwnerForTest(
	t *testing.T,
	ctx context.Context,
	db *gorm.DB,
	ownerID uint64,
) *model.Owner {
	t.Helper()
	var record model.Owner
	require.NoError(t, persistence.DBOrTx(ctx, db).
		Where("id = ?", ownerID).
		First(&record).Error)
	return &record
}

func assertOptionalTimeEqual(t *testing.T, want, got *time.Time) {
	t.Helper()
	if want == nil {
		assert.Nil(t, got)
		return
	}
	require.NotNil(t, got)
	assert.Equal(t, want.UTC(), got.UTC())
}

func timePointer(value time.Time) *time.Time {
	return &value
}
