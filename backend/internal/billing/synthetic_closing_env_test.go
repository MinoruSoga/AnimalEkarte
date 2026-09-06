package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowUATSyntheticClosing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		appEnv  string
		dbHost  string
		wantErr string
	}{
		{name: "development + compose db", appEnv: "development", dbHost: "db"},
		{name: "test + localhost", appEnv: "test", dbHost: "localhost"},
		{name: "local + 127.0.0.1", appEnv: "local", dbHost: "127.0.0.1"},
		{name: "dev + db", appEnv: "dev", dbHost: "db"},
		{name: "empty env", appEnv: "", dbHost: "db", wantErr: "APP_ENV"},
		{name: "staging", appEnv: "staging", dbHost: "db", wantErr: "APP_ENV"},
		{name: "production", appEnv: "production", dbHost: "localhost", wantErr: "APP_ENV"},
		{name: "planetscale host", appEnv: "development", dbHost: "aws.connect.psdb.cloud", wantErr: "db host"},
		{name: "empty host", appEnv: "development", dbHost: "", wantErr: "db host"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := AllowUATSyntheticClosing(tt.appEnv, tt.dbHost)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestRejectExistingBillingIDs(t *testing.T) {
	t.Parallel()
	assert.NoError(t, RejectExistingBillingIDs(nil))
	assert.NoError(t, RejectExistingBillingIDs([]uint64{}))
	err := RejectExistingBillingIDs([]uint64{9})
	require.Error(t, err)
	assert.ErrorContains(t, err, "existing billing")
}

func TestRejectReservedClinicID(t *testing.T) {
	t.Parallel()
	assert.NoError(t, RejectReservedClinicID(910001))
	require.Error(t, RejectReservedClinicID(1))
	require.Error(t, RejectReservedClinicID(2))
}
