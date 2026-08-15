package staff

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type compileStaffAccountStore struct{}

func (*compileStaffAccountStore) FindByEmail(
	context.Context,
	string,
) (*model.Account, error) {
	return nil, nil
}

func (*compileStaffAccountStore) Create(context.Context, *model.Account) error {
	return nil
}

func (*compileStaffAccountStore) UpdatePasswordHash(
	context.Context,
	uint64,
	string,
	time.Time,
) error {
	return nil
}

func (*compileStaffAccountStore) DeletePasswordResetTokens(
	context.Context,
	uint64,
) error {
	return nil
}

var _ StaffAccountStore = (*compileStaffAccountStore)(nil)

func TestStaffService_AccountStoreSourceContract(t *testing.T) {
	productionFiles, err := filepath.Glob("staff_service*.go")
	require.NoError(t, err)
	require.NotEmpty(t, productionFiles)
	for _, productionFile := range productionFiles {
		if strings.HasSuffix(productionFile, "_test.go") {
			continue
		}
		source, readErr := os.ReadFile(productionFile)
		require.NoError(t, readErr)
		assert.NotContains(t, string(source), "repository.AccountRepository", productionFile)
	}

	portType := reflect.TypeOf((*StaffAccountStore)(nil)).Elem()
	assert.Equal(t, 4, portType.NumMethod(), "staff account port must remain consumer-minimal")
	methodNames := make([]string, 0, portType.NumMethod())
	for i := 0; i < portType.NumMethod(); i++ {
		methodNames = append(methodNames, portType.Method(i).Name)
	}
	assert.ElementsMatch(
		t,
		[]string{
			"FindByEmail",
			"Create",
			"UpdatePasswordHash",
			"DeletePasswordResetTokens",
		},
		methodNames,
	)
}
