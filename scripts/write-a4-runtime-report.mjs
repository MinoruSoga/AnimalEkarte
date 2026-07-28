#!/usr/bin/env node

import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

import { validateA4EnvFile } from "./lib/a4-env-file.mjs";
import { validateA4ComposeConfig } from "./lib/a4-rehearsal-compose.mjs";
import { buildA4RuntimeReport } from "./lib/a4-runtime-report.mjs";

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const project = process.env.A4_COMPOSE_PROJECT;
const runId = process.env.A4_RUN_ID;
const revision = process.env.A4_TARGET_RELEASE_COMMIT;
const applyPath = path.resolve(repo, process.env.A4_APPLY_REPORT ?? "");
const sensitiveRoot = path.join(repo, "sensitive-local");
const clinicCodePattern = /^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$/;
const envFile = validateA4EnvFile({ repo, input: process.env.A4_ENV_FILE });

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: repo,
    encoding: "utf8",
    shell: false,
    ...options,
  });
  if (result.status !== 0) throw new Error(`${command} command failed`);
  return result.stdout.trim();
}

if (!project || !runId || !revision || !process.env.A4_APPLY_REPORT) {
  throw new Error("A4_COMPOSE_PROJECT, A4_RUN_ID, A4_TARGET_RELEASE_COMMIT, and A4_APPLY_REPORT are required");
}
const sensitiveStat = fs.lstatSync(sensitiveRoot);
if (!sensitiveStat.isDirectory() || sensitiveStat.isSymbolicLink()) {
  throw new Error("sensitive-local must be a real directory");
}
const canonicalSensitiveRoot = fs.realpathSync(sensitiveRoot);
if (applyPath !== sensitiveRoot && !applyPath.startsWith(`${sensitiveRoot}${path.sep}`)) {
  throw new Error("A4_APPLY_REPORT must be under sensitive-local");
}
let currentParent = sensitiveRoot;
for (const component of path.relative(sensitiveRoot, path.dirname(applyPath)).split(path.sep)) {
  if (component === "") continue;
  currentParent = path.join(currentParent, component);
  const parentStat = fs.lstatSync(currentParent);
  if (!parentStat.isDirectory() || parentStat.isSymbolicLink()) {
    throw new Error("A4_APPLY_REPORT parent directories may not be symlinks");
  }
}
const stat = fs.lstatSync(applyPath);
if (!stat.isFile() || stat.isSymbolicLink() || (stat.mode & 0o077) !== 0) {
  throw new Error("A4_APPLY_REPORT must be an owner-only regular file");
}
const canonicalApplyPath = fs.realpathSync(applyPath);
if (!canonicalApplyPath.startsWith(`${canonicalSensitiveRoot}${path.sep}`)) {
  throw new Error("A4_APPLY_REPORT may not traverse a symlink outside sensitive-local");
}
if (run("git", ["rev-parse", "HEAD"]) !== revision || run("git", ["status", "--porcelain"]) !== "") {
  throw new Error("A4 runtime report requires the exact clean target release checkout");
}

const composeArgs = [
  "compose", "--env-file", envFile, "-p", project,
  "-f", "docker-compose.yml", "-f", "docker-compose.a4-rehearsal.yml",
];
const composeEnv = {
  ...process.env,
  COMPOSE_PROJECT_NAME: project,
  A4_RUN_ID: runId,
  A4_TARGET_RELEASE_COMMIT: revision,
  A4_ENV_FILE: envFile,
};
const configJson = run("docker", [...composeArgs, "config", "--format", "json"], { env: composeEnv });
const composeSummary = validateA4ComposeConfig(JSON.parse(configJson));
const containerIds = ["backend", "frontend", "db"].map((service) => (
  run("docker", [...composeArgs, "ps", "-q", service], { env: composeEnv })
));
if (containerIds.some((id) => !/^[a-f0-9]{64}$/.test(id))) {
  throw new Error("A4 runtime must have exactly one container for each service");
}

const inspections = JSON.parse(run("docker", ["inspect", ...containerIds]));
const imageIds = inspections.map(({ Image }) => Image);
const imageInspections = JSON.parse(run("docker", ["image", "inspect", ...imageIds]));
const [networkInspection] = JSON.parse(run("docker", ["network", "inspect", composeSummary.networkName]));
const [volumeInspection] = JSON.parse(run("docker", ["volume", "inspect", composeSummary.databaseVolumeName]));
const applyBytes = fs.readFileSync(applyPath);
const applyReport = JSON.parse(applyBytes.toString("utf8"));
if (!clinicCodePattern.test(applyReport.clinicCode ?? "") || applyReport.runId !== runId) {
  throw new Error("apply report clinic/run identity is invalid");
}
const report = buildA4RuntimeReport({
  applyReport,
  applyReportSha256: crypto.createHash("sha256").update(applyBytes).digest("hex"),
  composeSummary,
  inspections,
  imageInspections,
  networkInspection,
  volumeInspection,
  generatedAt: new Date().toISOString(),
  emptyBandPreflight: process.env.A4_EMPTY_BAND_PREFLIGHT,
  backupRestorePreflight: process.env.A4_BACKUP_RESTORE_PREFLIGHT,
});

const outputDirectory = path.join(sensitiveRoot, "a4-rehearsal-reports");
try {
  fs.mkdirSync(outputDirectory, { mode: 0o700 });
} catch (error) {
  if (error?.code !== "EEXIST") throw error;
}
const outputStat = fs.lstatSync(outputDirectory);
if (!outputStat.isDirectory() || outputStat.isSymbolicLink()
  || fs.realpathSync(outputDirectory) !== path.join(canonicalSensitiveRoot, "a4-rehearsal-reports")) {
  throw new Error("A4 runtime report directory must be a real directory under sensitive-local");
}
fs.chmodSync(outputDirectory, 0o700);
const output = path.join(outputDirectory, `${applyReport.clinicCode}-${runId}-runtime.json`);
const temporary = fs.mkdtempSync(path.join(outputDirectory, ".runtime-"), { encoding: "utf8" });
fs.chmodSync(temporary, 0o700);
try {
  const temporaryFile = path.join(temporary, "report.json");
  fs.writeFileSync(temporaryFile, `${JSON.stringify(report, null, 2)}\n`, {
    flag: "wx",
    mode: 0o600,
  });
  fs.linkSync(temporaryFile, output);
  fs.unlinkSync(temporaryFile);
} finally {
  fs.rmSync(temporary, { recursive: true, force: true });
}
process.stdout.write(`A4 RUNTIME REPORT: PASS ${output}\n`);
