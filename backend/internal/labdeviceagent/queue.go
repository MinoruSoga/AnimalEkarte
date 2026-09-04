package labdeviceagent

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrEmptyFrame        = errors.New("empty lab device frame")
	ErrQueueFull         = errors.New("lab device frame queue is full")
	ErrRejectedQueueFull = errors.New("lab device rejected frame queue is full")
	ErrFrameNotFound     = errors.New("lab device frame not found")
)

type Frame struct {
	ID         string
	Raw        []byte
	ReceivedAt time.Time
}

type QueueStats struct {
	Capacity int    `json:"capacity"`
	Pending  int    `json:"pending"`
	Rejected int    `json:"rejected"`
	Overflow uint64 `json:"overflow"`
}

type Queue struct {
	mu       sync.RWMutex
	capacity int
	nextID   uint64
	pending  []Frame
	rejected []Frame
	overflow uint64
}

func NewQueue(capacity int) *Queue {
	if capacity < 1 {
		capacity = 1
	}
	return &Queue{capacity: capacity}
}

func (q *Queue) Enqueue(raw []byte, receivedAt time.Time) (Frame, error) {
	if len(raw) == 0 {
		return Frame{}, ErrEmptyFrame
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.pending) >= q.capacity {
		q.overflow++
		return Frame{}, ErrQueueFull
	}
	q.nextID++
	frame := Frame{
		ID:         fmt.Sprintf("frame-%d", q.nextID),
		Raw:        append([]byte(nil), raw...),
		ReceivedAt: receivedAt.UTC(),
	}
	q.pending = append(q.pending, frame)
	return cloneFrame(frame), nil
}

func (q *Queue) Snapshot() []Frame {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return cloneFrames(q.pending)
}

func (q *Queue) RejectedSnapshot() []Frame {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return cloneFrames(q.rejected)
}

func (q *Queue) Ack(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	index := frameIndex(q.pending, id)
	if index < 0 {
		return ErrFrameNotFound
	}
	q.pending = append(q.pending[:index], q.pending[index+1:]...)
	return nil
}

func (q *Queue) Reject(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	index := frameIndex(q.pending, id)
	if index < 0 {
		return ErrFrameNotFound
	}
	if len(q.rejected) >= q.capacity {
		return ErrRejectedQueueFull
	}
	q.rejected = append(q.rejected, q.pending[index])
	q.pending = append(q.pending[:index], q.pending[index+1:]...)
	return nil
}

func (q *Queue) Stats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return QueueStats{
		Capacity: q.capacity,
		Pending:  len(q.pending),
		Rejected: len(q.rejected),
		Overflow: q.overflow,
	}
}

func frameIndex(frames []Frame, id string) int {
	for index := range frames {
		if frames[index].ID == id {
			return index
		}
	}
	return -1
}

func cloneFrame(frame Frame) Frame {
	frame.Raw = append([]byte(nil), frame.Raw...)
	return frame
}

func cloneFrames(frames []Frame) []Frame {
	cloned := make([]Frame, len(frames))
	for index := range frames {
		cloned[index] = cloneFrame(frames[index])
	}
	return cloned
}
