import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import { validateA4EnvFile } from "./lib/a4-env-file.mjs";

function fixture(content) {
  const repo = fs.mkdtempSync(path.join(os.tmpdir(), "a4-env-test-"));
  const sensitive = path.join(repo, "sensitive-local");
  fs.mkdirSync(sensitive, { mode: 0o700 });
  const file = path.join(sensitive, "a4.env");
  fs.writeFileSync(file, content, { mode: 0o600 });
  return { repo, file };
}

const valid = [
  "DB_USER=a4_user",
  "DB_PASSWORD=disposable-password",
  "DB_NAME=animalekarte_a4",
  "JWT_SECRET=local-disposable-secret-at-least-32-bytes",
  "APP_ENV=development",
].join("\n");

test("accepts an owner-only minimal env", (t) => {
  const value = fixture(valid);
  t.after(() => fs.rmSync(value.repo, { recursive: true, force: true }));
  assert.equal(validateA4EnvFile({ repo: value.repo, input: value.file }), value.file);
});

test("rejects cloud and integration credentials", (t) => {
  const value = fixture(`${valid}\nLINE_CHANNEL_SECRET=forbidden\n`);
  t.after(() => fs.rmSync(value.repo, { recursive: true, force: true }));
  assert.throws(
    () => validateA4EnvFile({ repo: value.repo, input: value.file }),
    /forbidden key/,
  );
});

test("rejects a non-owner-only env", (t) => {
  const value = fixture(valid);
  t.after(() => fs.rmSync(value.repo, { recursive: true, force: true }));
  fs.chmodSync(value.file, 0o644);
  assert.throws(
    () => validateA4EnvFile({ repo: value.repo, input: value.file }),
    /owner-only/,
  );
});
