import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const catalogPath = path.join(__dirname, 'fe-feature-be-domain-map.md');

test('catalog declares required FE feature BE domain mappings', () => {
  assert.ok(fs.existsSync(catalogPath), `expected catalog at ${catalogPath}`);
  const body = fs.readFileSync(catalogPath, 'utf8');
  assert.ok(body.length > 0, 'catalog must not be empty');

  const lower = body.toLowerCase();
  for (const marker of [
    'features/',
    'rbac',
    'shared-liff',
    'line-reservation',
    'feature indexing',
    'no bulk',
  ]) {
    assert.ok(lower.includes(marker), `catalog must mention: ${marker}`);
  }

  // Core feature ↔ domain pins.
  for (const token of [
    'owners',
    'owner',
    'reservations',
    'reservation',
    'medical-records',
    'medicalrecord',
    'accounting',
    'billing',
    'trimming',
    'lstep',
    'identity-links',
    'identitylink',
    'reception',
    'model.Resource',
  ]) {
    assert.ok(body.includes(token), `catalog must cover mapping token: ${token}`);
  }

  assert.ok(
    body.includes('≥ 2') || lower.includes('2+') || body.includes('2 以上') || body.includes('≥2'),
    'catalog must state shared promotion needs multiple consumers',
  );
});
