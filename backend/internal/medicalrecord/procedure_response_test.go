package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-454: 品目別保険設定は nullable のまま返し、未保存・旧欠落（NULL）を
// false（保険対象）へ読み替えない。
func TestToProcedureResponse_IsNonInsurance(t *testing.T) {
	t.Run("explicit value is preserved", func(t *testing.T) {
		isNonInsurance := true
		p := &model.Procedure{ID: 1, ClinicID: 2, Name: "処置", IsNonInsurance: &isNonInsurance}

		resp := toProcedureResponse(p)

		require.NotNil(t, resp.IsNonInsurance)
		assert.True(t, *resp.IsNonInsurance)
	})

	t.Run("NULL stays nil", func(t *testing.T) {
		p := &model.Procedure{ID: 1, ClinicID: 2, Name: "処置"}

		resp := toProcedureResponse(p)

		assert.Nil(t, resp.IsNonInsurance)
	})
}
