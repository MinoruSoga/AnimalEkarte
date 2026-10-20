package medicalrecord

import (
	"encoding/json"
	"testing"
)

func TestTreatmentResponse_NullableInsuranceJSON(t *testing.T) {
	insured, excluded := true, false
	for _, tc := range []struct {
		name  string
		value *bool
		want  string
	}{
		{name: "unset", want: "null"}, {name: "insured", value: &insured, want: "true"}, {name: "excluded", value: &excluded, want: "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(treatmentResponse{Insurance: tc.value})
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			if got := string(fields["is_insurance"]); got != tc.want {
				t.Fatalf("insurance JSON = %s, want %s", got, tc.want)
			}
		})
	}
}
