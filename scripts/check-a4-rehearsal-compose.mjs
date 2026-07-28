#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import path from "node:path";

import { validateA4EnvFile } from "./lib/a4-env-file.mjs";
import { validateA4ComposeConfig } from "./lib/a4-rehearsal-compose.mjs";

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const project = process.env.A4_COMPOSE_PROJECT;
const runId = process.env.A4_RUN_ID;
const revision = process.env.A4_TARGET_RELEASE_COMMIT;
const envFile = validateA4EnvFile({ repo, input: process.env.A4_ENV_FILE });

if (!project || !runId || !revision) {
  throw new Error("A4_COMPOSE_PROJECT, A4_RUN_ID, and A4_TARGET_RELEASE_COMMIT are required");
}

const result = spawnSync("docker", [
  "compose",
  "--env-file", envFile,
  "-p", project,
  "-f", "docker-compose.yml",
  "-f", "docker-compose.a4-rehearsal.yml",
  "config",
  "--format", "json",
], {
  cwd: repo,
  encoding: "utf8",
  shell: false,
  env: {
    ...process.env,
    COMPOSE_PROJECT_NAME: project,
    A4_RUN_ID: runId,
    A4_TARGET_RELEASE_COMMIT: revision,
    A4_ENV_FILE: envFile,
  },
});

if (result.status !== 0) {
  process.stderr.write(result.stderr);
  throw new Error("Docker Compose could not render the A4 rehearsal configuration");
}

const summary = validateA4ComposeConfig(JSON.parse(result.stdout));
process.stdout.write(`A4 COMPOSE CONTRACT: PASS ${JSON.stringify(summary)}\n`);
