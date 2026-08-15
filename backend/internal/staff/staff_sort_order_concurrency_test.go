package staff_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	. "github.com/animal-ekarte/backend/internal/staff"
)

type swapBarrierContextKey struct{}

type swapUpdateBarrier struct {
	mu       sync.Mutex
	seen     map[string]struct{}
	arrivals int
	release  chan struct{}
	once     sync.Once
}

func newSwapUpdateBarrier() *swapUpdateBarrier {
	return &swapUpdateBarrier{
		seen:    make(map[string]struct{}),
		release: make(chan struct{}),
	}
}

func (b *swapUpdateBarrier) beforeUpdate(tx *gorm.DB) {
	token, ok := tx.Statement.Context.Value(swapBarrierContextKey{}).(string)
	if !ok {
		return
	}
	b.mu.Lock()
	if _, seen := b.seen[token]; seen {
		b.mu.Unlock()
		return
	}
	b.seen[token] = struct{}{}
	b.arrivals++
	if b.arrivals == 2 {
		b.once.Do(func() { close(b.release) })
	}
	b.mu.Unlock()

	select {
	case <-b.release:
	case <-time.After(250 * time.Millisecond):
	}
}

func TestStaffRepository_SwapSortOrderForReservation_SerializesOverlappingSwaps(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	const clinicID = uint64(1)
	first := makeAssignedDoctor(t, db, clinicID, "並行swap-1", 1)
	second := makeAssignedDoctor(t, db, clinicID, "並行swap-2", 2)
	third := makeAssignedDoctor(t, db, clinicID, "並行swap-3", 3)

	barrier := newSwapUpdateBarrier()
	callbackName := "staff_test:swap_update_barrier"
	require.NoError(t, db.Callback().Update().
		Before("gorm:update").
		Register(callbackName, barrier.beforeUpdate))
	t.Cleanup(func() {
		_ = db.Callback().Update().Remove(callbackName)
	})

	repo := NewStaffRepository(db)
	results := make(chan error, 2)
	start := make(chan struct{})
	for token, request := range map[string]struct {
		staffID   uint64
		direction string
	}{
		"swap-second-up": {staffID: second.ID, direction: "up"},
		"swap-third-up":  {staffID: third.ID, direction: "up"},
	} {
		token, request := token, request
		go func() {
			<-start
			ctx := context.WithValue(context.Background(), swapBarrierContextKey{}, token)
			results <- repo.SwapSortOrderForReservation(
				ctx,
				clinicID,
				request.staffID,
				request.direction,
			)
		}()
	}
	close(start)
	for range 2 {
		require.NoError(t, <-results)
	}

	orders := make([]int, 0, 3)
	for _, staffID := range []uint64{first.ID, second.ID, third.ID} {
		var sortOrder int
		require.NoError(t, db.Model(&struct {
			SortOrder int
		}{}).
			Table("staffs").
			Select("sort_order").
			Where("id = ?", staffID).
			Scan(&sortOrder).Error)
		orders = append(orders, sortOrder)
	}
	assert.ElementsMatch(t, []int{1, 2, 3}, orders, fmt.Sprintf("duplicate sort orders: %v", orders))
}
