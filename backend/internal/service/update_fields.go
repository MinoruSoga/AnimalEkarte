package service

// setNullableUint64Field applies a nullable uint64 FK/parent column to a PATCH
// field map following the buildXxxUpdate convention used across master-data
// services:
//   - clear == true  → write an explicit NULL (clear the association)
//   - clear == false, id != nil → write *id
//   - clear == false, id == nil → leave the column untouched (omit from update)
//
// fields is the existing map[string]any PATCH payload (P10/P13); this helper
// introduces no new dynamic typing beyond that established shape.
func setNullableUint64Field(fields map[string]any, col string, clear bool, id *uint64) {
	if clear {
		fields[col] = nil
	} else if id != nil {
		fields[col] = *id
	}
}
