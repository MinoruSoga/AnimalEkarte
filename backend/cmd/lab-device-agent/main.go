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
	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

type stringListFlag []string

func (f *stringListFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringListFlag) Set(value string) error {
	if _, ok := labdeviceagent.NormalizeAllowedOrigin(value); !ok {
		return errors.New("allowed-origin must use lowercase http(s) with canonical IPv4, non-mapped IPv6, or strict ASCII DNS")
	}
	*f = append(*f, value)
	return nil
}

func main() {
	clinicID := flag.String("clinic-id", "", "clinic ID bound to this workstation")
	portsFile := flag.String("ports-file", "", "newline-delimited allowlist of serial ports")
	consumerToken := flag.String("consumer-token", "", "consumer token required for protected loopback operations")
	var allowedOrigins stringListFlag
	flag.Var(&allowedOrigins, "allowed-origin", "exact supported deployed frontend origin allowed to use the loopback agent (repeatable)")
	pimsReply := flag.Bool("pims-reply", false, "write IDEXX ACK+A+IM/SM on the same usbserial port; do not use on hospital VetLab")
	flag.Parse()
	if *clinicID == "" || *portsFile == "" || *consumerToken == "" {
		slog.Error("clinic-id, ports-file, and consumer-token are required")
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
	if *pimsReply {
		slog.Info("idexx pims reply enabled; do not use on hospital VetLab")
		agent.UseReadWriteSerial()
		agent.EnablePIMSReply(func(buf []byte) ([][]byte, int) {
			replies, rest := medicalrecord.DrainIDEXXPIMSReplies(buf, medicalrecord.IDEXXPIMSJoutoHost, time.Now())
			return replies, len(buf) - len(rest)
		})
	}
	server := &http.Server{
		Addr:              labdeviceagent.ListenAddress,
		Handler:           labdeviceagent.NewHandler(queue, status, *clinicID, *consumerToken, allowedOrigins...),
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
