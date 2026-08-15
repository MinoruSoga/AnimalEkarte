package medicalrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func sampleCheckupPackageManifestJSON(t *testing.T, namespace, version string) []byte {
	t.Helper()
	typeName := "Dental Basic " + namespace
	raw, err := json.Marshal(map[string]any{
		"namespace":             namespace,
		"version":               version,
		"clinical_approval_ref": "clinical-ref-opaque-001",
		"types": []map[string]any{
			{
				"key": "dental_basic", "name": typeName, "description": "d",
				"interval": "1y", "target_age": "all", "sort_order": 1, "is_active": true,
			},
		},
		"fields": []map[string]any{
			{
				"key": "dental_score", "type_key": "dental_basic", "name": "Score",
				"field_type": "number", "unit": "pt", "min_value": "0.0000", "max_value": "10.0000",
				"options": []string{}, "is_provisional": false, "sort_order": 1,
			},
		},
	})
	require.NoError(t, err)
	return raw
}

func TestCheckupPackageImport_CanonicalizeAndRejectUnknownField(t *testing.T) {
	ok := sampleCheckupPackageManifestJSON(t, "ns.demo", "1.0.0")
	canonical, err := ParseAndCanonicalizeCheckupPackage(ok)
	require.NoError(t, err)
	require.NotNil(t, canonical)
	assert.NotEmpty(t, canonical.Digest)
	assert.Equal(t, "dental_basic", canonical.Manifest.Types[0].Key)

	bad := []byte(`{"namespace":"n","version":"1","clinical_approval_ref":"r","types":[],"fields":[],"evil":true}`)
	_, err = ParseAndCanonicalizeCheckupPackage(bad)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestCheckupPackageImport_OperatorReceiptAllowlistExcludesInternalFields(t *testing.T) {
	receipt := CheckupPackageImportOperatorReceipt{
		ReceiptID: "rcp_abc", Result: "applied", TypesCreated: 1, FieldsCreated: 1,
	}
	b, err := json.Marshal(receipt)
	require.NoError(t, err)
	var asMap map[string]any
	require.NoError(t, json.Unmarshal(b, &asMap))
	for _, forbidden := range []string{
		"actor_id", "clinic_id", "content_digest", "digest",
		"resource_mapping", "before", "after", "namespace", "version",
	} {
		_, present := asMap[forbidden]
		assert.Falsef(t, present, "operator receipt must not include %s", forbidden)
	}
}

func TestCheckupPackageImport_DryRunApplyReplayAndCrossClinic(t *testing.T) {
	db := setupExaminationTestDB(t)
	// Ensure import columns/table exist for this test schema.
	require.NoError(t, db.Exec(`
ALTER TABLE checkup_types ADD COLUMN IF NOT EXISTS import_namespace text;
ALTER TABLE checkup_types ADD COLUMN IF NOT EXISTS import_key text;
ALTER TABLE checkup_type_fields ADD COLUMN IF NOT EXISTS import_namespace text;
ALTER TABLE checkup_type_fields ADD COLUMN IF NOT EXISTS import_key text;
CREATE TABLE IF NOT EXISTS checkup_package_import_receipts (
  id bigserial PRIMARY KEY,
  clinic_id bigint NOT NULL,
  namespace text NOT NULL,
  version text NOT NULL,
  content_digest text NOT NULL,
  status text NOT NULL,
  actor_id bigint NOT NULL,
  types_created integer NOT NULL DEFAULT 0,
  fields_created integer NOT NULL DEFAULT 0,
  resource_mapping jsonb NOT NULL DEFAULT '{}'::jsonb,
  clinical_approval_ref text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (clinic_id, namespace, version)
);
`).Error)

	ctx := context.Background()
	const clinicA = uint64(1)
	const clinicB = uint64(2)
	actorA := makeExaminationActor(t, db, clinicA, "import actor a")
	actorB := makeExaminationActor(t, db, clinicB, "import actor b")

	svc := NewCheckupPackageImportService(db, persistence.NewTransactor(db), &mockAuditTxLogger{})
	// Unique per run: test DB schema is shared across package test invocations.
	ns := fmt.Sprintf("pkg.demo.%s.%d", t.Name(), time.Now().UnixNano())
	raw := sampleCheckupPackageManifestJSON(t, ns, "1.0.0")

	// Dry-run: zero domain write for this package namespace
	preview, err := svc.Preview(ctx, clinicA, actorA, raw)
	require.NoError(t, err)
	assert.Equal(t, "dry_run_ok", preview.Result)
	var typeCount int64
	require.NoError(t, db.Model(&model.CheckupType{}).
		Where("clinic_id = ? AND import_namespace = ?", clinicA, ns).
		Count(&typeCount).Error)
	assert.Zero(t, typeCount)

	// Apply
	applied, err := svc.Apply(ctx, clinicA, actorA, raw)
	require.NoError(t, err)
	assert.Equal(t, "applied", applied.Result)
	assert.Equal(t, 1, applied.TypesCreated)
	assert.Equal(t, 1, applied.FieldsCreated)
	require.NoError(t, db.Model(&model.CheckupType{}).
		Where("clinic_id = ? AND import_namespace = ?", clinicA, ns).
		Count(&typeCount).Error)
	assert.Equal(t, int64(1), typeCount)

	// Replay same content → noop
	noop, err := svc.Apply(ctx, clinicA, actorA, raw)
	require.NoError(t, err)
	assert.Equal(t, "noop", noop.Result)

	// Different content same version → conflict
	conflictRaw := sampleCheckupPackageManifestJSON(t, ns, "1.0.0")
	// change name to alter digest while keeping namespace/version
	var m map[string]any
	require.NoError(t, json.Unmarshal(conflictRaw, &m))
	types := m["types"].([]any)
	types[0].(map[string]any)["name"] = "Dental Changed"
	conflictRaw, err = json.Marshal(m)
	require.NoError(t, err)
	_, err = svc.Apply(ctx, clinicA, actorA, conflictRaw)
	assert.True(t, apperrors.IsConflict(err))

	// Cross-clinic actor cannot apply to clinic A
	_, err = svc.Apply(ctx, clinicA, actorB, sampleCheckupPackageManifestJSON(t, ns, "2.0.0"))
	assert.True(t, apperrors.IsNotFound(err), "foreign actor must non-leak as NotFound")

	// Clinic B can apply same stable keys independently
	appliedB, err := svc.Apply(ctx, clinicB, actorB, sampleCheckupPackageManifestJSON(t, ns, "1.0.0"))
	require.NoError(t, err)
	assert.Equal(t, "applied", appliedB.Result)
	var typeCountB int64
	require.NoError(t, db.Model(&model.CheckupType{}).
		Where("clinic_id = ? AND import_namespace = ?", clinicB, ns).
		Count(&typeCountB).Error)
	assert.Equal(t, int64(1), typeCountB)
}

func TestCheckupPackageImport_PermissionAndRollback(t *testing.T) {
	db := setupExaminationTestDB(t)
	require.NoError(t, db.Exec(`
ALTER TABLE checkup_types ADD COLUMN IF NOT EXISTS import_namespace text;
ALTER TABLE checkup_types ADD COLUMN IF NOT EXISTS import_key text;
ALTER TABLE checkup_type_fields ADD COLUMN IF NOT EXISTS import_namespace text;
ALTER TABLE checkup_type_fields ADD COLUMN IF NOT EXISTS import_key text;
CREATE TABLE IF NOT EXISTS checkup_package_import_receipts (
  id bigserial PRIMARY KEY,
  clinic_id bigint NOT NULL,
  namespace text NOT NULL,
  version text NOT NULL,
  content_digest text NOT NULL,
  status text NOT NULL,
  actor_id bigint NOT NULL,
  types_created integer NOT NULL DEFAULT 0,
  fields_created integer NOT NULL DEFAULT 0,
  resource_mapping jsonb NOT NULL DEFAULT '{}'::jsonb,
  clinical_approval_ref text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (clinic_id, namespace, version)
);
`).Error)
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "rollback actor")

	failingAudit := &mockAuditTxLogger{logEntryTxFn: func(context.Context, *AuditEntry) error {
		return apperrors.WrapInternalServerError("audit failed")
	}}
	svc := NewCheckupPackageImportService(db, persistence.NewTransactor(db), failingAudit)
	ns := fmt.Sprintf("pkg.rollback.%d", time.Now().UnixNano())
	_, err := svc.Apply(ctx, clinicID, actorID, sampleCheckupPackageManifestJSON(t, ns, "1.0.0"))
	require.Error(t, err)

	var typeCount int64
	require.NoError(t, db.Model(&model.CheckupType{}).
		Where("clinic_id = ? AND import_namespace = ?", clinicID, ns).
		Count(&typeCount).Error)
	assert.Zero(t, typeCount, "audit failure must rollback domain writes")

	var receiptCount int64
	require.NoError(t, db.Model(&model.CheckupPackageImportReceipt{}).
		Where("clinic_id = ? AND namespace = ?", clinicID, ns).
		Count(&receiptCount).Error)
	assert.Zero(t, receiptCount)
}
