import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const catalogPath = path.join(__dirname, 'cross-domain-orchestration-catalog.md');

test('catalog declares automation and batch orchestration contracts', () => {
  assert.ok(
    fs.existsSync(catalogPath),
    `expected catalog at ${catalogPath}`,
  );

  const body = fs.readFileSync(catalogPath, 'utf8');
  assert.ok(body.length > 0, 'catalog must not be empty');

  // Table column coverage (headers may use slight wording variants).
  const headerNeedles = [
    'initiator',
    'owner operation',
    'transaction boundary',
    'fail-closed',
    'best-effort',
    'failure recovery',
    'audit',
    'test anchor',
  ];
  for (const needle of headerNeedles) {
    assert.ok(
      body.toLowerCase().includes(needle),
      `catalog table/header must mention: ${needle}`,
    );
  }

  // Prior A3 request-path markers must remain (existing A3 TAP compatibility).
  const a3PathMarkers = [
    'AutoCreateFromReservation',
    'DeleteDraftFromReservation',
    'BackfillForMedicalRecord',
    'PrepareForMedicalRecordFinalization',
    'CompleteForAccounting',
    'DischargeWithBilling',
    'CreateForTrimming',
  ];
  for (const marker of a3PathMarkers) {
    assert.ok(
      body.includes(marker),
      `catalog must preserve A3 path marker: ${marker}`,
    );
  }

  // Intentional best-effort separate-tx contracts must not be silently upgraded.
  assert.ok(
    body.includes('PATH-RES-MR-AUTOCREATE'),
    'catalog must retain PATH-RES-MR-AUTOCREATE Path ID',
  );
  assert.ok(
    body.includes('PATH-RES-MR-CANCEL-CLEANUP'),
    'catalog must retain PATH-RES-MR-CANCEL-CLEANUP Path ID',
  );
  assert.ok(
    body.includes('best-effort') || body.includes('best effort'),
    'catalog must label intentional best-effort contracts',
  );
  assert.ok(
    body.includes('separate-tx') || body.includes('separate tx') || body.includes('separate transaction'),
    'catalog must document separate-tx boundaries for best-effort paths',
  );
  assert.ok(
    body.includes('same-tx') || body.includes('same tx') || body.includes('same transaction'),
    'catalog must document same-tx boundaries for fail-closed paths',
  );

  // Automation / batch orchestration markers (ARCH-A3-4 residual).
  const automationMarkers = [
    'JobNoShow',
    'JobDelivery',
    'JobDormant',
    'batch_scheduler',
    'RunNoShowCheckAllClinicsAt',
    'MarkNoShow',
    'LogNoShowTransitionTx',
    'RunDeliveryTriggerBatchAllClinicsAt',
    'RunDormantDetectionAllClinicsAt',
    'RunHealthPreventionTagSyncAllClinics',
    'SyncCPMStageTag',
    'BatchRunResult',
  ];
  for (const marker of automationMarkers) {
    assert.ok(
      body.includes(marker),
      `catalog must cover automation/batch marker: ${marker}`,
    );
  }

  // New-path rules section remains and applies to automation/batch paths.
  const lower = body.toLowerCase();
  assert.ok(
    lower.includes('new-path') || lower.includes('new path'),
    'catalog must include a new-path rules section',
  );
  assert.ok(
    lower.includes('typed intent') || lower.includes('owner typed') || lower.includes('typed intents'),
    'new-path rules must require owner typed intents',
  );
  assert.ok(
    lower.includes('ambient') || lower.includes('explicit orchestration'),
    'new-path rules must require ambient-tx participation or explicit orchestration',
  );
  assert.ok(
    lower.includes('fail-closed') || lower.includes('fail closed'),
    'new-path rules must default to fail-closed',
  );
  assert.ok(
    lower.includes('no silent partial') || lower.includes('silent partial success') || lower.includes('partial success'),
    'new-path rules must forbid silent partial success',
  );
  assert.ok(
    lower.includes('automation') || lower.includes('batch'),
    'new-path rules or catalog must acknowledge automation/batch paths',
  );
});
