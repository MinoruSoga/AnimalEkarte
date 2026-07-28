package owner

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestOwnerRegistrationPetDrafts_PreservesPhone(t *testing.T) {
	drafts := ownerRegistrationPetDrafts([]model.Pet{{Phone: "090-1234-5678"}})

	require.Len(t, drafts, 1)
	require.Equal(t, "090-1234-5678", drafts[0].Phone)
}
