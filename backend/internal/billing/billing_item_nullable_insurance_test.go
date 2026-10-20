package billing

import (
	"encoding/json"
	"testing"
)

// This tests response/revision JSON, not database persistence or receipt adoption.
func TestBillingItemNullableInsuranceJSON(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value *bool
		want  string
	}{
		{name: "unset", want: "null"},
		{name: "excluded", value: nullableInsuranceBool(false), want: "false"},
		{name: "insured", value: nullableInsuranceBool(true), want: "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range []interface{}{BillingItemResponse{IsInsuranceApplicable: tc.value}, unbilledRevisionItem{IsInsuranceApplicable: tc.value}} {
				raw, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(raw, &fields); err != nil {
					t.Fatal(err)
				}
				if got := string(fields["is_insurance_applicable"]); got != tc.want {
					t.Fatalf("insurance JSON = %s, want %s", got, tc.want)
				}
			}
		})
	}
}
func nullableInsuranceBool(value bool) *bool { return &value }

func TestUnbilledRevision_NullInsuranceDiffersFromExplicitFalse(t *testing.T) {
	unknown, err := json.Marshal(unbilledRevisionItem{})
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := json.Marshal(unbilledRevisionItem{IsInsuranceApplicable: nullableInsuranceBool(false)})
	if err != nil {
		t.Fatal(err)
	}
	if string(unknown) == string(explicit) {
		t.Fatal("revision canonical JSON collapsed unknown into false")
	}
}
