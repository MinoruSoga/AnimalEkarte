package smtp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRejectsEnvelopeInjectionBeforeDial(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{
			name: "mail from",
			from: "sender@example.com\r\nRCPT TO:<attacker@example.com>",
			to:   "owner@example.com",
		},
		{
			name: "recipient",
			from: "sender@example.com",
			to:   "owner@example.com\nDATA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Send(context.Background(), Config{
				Host: "127.0.0.1",
				Port: "1",
			}, tt.from, tt.to, nil)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "CR or LF")
		})
	}
}

func TestSendRespectsContextDeadlineDuringSMTPGreeting(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan struct{})
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		close(accepted)
		time.Sleep(2 * time.Second)
	}()

	host, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	startedAt := time.Now()
	sendErr := Send(ctx, Config{Host: host, Port: port}, "sender@example.com", "owner@example.com", nil)
	elapsed := time.Since(startedAt)

	<-accepted
	require.Error(t, sendErr)
	assert.Less(t, elapsed, time.Second)
}

func TestValidateEnvelopeAddress(t *testing.T) {
	assert.NoError(t, ValidateEnvelopeAddress("owner@example.com"))
	assert.Error(t, ValidateEnvelopeAddress("owner@example.com\rSubject: injected"))
	assert.Error(t, ValidateEnvelopeAddress("owner@example.com\nSubject: injected"))
}
