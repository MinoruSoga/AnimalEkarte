package auth

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialMutationAuditSourceContract(t *testing.T) {
	httpSource := readCredentialAuditSource(t, "http_password.go")
	assert.NotContains(t, httpSource, "BestEffort")
	assert.NotContains(t, httpSource, "context.WithoutCancel")

	accountSource := sourceFromMarker(
		t,
		readCredentialAuditSource(t, "account_service.go"),
		"func (s *accountService) ChangePassword(",
	)
	assertSourceOrder(
		t,
		accountSource,
		"s.transactor.WithTx(",
		"s.repo.CompareAndSwapPasswordHash(",
		"s.resetTokens.DeleteByAccountID(",
		"s.audit.LogEntryTx(",
	)

	resetSource := sourceFromMarker(
		t,
		readCredentialAuditSource(t, "password_reset_service.go"),
		"func (s *passwordResetService) consumeResetToken(",
	)
	assert.Contains(t, resetSource, "s.transactor.WithTx(")
	resetMutationSource := sourceFromMarker(
		t,
		resetSource,
		"s.credentialUpdater.UpdatePasswordHash(",
	)
	assertSourceOrder(
		t,
		resetMutationSource,
		"s.credentialUpdater.UpdatePasswordHash(",
		"s.tokenRepo.ConsumeByID(",
		"s.auditSubject.ResolveCredentialAuditSubject(",
		"s.audit.LogEntryTx(",
	)
}

func readCredentialAuditSource(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(name)
	require.NoError(t, err)
	return string(content)
}

func sourceFromMarker(t *testing.T, source, marker string) string {
	t.Helper()
	index := strings.Index(source, marker)
	require.NotEqual(t, -1, index, "source marker %q is missing", marker)
	return source[index:]
}

func assertSourceOrder(t *testing.T, source string, markers ...string) {
	t.Helper()
	previous := -1
	for _, marker := range markers {
		index := strings.Index(source, marker)
		require.NotEqual(t, -1, index, "source marker %q is missing", marker)
		assert.Greater(t, index, previous, "source marker %q is out of order", marker)
		previous = index
	}
}
