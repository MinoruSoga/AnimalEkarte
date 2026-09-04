package labdeviceagent

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testReadCloser struct {
	*bytes.Reader
	closed atomic.Bool
}

type blockingReadCloser struct {
	readOnce atomic.Bool
	closed   chan struct{}
}

func (r *blockingReadCloser) Read(target []byte) (int, error) {
	if !r.readOnce.Swap(true) {
		copy(target, []byte{0x02, 0x41, 0x03})
		return 3, nil
	}
	<-r.closed
	return 0, io.EOF
}

func (r *blockingReadCloser) Close() error {
	select {
	case <-r.closed:
	default:
		close(r.closed)
	}
	return nil
}

func (r *testReadCloser) Close() error {
	r.closed.Store(true)
	return nil
}

func TestDiscoverPortsReturnsOnlyStableCalloutPaths(t *testing.T) {
	ports, err := DiscoverPorts(func(pattern string) ([]string, error) {
		require.Equal(t, "/dev/cu.usbserial-*", pattern)
		return []string{
			"/dev/cu.usbserial-Z",
			"/dev/cu.usbserial-A",
		}, nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{"/dev/cu.usbserial-A", "/dev/cu.usbserial-Z"}, ports)
}

func TestNewAgentRecordsConfiguredPortCount(t *testing.T) {
	status := &Status{}
	NewAgent(NewQueue(1), status, []string{"port-a", "port-b"})
	require.Equal(t, 2, status.ConfiguredPorts())
}

func TestDiscoverPortsPropagatesDiscoveryFailure(t *testing.T) {
	want := errors.New("glob failed")
	_, err := DiscoverPorts(func(string) ([]string, error) { return nil, want })
	require.ErrorIs(t, err, want)
}

func TestAgentRunRecordsDiscoveryFailureWithoutPortIdentity(t *testing.T) {
	status := &Status{}
	agent := &Agent{
		queue:        NewQueue(1),
		status:       status,
		allowedPorts: map[string]struct{}{},
		glob:         func(string) ([]string, error) { return nil, errors.New("discovery failed") },
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	agent.Run(ctx)
	require.Equal(t, uint64(1), status.DiscoveryErrors())
	require.Equal(t, "discovery_failed", status.LastErrorCategory())
}

func TestFrameBufferPreservesChunksAndResetsAfterTake(t *testing.T) {
	buffer := &FrameBuffer{}
	buffer.Push([]byte{0x02})
	buffer.Push([]byte{0xff, 0x03})
	require.Equal(t, []byte{0x02, 0xff, 0x03}, buffer.Take())
	require.Nil(t, buffer.Take())
}

func TestFrameBufferFailsClosedAboveBackendLimit(t *testing.T) {
	buffer := &FrameBuffer{}
	require.False(t, buffer.Push(make([]byte, maxFrameBytes+1)))
	require.Nil(t, buffer.Take())
	require.True(t, buffer.Push([]byte{0x02, 0x03}))
	require.Equal(t, []byte{0x02, 0x03}, buffer.Take())
}

type rwPort struct {
	*bytes.Reader
	writes bytes.Buffer
	closed atomic.Bool
}

func (p *rwPort) Write(b []byte) (int, error) {
	return p.writes.Write(b)
}

func (p *rwPort) Close() error {
	p.closed.Store(true)
	return nil
}

func TestMonitorPortDoesNotWriteWithoutPIMSReply(t *testing.T) {
	inquiry := []byte{0x02, 0x31, 0x30, 0x80, 0x81, 0x73, 0xf3, 0x03}
	port := &rwPort{Reader: bytes.NewReader(inquiry)}
	agent := &Agent{
		queue:        NewQueue(2),
		status:       &Status{},
		allowedPorts: map[string]struct{}{"/dev/cu.usbserial-test": {}},
		open:         func(context.Context, string) (io.ReadCloser, error) { return port, nil },
	}
	agent.monitorPort(context.Background(), "/dev/cu.usbserial-test")
	require.Equal(t, 0, port.writes.Len())
}

func TestMonitorPortWritesPIMSRepliesWhenEnabled(t *testing.T) {
	inquiry := []byte{0x02, 0x31, 0x30, 0x80, 0x81, 0x73, 0xf3, 0x03}
	want := [][]byte{{0x06}, {0x41}, {0x53}}
	port := &rwPort{Reader: bytes.NewReader(inquiry)}
	agent := &Agent{
		queue:        NewQueue(2),
		status:       &Status{},
		allowedPorts: map[string]struct{}{"/dev/cu.usbserial-test": {}},
		open:         func(context.Context, string) (io.ReadCloser, error) { return port, nil },
	}
	agent.EnablePIMSReply(func(buf []byte) ([][]byte, int) {
		require.Equal(t, inquiry, buf)
		return want, len(buf)
	})
	agent.monitorPort(context.Background(), "/dev/cu.usbserial-test")
	require.Equal(t, append(append([]byte{0x06}, 0x41), 0x53), port.writes.Bytes())
}

func TestMonitorPortQueuesRawFrameAndClosesReader(t *testing.T) {
	queue := NewQueue(2)
	status := &Status{}
	reader := &testReadCloser{Reader: bytes.NewReader([]byte{0x02, 0xff, 0x03})}
	agent := &Agent{
		queue:        queue,
		status:       status,
		allowedPorts: map[string]struct{}{"/dev/cu.usbserial-test": {}},
		open:         func(context.Context, string) (io.ReadCloser, error) { return reader, nil },
	}

	agent.monitorPort(context.Background(), "/dev/cu.usbserial-test")

	require.Eventually(t, func() bool { return len(queue.Snapshot()) == 1 }, time.Second, time.Millisecond)
	require.Equal(t, []byte{0x02, 0xff, 0x03}, queue.Snapshot()[0].Raw)
	require.True(t, reader.closed.Load())
	require.Equal(t, 0, status.OpenPorts())
}

func TestAgentRunScansPortsAndStops(t *testing.T) {
	queue := NewQueue(2)
	status := &Status{}
	var scans atomic.Int64
	agent := &Agent{
		queue:        queue,
		status:       status,
		allowedPorts: map[string]struct{}{"/dev/cu.usbserial-test": {}},
		glob: func(string) ([]string, error) {
			if scans.Add(1) == 1 {
				return []string{"/dev/cu.usbserial-test"}, nil
			}
			return nil, nil
		},
		open: func(context.Context, string) (io.ReadCloser, error) {
			return &testReadCloser{Reader: bytes.NewReader([]byte{0x02, 0x03})}, nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		agent.Run(ctx)
		close(done)
	}()
	require.Eventually(t, func() bool { return len(queue.Snapshot()) == 1 }, time.Second, time.Millisecond)
	cancel()
	require.Eventually(t, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond)
}

func TestAgentRunCancelsBlockedPortConfiguration(t *testing.T) {
	status := &Status{}
	var openStarted atomic.Bool
	agent := &Agent{
		queue:        NewQueue(1),
		status:       status,
		allowedPorts: map[string]struct{}{"/dev/cu.usbserial-test": {}},
		glob: func(string) ([]string, error) {
			return []string{"/dev/cu.usbserial-test"}, nil
		},
		open: func(ctx context.Context, _ string) (io.ReadCloser, error) {
			openStarted.Store(true)
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		agent.Run(ctx)
		close(done)
	}()
	require.Eventually(t, openStarted.Load, time.Second, time.Millisecond)
	cancel()
	require.Eventually(t, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond)
	require.Zero(t, status.OpenErrors())
}

func TestMonitorPortIgnoresOpenFailure(t *testing.T) {
	agent := &Agent{
		queue:  NewQueue(1),
		status: &Status{},
		open:   func(context.Context, string) (io.ReadCloser, error) { return nil, errors.New("open failed") },
	}
	agent.monitorPort(context.Background(), "/dev/cu.usbserial-test")
	require.Empty(t, agent.queue.Snapshot())
	require.Equal(t, uint64(1), agent.status.OpenErrors())
	require.Equal(t, "port_open_failed", agent.status.LastErrorCategory())
}

func TestMonitorPortFlushesBufferedRawOnCancellation(t *testing.T) {
	queue := NewQueue(1)
	status := &Status{}
	reader := &blockingReadCloser{closed: make(chan struct{})}
	agent := &Agent{queue: queue, status: status, open: func(context.Context, string) (io.ReadCloser, error) { return reader, nil }}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		agent.monitorPort(ctx, "/dev/cu.usbserial-test")
		close(done)
	}()
	require.Eventually(t, func() bool { return status.OpenPorts() == 1 && reader.readOnce.Load() }, time.Second, time.Millisecond)
	cancel()
	require.Eventually(t, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond)
	require.Equal(t, []byte{0x02, 0x41, 0x03}, queue.Snapshot()[0].Raw)
}

func TestMonitorPortRecordsQueueFailureWithoutReplacingExistingRaw(t *testing.T) {
	queue := NewQueue(1)
	_, err := queue.Enqueue([]byte{0x02, 0x41, 0x03}, time.Now())
	require.NoError(t, err)
	status := &Status{}
	agent := &Agent{
		queue:  queue,
		status: status,
		open: func(context.Context, string) (io.ReadCloser, error) {
			return &testReadCloser{Reader: bytes.NewReader([]byte{0x02, 0x42, 0x03})}, nil
		},
	}
	agent.monitorPort(context.Background(), "/dev/cu.usbserial-test")
	require.Equal(t, uint64(1), status.QueueErrors())
	require.Equal(t, "queue_write_failed", status.LastErrorCategory())
	require.Equal(t, []byte{0x02, 0x41, 0x03}, queue.Snapshot()[0].Raw)
}

func TestMonitorPortBoundsIncompletePIMSBuffer(t *testing.T) {
	port := &rwPort{Reader: bytes.NewReader(make([]byte, maxFrameBytes+1))}
	status := &Status{}
	agent := &Agent{
		queue:        NewQueue(2),
		status:       status,
		allowedPorts: map[string]struct{}{"/dev/cu.usbserial-test": {}},
		open: func(context.Context, string) (io.ReadCloser, error) {
			return port, nil
		},
	}
	maxSeen := 0
	agent.EnablePIMSReply(func(buf []byte) ([][]byte, int) {
		if len(buf) > maxSeen {
			maxSeen = len(buf)
		}
		return nil, 0
	})
	agent.monitorPort(context.Background(), "/dev/cu.usbserial-test")
	require.LessOrEqual(t, maxSeen, maxFrameBytes)
	require.Greater(t, status.InputOverflow(), uint64(0))
}
