#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  rejectDockerEnvironmentOverrides,
  validateLocalDockerAttestation,
  validateLocalDockerEndpoint,
} from "./lib/f8-g4-host-safety.mjs";

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const project = process.env.F8_G4_COMPOSE_PROJECT;
const runId = process.env.F8_G4_RUN_ID;
const PROJECT_RE = /^animalekarte-f8-g4-[a-z0-9][a-z0-9-]{0,35}$/;
const RUN_RE = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/;
const dockerEnvironment = {
  ...process.env,
  F8_G4_BUILD_CONTEXT: path.join(repo, "backend"),
  F8_G4_BACKEND_TREE_ID: "cleanup-only",
  F8_G4_RUNNER_IMAGE: "cleanup-only:latest",
};
let pinnedDockerEndpoint;

function requireValue(condition, message) {
  if (!condition) throw new Error(`F8 G4 cleanup rejected: ${message}`);
}

function rawDocker(args, allowMissing = false) {
  const result = spawnSync("docker", args, {
    cwd: repo,
    encoding: "utf8",
    env: dockerEnvironment,
    shell: false,
    timeout: 120_000,
    killSignal: "SIGTERM",
  });
  if (result.error?.code === "ETIMEDOUT") {
    throw new Error("F8 G4 cleanup rejected: Docker command timed out");
  }
  if (allowMissing && result.status !== 0 && /no such|not found/i.test(result.stderr)) {
    return undefined;
  }
  requireValue(result.status === 0, "Docker inspection failed");
  return result.stdout.trim();
}

function docker(args, allowMissing = false) {
  requireValue(pinnedDockerEndpoint, "Docker endpoint is not attested");
  return rawDocker(["--host", pinnedDockerEndpoint, ...args], allowMissing);
}

function attestLocalDocker() {
  rejectDockerEnvironmentOverrides();
  const contextName = rawDocker(["context", "show"]);
  const [context] = JSON.parse(rawDocker(["context", "inspect", contextName]));
  const endpoint = context?.Endpoints?.docker?.Host;
  validateLocalDockerEndpoint({ contextName, endpoint });
  pinnedDockerEndpoint = endpoint;
  const daemonId = JSON.parse(docker(["info", "--format", "{{json .ID}}"]));
  return validateLocalDockerAttestation({ contextName, endpoint, daemonId });
}

function reattestPinnedDocker(expected) {
  const daemonId = JSON.parse(docker(["info", "--format", "{{json .ID}}"]));
  const current = validateLocalDockerAttestation({
    contextName: expected.contextName,
    endpoint: expected.endpoint,
    daemonId,
  });
  requireValue(
    JSON.stringify(current) === JSON.stringify(expected),
    "Docker daemon identity changed before cleanup",
  );
}

requireValue(PROJECT_RE.test(project ?? ""), "project is invalid");
requireValue(RUN_RE.test(runId ?? ""), "run ID is invalid");
const initialDocker = attestLocalDocker();
const ids = docker(["ps", "-aq", "--filter", `label=com.docker.compose.project=${project}`])
  .split(/\s+/)
  .filter(Boolean);
const resources = ids.length === 0 ? [] : JSON.parse(docker(["inspect", ...ids]));
for (const resource of resources) {
  const labels = resource.Config?.Labels ?? {};
  requireValue(labels["com.docker.compose.project"] === project, "container project mismatch");
  requireValue(labels["com.animalekarte.f8-g4.disposable"] === "true", "container is not disposable");
  requireValue(labels["com.animalekarte.f8-g4.run-id"] === runId, "container run ID mismatch");
}
for (const [kind, name] of [
  ["network", `${project}_f8-g4-network`],
  ["volume", `${project}_postgres_data`],
]) {
  const output = docker([kind, "inspect", name], true);
  if (output === undefined) continue;
  const resource = JSON.parse(output)[0];
  requireValue(resource.Labels?.["com.docker.compose.project"] === project, `${kind} project mismatch`);
  requireValue(resource.Labels?.["com.animalekarte.f8-g4.disposable"] === "true", `${kind} is not disposable`);
  requireValue(resource.Labels?.["com.animalekarte.f8-g4.run-id"] === runId, `${kind} run ID mismatch`);
  resources.push(resource);
}
requireValue(resources.length > 0, "no matching disposable resources exist");
reattestPinnedDocker(initialDocker);
const envFile = process.env.F8_G4_ENV_FILE;
requireValue(envFile && path.isAbsolute(envFile), "F8_G4_ENV_FILE must be absolute");
docker([
  "compose", "--env-file", envFile, "-p", project,
  "-f", "docker-compose.f8-g4-rehearsal.yml",
  "down", "--volumes", "--remove-orphans",
]);
process.stdout.write(`F8 G4 RESOURCE CLEANUP: PASS project=${project}\n`);
