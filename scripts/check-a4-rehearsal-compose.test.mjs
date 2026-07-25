import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { validateA4ComposeConfig } from "./lib/a4-rehearsal-compose.mjs";

function validConfig() {
  const project = "animalekarte-a4-knjo-20260725";
  const runId = "20260725T120000Z";
  const revision = "c46f5114175f6fcc3fde2965a2bac209ec97ba4e";
  const service = (name, target, published, extra = {}) => ({
    ...extra,
    labels: {
      "com.animalekarte.a4.disposable": "true",
      "com.animalekarte.a4.run-id": runId,
    },
    networks: { "ekarte-network": null },
    ports: [{
      mode: "ingress",
      target,
      published: String(published),
      host_ip: "127.0.0.1",
      protocol: "tcp",
    }],
  });
  return {
    name: project,
    services: {
      db: service("db", 5432, 15434, {
        environment: { POSTGRES_DB: "animalekarte_a4" },
        volumes: [{
          type: "volume",
          source: "postgres_data",
          target: "/var/lib/postgresql",
          volume: {},
        }],
      }),
      backend: service("backend", 8080, 18080, {
        image: `${project}_backend:${revision}`,
        build: {
          context: "/repo/backend",
          labels: { "org.opencontainers.image.revision": revision },
        },
        environment: { DB_HOST: "db", DB_NAME: "animalekarte_a4" },
      }),
      frontend: service("frontend", 3000, 13003, {
        image: `${project}_frontend:${revision}`,
        build: {
          context: "/repo/frontend",
          labels: { "org.opencontainers.image.revision": revision },
        },
      }),
    },
    networks: {
      "ekarte-network": {
        name: `${project}_ekarte-network`,
        internal: true,
        attachable: false,
        labels: { "com.animalekarte.a4.disposable": "true" },
      },
    },
    volumes: {
      postgres_data: {
        name: `${project}_postgres_data`,
        labels: { "com.animalekarte.a4.disposable": "true" },
      },
      frontend_node_modules: {
        name: `${project}_frontend_node_modules`,
        labels: { "com.animalekarte.a4.disposable": "true" },
      },
      go_mod_cache: {
        name: `${project}_go_mod_cache`,
        labels: { "com.animalekarte.a4.disposable": "true" },
      },
      go_build_cache: {
        name: `${project}_go_build_cache`,
        labels: { "com.animalekarte.a4.disposable": "true" },
      },
    },
  };
}

test("accepts the isolated disposable A4 compose contract", () => {
  assert.deepEqual(validateA4ComposeConfig(validConfig()), {
    composeProject: "animalekarte-a4-knjo-20260725",
    runId: "20260725T120000Z",
    targetReleaseCommit: "c46f5114175f6fcc3fde2965a2bac209ec97ba4e",
    networkName: "animalekarte-a4-knjo-20260725_ekarte-network",
    databaseVolumeName: "animalekarte-a4-knjo-20260725_postgres_data",
  });
});

test("renders and validates the real A4 Compose files", (t) => {
  const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
  const sensitive = path.join(repo, "sensitive-local");
  fs.mkdirSync(sensitive, { recursive: true, mode: 0o700 });
  const envFile = path.join(sensitive, `.a4-compose-test-${process.pid}.env`);
  fs.writeFileSync(envFile, [
    "DB_USER=a4_user",
    "DB_PASSWORD=disposable-password",
    "DB_NAME=animalekarte_a4",
    "JWT_SECRET=local-disposable-secret-at-least-32-bytes",
    "APP_ENV=development",
  ].join("\n"), { mode: 0o600 });
  t.after(() => fs.rmSync(envFile, { force: true }));
  const result = spawnSync(process.execPath, ["scripts/check-a4-rehearsal-compose.mjs"], {
    cwd: repo,
    encoding: "utf8",
    shell: false,
    env: {
      ...process.env,
      A4_COMPOSE_PROJECT: "animalekarte-a4-contract-test",
      A4_RUN_ID: "contract-test",
      A4_TARGET_RELEASE_COMMIT: "c".repeat(40),
      A4_ENV_FILE: envFile,
    },
  });
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /A4 COMPOSE CONTRACT: PASS/);
});

for (const [name, mutate, message] of [
  ["rejects public interfaces", (config) => {
    config.services.backend.ports[0].host_ip = "0.0.0.0";
  }, /localhost/],
  ["rejects a shared network", (config) => {
    config.networks["ekarte-network"].internal = false;
  }, /internal/],
  ["rejects an unscoped database volume", (config) => {
    config.volumes.postgres_data.name = "ekarte-postgres-data";
  }, /project-scoped/],
  ["rejects a missing disposable label", (config) => {
    delete config.services.frontend.labels["com.animalekarte.a4.disposable"];
  }, /disposable/],
  ["rejects a release-label mismatch", (config) => {
    config.services.frontend.build.labels["org.opencontainers.image.revision"] = "a".repeat(40);
  }, /revision/],
  ["rejects an unexpected published port", (config) => {
    config.services.db.ports.push({
      target: 9999,
      published: "19999",
      host_ip: "127.0.0.1",
      protocol: "tcp",
    });
  }, /exactly one/],
  ["rejects a database bind mount", (config) => {
    config.services.db.volumes[0].type = "bind";
  }, /named volume/],
  ["rejects a backend connected to another network", (config) => {
    config.services.backend.networks.other = null;
  }, /dedicated network/],
]) test(name, () => {
  const config = validConfig();
  mutate(config);
  assert.throws(() => validateA4ComposeConfig(config), message);
});
