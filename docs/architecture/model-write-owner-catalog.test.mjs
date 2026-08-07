import { test } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const catalogPath = path.join(__dirname, 'model-write-owner-catalog.md');

test('catalog declares required model write owners', () => {
  assert.ok(fs.existsSync(catalogPath), `expected catalog at ${catalogPath}`);
  const body = fs.readFileSync(catalogPath, 'utf8');
  assert.ok(body.length > 0, 'catalog must not be empty');

  const lower = body.toLowerCase();
  for (const marker of [
    'write owner',
    'reservation',
    'staff',
    'medicalrecord',
    'billing',
    'appointments',
    'no bulk',
  ]) {
    assert.ok(lower.includes(marker), `catalog must mention: ${marker}`);
  }

  // Critical ownership pins (type or table / package tokens).
  for (const token of [
    'Reservation',
    'appointments',
    'appointment_write_owner',
    'Staff',
    'shift_entries',
    'MedicalRecord',
    'Billing',
    'Owner',
    'Pet',
    'TrimmingCourse',
    'CreateForTrimming',
    'identitylink',
    'AuditLog',
  ]) {
    assert.ok(body.includes(token), `catalog must cover ownership token: ${token}`);
  }

  // New-type / PR rules section must stay actionable.
  assert.ok(
    lower.includes('command') || lower.includes('dto'),
    'catalog must keep command/DTO separation guidance',
  );
  assert.ok(
    lower.includes('review checklist') || lower.includes('pr description'),
    'catalog must include review/PR guidance',
  );
});
