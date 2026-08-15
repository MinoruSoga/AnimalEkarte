package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackgroundWorkGate_CloseAndWaitRejectsLaterRegistrations(t *testing.T) {
	var gate backgroundWorkGate

	require.True(t, gate.Register())

	waitReturned := make(chan struct{})
	go func() {
		gate.CloseAndWait()
		close(waitReturned)
	}()

	require.Eventually(t, func() bool {
		gate.mu.Lock()
		defer gate.mu.Unlock()
		return gate.closed
	}, time.Second, time.Millisecond)
	assert.False(t, gate.Register())

	select {
	case <-waitReturned:
		t.Fatal("CloseAndWait returned before registered work completed")
	default:
	}

	gate.Done()
	select {
	case <-waitReturned:
	case <-time.After(time.Second):
		t.Fatal("CloseAndWait did not return after registered work completed")
	}
}
