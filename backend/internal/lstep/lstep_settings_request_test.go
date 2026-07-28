package lstep

// lstep_settings_request_test.go — BE9-2C L①: handler/upsert_misc_request_test.go から
// updateLstepSettingsRequest のテストを実装と同 package へ移動。

import (
	"testing"
)

func TestUpdateLstepSettingsRequest_ToServiceInput(t *testing.T) {
	isSyncEnabled := false
	cpmVersion := "v2"
	noahLTV := int64(10000)
	lookbackDays := 30
	req := updateLstepSettingsRequest{
		LstepAPIKey:                  "key",
		IsSyncEnabled:                &isSyncEnabled,
		CPMVersion:                   &cpmVersion,
		CPMV1NoahLTV:                 &noahLTV,
		HealthPreventionLookbackDays: &lookbackDays,
	}

	input := req.toServiceInput()

	if input.LstepAPIKey != req.LstepAPIKey {
		t.Errorf("LstepAPIKey = %q, want %q", input.LstepAPIKey, req.LstepAPIKey)
	}
	if input.IsSyncEnabled == nil || *input.IsSyncEnabled {
		t.Errorf("IsSyncEnabled = %v, want false pointer", input.IsSyncEnabled)
	}
	if input.CPMVersion == nil || *input.CPMVersion != cpmVersion {
		t.Errorf("CPMVersion = %v, want %q", input.CPMVersion, cpmVersion)
	}
	if input.CPMV1NoahLTV == nil || *input.CPMV1NoahLTV != noahLTV {
		t.Errorf("CPMV1NoahLTV = %v, want %d", input.CPMV1NoahLTV, noahLTV)
	}
	if input.HealthPreventionLookbackDays == nil || *input.HealthPreventionLookbackDays != lookbackDays {
		t.Errorf("HealthPreventionLookbackDays = %v, want %d", input.HealthPreventionLookbackDays, lookbackDays)
	}
}
