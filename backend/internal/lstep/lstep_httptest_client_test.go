package lstep

import (
	"testing"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
)

func useHttptestLstepClient(t *testing.T) {
	t.Helper()
	// Swaps a package-level factory. Do not call t.Parallel() from tests that use this.
	orig := newLstepAPIClient
	newLstepAPIClient = lstepapi.NewInsecureTestClient
	t.Cleanup(func() { newLstepAPIClient = orig })
}
