package labdeviceagent

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueuePreservesRawFramesUntilTerminalDecision(t *testing.T) {
	queue := NewQueue(2)
	raw := []byte{0x02, 0xff, 0x03}
	frame, err := queue.Enqueue(raw, time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	raw[1] = 0x00

	snapshot := queue.Snapshot()
	require.Len(t, snapshot, 1)
	require.Equal(t, []byte{0x02, 0xff, 0x03}, snapshot[0].Raw)
	require.Equal(t, frame.ID, snapshot[0].ID)
	require.Len(t, queue.Snapshot(), 1, "reading must not remove an unacknowledged frame")

	require.NoError(t, queue.Ack(frame.ID))
	require.Empty(t, queue.Snapshot())
	require.ErrorIs(t, queue.Ack(frame.ID), ErrFrameNotFound)
}

func TestQueueFailsClosedWhenFullAndRetainsRejectedRaw(t *testing.T) {
	queue := NewQueue(1)
	frame, err := queue.Enqueue([]byte{0x02, 0x41, 0x03}, time.Now())
	require.NoError(t, err)

	_, err = queue.Enqueue([]byte{0x02, 0x42, 0x03}, time.Now())
	require.ErrorIs(t, err, ErrQueueFull)
	require.Len(t, queue.Snapshot(), 1)
	require.Equal(t, uint64(1), queue.Stats().Overflow)

	require.NoError(t, queue.Reject(frame.ID))
	require.Empty(t, queue.Snapshot())
	require.Equal(t, 1, queue.Stats().Rejected)
	require.Equal(t, []byte{0x02, 0x41, 0x03}, queue.RejectedSnapshot()[0].Raw)
}

func TestQueueRejectsInvalidInput(t *testing.T) {
	queue := NewQueue(1)
	_, err := queue.Enqueue(nil, time.Now())
	require.ErrorIs(t, err, ErrEmptyFrame)
	require.True(t, errors.Is(queue.Reject("missing"), ErrFrameNotFound))
}

func TestQueueCapacityBoundariesPreserveRejectedRaw(t *testing.T) {
	t.Run("rejected frame does not consume pending capacity", func(t *testing.T) {
		queue := NewQueue(1)
		first, err := queue.Enqueue([]byte{0x02, 0x41, 0x03}, time.Now())
		require.NoError(t, err)
		require.NoError(t, queue.Reject(first.ID))

		secondRaw := []byte{0x02, 0x42, 0x03}
		_, err = queue.Enqueue(secondRaw, time.Now())
		require.NoError(t, err)
		require.Equal(t, []byte{0x02, 0x41, 0x03}, queue.RejectedSnapshot()[0].Raw)
		require.Equal(t, secondRaw, queue.Snapshot()[0].Raw)
		require.Equal(t, QueueStats{Capacity: 1, Pending: 1, Rejected: 1}, queue.Stats())
	})

	t.Run("full rejected lane keeps the next raw frame pending", func(t *testing.T) {
		queue := NewQueue(1)
		first, err := queue.Enqueue([]byte{0x02, 0x41, 0x03}, time.Now())
		require.NoError(t, err)
		require.NoError(t, queue.Reject(first.ID))
		second, err := queue.Enqueue([]byte{0x02, 0x42, 0x03}, time.Now())
		require.NoError(t, err)

		err = queue.Reject(second.ID)
		require.ErrorIs(t, err, ErrRejectedQueueFull)
		require.Equal(t, []byte{0x02, 0x41, 0x03}, queue.RejectedSnapshot()[0].Raw)
		require.Equal(t, []byte{0x02, 0x42, 0x03}, queue.Snapshot()[0].Raw)
	})

	t.Run("full pending lane retains the original raw frame", func(t *testing.T) {
		queue := NewQueue(1)
		_, err := queue.Enqueue([]byte{0x02, 0x41, 0x03}, time.Now())
		require.NoError(t, err)
		_, err = queue.Enqueue([]byte{0x02, 0x42, 0x03}, time.Now())
		require.ErrorIs(t, err, ErrQueueFull)
		require.Equal(t, []byte{0x02, 0x41, 0x03}, queue.Snapshot()[0].Raw)
		require.Equal(t, uint64(1), queue.Stats().Overflow)
	})
}
