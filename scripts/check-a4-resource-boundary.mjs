#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import net from "node:net";
import path from "node:path";

import {
  assertDisposableA4Resources,
  assertFreshA4Resources,
  validateA4ResourceIdentity,
} from "./lib/a4-resource-boundary.mjs";

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const mode = process.argv[2];
const identity = validateA4ResourceIdentity({
  project: process.env.A4_COMPOSE_PROJECT,
  runId: process.env.A4_RUN_ID,
  revision: process.env.A4_TARGET_RELEASE_COMMIT,
});
if (mode !== "start" && mode !== "destroy") throw new Error("mode must be start or destroy");

function run(args, { allowMissing = false } = {}) {
  const result = spawnSync("docker", args, {
    cwd: repo,
    encoding: "utf8",
    shell: false,
  });
  if (result.status !== 0) {
    if (allowMissing && /no such|not found/i.test(result.stderr)) return undefined;
    throw new Error("Docker resource inspection failed");
  }
  return result.stdout.trim();
}

const idOutput = run([
  "ps", "-aq", "--filter", `label=com.docker.compose.project=${identity.project}`,
]);
const ids = idOutput === "" ? [] : idOutput.split(/\s+/);
const containers = ids.length === 0 ? [] : JSON.parse(run(["inspect", ...ids]));
function inspectOne(kind, name) {
  const output = run([kind, "inspect", name], { allowMissing: true });
  return output === undefined ? undefined : JSON.parse(output)[0];
}
const network = inspectOne("network", `${identity.project}_ekarte-network`);
const volumes = [
  "postgres_data",
  "frontend_node_modules",
  "go_mod_cache",
  "go_build_cache",
].map((suffix) => inspectOne("volume", `${identity.project}_${suffix}`));

if (mode === "start") {
  const head = spawnSync("git", ["rev-parse", "HEAD"], {
    cwd: repo,
    encoding: "utf8",
    shell: false,
  });
  const status = spawnSync("git", ["status", "--porcelain"], {
    cwd: repo,
    encoding: "utf8",
    shell: false,
  });
  if (head.status !== 0 || status.status !== 0
    || head.stdout.trim() !== identity.revision || status.stdout.trim() !== "") {
    throw new Error("A4 start requires the exact clean target release checkout");
  }
  assertFreshA4Resources({ containers, network, volumes });
  const ports = [
    process.env.A4_DB_PORT ?? "15434",
    process.env.A4_BACKEND_PORT ?? "18080",
    process.env.A4_FRONTEND_PORT ?? "13003",
  ].map((value) => Number(value));
  if (ports.some((port) => !Number.isSafeInteger(port) || port < 1 || port > 65535)
    || new Set(ports).size !== ports.length) {
    throw new Error("A4 host ports must be unique valid TCP ports");
  }
  for (const port of ports) {
    await new Promise((resolve, reject) => {
      const server = net.createServer();
      server.once("error", () => reject(new Error(`A4 host port ${port} is unavailable`)));
      server.listen({ host: "127.0.0.1", port, exclusive: true }, () => (
        server.close((error) => error ? reject(error) : resolve())
      ));
    });
  }
} else {
  assertDisposableA4Resources({
    containers,
    network,
    volumes,
    project: identity.project,
    runId: identity.runId,
  });
}
process.stdout.write(`A4 RESOURCE BOUNDARY: PASS mode=${mode} project=${identity.project}\n`);
