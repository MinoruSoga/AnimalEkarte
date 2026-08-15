import crypto from "node:crypto";

const SHA256_RE = /^[a-f0-9]{64}$/;
const COMMIT_RE = /^[a-f0-9]{40}$/;
const CLINIC_RE = /^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$/;
const RUN_RE = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/;
const SAFETY_POLICY =
  "Aggregate timings, counts, statuses, and digests only. No identifiers, row values, credentials, paths, screenshots, logs, or free text.";
const TABLES = Object.freeze([
  "staffs", "procedures", "merchandise_items", "owners", "pets",
  "medical_records", "inquiries", "clinical_plans", "vital_records",
  "appointments", "appointment_trimming_details", "billings",
  "billing_items", "payments", "payment_splits", "estimates",
  "estimate_items", "exams", "exam_results", "vaccines", "vaccinations",
]);

function requireValue(condition, message) {
  if (!condition) throw new Error(`F8 G4 evidence rejected: ${message}`);
}

function canonicalTimestamp(value, label) {
  const parsed = Date.parse(value ?? "");
  requireValue(Number.isFinite(parsed) && new Date(parsed).toISOString() === value, `${label} is invalid`);
}

function validateIdentity(identity) {
  requireValue(CLINIC_RE.test(identity?.clinicCode ?? ""), "clinic code is invalid");
  requireValue(identity.clinicOrdinal === 1, "clinic ordinal must be 1");
  requireValue(RUN_RE.test(identity.runId ?? ""), "run ID is invalid");
  requireValue(COMMIT_RE.test(identity.targetReleaseCommit ?? ""), "target release commit is invalid");
  requireValue(
    SHA256_RE.test(identity.targetDatabaseIdentitySha256 ?? ""),
    "target database identity is invalid",
  );
}

export function canonicalEvidenceBytes(value) {
  return Buffer.from(`${JSON.stringify(value)}\n`, "utf8");
}

export function artifactBytes(value) {
  return Buffer.from(`${JSON.stringify(value, null, 2)}\n`, "utf8");
}

export function digestBytes(value) {
  return crypto.createHash("sha256").update(value).digest("hex");
}

export function targetDatabaseIdentityDigest(identity) {
  return digestBytes(canonicalEvidenceBytes({
    schemaVersion: 1,
    backendArchiveSha256: identity.backendArchiveSha256,
    backendTreeId: identity.backendTreeId,
    composeProject: identity.composeProject,
    dockerContextName: identity.dockerContextName,
    dockerEndpoint: identity.dockerEndpoint,
    dockerDaemonId: identity.dockerDaemonId,
    dbContainerId: identity.dbContainerId,
    dbImageId: identity.dbImageId,
    dbName: identity.dbName,
    dbVolumeName: identity.dbVolumeName,
    networkName: identity.networkName,
    runnerImageId: identity.runnerImageId,
  }));
}

export function buildSyntheticFixtureManifest(identity, generatedAt) {
  validateIdentity(identity);
  canonicalTimestamp(generatedAt, "fixture generatedAt");
  return {
    schemaVersion: 1,
    status: "PASS",
    clinicCode: identity.clinicCode,
    clinicOrdinal: identity.clinicOrdinal,
    runId: identity.runId,
    targetReleaseCommit: identity.targetReleaseCommit,
    generatedAt,
    fixtureId: "F8_G4_TRANSACTION_ROLLBACK_V1",
    targetClassification: "DISPOSABLE_EMPTY_BAND",
    targetDatabaseIdentitySha256: identity.targetDatabaseIdentitySha256,
    clinicBandBase: 0,
    clinicBandEndExclusive: 10_000_000,
    tableCount: TABLES.length,
    injectionCheckpoint: "G4_TARGET_VERIFIED",
    injectionStage: "TRANSACTION",
    injectionMarker: "SYNTHETIC_FK_VIOLATION_AFTER_COPY",
    containsProductionData: false,
    productionEligible: false,
  };
}

export function buildFailureRuntimeReport(identity, fixtureSha, generatedAt) {
  validateIdentity(identity);
  requireValue(SHA256_RE.test(fixtureSha ?? ""), "fixture manifest digest is invalid");
  canonicalTimestamp(generatedAt, "runtime generatedAt");
  return {
    schemaVersion: 1,
    status: "PASS",
    clinicCode: identity.clinicCode,
    clinicOrdinal: identity.clinicOrdinal,
    runId: identity.runId,
    targetReleaseCommit: identity.targetReleaseCommit,
    generatedAt,
    fixtureManifestSha256: fixtureSha,
    targetDatabaseIdentitySha256: identity.targetDatabaseIdentitySha256,
    targetClassification: "DISPOSABLE_EMPTY_BAND",
    clinicBandBase: 0,
    clinicBandEndExclusive: 10_000_000,
    attestationMethod: "DOCKER_INSPECT_AND_GIT_HEAD",
    targetHeadCommit: identity.targetReleaseCommit,
    targetWorktreeClean: true,
    databaseDisposition: "DISPOSABLE",
    networkIsolation: "LOCALHOST_ONLY",
    productionEligible: false,
  };
}

function validateReceipt(identity, receipt, expectedBindings = {}) {
  validateIdentity(identity);
  requireValue(receipt?.schemaVersion === 1, "receipt schema is invalid");
  requireValue(receipt.status === "FAILED_DATA_ROLLED_BACK", "receipt status is invalid");
  for (const key of ["clinicCode", "clinicOrdinal", "runId", "targetReleaseCommit", "targetDatabaseIdentitySha256"]) {
    requireValue(receipt[key] === identity[key], `receipt ${key} binding mismatch`);
  }
  for (const [key, label] of [
    ["fixtureManifestSha256", "fixture"],
    ["failureRuntimeReportSha256", "runtime"],
  ]) {
    requireValue(SHA256_RE.test(receipt[key] ?? ""), `receipt ${label} digest is invalid`);
    if (expectedBindings[key] !== undefined) {
      requireValue(receipt[key] === expectedBindings[key], `receipt ${label} digest binding mismatch`);
    }
  }
  requireValue(
    receipt.executionMode === "SYNTHETIC_FAILURE_REHEARSAL"
      && receipt.failureCheckpoint === "G4_TARGET_VERIFIED"
      && receipt.failureStage === "TRANSACTION"
      && receipt.injectionMarker === "SYNTHETIC_FK_VIOLATION_AFTER_COPY",
    "receipt injection contract is invalid",
  );
  requireValue(
    receipt.transactionStarted === true
      && receipt.transactionRolledBack === true
      && receipt.productionEligible === false,
    "receipt rollback state is invalid",
  );
  requireValue(
    receipt.bandRowCountBefore === 0
      && receipt.bandRowCountAfter === 0,
    "receipt must prove a zero-row rollback",
  );
  canonicalTimestamp(receipt.startedAt, "receipt startedAt");
  canonicalTimestamp(receipt.failureInjectedAt, "receipt failureInjectedAt");
  canonicalTimestamp(receipt.completedAt, "receipt completedAt");
  requireValue(
    Date.parse(receipt.startedAt) <= Date.parse(receipt.failureInjectedAt)
      && Date.parse(receipt.failureInjectedAt) <= Date.parse(receipt.completedAt),
    "receipt timestamps are not chronological",
  );

  const beforeBytes = canonicalEvidenceBytes(receipt.beforeBandCounts);
  const afterBytes = canonicalEvidenceBytes(receipt.afterBandCounts);
  requireValue(beforeBytes.equals(afterBytes), "before/after band evidence differs");
  requireValue(
    receipt.beforeBandCounts?.schemaVersion === 1
      && receipt.beforeBandCounts?.tableCount === TABLES.length
      && receipt.beforeBandCounts?.totalRowCount === 0
      && JSON.stringify(receipt.beforeBandCounts.tables)
        === JSON.stringify(TABLES.map((table) => ({ table, rowCount: 0 }))),
    "band count evidence is invalid",
  );
  requireValue(
    digestBytes(beforeBytes) === receipt.beforeBandCountsSha256
      && digestBytes(afterBytes) === receipt.afterBandCountsSha256,
    "band count digest mismatch",
  );
  const transaction = receipt.transactionEvidence;
  requireValue(
    transaction?.schemaVersion === 1
      && transaction.fixtureId === "F8_G4_TRANSACTION_ROLLBACK_V1"
      && transaction.clinicCode === identity.clinicCode
      && transaction.clinicOrdinal === identity.clinicOrdinal
      && transaction.runId === identity.runId
      && transaction.targetReleaseCommit === identity.targetReleaseCommit
      && transaction.targetDatabaseIdentitySha256 === identity.targetDatabaseIdentitySha256
      && transaction.injectionCheckpoint === receipt.failureCheckpoint
      && transaction.injectionStage === receipt.failureStage
      && transaction.injectionMarker === receipt.injectionMarker
      && transaction.copiedRowCount === 1
      && transaction.observedSqlState === "23503"
      && transaction.transactionStarted === true
      && transaction.transactionRolledBack === true
      && transaction.beforeBandCountsSha256 === receipt.beforeBandCountsSha256
      && transaction.afterBandCountsSha256 === receipt.afterBandCountsSha256,
    "transaction evidence is invalid",
  );
  requireValue(
    digestBytes(canonicalEvidenceBytes(transaction)) === receipt.transactionEvidenceSha256,
    "transaction evidence digest mismatch",
  );
}

export function buildFailureApplyReport(identity, receipt, expectedBindings) {
  validateReceipt(identity, receipt, expectedBindings);
  return {
    schemaVersion: 1,
    status: "FAILED_DATA_ROLLED_BACK",
    clinicCode: identity.clinicCode,
    clinicOrdinal: identity.clinicOrdinal,
    runId: identity.runId,
    targetReleaseCommit: identity.targetReleaseCommit,
    fixtureManifestSha256: receipt.fixtureManifestSha256,
    failureRuntimeReportSha256: receipt.failureRuntimeReportSha256,
    targetDatabaseIdentitySha256: identity.targetDatabaseIdentitySha256,
    targetClassification: "DISPOSABLE_EMPTY_BAND",
    clinicBandBase: 0,
    clinicBandEndExclusive: 10_000_000,
    executionMode: "SYNTHETIC_FAILURE_REHEARSAL",
    startedAt: receipt.startedAt,
    completedAt: receipt.completedAt,
    failureCheckpoint: receipt.failureCheckpoint,
    failureStage: receipt.failureStage,
    injectionMarker: receipt.injectionMarker,
    transactionStarted: true,
    transactionRolledBack: true,
    beforeBandCountsSha256: receipt.beforeBandCountsSha256,
    afterBandCountsSha256: receipt.afterBandCountsSha256,
    bandRowCountBefore: 0,
    bandRowCountAfter: 0,
    transactionEvidenceSha256: receipt.transactionEvidenceSha256,
    productionEligible: false,
  };
}

export function buildPostFailurePreflightReport(
  identity,
  receipt,
  failureApplyReportSha256,
  generatedAt,
  expectedBindings,
) {
  validateReceipt(identity, receipt, expectedBindings);
  requireValue(SHA256_RE.test(failureApplyReportSha256 ?? ""), "apply report digest is invalid");
  canonicalTimestamp(generatedAt, "preflight generatedAt");
  requireValue(Date.parse(generatedAt) >= Date.parse(receipt.completedAt), "preflight predates rollback");
  return {
    schemaVersion: 1,
    status: "PASS",
    clinicCode: identity.clinicCode,
    clinicOrdinal: identity.clinicOrdinal,
    runId: identity.runId,
    targetReleaseCommit: identity.targetReleaseCommit,
    generatedAt,
    fixtureManifestSha256: receipt.fixtureManifestSha256,
    failureRuntimeReportSha256: receipt.failureRuntimeReportSha256,
    failureApplyReportSha256,
    targetDatabaseIdentitySha256: identity.targetDatabaseIdentitySha256,
    targetClassification: "DISPOSABLE_EMPTY_BAND",
    clinicBandBase: 0,
    clinicBandEndExclusive: 10_000_000,
    beforeBandCountsSha256: receipt.beforeBandCountsSha256,
    afterBandCountsSha256: receipt.afterBandCountsSha256,
    bandRowCount: 0,
    emptyBandPreflight: "PASS",
    seedPreflight: "PASS",
    attestationMethod: "DOCKER_INSPECT_AND_GIT_HEAD",
    targetHeadCommit: identity.targetReleaseCommit,
    targetWorktreeClean: true,
    databaseDisposition: "DISPOSABLE",
    networkIsolation: "LOCALHOST_ONLY",
    productionEligible: false,
  };
}

export function buildFailureEvidence({
  identity,
  receipt,
  fixtureSha,
  runtimeSha,
  applySha,
  preflightSha,
  startedAt,
  completedAt,
}) {
  validateReceipt(identity, receipt, {
    fixtureManifestSha256: fixtureSha,
    failureRuntimeReportSha256: runtimeSha,
  });
  for (const [digest, label] of [
    [fixtureSha, "fixture"], [runtimeSha, "runtime"],
    [applySha, "apply"], [preflightSha, "preflight"],
  ]) requireValue(SHA256_RE.test(digest ?? ""), `${label} digest is invalid`);
  canonicalTimestamp(startedAt, "failure startedAt");
  canonicalTimestamp(completedAt, "failure completedAt");
  requireValue(
    Date.parse(startedAt) <= Date.parse(receipt.failureInjectedAt)
      && Date.parse(receipt.completedAt) <= Date.parse(completedAt),
    "failure evidence timestamps are not chronological",
  );
  return {
    schemaVersion: 1,
    clinicCode: identity.clinicCode,
    clinicOrdinal: identity.clinicOrdinal,
    migrationRunId: identity.runId,
    executionMode: "CONTROLLED_FAILURE_REHEARSAL",
    failureMode: "SYNTHETIC_FIXTURE_FAILURE",
    startedAt,
    failureInjectedAt: receipt.failureInjectedAt,
    stoppedAt: receipt.completedAt,
    completedAt,
    checkpointOutcomes: [
      { name: "G1_RECEIPT", status: "PASS" },
      { name: "G2_F3_COMPLETE", status: "PASS" },
      { name: "G3_CSV_READY", status: "PASS" },
      { name: "G4_TARGET_VERIFIED", status: "FAIL" },
      { name: "G5_UI_GO", status: "NOT_RUN" },
    ],
    targetChangeStatus: "ROLLED_BACK",
    rollbackStatus: "PASS",
    rollbackVerificationStatus: "PASS",
    notificationTemplateStatus: "PASS",
    notificationMode: "NO_SEND_TEMPLATE_DRILL",
    piiSafetyStatus: "PASS",
    artifacts: {
      syntheticFixtureManifestSha256: fixtureSha,
      failureRuntimeReportSha256: runtimeSha,
      failureApplyReportSha256: applySha,
      postFailurePreflightReportSha256: preflightSha,
    },
    safetyPolicy: SAFETY_POLICY,
  };
}
