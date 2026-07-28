package medicalrecord

import "github.com/animal-ekarte/backend/internal/sharedkernel"

// BE9-2D ⑥: 移動した master service 群（cage/procedure/consultation/medicine）が使う enum 検証と
// 税率定数の呼び出し面互換 delegate（実装正本は sharedkernel）。

const DefaultTaxRate = sharedkernel.DefaultTaxRate

func validateAnesthesiaType(anesthesia string) error {
	return sharedkernel.ValidateAnesthesiaType(anesthesia)
}
func validateCageType(cageType string) error { return sharedkernel.ValidateCageType(cageType) }
func validateCageSize(cageSize string) error { return sharedkernel.ValidateCageSize(cageSize) }
func validateTaxType(taxType string) error   { return sharedkernel.ValidateTaxType(taxType) }
