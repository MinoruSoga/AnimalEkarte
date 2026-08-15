package auth

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountCredentialPersistence_SourceContract(t *testing.T) {
	repositoryBytes, err := os.ReadFile("account_repository.go")
	require.NoError(t, err)
	repositorySource := string(repositoryBytes)

	updatePasswordHash := authSourceMethod(
		t,
		repositorySource,
		"func (r *accountRepository) UpdatePasswordHash(",
	)
	assert.Contains(t, updatePasswordHash, "persistence.TxFromContext(ctx)")
	assert.Contains(t, updatePasswordHash, "logger.Silent")
	assert.Contains(t, updatePasswordHash, `"id = ? AND deleted_at IS NULL"`)
	assert.Contains(t, updatePasswordHash, `"password_hash": newHash`)
	assert.Contains(t, updatePasswordHash, "GREATEST")
	assert.Contains(t, updatePasswordHash, "switch result.RowsAffected")
	assert.Contains(t, updatePasswordHash, "case 1:")
	assert.Contains(t, updatePasswordHash, "case 0:")

	assert.NotContains(t, repositorySource, "func (r *accountRepository) Update(")
	assert.NotContains(t, repositorySource, "fields map[string]any")
	repositoryInterfaceStart := strings.Index(
		repositorySource,
		"type AccountRepository interface {",
	)
	require.NotEqual(t, -1, repositoryInterfaceStart)
	repositoryInterface := repositorySource[repositoryInterfaceStart:]
	repositoryInterfaceEnd := strings.Index(repositoryInterface, "\n}")
	require.NotEqual(t, -1, repositoryInterfaceEnd)
	assert.NotContains(
		t,
		repositoryInterface[:repositoryInterfaceEnd],
		"\n\tUpdate(ctx",
	)

	serviceBytes, err := os.ReadFile("account_service.go")
	require.NoError(t, err)
	accountService := string(serviceBytes)
	interfaceStart := strings.Index(accountService, "type AccountService interface {")
	require.NotEqual(t, -1, interfaceStart)
	interfaceSource := accountService[interfaceStart:]
	interfaceEnd := strings.Index(interfaceSource, "\n}")
	require.NotEqual(t, -1, interfaceEnd)
	interfaceSource = interfaceSource[:interfaceEnd]
	assert.NotContains(t, interfaceSource, "UpdatePasswordHash")
	assert.NotContains(t, accountService, "s.repo.Update(ctx, accountID")
}
