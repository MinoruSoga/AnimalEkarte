package medicalrecord

// EMR-521 request DTO tests (NOT_RUN — fixture-aligned, execution deferred).
// Covers the 8-branch mix, canonical wire enforcement, and PATCH null semantics.

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTreatmentPackageRequest_EightBranches(t *testing.T) {
	var req createTreatmentPackageRequest
	err := json.Unmarshal([]byte(`{
		"name":"セット","description":"","is_active":true,
		"components":[
			{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"},
			{"component_key":"a0000000-0000-4000-8000-000000000012","item":{"kind":"procedure","procedure_id":"102"},"quantity":"1"},
			{"component_key":"a0000000-0000-4000-8000-000000000013","item":{"kind":"medicine","medicine_id":"103"},"quantity":"0.25"},
			{"component_key":"a0000000-0000-4000-8000-000000000014","item":{"kind":"examination_type","examination_type_id":"104"},"quantity":"1"},
			{"component_key":"a0000000-0000-4000-8000-000000000015","item":{"kind":"vaccine","vaccine_id":"105"},"quantity":"1"},
			{"component_key":"a0000000-0000-4000-8000-000000000016","item":{"kind":"checkup_type","checkup_type_id":"106"},"quantity":"1"},
			{"component_key":"a0000000-0000-4000-8000-000000000017","item":{"kind":"merchandise_item","merchandise_item_id":"107"},"quantity":"1"},
			{"component_key":"a0000000-0000-4000-8000-000000000018","item":{"kind":"manual_other","manual":{"name":"手入力","reason":"明示","unit_label":null,"unit_price":"110","tax_type":"included","tax_rate":"0.1","is_non_insurance":true}},"quantity":"1"}
		]}`), &req)
	require.NoError(t, err)
	assert.Len(t, req.Components, 8)
	assert.Equal(t, "103", formatUint64String(req.Components[2].TypedID))
	assert.Equal(t, "0.25", req.Components[2].QuantityText)
	assert.Equal(t, "manual_other", req.Components[7].Kind)
	assert.EqualValues(t, 110, req.Components[7].Manual.UnitPrice)
	assert.Nil(t, req.Components[7].Manual.UnitLabel)
}

func TestCreateTreatmentPackageRequest_Rejects(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"mixed FK in one branch", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101","medicine_id":"103"},"quantity":"1"}]}`},
		{"JSON number id", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":101},"quantity":"1"}]}`},
		{"noncanonical quantity trailing zero", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"0.250"}]}`},
		{"zero quantity", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"0"}]}`},
		{"unknown top-level key", `{"name":"x","description":"","is_active":true,"clinic_id":"7","components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"}]}`},
		{"missing is_active", `{"name":"x","description":"","components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"}]}`},
		{"null name", `{"name":null,"description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"}]}`},
		{"noncanonical uuid uppercase", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"A0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"}]}`},
		{"nil uuid", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"00000000-0000-0000-0000-000000000000","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"}]}`},
		{"manual_other with FK key", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"manual_other","consultation_id":"101","manual":{"name":"m","reason":"r","unit_label":null,"unit_price":"0","tax_type":"exempt","tax_rate":"0.1","is_non_insurance":false}},"quantity":"1"}]}`},
		{"manual missing reason", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"manual_other","manual":{"name":"m","unit_label":null,"unit_price":"0","tax_type":"exempt","tax_rate":"0.1","is_non_insurance":false}},"quantity":"1"}]}`},
		{"duplicate component_key", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"consultation","consultation_id":"101"},"quantity":"1"},{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"procedure","procedure_id":"102"},"quantity":"1"}]}`},
		{"unknown kind", `{"name":"x","description":"","is_active":true,"components":[{"component_key":"a0000000-0000-4000-8000-000000000011","item":{"kind":"unknown","consultation_id":"101"},"quantity":"1"}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req createTreatmentPackageRequest
			assert.Error(t, json.Unmarshal([]byte(tc.body), &req))
		})
	}
}

func TestUpdateTreatmentPackageRequest_ExplicitNullRejected(t *testing.T) {
	var req updateTreatmentPackageRequest
	assert.Error(t, json.Unmarshal([]byte(`{"version":"3","name":null}`), &req))
	assert.Error(t, json.Unmarshal([]byte(`{"version":"3","components":null}`), &req))
	assert.Error(t, json.Unmarshal([]byte(`{"version":3}`), &req)) // JSON number rejected.
	assert.Error(t, json.Unmarshal([]byte(`{"name":"x"}`), &req))  // version required.
}

func TestUpdateTreatmentPackageRequest_OmittedMeansUnchanged(t *testing.T) {
	var req updateTreatmentPackageRequest
	require.NoError(t, json.Unmarshal([]byte(`{"version":"3","description":"明示空も許容"}`), &req))
	assert.EqualValues(t, 3, req.Version)
	assert.Nil(t, req.Name)
	assert.Nil(t, req.IsActive)
	assert.Nil(t, req.Components)
	require.NotNil(t, req.Description)
	assert.Equal(t, "明示空も許容", *req.Description)
}

func TestReorderTreatmentPackagesRequest(t *testing.T) {
	var req reorderTreatmentPackagesRequest
	require.NoError(t, json.Unmarshal([]byte(`{"order_version":"7","ids":["201","202"]}`), &req))
	assert.EqualValues(t, 7, req.OrderVersion)
	assert.Equal(t, []uint64{201, 202}, req.IDs)
	assert.Error(t, json.Unmarshal([]byte(`{"order_version":"7","ids":["201","201"]}`), &req)) // duplicate
	assert.Error(t, json.Unmarshal([]byte(`{"order_version":"7","ids":[201]}`), &req))          // JSON number
	assert.Error(t, json.Unmarshal([]byte(`{"ids":["201"]}`), &req))                          // order_version required
	assert.Error(t, json.Unmarshal([]byte(`{"order_version":"0","ids":["201"]}`), &req))        // noncanonical
}
