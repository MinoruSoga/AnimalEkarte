package billing

import "testing"

func TestInsuranceApplicableFromMasterSetting(t *testing.T) {
	for _, setting := range []bool{true, false} {
		t.Run(map[bool]string{true: "excluded", false: "insured"}[setting], func(t *testing.T) {
			got := insuranceApplicableFromMasterSetting(&setting)
			if got == nil || *got != !setting {
				t.Fatalf("setting %v: got %v, want %v", setting, got, !setting)
			}
			original := setting
			*got = setting
			if setting != original {
				t.Fatal("candidate mutation changed master input")
			}
		})
	}
	if got := insuranceApplicableFromMasterSetting(nil); got != nil {
		t.Fatalf("legacy NULL: got %v, want nil", *got)
	}
}
