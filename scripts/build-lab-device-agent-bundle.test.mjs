import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import test from "node:test";

const scriptURL = new URL("./build-lab-device-agent-bundle.sh", import.meta.url);
const scriptPath = fileURLToPath(scriptURL);
test("plist configurator writes the exact seven launch arguments without placeholders", () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-plist-"));
  const plistPath = path.join(tempDir, "agent.plist");
  const templatePath = fileURLToPath(new URL("../packaging/macos/com.animalekarte.lab-device-agent.plist", import.meta.url));
  const configuratorSource = fileURLToPath(new URL("../packaging/macos/configure-lab-device-agent-plist.sh", import.meta.url));
  const configuratorPath = path.join(tempDir, "configure-plist.sh");
  const plistBuddyPath = path.join(tempDir, "PlistBuddy");
  const plutilPath = path.join(tempDir, "plutil");
  fs.writeFileSync(plistBuddyPath, `#!/usr/bin/env python3
import plistlib, sys
command = sys.argv[2]
plist_path = sys.argv[3]
_, entry, value = command.split(" ", 2)
with open(plist_path, "rb") as stream:
    data = plistlib.load(stream)
parts = [part for part in entry.split(":") if part]
target = data
for part in parts[:-1]:
    target = target[int(part)] if isinstance(target, list) else target[part]
last = parts[-1]
if isinstance(target, list):
    target[int(last)] = value
else:
    target[last] = value
with open(plist_path, "wb") as stream:
    plistlib.dump(data, stream)
`, { mode: 0o700 });
  fs.writeFileSync(plutilPath, `#!/usr/bin/env python3
import plistlib, sys
with open(sys.argv[-1], "rb") as stream:
    plistlib.load(stream)
`, { mode: 0o700 });
  fs.writeFileSync(
    configuratorPath,
    fs.readFileSync(configuratorSource, "utf8").replace("/usr/libexec/PlistBuddy", plistBuddyPath),
    { mode: 0o700 },
  );
  const expected = [
    "/Users/Test User/Library/Application Support/AnimalEkarte/lab-device-agent",
    "--clinic-id",
    "42",
    "--ports-file",
    "/Users/Test User/Library/Application Support/AnimalEkarte/lab-device-agent-ports",
    "--allowed-origin",
    "https://clinic.example",
  ];

  try {
    fs.copyFileSync(templatePath, plistPath);
    const result = spawnSync("sh", [configuratorPath, plistPath, expected[0], expected[2], expected[4], expected[6], "/tmp/out.log", "/tmp/error.log"], {
      encoding: "utf8",
      env: { ...process.env, PATH: `${tempDir}:${process.env.PATH}` },
    });
    assert.equal(result.status, 0, result.stderr);
    const extract = spawnSync("python3", ["-c", "import json, plistlib, sys; print(json.dumps(plistlib.load(open(sys.argv[1], 'rb'))['ProgramArguments']))", plistPath], { encoding: "utf8" });
    assert.equal(extract.status, 0, extract.stderr);
    assert.deepEqual(JSON.parse(extract.stdout), expected);
    assert.doesNotMatch(fs.readFileSync(plistPath, "utf8"), /__[A-Z_]+__/);

    for (const installer of [
      new URL("../packaging/macos/install-lab-device-agent.sh", import.meta.url),
      new URL("./install-lab-device-agent.sh", import.meta.url),
    ]) {
      assert.match(fs.readFileSync(installer, "utf8"), /configure-(?:lab-device-agent-)?plist\.sh/);
    }
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
});

test("bundle rejects credential and malformed-port origins before Docker", () => {
  const binDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-bundle-bin-"));
  const dockerMarker = path.join(binDir, "docker-called");
  const dockerStub = path.join(binDir, "docker");
  fs.writeFileSync(dockerStub, `#!/bin/sh
: > "${dockerMarker}"
exit 77
`, { mode: 0o700 });

  try {
    for (const origin of [
      "https://user@example.test",
      "https://example.test:bad",
      "https://example.test:",
      "https://example.test:70000",
    ]) {
      fs.rmSync(dockerMarker, { force: true });
      const outputDir = path.join(binDir, `bundle-${Math.random().toString(16).slice(2)}`);
      const result = spawnSync("sh", [scriptPath, "123", origin, outputDir], {
        encoding: "utf8",
        env: { ...process.env, PATH: `${binDir}:${process.env.PATH}` },
      });

      assert.equal(result.status, 2, `${origin}: ${result.stderr}`);
      assert.equal(fs.existsSync(dockerMarker), false, `${origin} reached Docker`);
      assert.equal(fs.existsSync(outputDir), false, `${origin} created output`);
    }
  } finally {
    fs.rmSync(binDir, { recursive: true, force: true });
  }
});
