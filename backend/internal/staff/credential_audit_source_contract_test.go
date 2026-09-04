package staff

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaffCredentialMutationAuditSourceContract(t *testing.T) {
	handlerSource := readStaffCredentialAuditSource(t, "staff_handler.go")
	assert.NotContains(t, handlerSource, "BestEffort")
	assert.NotContains(t, handlerSource, "context.WithoutCancel")

	serviceSource := sourceAfterStaffAuditMarker(
		t,
		readStaffCredentialAuditSource(t, "staff_service_core.go")+
			"\n"+
			readStaffCredentialAuditSource(t, "staff_service_update.go"),
		"func (s *staffService) Update(",
	)
	assertStaffAuditSourceOrder(
		t,
		serviceSource,
		"s.tx.WithTx(",
		"s.accountRepo.UpdatePasswordHash(",
		"s.accountRepo.DeletePasswordResetTokens(",
		"s.credentialAudit.LogEntryTx(",
	)
}

func readStaffCredentialAuditSource(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(name)
	require.NoError(t, err)
	return string(content)
}

func sourceAfterStaffAuditMarker(
	t *testing.T,
	source, marker string,
) string {
	t.Helper()
	index := strings.Index(source, marker)
	require.NotEqual(t, -1, index, "source marker %q is missing", marker)
	return source[index:]
}

func assertStaffAuditSourceOrder(
	t *testing.T,
	source string,
	markers ...string,
) {
	t.Helper()
	previous := -1
	for _, marker := range markers {
		index := strings.Index(source, marker)
		require.NotEqual(t, -1, index, "source marker %q is missing", marker)
		assert.Greater(t, index, previous, "source marker %q is out of order", marker)
		previous = index
	}
}
