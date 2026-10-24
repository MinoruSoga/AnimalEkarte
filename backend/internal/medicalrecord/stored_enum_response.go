package medicalrecord

// stored_enum_response.go — NULL 許容の文字列系 enum を応答 DTO へ写す共有 helper（EMR-549 C3）。

// nullableTextFromStored は NULL 許容の文字列系 enum（status・麻酔種別等）を応答用の *string へ写す。
// nil（未記録）は nil（JSON null）のまま返し、"" や pending 等の既定値で補完しない。
func nullableTextFromStored[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	text := string(*value)
	return &text
}
