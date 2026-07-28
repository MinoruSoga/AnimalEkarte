#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

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
} from "./lib/f8-g4-evidence.mjs";
import {
  rejectDockerEnvironmentOverrides,
  sanitizedGitEnvironment,
  validateLocalDockerAttestation,
  validateLocalDockerEndpoint,
} from "./lib/f8-g4-host-safety.mjs";

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const sensitiveRoot = path.join(repo, "sensitive-local");
const outputRoot = path.join(sensitiveRoot, "f8-g4-rehearsal");
const composeFile = "docker-compose.f8-g4-rehearsal.yml";
const PROJECT_RE = /^animalekarte-f8-g4-[a-z0-9][a-z0-9-]{0,35}$/;
const RUN_RE = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/;
const COMMIT_RE = /^[a-f0-9]{40}$/;
const CLINIC_RE = /^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$/;
const DB_RE = /^animalekarte_f8_g4_[a-z0-9_]+$/;
const SHA_RE = /^[a-f0-9]{64}$/;
const IMAGE_ID_RE = /^sha256:[a-f0-9]{64}$/;
const ALLOWED_ENV_KEYS = new Set(["DB_NAME", "DB_PASSWORD", "DB_USER"]);
let pinnedDockerEndpoint;

function requireValue(condition, message) {
  if (!condition) throw new Error(`F8 G4 runner rejected: ${message}`);
}

function run(command, args, options = {}) {
  const { timeout = 120_000, ...spawnOptions } = options;
  const result = spawnSync(command, args, {
    cwd: repo,
    encoding: "utf8",
    shell: false,
    maxBuffer: 16 * 1024 * 1024,
    timeout,
    killSignal: "SIGTERM",
    ...spawnOptions,
  });
  if (result.error?.code === "ETIMEDOUT") {
    throw new Error(`F8 G4 runner command timed out: ${command} ${args[0] ?? ""}`);
  }
  if (result.status !== 0) {
    throw new Error(`F8 G4 runner command failed: ${command} ${args[0] ?? ""}`);
  }
  return result.stdout.trim();
}

function runDocker(args, options = {}) {
  requireValue(pinnedDockerEndpoint, "Docker endpoint is not attested");
  return run("docker", ["--host", pinnedDockerEndpoint, ...args], options);
}

function assertCleanHead(revision) {
  const env = sanitizedGitEnvironment();
  const topLevel = fs.realpathSync(run("git", ["-C", repo, "rev-parse", "--show-toplevel"], { env }));
  requireValue(topLevel === fs.realpathSync(repo), "Git top-level does not match the AnimalEkarte repository");
  requireValue(run("git", ["-C", repo, "rev-parse", "HEAD"], { env }) === revision, "target HEAD mismatch");
  requireValue(
    run("git", ["-C", repo, "status", "--porcelain", "--untracked-files=all"], { env }) === "",
    "target worktree must be clean",
  );
}

function attestLocalDocker() {
  rejectDockerEnvironmentOverrides();
  const contextName = run("docker", ["context", "show"]);
  const [context] = JSON.parse(run("docker", ["context", "inspect", contextName]));
  const endpoint = context?.Endpoints?.docker?.Host;
  validateLocalDockerEndpoint({ contextName, endpoint });
  pinnedDockerEndpoint = endpoint;
  const daemonId = JSON.parse(runDocker(["info", "--format", "{{json .ID}}"]));
  return validateLocalDockerAttestation({ contextName, endpoint, daemonId });
}

function reattestPinnedDocker(expected) {
  const daemonId = JSON.parse(runDocker(["info", "--format", "{{json .ID}}"]));
  const current = validateLocalDockerAttestation({
    contextName: expected.contextName,
    endpoint: expected.endpoint,
    daemonId,
  });
  requireValue(
    JSON.stringify(current) === JSON.stringify(expected),
    "Docker daemon identity changed during the rehearsal",
  );
}

function requireRealOwnerOnlyDirectory(directory, label) {
  const stat = fs.lstatSync(directory);
  requireValue(stat.isDirectory() && !stat.isSymbolicLink(), `${label} must be a real directory`);
  requireValue((stat.mode & 0o077) === 0, `${label} must be owner-only`);
}

function validateEnvFile(input) {
  requireValue(input, "F8_G4_ENV_FILE is required");
  requireRealOwnerOnlyDirectory(sensitiveRoot, "sensitive-local");
  const file = path.resolve(repo, input);
  requireValue(file.startsWith(`${sensitiveRoot}${path.sep}`), "environment file must be under sensitive-local");
  let parent = sensitiveRoot;
  for (const component of path.relative(sensitiveRoot, path.dirname(file)).split(path.sep)) {
    if (component === "") continue;
    parent = path.join(parent, component);
    requireRealOwnerOnlyDirectory(parent, "environment parent");
  }
  requireValue(fs.realpathSync(file) === file, "environment file must not be a symlink");
  const descriptor = fs.openSync(file, fs.constants.O_RDONLY | fs.constants.O_NOFOLLOW);
  let contents;
  try {
    const stat = fs.fstatSync(descriptor);
    requireValue(
      stat.isFile() && (stat.mode & 0o077) === 0 && stat.uid === process.getuid(),
      "environment file must be an owner-only regular file owned by the operator",
    );
    contents = fs.readFileSync(descriptor, "utf8");
  } finally {
    fs.closeSync(descriptor);
  }
  const values = new Map();
  for (const rawLine of contents.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (line === "" || line.startsWith("#")) continue;
    const match = /^([A-Z][A-Z0-9_]*)=(.*)$/.exec(line);
    requireValue(match && ALLOWED_ENV_KEYS.has(match[1]), "environment file contains a forbidden key");
    requireValue(!values.has(match[1]), "environment file contains a duplicate key");
    values.set(match[1], match[2]);
  }
  for (const key of ALLOWED_ENV_KEYS) requireValue(values.get(key), `environment file is missing ${key}`);
  requireValue(DB_RE.test(values.get("DB_NAME")), "DB_NAME is not disposable F8 G4 identity");
  return { contents, values };
}

function positiveInteger(name) {
  const value = Number(process.env[name]);
  requireValue(Number.isSafeInteger(value) && value > 0, `${name} must be a positive integer`);
  return value;
}

function readIdentityInput() {
  const identity = {
    composeProject: process.env.F8_G4_COMPOSE_PROJECT,
    clinicCode: process.env.F8_G4_CLINIC_CODE,
    clinicOrdinal: Number(process.env.F8_G4_CLINIC_ORDINAL),
    runId: process.env.F8_G4_RUN_ID,
    targetReleaseCommit: process.env.F8_G4_TARGET_RELEASE_COMMIT,
  };
  requireValue(PROJECT_RE.test(identity.composeProject ?? ""), "Compose project is invalid");
  requireValue(CLINIC_RE.test(identity.clinicCode ?? ""), "clinic code is invalid");
  requireValue(identity.clinicOrdinal === 1, "clinic ordinal must be 1");
  requireValue(RUN_RE.test(identity.runId ?? ""), "run ID is invalid");
  requireValue(COMMIT_RE.test(identity.targetReleaseCommit ?? ""), "target release commit is invalid");
  return {
    identity,
    seeds: {
      clinicId: positiveInteger("TARGET_CLINIC_ID"),
      animalSpeciesId: positiveInteger("FALLBACK_ANIMAL_SPECIES_ID"),
      examTypeId: positiveInteger("FALLBACK_EXAM_TYPE_ID"),
      trimmingReservationTypeId: positiveInteger("TRIMMING_RESERVATION_TYPE_ID"),
      cashPaymentMethodId: positiveInteger("PAYMENT_METHOD_CASH_ID"),
      creditCardPaymentMethodId: positiveInteger("PAYMENT_METHOD_CREDIT_CARD_ID"),
    },
  };
}

function inspectOptional(kind, name) {
  requireValue(pinnedDockerEndpoint, "Docker endpoint is not attested");
  const result = spawnSync("docker", ["--host", pinnedDockerEndpoint, kind, "inspect", name], {
    cwd: repo,
    encoding: "utf8",
    shell: false,
    timeout: 120_000,
    killSignal: "SIGTERM",
  });
  if (result.error?.code === "ETIMEDOUT") {
    throw new Error(`F8 G4 runner Docker ${kind} inspection timed out`);
  }
  if (result.status !== 0 && /no such|not found/i.test(result.stderr)) return undefined;
  requireValue(result.status === 0, `cannot inspect Docker ${kind}`);
  return JSON.parse(result.stdout)[0];
}

function assertFreshProject(project) {
  const containers = runDocker([
    "ps", "-aq", "--filter", `label=com.docker.compose.project=${project}`,
  ]);
  requireValue(containers === "", "Compose project containers already exist");
  requireValue(inspectOptional("network", `${project}_f8-g4-network`) === undefined, "project network already exists");
  for (const suffix of ["postgres_data"]) {
    requireValue(inspectOptional("volume", `${project}_${suffix}`) === undefined, "project volume already exists");
  }
}

function validateRunnerImage(identity, backendTreeId) {
  const imageName = `${identity.composeProject}_runner:${identity.targetReleaseCommit}`;
  const image = inspectOptional("image", imageName);
  requireValue(image !== undefined && IMAGE_ID_RE.test(image.Id ?? ""), "runner image is missing");
  const labels = image.Config?.Labels ?? {};
  requireValue(
    labels["org.opencontainers.image.revision"] === identity.targetReleaseCommit
      && labels["com.animalekarte.f8-g4.backend-tree"] === backendTreeId,
    "runner image release labels mismatch",
  );
  return { runnerImageId: image.Id, runnerImageName: imageName };
}

function validateRuntimeResources({
  composeArgs,
  composeEnv,
  identity,
  dbName,
  dockerAttestation,
  runnerImage,
  immutableSource,
}) {
  const containerId = runDocker([...composeArgs, "ps", "-q", "db"], { env: composeEnv });
  requireValue(/^[a-f0-9]{64}$/.test(containerId), "database container is missing");
  const [container] = JSON.parse(runDocker(["inspect", containerId]));
  const labels = container.Config?.Labels ?? {};
  const networkName = `${identity.composeProject}_f8-g4-network`;
  const volumeName = `${identity.composeProject}_postgres_data`;
  requireValue(container.State?.Running === true && container.State?.Health?.Status === "healthy", "database is not healthy");
  requireValue(IMAGE_ID_RE.test(container.Image ?? ""), "database image identity is invalid");
  requireValue(
    labels["com.docker.compose.project"] === identity.composeProject
      && labels["com.animalekarte.f8-g4.disposable"] === "true"
      && labels["com.animalekarte.f8-g4.run-id"] === identity.runId,
    "database disposable labels mismatch",
  );
  requireValue(container.HostConfig?.NetworkMode === networkName, "database network mismatch");
  requireValue(
    JSON.stringify(Object.keys(container.NetworkSettings?.Networks ?? {})) === JSON.stringify([networkName]),
    "database must use only the dedicated network",
  );
  const bindings = container.HostConfig?.PortBindings ?? {};
  requireValue(
    JSON.stringify(Object.keys(bindings)) === JSON.stringify(["5432/tcp"])
      && bindings["5432/tcp"].every(({ HostIp }) => HostIp === "127.0.0.1" || HostIp === "::1"),
    "database port must bind only to localhost",
  );
  requireValue(
    container.Mounts?.length === 1
      && container.Mounts[0].Type === "volume"
      && container.Mounts[0].Name === volumeName
      && container.Mounts[0].Destination === "/var/lib/postgresql"
      && container.Mounts[0].RW === true,
    "database volume mismatch",
  );
  const network = inspectOptional("network", networkName);
  const volume = inspectOptional("volume", volumeName);
  requireValue(
    network?.Internal === true
      && network?.Attachable === false
      && network.Labels?.["com.animalekarte.f8-g4.disposable"] === "true",
    "dedicated internal network mismatch",
  );
  requireValue(
    volume?.Scope === "local"
      && volume.Labels?.["com.animalekarte.f8-g4.disposable"] === "true",
    "dedicated local database volume mismatch",
  );
  const databaseEnv = new Map(
    (container.Config?.Env ?? []).map((entry) => {
      const separator = entry.indexOf("=");
      return [entry.slice(0, separator), entry.slice(separator + 1)];
    }),
  );
  requireValue(databaseEnv.get("POSTGRES_DB") === dbName, "database name mismatch");
  return {
    ...identity,
    ...immutableSource,
    dockerContextName: dockerAttestation.contextName,
    dockerEndpoint: dockerAttestation.endpoint,
    dockerDaemonId: dockerAttestation.daemonId,
    dbContainerId: containerId,
    dbImageId: container.Image,
    dbName,
    dbVolumeName: volumeName,
    networkName,
    runnerImageId: runnerImage.runnerImageId,
  };
}

function writeExclusive(file, contents) {
  fs.writeFileSync(file, contents, { flag: "wx", mode: 0o600 });
  fs.chmodSync(file, 0o600);
}

function createImmutableBackendContext(revision, workingDirectory) {
  const gitEnv = sanitizedGitEnvironment();
  const backendTreeId = run("git", ["-C", repo, "rev-parse", `${revision}:backend`], { env: gitEnv });
  requireValue(COMMIT_RE.test(backendTreeId), "backend tree identity is invalid");
  const archive = path.join(workingDirectory, "backend.tar");
  const descriptor = fs.openSync(archive, "wx", 0o600);
  try {
    const result = spawnSync(
      "git",
      ["-C", repo, "archive", "--format=tar", `${revision}:backend`],
      {
        cwd: repo,
        env: gitEnv,
        shell: false,
        stdio: ["ignore", descriptor, "pipe"],
        timeout: 120_000,
        killSignal: "SIGTERM",
      },
    );
    if (result.error?.code === "ETIMEDOUT") {
      throw new Error("F8 G4 runner command timed out: git archive");
    }
    requireValue(result.status === 0, "cannot archive immutable backend source");
  } finally {
    fs.closeSync(descriptor);
  }
  const backendArchiveSha256 = digestBytes(fs.readFileSync(archive));
  const buildContext = path.join(workingDirectory, "backend");
  fs.mkdirSync(buildContext, { mode: 0o700 });
  run("tar", ["-xf", archive, "-C", buildContext]);
  fs.rmSync(archive);
  return { backendArchiveSha256, backendTreeId, buildContext };
}

const { identity: inputIdentity, seeds } = readIdentityInput();
const { contents: envContents, values: envValues } = validateEnvFile(process.env.F8_G4_ENV_FILE);
const dockerAttestation = attestLocalDocker();
assertCleanHead(inputIdentity.targetReleaseCommit);
assertFreshProject(inputIdentity.composeProject);
try {
  fs.mkdirSync(outputRoot, { recursive: true, mode: 0o700 });
  fs.chmodSync(outputRoot, 0o700);
} catch (error) {
  if (error?.code !== "EEXIST") throw error;
}
requireRealOwnerOnlyDirectory(outputRoot, "F8 G4 output root");
const finalDirectory = path.join(outputRoot, inputIdentity.runId);
requireValue(!fs.existsSync(finalDirectory), "run output already exists");
const workingDirectory = fs.mkdtempSync(path.join(outputRoot, ".work-"));
fs.chmodSync(workingDirectory, 0o700);

try {
  const stagingDirectory = path.join(workingDirectory, "artifacts");
  fs.mkdirSync(stagingDirectory, { mode: 0o700 });
  const envFile = path.join(workingDirectory, "rehearsal.env");
  writeExclusive(envFile, envContents);
  const immutableSource = createImmutableBackendContext(
    inputIdentity.targetReleaseCommit,
    workingDirectory,
  );
  const composeArgs = [
    "compose", "--env-file", envFile, "-p", inputIdentity.composeProject,
    "-f", composeFile,
  ];
  const composeEnv = {
    ...process.env,
    F8_G4_ENV_FILE: envFile,
    F8_G4_COMPOSE_PROJECT: inputIdentity.composeProject,
    F8_G4_RUN_ID: inputIdentity.runId,
    F8_G4_TARGET_RELEASE_COMMIT: inputIdentity.targetReleaseCommit,
    F8_G4_BUILD_CONTEXT: immutableSource.buildContext,
    F8_G4_BACKEND_TREE_ID: immutableSource.backendTreeId,
    F8_G4_RUNNER_IMAGE:
      `${inputIdentity.composeProject}_runner:${inputIdentity.targetReleaseCommit}`,
  };
  runDocker([...composeArgs, "config", "--quiet"], { env: composeEnv });
  runDocker([...composeArgs, "build", "--pull", "--no-cache", "runner"], {
    env: composeEnv,
    timeout: 900_000,
  });
  const runnerImage = validateRunnerImage(inputIdentity, immutableSource.backendTreeId);
  composeEnv.F8_G4_RUNNER_IMAGE = runnerImage.runnerImageId;
  runDocker([...composeArgs, "up", "-d", "--pull", "always", "--wait", "--wait-timeout", "600", "db"], {
    env: composeEnv,
    timeout: 660_000,
  });
  const migrationOutput = runDocker(
    [...composeArgs, "run", "--rm", "--no-deps", "--pull", "never", "migrate"],
    {
      env: composeEnv,
      timeout: 600_000,
    },
  );
  requireValue(
    migrationOutput.includes("--- PASS: TestRunSQLMigrationsAgainstDisposablePostgres"),
    "DDL migration test did not execute",
  );
  const seedOutput = runDocker(
    [...composeArgs, "run", "--rm", "--no-deps", "--pull", "never", "seed"],
    {
      env: composeEnv,
      timeout: 600_000,
    },
  );
  requireValue(
    seedOutput.includes("--- PASS: TestSeedCutoverRehearsalAgainstDisposablePostgres"),
    "cutover seed test did not execute",
  );
  assertCleanHead(inputIdentity.targetReleaseCommit);
  reattestPinnedDocker(dockerAttestation);
  const inspectedIdentity = validateRuntimeResources({
    composeArgs,
    composeEnv,
    identity: inputIdentity,
    dbName: envValues.get("DB_NAME"),
    dockerAttestation,
    runnerImage,
    immutableSource: {
      backendArchiveSha256: immutableSource.backendArchiveSha256,
      backendTreeId: immutableSource.backendTreeId,
    },
  });
  const identity = {
    ...inspectedIdentity,
    targetDatabaseIdentitySha256: targetDatabaseIdentityDigest(inspectedIdentity),
  };

  const fixture = buildSyntheticFixtureManifest(identity, new Date().toISOString());
  const fixtureBytes = artifactBytes(fixture);
  const fixtureSha = digestBytes(fixtureBytes);
  const runtime = buildFailureRuntimeReport(identity, fixtureSha, new Date().toISOString());
  const runtimeBytes = artifactBytes(runtime);
  const runtimeSha = digestBytes(runtimeBytes);

  const runnerArgs = [
    ...composeArgs, "run", "--rm", "--no-deps", "--pull", "never", "runner", "run",
    "--clinic-code", identity.clinicCode,
    "--clinic-ordinal", String(identity.clinicOrdinal),
    "--run-id", identity.runId,
    "--target-release-commit", identity.targetReleaseCommit,
    "--target-database-identity-sha256", identity.targetDatabaseIdentitySha256,
    "--fixture-manifest-sha256", fixtureSha,
    "--failure-runtime-report-sha256", runtimeSha,
    "--clinic-id", String(seeds.clinicId),
    "--fallback-animal-species-id", String(seeds.animalSpeciesId),
    "--fallback-exam-type-id", String(seeds.examTypeId),
    "--trimming-reservation-type-id", String(seeds.trimmingReservationTypeId),
    "--cash-payment-method-id", String(seeds.cashPaymentMethodId),
    "--credit-card-payment-method-id", String(seeds.creditCardPaymentMethodId),
    "--confirm-target-database", identity.dbName,
    "--confirm-disposable-rehearsal", "F8_G4_DISPOSABLE_ONLY",
  ];
  const receipt = JSON.parse(runDocker(runnerArgs, {
    env: composeEnv,
    timeout: 360_000,
  }));
  reattestPinnedDocker(dockerAttestation);
  const expectedBindings = {
    fixtureManifestSha256: fixtureSha,
    failureRuntimeReportSha256: runtimeSha,
  };
  const apply = buildFailureApplyReport(identity, receipt, expectedBindings);
  const applyBytes = artifactBytes(apply);
  const applySha = digestBytes(applyBytes);
  assertCleanHead(identity.targetReleaseCommit);
  const preflight = buildPostFailurePreflightReport(
    identity,
    receipt,
    applySha,
    new Date().toISOString(),
    expectedBindings,
  );
  const preflightBytes = artifactBytes(preflight);
  const preflightSha = digestBytes(preflightBytes);
  const failure = buildFailureEvidence({
    identity,
    receipt,
    fixtureSha,
    runtimeSha,
    applySha,
    preflightSha,
    startedAt: fixture.generatedAt,
    completedAt: preflight.generatedAt,
  });

  const outputs = new Map([
    ["synthetic-fixture-manifest.json", fixtureBytes],
    ["failure-runtime-report.json", runtimeBytes],
    ["failure-apply-report.json", applyBytes],
    ["failure-preflight-report.json", preflightBytes],
    ["failure-evidence.json", artifactBytes(failure)],
    ["failure-before-band-counts.json", canonicalEvidenceBytes(receipt.beforeBandCounts)],
    ["failure-after-band-counts.json", canonicalEvidenceBytes(receipt.afterBandCounts)],
    ["failure-transaction-evidence.json", canonicalEvidenceBytes(receipt.transactionEvidence)],
  ]);
  for (const [file, contents] of outputs) {
    writeExclusive(path.join(stagingDirectory, file), contents);
  }
  assertCleanHead(identity.targetReleaseCommit);
  reattestPinnedDocker(dockerAttestation);
  fs.renameSync(stagingDirectory, finalDirectory);
  requireValue(SHA_RE.test(receipt.transactionEvidenceSha256), "transaction evidence digest is invalid");
  process.stdout.write([
    `F8 G4 REHEARSAL PRODUCER: PASS ${finalDirectory}`,
    `F8_FAILURE_TRANSACTION_EVIDENCE_EXPECTED_SHA256=${receipt.transactionEvidenceSha256}`,
    `F8_FAILURE_BEFORE_COUNTS_EXPECTED_SHA256=${receipt.beforeBandCountsSha256}`,
    `F8_FAILURE_AFTER_COUNTS_EXPECTED_SHA256=${receipt.afterBandCountsSha256}`,
    "",
  ].join("\n"));
} finally {
  fs.rmSync(workingDirectory, { recursive: true, force: true });
}
