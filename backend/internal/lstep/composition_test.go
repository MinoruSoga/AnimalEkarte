package lstep

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestNewApplication_ReturnsTypedConsumers(t *testing.T) {
	app := NewApplication(&Dependencies{})

	require.NotNil(t, app)
	require.NotNil(t, app.TagSync)
	require.NotNil(t, app.DeliveryTrigger)
	require.NotNil(t, app.Batch)
	require.NotNil(t, app.LineCustomers)
}

func TestApplication_NewHandlerOwnsSharedFileRoutes(t *testing.T) {
	app := NewApplication(&Dependencies{})
	h := app.NewHandler(HandlerDependencies{
		RequirePermission: func(string, string) gin.HandlerFunc {
			return func(c *gin.Context) { c.Next() }
		},
		RequireAnyPermission: func(...PermissionRequirement) gin.HandlerFunc {
			return func(c *gin.Context) { c.Next() }
		},
	})

	require.NotNil(t, h)
	require.NotNil(t, h.sharedFile)
}

func TestNewApplication_PropagatesCipherToSettings(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)
	encryptedAPIKey, err := cipher.Encrypt("secret-lstep-api-key")
	require.NoError(t, err)

	settingsRepo := &mockLstepSettingsRepository{
		findByClinicAndServiceFn: func(_ context.Context, _ uint64, _ string) ([]*model.ClinicIntegration, error) {
			return []*model.ClinicIntegration{{
				KeyName:  model.IntegrationKeyLstepAPIKey,
				KeyValue: encryptedAPIKey,
			}}, nil
		},
	}
	app := newApplication(&Dependencies{Cipher: cipher}, &applicationRepositories{
		settings:     settingsRepo,
		syncSettings: &mockLstepSyncSettingsRepository{},
	})

	apiKey, _, _, err := app.graph.settings.GetRawCredentials(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "secret-lstep-api-key", apiKey)
}
