package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-196②: revision は画面に表示した未請求集約の決定的フィンガープリント。
// 追加・削除・変更・順序入替の検出性と決定性を固定する。
func TestComputeUnbilledRevision(t *testing.T) {
	treatmentID := uint64(41)
	mrID := uint64(9)
	vaccinationID := uint64(41)

	baseItem := func() model.BillingItem {
		return model.BillingItem{
			ID:              treatmentID,
			Name:            "診察",
			Category:        model.ItemCategoryExamination,
			UnitPrice:       1000,
			Quantity:        1,
			TaxType:         model.TaxTypeExcluded,
			TaxRate:         0.1,
			Source:          model.ItemSourceMedicalRecord,
			TreatmentID:     &treatmentID,
			MedicalRecordID: &mrID,
		}
	}
	vaccItem := model.BillingItem{
		ID:            vaccinationID,
		Name:          "混合ワクチン",
		Category:      model.ItemCategoryVaccine,
		UnitPrice:     5000,
		Quantity:      1,
		TaxType:       model.TaxTypeExcluded,
		TaxRate:       0.1,
		Source:        model.ItemSourceMedicalRecord,
		VaccinationID: &vaccinationID,
	}

	t.Run("同一集約は常に同一 revision（決定的）", func(t *testing.T) {
		items := []model.BillingItem{baseItem(), vaccItem}
		r1 := computeUnbilledRevision(items, nil)
		r2 := computeUnbilledRevision(items, nil)
		require.Equal(t, r1, r2)
		assert.Contains(t, r1, "u1:", "algorithm 接頭辞で方式版を識別する")
	})

	t.Run("集約内の順序差は同一 revision（順序非依存）", func(t *testing.T) {
		r1 := computeUnbilledRevision([]model.BillingItem{baseItem(), vaccItem}, nil)
		r2 := computeUnbilledRevision([]model.BillingItem{vaccItem, baseItem()}, nil)
		assert.Equal(t, r1, r2)
	})

	t.Run("明細の追加を検出する（EMR-196② の本体ケース）", func(t *testing.T) {
		before := computeUnbilledRevision([]model.BillingItem{baseItem()}, nil)
		after := computeUnbilledRevision([]model.BillingItem{baseItem(), vaccItem}, nil)
		assert.NotEqual(t, before, after)
	})

	t.Run("明細の値変更を検出する（価格・数量・名称）", func(t *testing.T) {
		before := computeUnbilledRevision([]model.BillingItem{baseItem()}, nil)
		changed := baseItem()
		changed.UnitPrice = 2000
		assert.NotEqual(t, before, computeUnbilledRevision([]model.BillingItem{changed}, nil))
		changed = baseItem()
		changed.Quantity = 2
		assert.NotEqual(t, before, computeUnbilledRevision([]model.BillingItem{changed}, nil))
		changed = baseItem()
		changed.Name = "再診察"
		assert.NotEqual(t, before, computeUnbilledRevision([]model.BillingItem{changed}, nil))
	})

	t.Run("同じ数値 ID でも provenance 種別で区別する（treatment vs vaccination）", func(t *testing.T) {
		asTreatment := computeUnbilledRevision([]model.BillingItem{baseItem()}, nil)
		asVaccination := computeUnbilledRevision([]model.BillingItem{vaccItem}, nil)
		assert.NotEqual(t, asTreatment, asVaccination)
	})

	t.Run("税・保険・リンク先の変更も検出する", func(t *testing.T) {
		before := computeUnbilledRevision([]model.BillingItem{baseItem()}, nil)
		changed := baseItem()
		changed.TaxType = model.TaxTypeIncluded
		assert.NotEqual(t, before, computeUnbilledRevision([]model.BillingItem{changed}, nil))
		changed = baseItem()
		changed.IsInsuranceApplicable = true
		assert.NotEqual(t, before, computeUnbilledRevision([]model.BillingItem{changed}, nil))
		changed = baseItem()
		changed.MedicalRecordID = ptrU64(10)
		assert.NotEqual(t, before, computeUnbilledRevision([]model.BillingItem{changed}, nil))
	})

	t.Run("warnings も revision に含める（blocking 出現は確定可否に直結）", func(t *testing.T) {
		items := []model.BillingItem{baseItem()}
		without := computeUnbilledRevision(items, nil)
		with := computeUnbilledRevision(items, []UnbilledWarning{{
			Source: UnbilledWarningSourceVaccination,
			Code:   UnbilledWarningCodeVaccinationMasterUnbillable,
			Count:  1, Blocking: true,
		}})
		assert.NotEqual(t, without, with)
	})

	t.Run("空集約でも安定した空集合版を返す", func(t *testing.T) {
		r1 := computeUnbilledRevision(nil, nil)
		r2 := computeUnbilledRevision([]model.BillingItem{}, []UnbilledWarning{})
		assert.Equal(t, r1, r2)
		assert.Contains(t, r1, "u1:")
		// 明細が1件でもあれば空集合版とは一致しない（未指定トークン相当のすり抜けなし）。
		assert.NotEqual(t, r1, computeUnbilledRevision([]model.BillingItem{baseItem()}, nil))
	})
}
