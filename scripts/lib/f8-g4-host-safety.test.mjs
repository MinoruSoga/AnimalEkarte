import assert from "node:assert/strict";
import { test } from "node:test";

import {
  rejectDockerEnvironmentOverrides,
  sanitizedGitEnvironment,
  validateLocalDockerEndpoint,
} from "./f8-g4-host-safety.mjs";

test("removes every Git control variable from release attestation", () => {
  assert.deepEqual(
    sanitizedGitEnvironment({
      PATH: "/bin",
      GIT_DIR: "/attacker",
      GIT_WORK_TREE: "/other",
      GIT_INDEX_FILE: "/tmp/index",
      F8_G4_RUN_ID: "run-1",
    }),
    { PATH: "/bin", F8_G4_RUN_ID: "run-1" },
  );
});

test("rejects Docker daemon and context overrides", () => {
  for (const key of [
    "DOCKER_CERT_PATH", "DOCKER_CONFIG", "DOCKER_CONTEXT",
    "DOCKER_HOST", "DOCKER_TLS", "DOCKER_TLS_VERIFY",
  ]) {
    assert.throws(
      () => rejectDockerEnvironmentOverrides({ [key]: "attacker-controlled" }),
      new RegExp(key),
    );
  }
  assert.doesNotThrow(() => rejectDockerEnvironmentOverrides({ PATH: "/bin" }));
});

test("rejects remote Docker endpoints before touching a daemon", () => {
  assert.throws(
    () => validateLocalDockerEndpoint({
      contextName: "remote",
      endpoint: "ssh://production.example",
    }),
    /local Unix socket/,
  );
  assert.throws(
    () => validateLocalDockerEndpoint({
      contextName: "remote",
      endpoint: "tcp://127.0.0.1:2375",
    }),
    /local Unix socket/,
  );
});
