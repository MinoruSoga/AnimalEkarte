package config

// config_plane_test.go — Plane 連携の APP_ENV 名前空間（PR650-R3）と
// label UUID 設定（PR650-R4）の Load/Validate 契約。
//
// PlaneEnvironment は APP_ENV から allowlist (development/staging/production) で
// 導出する。debug では未設定/未知値を development 既定とするが、release で
// PLANE_API_KEY が有効な場合は空のまま Validate が拒否する（黙って
// development/staging に寄せない）。label ID は enabled 連携で必須の UUID。

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPlaneSourceLabelID = "31530203-1a6e-445f-9952-01c6eac98fe1"
	testPlaneLaneLabelID   = "7a38738a-d932-4166-a7fa-87524f29b824"
)

func TestLoad_PlaneEnvironmentFollowsAllowlistedAppEnv(t *testing.T) {
	for _, appEnv := range []string{"development", "staging", "production"} {
		t.Run(appEnv, func(t *testing.T) {
			t.Setenv("GIN_MODE", "debug")
			t.Setenv("APP_ENV", appEnv)
			assert.Equal(t, appEnv, Load().PlaneEnvironment)
		})
	}
}

func TestLoad_PlaneEnvironmentDebugDefaultsMissingOrUnknownToDevelopment(t *testing.T) {
	t.Setenv("GIN_MODE", "debug")
	for _, appEnv := range []string{"", "test", "LOCAL", "e2e"} {
		t.Setenv("APP_ENV", appEnv)
		assert.Equal(t, "development", Load().PlaneEnvironment, "APP_ENV=%q", appEnv)
	}
}

// release では未知/未設定の APP_ENV を黙って development/staging に寄せない —
// 空のまま残して enabled 連携の Validate が拒否できるようにする。
func TestLoad_PlaneEnvironmentReleaseLeavesUnknownEmpty(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	for _, appEnv := range []string{"", "test", "e2e"} {
		t.Setenv("APP_ENV", appEnv)
		assert.Empty(t, Load().PlaneEnvironment, "APP_ENV=%q", appEnv)
	}
	t.Setenv("APP_ENV", "production")
	assert.Equal(t, "production", Load().PlaneEnvironment)
}

func TestLoad_PlaneLabelIDs(t *testing.T) {
	t.Setenv("PLANE_SOURCE_LABEL_ID", testPlaneSourceLabelID)
	t.Setenv("PLANE_LANE_LABEL_ID", testPlaneLaneLabelID)
	cfg := Load()
	assert.Equal(t, testPlaneSourceLabelID, cfg.PlaneSourceLabelID)
	assert.Equal(t, testPlaneLaneLabelID, cfg.PlaneLaneLabelID)
}

// enabled 連携（PLANE_API_KEY 設定）はモードに関わらず label UUID を要求する。
func TestValidate_EnabledPlaneRequiresLabelUUIDs(t *testing.T) {
	tests := []struct {
		name     string
		sourceID string
		laneID   string
		wantErr  string
	}{
		{name: "missing source label", sourceID: "", laneID: testPlaneLaneLabelID, wantErr: "PLANE_SOURCE_LABEL_ID"},
		{name: "missing lane label", sourceID: testPlaneSourceLabelID, laneID: "", wantErr: "PLANE_LANE_LABEL_ID"},
		{name: "malformed source label", sourceID: "not-a-uuid", laneID: testPlaneLaneLabelID, wantErr: "PLANE_SOURCE_LABEL_ID"},
		{name: "malformed lane label", sourceID: testPlaneSourceLabelID, laneID: "lane:dev", wantErr: "PLANE_LANE_LABEL_ID"},
		{name: "valid label UUIDs pass", sourceID: testPlaneSourceLabelID, laneID: testPlaneLaneLabelID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				GinMode:            "debug",
				JWTSecret:          "test-secret-not-dev-default",
				PlaneAPIKey:        "plane-api-key",
				PlaneSourceLabelID: tt.sourceID,
				PlaneLaneLabelID:   tt.laneID,
			}
			err := cfg.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

// 連携無効（PLANE_API_KEY 未設定）なら label 未設定でも起動できる。
func TestValidate_DisabledPlaneSkipsLabelCheck(t *testing.T) {
	cfg := &Config{
		GinMode:   "debug",
		JWTSecret: "test-secret-not-dev-default",
	}
	assert.NoError(t, cfg.Validate())
}

// release で enabled 連携は allowlist 外/未設定の APP_ENV を拒否する。
func TestValidate_ReleaseEnabledPlaneRequiresAllowlistedAppEnv(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.PlaneAPIKey = "plane-api-key"
	cfg.PlaneWorkspaceSlug = "baritechllc"
	cfg.PlaneProjectID = "6a009324-8a8a-455c-b16a-41b3f6207b0e"
	cfg.PlaneSourceLabelID = testPlaneSourceLabelID
	cfg.PlaneLaneLabelID = testPlaneLaneLabelID

	cfg.PlaneEnvironment = ""
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APP_ENV")

	for _, env := range []string{"development", "staging", "production"} {
		cfg.PlaneEnvironment = env
		assert.NoError(t, cfg.Validate(), "PlaneEnvironment=%q", env)
	}
}
