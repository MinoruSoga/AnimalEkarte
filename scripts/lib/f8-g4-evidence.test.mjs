import assert from "node:assert/strict";
import { test } from "node:test";

import {
  artifactBytes,
  buildFailureApplyReport,
  buildFailureEvidence,
  buildFailureRuntimeReport,
  buildPostFailurePreflightReport,
  buildSyntheticFixtureManifest,
  canonicalEvidenceBytes,
  digestBytes,
  targetDatabaseIdentityDigest,
} from "./f8-g4-evidence.mjs";

const identity = {
  clinicCode: "hachioji",
  clinicOrdinal: 1,
  runId: "hachioji-f8-failure-20260727",
  targetReleaseCommit: "a".repeat(40),
  targetDatabaseIdentitySha256: "b".repeat(64),
  backendArchiveSha256: "f".repeat(64),
  backendTreeId: "3".repeat(40),
  composeProject: "animalekarte-f8-g4-run-1",
  dockerContextName: "orbstack",
  dockerEndpoint: "unix:///tmp/docker.sock",
  dockerDaemonId: "daemon-1",
  dbContainerId: "c".repeat(64),
  dbImageId: `sha256:${"1".repeat(64)}`,
  dbName: "animalekarte_f8_g4_run_1",
  dbVolumeName: "animalekarte-f8-g4-run-1_postgres_data",
  networkName: "animalekarte-f8-g4-run-1_f8-g4-network",
  runnerImageId: `sha256:${"2".repeat(64)}`,
};

function receipt() {
  const counts = {
    schemaVersion: 1,
    tableCount: 21,
    tables: [
      "staffs", "procedures", "merchandise_items", "owners", "pets",
      "medical_records", "inquiries", "clinical_plans", "vital_records",
      "appointments", "appointment_trimming_details", "billings",
      "billing_items", "payments", "payment_splits", "estimates",
      "estimate_items", "exams", "exam_results", "vaccines", "vaccinations",
    ].map((table) => ({ table, rowCount: 0 })),
    totalRowCount: 0,
  };
  const before = digestBytes(canonicalEvidenceBytes(counts));
  const transactionEvidence = {
    schemaVersion: 1,
    fixtureId: "F8_G4_TRANSACTION_ROLLBACK_V1",
    clinicCode: identity.clinicCode,
    clinicOrdinal: identity.clinicOrdinal,
    runId: identity.runId,
    targetReleaseCommit: identity.targetReleaseCommit,
    targetDatabaseIdentitySha256: identity.targetDatabaseIdentitySha256,
    injectionCheckpoint: "G4_TARGET_VERIFIED",
    injectionStage: "TRANSACTION",
    injectionMarker: "SYNTHETIC_FK_VIOLATION_AFTER_COPY",
    copiedRowCount: 1,
    observedSqlState: "23503",
    transactionStarted: true,
    transactionRolledBack: true,
    beforeBandCountsSha256: before,
    afterBandCountsSha256: before,
  };
  return {
    schemaVersion: 1,
    status: "FAILED_DATA_ROLLED_BACK",
    ...identity,
    fixtureManifestSha256: "d".repeat(64),
    failureRuntimeReportSha256: "e".repeat(64),
    executionMode: "SYNTHETIC_FAILURE_REHEARSAL",
    failureCheckpoint: "G4_TARGET_VERIFIED",
    failureStage: "TRANSACTION",
    injectionMarker: "SYNTHETIC_FK_VIOLATION_AFTER_COPY",
    startedAt: "2026-07-27T10:10:00.000Z",
    failureInjectedAt: "2026-07-27T10:20:00.000Z",
    completedAt: "2026-07-27T10:21:00.000Z",
    transactionStarted: true,
    transactionRolledBack: true,
    beforeBandCountsSha256: before,
    afterBandCountsSha256: before,
    bandRowCountBefore: 0,
    bandRowCountAfter: 0,
    transactionEvidenceSha256: digestBytes(canonicalEvidenceBytes(transactionEvidence)),
    beforeBandCounts: counts,
    afterBandCounts: counts,
    transactionEvidence,
    productionEligible: false,
  };
}

test("builds the exact old_db G4 artifact chain", () => {
  const fixture = buildSyntheticFixtureManifest(identity, "2026-07-27T10:01:00.000Z");
  const fixtureSha = digestBytes(artifactBytes(fixture));
  const runtime = buildFailureRuntimeReport(identity, fixtureSha, "2026-07-27T10:05:00.000Z");
  const runtimeSha = digestBytes(artifactBytes(runtime));
  const sourceReceipt = {
    ...receipt(),
    fixtureManifestSha256: fixtureSha,
    failureRuntimeReportSha256: runtimeSha,
  };
  const expectedBindings = {
    fixtureManifestSha256: fixtureSha,
    failureRuntimeReportSha256: runtimeSha,
  };
  const apply = buildFailureApplyReport(identity, sourceReceipt, expectedBindings);
  const applySha = digestBytes(artifactBytes(apply));
  const preflight = buildPostFailurePreflightReport(
    identity,
    sourceReceipt,
    applySha,
    "2026-07-27T10:25:00.000Z",
    expectedBindings,
  );
  const preflightSha = digestBytes(artifactBytes(preflight));
  const failure = buildFailureEvidence({
    identity,
    receipt: sourceReceipt,
    fixtureSha,
    runtimeSha,
    applySha,
    preflightSha,
    startedAt: fixture.generatedAt,
    completedAt: preflight.generatedAt,
  });

  assert.deepEqual(Object.keys(fixture), [
    "schemaVersion", "status", "clinicCode", "clinicOrdinal", "runId",
    "targetReleaseCommit", "generatedAt", "fixtureId", "targetClassification",
    "targetDatabaseIdentitySha256", "clinicBandBase", "clinicBandEndExclusive",
    "tableCount", "injectionCheckpoint", "injectionStage", "injectionMarker",
    "containsProductionData", "productionEligible",
  ]);
  assert.deepEqual(Object.keys(runtime), [
    "schemaVersion", "status", "clinicCode", "clinicOrdinal", "runId",
    "targetReleaseCommit", "generatedAt", "fixtureManifestSha256",
    "targetDatabaseIdentitySha256", "targetClassification", "clinicBandBase",
    "clinicBandEndExclusive", "attestationMethod", "targetHeadCommit",
    "targetWorktreeClean", "databaseDisposition", "networkIsolation",
    "productionEligible",
  ]);
  assert.deepEqual(Object.keys(apply), [
    "schemaVersion", "status", "clinicCode", "clinicOrdinal", "runId",
    "targetReleaseCommit", "fixtureManifestSha256", "failureRuntimeReportSha256",
    "targetDatabaseIdentitySha256", "targetClassification", "clinicBandBase",
    "clinicBandEndExclusive", "executionMode", "startedAt", "completedAt",
    "failureCheckpoint", "failureStage", "injectionMarker",
    "transactionStarted", "transactionRolledBack", "beforeBandCountsSha256",
    "afterBandCountsSha256", "bandRowCountBefore", "bandRowCountAfter",
    "transactionEvidenceSha256", "productionEligible",
  ]);
  assert.deepEqual(Object.keys(preflight), [
    "schemaVersion", "status", "clinicCode", "clinicOrdinal", "runId",
    "targetReleaseCommit", "generatedAt", "fixtureManifestSha256",
    "failureRuntimeReportSha256", "failureApplyReportSha256",
    "targetDatabaseIdentitySha256", "targetClassification", "clinicBandBase",
    "clinicBandEndExclusive", "beforeBandCountsSha256", "afterBandCountsSha256",
    "bandRowCount", "emptyBandPreflight", "seedPreflight", "attestationMethod",
    "targetHeadCommit", "targetWorktreeClean", "databaseDisposition",
    "networkIsolation", "productionEligible",
  ]);
  assert.equal(runtime.fixtureManifestSha256, fixtureSha);
  assert.equal(runtime.attestationMethod, "DOCKER_INSPECT_AND_GIT_HEAD");
  assert.equal(apply.failureRuntimeReportSha256, runtimeSha);
  assert.equal(preflight.failureApplyReportSha256, applySha);
  assert.equal(preflight.attestationMethod, "DOCKER_INSPECT_AND_GIT_HEAD");
  assert.equal(failure.artifacts.postFailurePreflightReportSha256, preflightSha);
  assert.equal(failure.checkpointOutcomes[3].status, "FAIL");
  assert.equal(failure.checkpointOutcomes[4].status, "NOT_RUN");
});

test("rejects receipts that do not independently prove rollback", () => {
  const invalid = receipt();
  invalid.bandRowCountAfter = 1;
  assert.throws(
    () => buildFailureApplyReport(identity, invalid, {
      fixtureManifestSha256: invalid.fixtureManifestSha256,
      failureRuntimeReportSha256: invalid.failureRuntimeReportSha256,
    }),
    /zero-row rollback/,
  );

  const digestMismatch = receipt();
  digestMismatch.transactionEvidenceSha256 = "0".repeat(64);
  assert.throws(
    () => buildFailureApplyReport(identity, digestMismatch, {
      fixtureManifestSha256: digestMismatch.fixtureManifestSha256,
      failureRuntimeReportSha256: digestMismatch.failureRuntimeReportSha256,
    }),
    /transaction evidence digest/,
  );
});

test("rejects a receipt bound to stale local fixture or runtime bytes", () => {
  const sourceReceipt = receipt();
  assert.throws(
    () => buildFailureApplyReport(identity, sourceReceipt, {
      fixtureManifestSha256: "0".repeat(64),
      failureRuntimeReportSha256: sourceReceipt.failureRuntimeReportSha256,
    }),
    /fixture digest binding mismatch/,
  );
  assert.throws(
    () => buildFailureApplyReport(identity, sourceReceipt, {
      fixtureManifestSha256: sourceReceipt.fixtureManifestSha256,
      failureRuntimeReportSha256: "0".repeat(64),
    }),
    /runtime digest binding mismatch/,
  );
});

test("database identity digest is canonical and excludes secrets", () => {
  const digest = targetDatabaseIdentityDigest(identity);
  assert.match(digest, /^[a-f0-9]{64}$/);
  assert.equal(
    digest,
    targetDatabaseIdentityDigest({ ...identity, ignoredPassword: "do-not-hash" }),
  );
});
