package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lifecycleTestListener struct {
	closeOnce sync.Once
	closed    chan struct{}
}

func newLifecycleTestListener() *lifecycleTestListener {
	return &lifecycleTestListener{closed: make(chan struct{})}
}

func (l *lifecycleTestListener) Accept() (net.Conn, error) {
	<-l.closed
	return nil, net.ErrClosed
}

func (l *lifecycleTestListener) Close() error {
	l.closeOnce.Do(func() {
		close(l.closed)
	})
	return nil
}

func (l *lifecycleTestListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4zero, Port: 8080}
}

type lifecycleTestServer struct {
	serveStarted chan struct{}
	serveRelease chan struct{}
	releaseOnce  sync.Once

	serveErr    error
	shutdownErr error
	closeErr    error

	shutdownCalled      atomic.Bool
	shutdownHasDeadline atomic.Bool
	closeCalled         atomic.Bool
}

func newLifecycleTestServer() *lifecycleTestServer {
	return &lifecycleTestServer{
		serveStarted: make(chan struct{}),
		serveRelease: make(chan struct{}),
		serveErr:     http.ErrServerClosed,
	}
}

func (s *lifecycleTestServer) Serve(net.Listener) error {
	close(s.serveStarted)
	<-s.serveRelease
	return s.serveErr
}

func (s *lifecycleTestServer) Shutdown(ctx context.Context) error {
	s.shutdownCalled.Store(true)
	if _, ok := ctx.Deadline(); ok {
		s.shutdownHasDeadline.Store(true)
	}
	if s.shutdownErr == nil {
		s.release()
	}
	return s.shutdownErr
}

func (s *lifecycleTestServer) Close() error {
	s.closeCalled.Store(true)
	s.release()
	return s.closeErr
}

func (s *lifecycleTestServer) release() {
	s.releaseOnce.Do(func() {
		close(s.serveRelease)
	})
}

type lifecycleTestCloser struct {
	closed  atomic.Bool
	err     error
	onClose func()
}

func (c *lifecycleTestCloser) Close() error {
	if c.onClose != nil {
		c.onClose()
	}
	c.closed.Store(true)
	return c.err
}

func waitForLifecycleRun(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(time.Second):
		t.Fatal("server runner did not stop")
		return nil
	}
}

func TestServerRunner_BindFailureCancelsBackgroundContext(t *testing.T) {
	bindErr := errors.New("address already in use")
	backgroundCtx, cancelBackground := context.WithCancel(context.Background())
	t.Cleanup(cancelBackground)

	resource := &lifecycleTestCloser{}
	server := newLifecycleTestServer()
	runner := serverRunner{
		server:           server,
		address:          ":8080",
		backgroundCtx:    backgroundCtx,
		cancelBackground: cancelBackground,
		listen: func(string, string) (net.Listener, error) {
			return nil, bindErr
		},
		resources: []resourceCloser{resource},
	}

	err := runner.run(context.Background())

	require.ErrorIs(t, err, bindErr)
	assert.False(t, server.shutdownCalled.Load())
	assert.False(t, server.closeCalled.Load())
	assert.True(t, resource.closed.Load())
	select {
	case <-backgroundCtx.Done():
	default:
		t.Fatal("background context was not canceled after bind failure")
	}
}

func TestServerRunner_GracefulShutdownCancelsBackgroundBeforeDrainersAndResources(t *testing.T) {
	backgroundCtx, cancelBackground := context.WithCancel(context.Background())
	server := newLifecycleTestServer()
	listener := newLifecycleTestListener()

	backgroundStopped := make(chan struct{})
	go func() {
		<-backgroundCtx.Done()
		close(backgroundStopped)
	}()

	var drainCalled atomic.Bool
	var resourceClosedTooEarly atomic.Bool
	resource := &lifecycleTestCloser{
		onClose: func() {
			if backgroundCtx.Err() == nil || !drainCalled.Load() {
				resourceClosedTooEarly.Store(true)
			}
		},
	}
	runner := serverRunner{
		server:           server,
		address:          ":8080",
		backgroundCtx:    backgroundCtx,
		cancelBackground: cancelBackground,
		listen: func(string, string) (net.Listener, error) {
			return listener, nil
		},
		shutdownTimeout: time.Second,
		drainers: []func(){
			func() {
				select {
				case <-backgroundStopped:
				case <-time.After(time.Second):
					return
				}
				drainCalled.Store(true)
			},
		},
		resources: []resourceCloser{resource},
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- runner.run(runCtx)
	}()

	select {
	case <-server.serveStarted:
	case <-time.After(time.Second):
		t.Fatal("server did not start serving")
	}

	cancelRun()
	require.NoError(t, waitForLifecycleRun(t, runDone))

	assert.True(t, server.shutdownCalled.Load())
	assert.True(t, server.shutdownHasDeadline.Load())
	assert.False(t, server.closeCalled.Load())
	assert.True(t, drainCalled.Load())
	assert.True(t, resource.closed.Load())
	assert.False(t, resourceClosedTooEarly.Load())
}

func TestServerRunner_ShutdownFailureFallsBackToClose(t *testing.T) {
	shutdownErr := errors.New("shutdown deadline exceeded")
	server := newLifecycleTestServer()
	server.shutdownErr = shutdownErr
	listener := newLifecycleTestListener()

	backgroundCtx, cancelBackground := context.WithCancel(context.Background())
	backgroundStopped := make(chan struct{})
	var backgroundCanceled atomic.Bool
	go func() {
		<-backgroundCtx.Done()
		backgroundCanceled.Store(true)
		close(backgroundStopped)
	}()
	var resourceClosedTooEarly atomic.Bool
	resource := &lifecycleTestCloser{
		onClose: func() {
			if !backgroundCanceled.Load() {
				resourceClosedTooEarly.Store(true)
			}
		},
	}
	runner := serverRunner{
		server:           server,
		address:          ":8080",
		backgroundCtx:    backgroundCtx,
		cancelBackground: cancelBackground,
		listen: func(string, string) (net.Listener, error) {
			return listener, nil
		},
		shutdownTimeout: time.Second,
		drainers: []func(){
			func() {
				select {
				case <-backgroundStopped:
				case <-time.After(time.Second):
				}
			},
		},
		resources: []resourceCloser{resource},
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- runner.run(runCtx)
	}()

	select {
	case <-server.serveStarted:
	case <-time.After(time.Second):
		t.Fatal("server did not start serving")
	}
	cancelRun()

	err := waitForLifecycleRun(t, runDone)
	require.ErrorIs(t, err, shutdownErr)
	assert.True(t, server.shutdownCalled.Load())
	assert.True(t, server.closeCalled.Load())
	assert.True(t, backgroundCanceled.Load())
	assert.True(t, resource.closed.Load())
	assert.False(t, resourceClosedTooEarly.Load())
}

func TestServerRunner_ServeFailureStopsBackgroundWorkAndReturns(t *testing.T) {
	serveErr := errors.New("accept failed")
	server := newLifecycleTestServer()
	server.serveErr = serveErr
	server.release()
	listener := newLifecycleTestListener()

	backgroundCtx, cancelBackground := context.WithCancel(context.Background())
	backgroundStopped := make(chan struct{})
	var backgroundCanceled atomic.Bool
	go func() {
		<-backgroundCtx.Done()
		backgroundCanceled.Store(true)
		close(backgroundStopped)
	}()
	resource := &lifecycleTestCloser{}
	runner := serverRunner{
		server:           server,
		address:          ":8080",
		backgroundCtx:    backgroundCtx,
		cancelBackground: cancelBackground,
		listen: func(string, string) (net.Listener, error) {
			return listener, nil
		},
		shutdownTimeout: time.Second,
		drainers: []func(){
			func() {
				select {
				case <-backgroundStopped:
				case <-time.After(time.Second):
				}
			},
		},
		resources: []resourceCloser{resource},
	}

	err := runner.run(context.Background())

	require.ErrorIs(t, err, serveErr)
	assert.True(t, server.shutdownCalled.Load())
	assert.False(t, server.closeCalled.Load())
	assert.True(t, backgroundCanceled.Load())
	assert.True(t, resource.closed.Load())
}

func TestServerRunner_DefaultShutdownTimeoutCoversScheduledRequestDeadline(t *testing.T) {
	runner := &serverRunner{}

	assert.GreaterOrEqual(
		t,
		runner.effectiveShutdownTimeout(),
		130*time.Second,
		"100s scheduled job deadline requires at least a 30s shutdown margin",
	)
}
