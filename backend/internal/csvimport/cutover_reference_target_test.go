package csvimport

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPreflightCutoverTargetRequiresEstimatePetForeignKeys(t *testing.T) {
	for _, tc := range []struct {
		name      string
		target    validTargetQuerier
		wantError bool
	}{
		{"exact ordered constraints", validTargetQuerier{estimatePetChildColumns: []string{"clinic_id", "pet_id"}, estimatePetParentColumns: []string{"clinic_id", "id"}}, false},
		{"missing composite", validTargetQuerier{missingEstimatePetComposite: true}, true},
		{"reversed child columns", validTargetQuerier{estimatePetChildColumns: []string{"pet_id", "clinic_id"}}, true},
		{"reversed parent columns", validTargetQuerier{estimatePetParentColumns: []string{"id", "clinic_id"}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := PreflightCutoverTarget(context.Background(), tc.target, cutoverManifestForTargetTests(), validCutoverSeeds())
			if (err != nil) != tc.wantError {
				t.Fatalf("PreflightCutoverTarget() error = %v, want error %v", err, tc.wantError)
			}
			if err != nil && (!strings.Contains(err.Error(), "estimates") || !strings.Contains(err.Error(), "pet_id")) {
				t.Fatalf("wrong target constraint rejection: %v", err)
			}
		})
	}
}

func TestCutoverEstimatePetInventoryMatchesCanonicalDDL(t *testing.T) {
	for _, spec := range cutoverRequiredForeignKeys() {
		if spec.childTable == "estimates" && spec.childColumn == "pet_id" {
			t.Fatal("canonical DDL has no single-column estimates.pet_id FK")
		}
	}
	composite := cutoverCompositeForeignKeySpec{"estimates", []string{"clinic_id", "pet_id"}, "pets", []string{"clinic_id", "id"}}
	compositeFound := false
	for _, spec := range cutoverRequiredCompositeForeignKeys() {
		if reflect.DeepEqual(spec, composite) {
			compositeFound = true
		}
	}
	if !compositeFound {
		t.Fatal("estimate pet composite target inventory is incomplete")
	}
	ddl, err := os.ReadFile("../../migrations/001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"ADD COLUMN IF NOT EXISTS pet_id bigint;",
		"ADD CONSTRAINT fk_estimates_pet_clinic",
		"FOREIGN KEY (clinic_id, pet_id)\n  REFERENCES pets (clinic_id, id)",
	} {
		if !strings.Contains(string(ddl), fragment) {
			t.Errorf("canonical DDL is missing pinned estimate pet contract %q", fragment)
		}
	}
	for _, fragment := range []string{"c.convalidated = true", "c.conenforced = true", "array_agg(a.attname ORDER BY cols.ordinality)", "= $3::text[]", "= $4::text[]"} {
		if !strings.Contains(cutoverCompositeForeignKeyQuery, fragment) {
			t.Errorf("ordered/enforced composite query contract is missing %q", fragment)
		}
	}
}
