package medicalrecord

// EMR-521 response/wire tests (NOT_RUN).

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func mustNumeric(t *testing.T, s string) model.ExactNumeric {
	var n model.ExactNumeric
	require.NoError(t, n.Scan(s))
	return n
}

func TestExactNumericCanonical(t *testing.T) {
	assert.Equal(t, "0.25", exactNumericCanonical(mustNumeric(t, "0.25")))
	assert.Equal(t, "1", exactNumericCanonical(mustNumeric(t, "1")))
	assert.Equal(t, "110", exactNumericCanonical(mustNumeric(t, "110")))
	assert.Equal(t, "0", exactNumericCanonical(model.ExactNumeric{Numeric: pgtype.Numeric{Valid: false}}))
	// Stored value with internal zero-padded exponent still renders canonical.
	n := model.ExactNumeric{Numeric: pgtype.Numeric{Int: big.NewInt(1), Exp: 0, Valid: true}}
	assert.Equal(t, "1", exactNumericCanonical(n))
}

func TestPackageItemResponseMarshal(t *testing.T) {
	out, err := json.Marshal(treatmentPackageItemResponse{Kind: model.TreatmentPackageItemKindConsultation, TypedID: 101})
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind":"consultation","consultation_id":"101"}`, string(out))

	out, err = json.Marshal(treatmentPackageItemResponse{
		Kind:   model.TreatmentPackageItemKindManualOther,
		Manual: json.RawMessage(`{"name":"m","reason":"r","unit_label":null,"unit_price":"110","tax_type":"included","tax_rate":"0.1","is_non_insurance":true}`),
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind":"manual_other","manual":{"name":"m","reason":"r","unit_label":null,"unit_price":"110","tax_type":"included","tax_rate":"0.1","is_non_insurance":true}}`, string(out))
}

func TestTreatmentPackageErrorCodes(t *testing.T) {
	assert.Equal(t, "invalid-input", treatmentPackageErrorCode(400))
	assert.Equal(t, "not-authorized", treatmentPackageErrorCode(403))
	assert.Equal(t, "scope-not-found", treatmentPackageErrorCode(404))
	assert.Equal(t, "version-conflict", treatmentPackageErrorCode(409))
	assert.Equal(t, "validation", treatmentPackageErrorCode(422))
	assert.Equal(t, "internal-error", treatmentPackageErrorCode(500))
	assert.Equal(t, "temporarily-unavailable", treatmentPackageErrorCode(503))
}
