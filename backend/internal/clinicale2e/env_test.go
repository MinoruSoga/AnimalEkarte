package clinicale2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		appEnv  string
		dbHost  string
		wantErr string
	}{
		{name: "test + compose db", appEnv: "test", dbHost: "db"},
		{name: "test + localhost", appEnv: "test", dbHost: "localhost"},
		{name: "test + 127.0.0.1", appEnv: "test", dbHost: "127.0.0.1"},
		{name: "TEST upper", appEnv: "TEST", dbHost: "db"},
		{name: "empty env", appEnv: "", dbHost: "db", wantErr: "APP_ENV"},
		{name: "development", appEnv: "development", dbHost: "db", wantErr: "APP_ENV"},
		{name: "staging", appEnv: "staging", dbHost: "db", wantErr: "APP_ENV"},
		{name: "production", appEnv: "production", dbHost: "localhost", wantErr: "APP_ENV"},
		{name: "planetscale host", appEnv: "test", dbHost: "aws.connect.psdb.cloud", wantErr: "db host"},
		{name: "empty host", appEnv: "test", dbHost: "", wantErr: "db host"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Allow(tt.appEnv, tt.dbHost)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestRejectReservedClinicID(t *testing.T) {
	t.Parallel()
	assert.NoError(t, RejectReservedClinicID(991001))
	require.Error(t, RejectReservedClinicID(1))
	require.Error(t, RejectReservedClinicID(2))
}

func TestLoginEmail(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "e2e-clinical-991234@example.test", LoginEmail(991234))
}
