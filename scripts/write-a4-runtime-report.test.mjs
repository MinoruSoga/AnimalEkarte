import assert from "node:assert/strict";
import test from "node:test";

import { buildA4RuntimeReport } from "./lib/a4-runtime-report.mjs";

const commit = "c".repeat(40);
const project = "animalekarte-a4-knjo-20260725";
const runId = "20260725T120000Z";
const ids = { backend: "1".repeat(64), frontend: "2".repeat(64), db: "3".repeat(64) };
const labels = {
  "com.docker.compose.project": project,
  "com.animalekarte.a4.disposable": "true",
  "com.animalekarte.a4.run-id": runId,
};

function inputs() {
  return {
    applyReport: {
      status: "PASS",
      clinicCode: "KNJO",
      clinicOrdinal: 1,
      runId,
      targetHost: "db",
      targetDatabase: "animalekarte_a4",
    },
    applyReportSha256: "a".repeat(64),
    composeSummary: {
      composeProject: project,
      runId,
      targetReleaseCommit: commit,
      networkName: `${project}_ekarte-network`,
      databaseVolumeName: `${project}_postgres_data`,
    },
    inspections: Object.entries(ids).map(([service, id]) => ({
      Id: id,
      Image: `sha256:${id}`,
      State: {
        Running: true,
        Health: { Status: "healthy" },
        StartedAt: "2026-07-25T12:00:00.000Z",
      },
      HostConfig: {
        NetworkMode: `${project}_ekarte-network`,
        PortBindings: {
          [`${service === "backend" ? 8080 : service === "frontend" ? 3000 : 5432}/tcp`]:
            [{ HostIp: "127.0.0.1", HostPort: "12345" }],
        },
      },
      NetworkSettings: { Networks: { [`${project}_ekarte-network`]: {} } },
      Mounts: service === "db" ? [{
        Type: "volume",
        Name: `${project}_postgres_data`,
        Destination: "/var/lib/postgresql",
        RW: true,
      }] : [],
      Config: {
        Env: service === "db"
          ? ["POSTGRES_DB=animalekarte_a4"]
          : service === "backend" ? ["DB_HOST=db", "DB_NAME=animalekarte_a4"] : [],
        Labels: {
          "com.docker.compose.project": project,
          "com.docker.compose.service": service,
          "com.animalekarte.a4.disposable": "true",
          "com.animalekarte.a4.run-id": runId,
        },
      },
    })),
    imageInspections: Object.entries(ids).map(([service, id]) => ({
      Id: `sha256:${id}`,
      Config: {
        Labels: service === "db" ? {} : { "org.opencontainers.image.revision": commit },
      },
    })),
    networkInspection: {
      Name: `${project}_ekarte-network`,
      Internal: true,
      Attachable: false,
      Labels: labels,
      Containers: Object.fromEntries(Object.values(ids).map((id) => [id, {}])),
    },
    volumeInspection: {
      Name: `${project}_postgres_data`,
      Scope: "local",
      Labels: labels,
    },
    generatedAt: "2026-07-25T12:30:00.000Z",
    emptyBandPreflight: "PASS",
    backupRestorePreflight: "PASS",
  };
}

test("builds the aggregate-only runtime report", () => {
  const report = buildA4RuntimeReport(inputs());
  assert.equal(report.status, "PASS");
  assert.equal(report.containerIds.database, ids.db);
  assert.equal(report.backendImageDigest, `sha256:${ids.backend}`);
  assert.equal(report.attestationMethod, "DOCKER_INSPECT_AND_GIT_HEAD");
});

for (const [name, mutate, expected] of [
  ["rejects failed apply", (value) => { value.applyReport.status = "FAILED"; }, /apply report/],
  ["rejects unhealthy runtime", (value) => {
    value.inspections[0].State.Health.Status = "unhealthy";
  }, /healthy/],
  ["rejects wrong run labels", (value) => {
    value.inspections[1].Config.Labels["com.animalekarte.a4.run-id"] = "other";
  }, /run ID/],
  ["rejects a public network", (value) => { value.networkInspection.Internal = false; }, /network/],
  ["rejects a shared volume", (value) => { value.volumeInspection.Name = "shared"; }, /volume/],
  ["rejects missing operator preflight", (value) => {
    value.backupRestorePreflight = "BLOCKED";
  }, /backup/],
]) test(name, () => {
  const value = inputs();
  mutate(value);
  assert.throws(() => buildA4RuntimeReport(value), expected);
});
