package csvimport

import "testing"

func TestTableSpecResolvesTargetOnlyReferences(t *testing.T) {
	s, err := tableSpec("pets")
	if err != nil {
		t.Fatal(err)
	}
	row := make([]string, len(s.columns))
	row[0], row[1], row[2], row[5] = "10", "{{CLINIC_ID}}", "20", ""
	values, err := s.values(row, 1, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if values[1] != int64(1) {
		t.Fatalf("clinic placeholder was not resolved: %#v", values[1])
	}
	if values[5] != int64(2) {
		t.Fatalf("species fallback was not resolved: %#v", values[5])
	}
}

func TestTypedValueConvertsBoundaryTypes(t *testing.T) {
	if got, err := typedValue("id", "42"); err != nil || got != int64(42) {
		t.Fatalf("id conversion: %#v %v", got, err)
	}
	if got, err := typedValue("is_dangerous", "t"); err != nil || got != true {
		t.Fatalf("bool conversion: %#v %v", got, err)
	}
}
