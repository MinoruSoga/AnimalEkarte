package auth

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func authSourceMethod(t *testing.T, source, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	require.NotEqual(t, -1, start, "missing method %q", signature)
	method := source[start:]
	if next := strings.Index(method[len(signature):], "\nfunc "); next >= 0 {
		method = method[:len(signature)+next]
	}
	return method
}

func TestAccountPasswordChange_SourceContract(t *testing.T) {
	serviceBytes, err := os.ReadFile("account_service.go")
	require.NoError(t, err)
	serviceMethod := authSourceMethod(
		t,
		string(serviceBytes),
		"func (s *accountService) ChangePassword(",
	)
	find := strings.Index(serviceMethod, "s.repo.FindByID")
	verify := strings.Index(serviceMethod, "bcrypt.CompareHashAndPassword")
	hash := strings.Index(serviceMethod, "bcrypt.GenerateFromPassword")
	compareAndSwap := strings.Index(serviceMethod, "s.repo.CompareAndSwapPasswordHash")
	for _, position := range []int{find, verify, hash, compareAndSwap} {
		require.NotEqual(t, -1, position)
	}
	assert.Less(t, find, verify)
	assert.Less(t, verify, hash)
	assert.Less(t, hash, compareAndSwap)

	repositoryBytes, err := os.ReadFile("account_repository.go")
	require.NoError(t, err)
	repositoryMethod := authSourceMethod(
		t,
		string(repositoryBytes),
		"func (r *accountRepository) CompareAndSwapPasswordHash(",
	)
	assert.Contains(
		t,
		repositoryMethod,
		`id = ? AND deleted_at IS NULL AND password_hash = ?`,
	)
	assert.Contains(t, repositoryMethod, `"password_hash": newHash`)
	assert.Contains(t, repositoryMethod, `"updated_at":    updatedAt`)
	assert.Contains(t, repositoryMethod, "result.RowsAffected == 1")

	handlerBytes, err := os.ReadFile("http_password.go")
	require.NoError(t, err)
	handlerMethod := authSourceMethod(
		t,
		string(handlerBytes),
		"func (h *HTTPHandler) ChangeMyPassword(",
	)
	assert.Contains(t, handlerMethod, "passwordChanger.ChangePassword")
	assert.NotContains(t, handlerMethod, "h.deps.Accounts.GetByID(")
	assert.NotContains(t, handlerMethod, "h.deps.Accounts.UpdatePasswordHash(")
}
