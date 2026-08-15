import assert from "node:assert/strict";
import test from "node:test";

import {
  assertDisposableA4Resources,
  assertFreshA4Resources,
  validateA4ResourceIdentity,
} from "./lib/a4-resource-boundary.mjs";

const project = "animalekarte-a4-knjo-20260725";
const runId = "20260725T120000Z";
const revision = "c".repeat(40);
const labels = {
  "com.docker.compose.project": project,
  "com.animalekarte.a4.disposable": "true",
  "com.animalekarte.a4.run-id": runId,
};

test("accepts an A4 resource identity", () => {
  assert.deepEqual(validateA4ResourceIdentity({ project, runId, revision }), {
    project,
    runId,
    revision,
  });
});

test("accepts a fresh project", () => {
  assert.doesNotThrow(() => assertFreshA4Resources({
    containers: [],
    network: undefined,
    volumes: [undefined, undefined],
  }));
});

test("accepts only labeled resources for cleanup", () => {
  assert.doesNotThrow(() => assertDisposableA4Resources({
    project,
    runId,
    containers: [{ Config: { Labels: labels } }],
    network: { Labels: labels },
    volumes: [{ Labels: labels }, { Labels: labels }],
  }));
});

for (const [name, action, expected] of [
  ["rejects a normal project", () => validateA4ResourceIdentity({
    project: "animalekarte",
    runId,
    revision,
  }), /project/],
  ["rejects an existing project at start", () => assertFreshA4Resources({
    containers: [{}],
    network: undefined,
    volumes: [undefined],
  }), /already exist/],
  ["rejects unlabeled cleanup", () => assertDisposableA4Resources({
    project,
    runId,
    containers: [{ Config: { Labels: {} } }],
    network: undefined,
    volumes: [undefined],
  }), /project label/],
  ["rejects another run at cleanup", () => assertDisposableA4Resources({
    project,
    runId,
    containers: [{ Config: { Labels: { ...labels, "com.animalekarte.a4.run-id": "other" } } }],
    network: undefined,
    volumes: [undefined],
  }), /run ID/],
  ["rejects cleanup when nothing exists", () => assertDisposableA4Resources({
    project,
    runId,
    containers: [],
    network: undefined,
    volumes: [undefined],
  }), /no matching/],
]) test(name, () => assert.throws(action, expected));
