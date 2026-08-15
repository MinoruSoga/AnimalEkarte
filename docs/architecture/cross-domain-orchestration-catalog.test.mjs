import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const catalogPath = path.join(__dirname, 'cross-domain-orchestration-catalog.md');

test('catalog declares required cross-domain orchestration contracts', () => {
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

  // Required cross-domain orchestration path markers.
  const pathMarkers = [
    'AutoCreateFromReservation',
    'DeleteDraftFromReservation',
    'BackfillForMedicalRecord',
    'PrepareForMedicalRecordFinalization',
    'CompleteForAccounting',
    'DischargeWithBilling',
    'CreateForTrimming',
  ];
  for (const marker of pathMarkers) {
    assert.ok(
      body.includes(marker),
      `catalog must cover path marker: ${marker}`,
    );
  }

  // Intentional best-effort contracts must not be silently upgraded to same-tx.
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

  // New-path rules section.
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
});
