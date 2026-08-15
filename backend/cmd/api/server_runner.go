package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const defaultShutdownTimeout = 130 * time.Second

type lifecycleServer interface {
	Serve(net.Listener) error
	Shutdown(context.Context) error
	Close() error
}

type resourceCloser interface {
	Close() error
}

type serverRunner struct {
	server  lifecycleServer
	network string
	address string
	listen  func(network, address string) (net.Listener, error)

	backgroundCtx    context.Context
	cancelBackground context.CancelFunc
	drainers         []func()
	resources        []resourceCloser
	shutdownTimeout  time.Duration
}

func (r *serverRunner) run(runCtx context.Context) (resultErr error) {
	if runCtx == nil {
		return errors.New("run HTTP server: context is required")
	}
	_, cancelBackground := r.backgroundLifecycle(runCtx)
	defer func() {
		cancelBackground()
		resultErr = errors.Join(resultErr, closeResources(r.resources))
	}()

	if r.server == nil {
		return errors.New("run HTTP server: server is required")
	}

	listener, err := r.listenerFactory()(r.listenerNetwork(), r.address)
	if err != nil {
		return fmt.Errorf("bind HTTP listener %q: %w", r.address, err)
	}
	defer func() {
		_ = listener.Close()
	}()

	serveResult := make(chan error, 1)
	go func() {
		serveResult <- r.server.Serve(listener)
	}()

	serveFinished := false
	select {
	case <-runCtx.Done():
	case err := <-serveResult:
		serveFinished = true
		resultErr = errors.Join(resultErr, normalizeServeError(err))
	}

	cancelBackground()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.WithoutCancel(runCtx), r.effectiveShutdownTimeout())
	shutdownErr := r.server.Shutdown(shutdownCtx)
	cancelShutdown()
	if shutdownErr != nil {
		resultErr = errors.Join(resultErr, fmt.Errorf("shutdown HTTP server: %w", shutdownErr))
		if closeErr := r.server.Close(); closeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close HTTP server after shutdown failure: %w", closeErr))
		}
	}

	if !serveFinished {
		resultErr = errors.Join(resultErr, normalizeServeError(<-serveResult))
	}

	for _, drain := range r.drainers {
		if drain != nil {
			drain()
		}
	}

	return resultErr
}

func (r *serverRunner) backgroundLifecycle(parent context.Context) (context.Context, context.CancelFunc) {
	if r.backgroundCtx != nil && r.cancelBackground != nil {
		return r.backgroundCtx, r.cancelBackground
	}
	return context.WithCancel(parent)
}

func (r *serverRunner) listenerFactory() func(string, string) (net.Listener, error) {
	if r.listen != nil {
		return r.listen
	}
	return net.Listen
}

func (r *serverRunner) listenerNetwork() string {
	if r.network != "" {
		return r.network
	}
	return "tcp"
}

func (r *serverRunner) effectiveShutdownTimeout() time.Duration {
	if r.shutdownTimeout > 0 {
		return r.shutdownTimeout
	}
	return defaultShutdownTimeout
}

func normalizeServeError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("serve HTTP: %w", err)
}

func closeResources(resources []resourceCloser) error {
	var resultErr error
	for index := len(resources) - 1; index >= 0; index-- {
		if resources[index] == nil {
			continue
		}
		if err := resources[index].Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close process resource: %w", err))
		}
	}
	return resultErr
}
