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
      "https://example.test\\evil",
      "https://0x7f000001",
      "https://0x7f.0.0.1",
      "https://0x7f.1",
      "https://2130706433",
      "https://127.1",
      "https://0177.0.0.1",
      "https://127.0.0.1.",
      "https://[::ffff:192.0.2.128]",
      "https://[::ffff:c000:280]",
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


test("canonical origin helper matches browser Origin serialization", () => {
  const helperPath = fileURLToPath(new URL("./canonicalize-lab-device-origin.mjs", import.meta.url));
  for (const [raw, expected] of [
    ["https://EXAMPLE.test", "https://example.test"],
    ["https://Example.test:443", "https://example.test"],
    ["https://Example.test:0443", "https://example.test"],
    ["https://Example.test:8443", "https://example.test:8443"],
    ["https://[2001:0DB8:0:0::1]:443", "https://[2001:db8::1]"],
    ["https://[2001:DB8::1]:8443", "https://[2001:db8::1]:8443"],
    ["https://127.0.0.1", "https://127.0.0.1"],
    ["https://service.123.example", "https://service.123.example"],
  ]) {
    const result = spawnSync("node", [helperPath, raw], { encoding: "utf8" });
    assert.equal(result.status, 0, `${raw}: ${result.stderr}`);
    assert.equal(result.stdout.trim(), expected);
  }
});

test("canonical origin helper rejects rewritten numeric and mapped IP hosts", () => {
  const helperPath = fileURLToPath(new URL("./canonicalize-lab-device-origin.mjs", import.meta.url));
  for (const raw of [
    "https://0x7f000001",
    "https://0x7f.0.0.1",
    "https://0x7f.1",
    "https://127.1",
    "https://2130706433",
    "https://0177.0.0.1",
    "https://127.0.0.1.",
    "https://[::ffff:192.0.2.128]",
    "https://[::ffff:c000:280]",
  ]) {
    const result = spawnSync("node", [helperPath, raw], { encoding: "utf8" });
    assert.notEqual(result.status, 0, raw);
    assert.equal(result.stdout, "", raw);
  }
});

test("both macOS entry points canonicalize before later installation work", () => {
  const directInstallPath = fileURLToPath(new URL("./install-lab-device-agent.sh", import.meta.url));
  const cases = [
    ["https://EXAMPLE.test", "https://example.test"],
    ["https://Example.test:443", "https://example.test"],
    ["https://Example.test:0443", "https://example.test"],
  ];
  for (const entry of [
    { path: scriptPath, args: (origin, temp) => ["123", origin, path.join(temp, "bundle")] },
    { path: directInstallPath, args: (origin) => ["123", origin] },
  ]) {
    for (const [raw, expected] of cases) {
      const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-origin-entry-"));
      try {
        const result = spawnSync("sh", ["-x", entry.path, ...entry.args(raw, tempDir)], {
          encoding: "utf8",
          env: { ...process.env, HOME: tempDir },
        });
        assert.match(result.stderr, new RegExp(`allowed_origin=${expected.replaceAll(".", "\\.")}(?:\\n|$)`), `${entry.path}: ${result.stderr}`);
      } finally {
        fs.rmSync(tempDir, { recursive: true, force: true });
      }
    }
  }
});


test("direct installer rejects unsafe origins before device discovery or Docker", () => {
  const directInstallPath = fileURLToPath(new URL("./install-lab-device-agent.sh", import.meta.url));
  const binDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-install-bin-"));
  const dockerMarker = path.join(binDir, "docker-called");
  fs.writeFileSync(path.join(binDir, "docker"), `#!/bin/sh
: > "${dockerMarker}"
exit 77
`, { mode: 0o700 });
  try {
    for (const origin of [
      "https://user@example.test",
      "https://example.test:bad",
      "https://example.test:",
      "https://example.test:70000",
      "https://example.test\\evil",
      "https://0x7f000001",
      "https://0x7f.0.0.1",
      "https://0x7f.1",
      "https://2130706433",
      "https://127.1",
      "https://0177.0.0.1",
      "https://127.0.0.1.",
      "https://[::ffff:192.0.2.128]",
      "https://[::ffff:c000:280]",
    ]) {
      fs.rmSync(dockerMarker, { force: true });
      const result = spawnSync("sh", [directInstallPath, "123", origin], {
        encoding: "utf8",
        env: { ...process.env, HOME: binDir, PATH: `${binDir}:${process.env.PATH}` },
      });
      assert.equal(result.status, 2, `${origin}: ${result.stderr}`);
      assert.equal(fs.existsSync(dockerMarker), false, `${origin} reached Docker`);
    }
  } finally {
    fs.rmSync(binDir, { recursive: true, force: true });
  }
});

test("bundle writes only the canonical origin to its configuration", () => {
  const binDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-bundle-canonical-"));
  const dockerStub = path.join(binDir, "docker");
  const lipoStub = path.join(binDir, "lipo");
  fs.writeFileSync(dockerStub, `#!/bin/sh
if [ "$2" = "cp" ]; then
  for last do :; done
  printf binary > "$last"
fi
`, { mode: 0o700 });
  fs.writeFileSync(lipoStub, `#!/bin/sh
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-output" ]; then
    shift
    printf binary > "$1"
    exit 0
  fi
  shift
done
exit 1
`, { mode: 0o700 });
  const outputDir = path.join(binDir, "bundle");
  try {
    const result = spawnSync("sh", [scriptPath, "123", "https://EXAMPLE.test:443", outputDir], {
      encoding: "utf8",
      env: { ...process.env, PATH: `${binDir}:${process.env.PATH}` },
    });
    assert.equal(result.status, 0, result.stderr);
    assert.equal(fs.readFileSync(path.join(outputDir, "lab-device-agent.conf"), "utf8"), "123\nhttps://example.test\n");
  } finally {
    fs.rmSync(binDir, { recursive: true, force: true });
  }
});
