package apicontract

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const billDiscountExcludedField = "is_bill_discount_excluded"

// TestOpenAPIMasterBillDiscountExcludedContract は EMR-445 A01（品目別の全体割引除外）の API 正本を固定する。
// request は省略可・null 不可の boolean、response は required かつ nullable の保存値、
// 版を追加する3種は response の version と PATCH の任意 version・409 を持つ（runtime/DB は別試験）。
func TestOpenAPIMasterBillDiscountExcludedContract(t *testing.T) {
	spec := loadMasterPatchContractSpec(t)
	schemas := spec.Components.Schemas

	for _, tc := range []struct {
		path, response, create, update string
		addsVersion                    bool
	}{
		{"/masters/consultations", "Consultation", "CreateConsultationRequest", "UpdateConsultationRequest", false},
		{"/masters/examination-types", "ExaminationType", "CreateExaminationTypeRequest", "UpdateExaminationTypeRequest", true},
		{"/masters/procedures", "Procedure", "CreateProcedureRequest", "UpdateProcedureRequest", false},
		{"/masters/vaccines", "Vaccine", "CreateVaccineRequest", "UpdateVaccineRequest", false},
		{"/masters/checkup-types", "CheckupType", "CreateCheckupTypeRequest", "UpdateCheckupTypeRequest", false},
		{"/masters/medicines", "Medicine", "CreateMedicineRequest", "UpdateMedicineRequest", true},
		{"/masters/merchandise-items", "MerchandiseItem", "CreateMerchandiseItemRequest", "UpdateMerchandiseItemRequest", true},
	} {
		t.Run(tc.response, func(t *testing.T) {
			assert.Equal(t, "#/components/schemas/"+tc.create, requestSchemaRef(t, spec, tc.path, "post"))
			assert.Equal(t, "#/components/schemas/"+tc.update, requestSchemaRef(t, spec, tc.path+"/{id}", "patch"))

			response := schemas[tc.response]
			field, ok := response.Properties[billDiscountExcludedField]
			require.True(t, ok, "%s must return the stored setting", tc.response)
			assert.Equal(t, "boolean", field["type"])
			assert.Equal(t, true, field["nullable"], "known-unset NULL is returned raw")
			assert.Contains(t, response.Required, billDiscountExcludedField, "key is always present (no omitempty)")

			for _, request := range []string{tc.create, tc.update} {
				prop, ok := schemas[request].Properties[billDiscountExcludedField]
				require.True(t, ok, "%s must accept the setting", request)
				assert.Equal(t, "boolean", prop["type"])
				assert.NotContains(t, prop, "nullable", "%s rejects null with 400", request)
				assert.NotContains(t, schemas[request].Required, billDiscountExcludedField, "%s keeps the field optional", request)
			}

			if !tc.addsVersion {
				return
			}
			version, ok := response.Properties["version"]
			require.True(t, ok, "%s must return its version", tc.response)
			assert.Equal(t, "integer", version["type"])
			assert.Equal(t, "int64", version["format"])
			assert.Contains(t, response.Required, "version")
			patchVersion, ok := schemas[tc.update].Properties["version"]
			require.True(t, ok, "%s must accept the expected version", tc.update)
			assert.Equal(t, "integer", patchVersion["type"])
			assert.NotContains(t, schemas[tc.update].Required, "version", "legacy PATCH without version stays compatible")
			assert.Contains(t, masterPatchOperationOf(t, spec, tc.path+"/{id}", "patch").Responses, "409")
		})
	}

	for _, update := range []string{"UpdateConsultationRequest", "UpdateProcedureRequest", "UpdateVaccineRequest", "UpdateCheckupTypeRequest"} {
		assert.Contains(t, schemas[update].Required, "version", "existing %s version requirement must not be relaxed", update)
	}
}

// TestOpenAPIMasterSplitRequestsCopySharedContract は A01 で response から分離した request schema が、
// 追加 field 以外は共用 schema の既存契約と同一であることを固定する。
func TestOpenAPIMasterSplitRequestsCopySharedContract(t *testing.T) {
	schemas := loadMasterPatchContractSpec(t).Components.Schemas
	added := []string{billDiscountExcludedField, "version"}

	for _, tc := range []struct{ shared, request string }{
		{"Medicine", "CreateMedicineRequest"},
		{"ExaminationType", "CreateExaminationTypeRequest"},
		{"ExaminationType", "UpdateExaminationTypeRequest"},
		{"MerchandiseItem", "CreateMerchandiseItemRequest"},
		{"MerchandiseItem", "UpdateMerchandiseItemRequest"},
	} {
		t.Run(tc.request, func(t *testing.T) {
			shared := schemas[tc.shared]
			request := schemas[tc.request]
			var writableRequired []string
			for _, name := range shared.Required {
				if shared.Properties[name]["readOnly"] != true {
					writableRequired = append(writableRequired, name)
				}
			}
			assert.Equal(t, writableRequired, request.Required)
			for name, prop := range shared.Properties {
				if slices.Contains(added, name) {
					continue
				}
				if prop["readOnly"] == true {
					assert.NotContains(t, request.Properties, name)
					continue
				}
				assert.Equal(t, prop, request.Properties[name], "%s.%s must equal the shared contract", tc.request, name)
			}
		})
	}
}
