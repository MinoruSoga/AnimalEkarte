package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/animal-ekarte/backend/internal/labdeviceagent"
)

func main() {
	clinicID := flag.String("clinic-id", "", "clinic ID bound to this workstation")
	portsFile := flag.String("ports-file", "", "newline-delimited allowlist of serial ports")
	flag.Parse()
	if *clinicID == "" || *portsFile == "" {
		slog.Error("clinic-id and ports-file are required")
		os.Exit(2)
	}
	portsRaw, err := os.ReadFile(*portsFile)
	if err != nil {
		slog.Error("serial allowlist is unavailable")
		os.Exit(2)
	}
	var allowedPorts []string
	for _, path := range strings.Split(string(portsRaw), "\n") {
		path = strings.TrimSpace(path)
		if strings.HasPrefix(path, "/dev/cu.usbserial-") {
			allowedPorts = append(allowedPorts, path)
		}
	}
	if len(allowedPorts) == 0 {
		slog.Error("serial allowlist is empty")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	queue := labdeviceagent.NewQueue(100)
	status := &labdeviceagent.Status{}
	agent := labdeviceagent.NewAgent(queue, status, allowedPorts)
	server := &http.Server{
		Addr:              labdeviceagent.ListenAddress,
		Handler:           labdeviceagent.NewHandler(queue, status, *clinicID),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	agentDone := make(chan struct{})
	go func() {
		defer close(agentDone)
		agent.Run(ctx)
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			slog.Error("lab device agent shutdown failed")
		}
	}()

	slog.Info("lab device agent started", "listen", labdeviceagent.ListenAddress)
	serveErr := server.ListenAndServe()
	stop()
	<-agentDone
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		slog.Error("lab device agent stopped unexpectedly")
		os.Exit(1)
	}
}
