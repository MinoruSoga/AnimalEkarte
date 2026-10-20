package medicalrecord

import (
	"encoding/json"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestMasterNullableRead_PreservesNullAndExplicitValues(t *testing.T) {
	zero := 0
	none := model.AnesthesiaType("none")
	for _, tc := range []struct {
		name     string
		response any
		want     map[string]string
	}{
		{"consultation null", toConsultationResponse(&model.Consultation{}), map[string]string{"sort_order": "null", "is_non_insurance": "null"}},
		{"consultation zero", toConsultationResponse(&model.Consultation{SortOrder: &zero}), map[string]string{"sort_order": "0"}},
		{"procedure null", toProcedureResponse(&model.Procedure{}), map[string]string{"sort_order": "null", "anesthesia": "null", "is_non_insurance": "null"}},
		{"procedure explicit", toProcedureResponse(&model.Procedure{SortOrder: &zero, Anesthesia: &none}), map[string]string{"sort_order": "0", "anesthesia": "\"none\""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.response)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err = json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			for key, want := range tc.want {
				if string(fields[key]) != want {
					t.Fatalf("%s=%s want %s", key, fields[key], want)
				}
			}
		})
	}
}
