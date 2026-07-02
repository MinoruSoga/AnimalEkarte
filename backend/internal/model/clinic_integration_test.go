package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsEncryptedKey guards which clinic_integrations.key_name values are
// AES-256-GCM encrypted at rest. A false negative here would mean a secret
// (API key / channel token) is persisted in plaintext.
func TestIsEncryptedKey(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
		want    bool
	}{
		{"lstep api key is encrypted", IntegrationKeyLstepAPIKey, true},
		{"line channel access token is encrypted", IntegrationKeyLineChannelAccessToken, true},
		{"line channel secret is encrypted", IntegrationKeyLineChannelSecret, true},
		{"lstep base url is not encrypted", IntegrationKeyLstepBaseURL, false},
		{"liff id is not encrypted", IntegrationKeyLiffID, false},
		{"line account name is not encrypted", IntegrationKeyLineAccountName, false},
		{"unknown key name is not encrypted", "unknown_key", false},
		{"empty key name is not encrypted", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsEncryptedKey(tt.keyName))
		})
	}
}
