import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import test from "node:test";

const scriptURL = new URL("./build-lab-device-agent-bundle.sh", import.meta.url);
const scriptPath = fileURLToPath(scriptURL);
const originHelperPath = fileURLToPath(new URL("./canonicalize-lab-device-origin.mjs", import.meta.url));
const originParityCorpus = JSON.parse(fs.readFileSync(
  fileURLToPath(new URL("../backend/internal/labdeviceagent/testdata/origin_parity.json", import.meta.url)),
  "utf8",
));
const acceptedOrigins = originParityCorpus.cases.filter(({ canonical }) => canonical !== undefined);
const rejectedOrigins = originParityCorpus.cases.filter(({ canonical }) => canonical === undefined).map(({ raw }) => raw);

assert.equal(originParityCorpus.cases.length, 66, "shared origin corpus size");
assert.equal(acceptedOrigins.length, 16, "shared accepted origin count");
assert.equal(rejectedOrigins.length, 50, "shared rejected origin count");

test("canonical origin helper matches the shared Go/browser parity corpus", () => {
  for (const { raw, canonical } of originParityCorpus.cases) {
    const result = spawnSync("node", [originHelperPath, raw], { encoding: "utf8" });
    if (canonical === undefined) {
      assert.notEqual(result.status, 0, raw);
      assert.equal(result.stdout, "", raw);
    } else {
      assert.equal(new URL(raw).origin, canonical, `${raw}: browser serialization`);
      assert.equal(result.status, 0, `${raw}: ${result.stderr}`);
      assert.equal(result.stdout.trim(), canonical, raw);
    }
  }
});

test("plist configurator writes the exact launch arguments including consumer token", () => {
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
    "--consumer-token",
    "test-consumer-token",
  ];

  try {
    fs.copyFileSync(templatePath, plistPath);
    const result = spawnSync("sh", [configuratorPath, plistPath, expected[0], expected[2], expected[4], expected[6], expected[8], "/tmp/out.log", "/tmp/error.log"], {
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

test("bundle rejects every unsupported shared origin before Docker", () => {
  const binDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-bundle-bin-"));
  const dockerMarker = path.join(binDir, "docker-called");
  const dockerStub = path.join(binDir, "docker");
  fs.writeFileSync(dockerStub, `#!/bin/sh
: > "${dockerMarker}"
exit 77
`, { mode: 0o700 });

  try {
    for (const origin of rejectedOrigins) {
      fs.rmSync(dockerMarker, { force: true });
      const outputDir = path.join(binDir, `bundle-${Math.random().toString(16).slice(2)}`);
      const result = spawnSync("sh", [scriptPath, "123", origin, outputDir], {
        encoding: "utf8",
        env: { ...process.env, PATH: `${binDir}:${process.env.PATH}`, LAB_DEVICE_AGENT_CONSUMER_TOKEN: "" },
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
  const cases = acceptedOrigins.map(({ raw, canonical }) => [raw, canonical]);
  for (const entry of [
    { path: scriptPath, args: (origin, temp) => ["123", origin, path.join(temp, "bundle")] },
    { path: directInstallPath, args: (origin) => ["123", origin] },
  ]) {
    for (const [raw, expected] of cases) {
      const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-origin-entry-"));
      try {
        const result = spawnSync("sh", ["-x", entry.path, ...entry.args(raw, tempDir)], {
          encoding: "utf8",
          env: { ...process.env, HOME: tempDir, LAB_DEVICE_AGENT_CONSUMER_TOKEN: "" },
        });
        const escapedExpected = expected.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        assert.match(result.stderr, new RegExp(`allowed_origin='?(?:${escapedExpected})'?(?:\\n|$)`), `${entry.path}: ${result.stderr}`);
      } finally {
        fs.rmSync(tempDir, { recursive: true, force: true });
      }
    }
  }
});


test("direct installer rejects every unsupported shared origin before device discovery or Docker", () => {
  const directInstallPath = fileURLToPath(new URL("./install-lab-device-agent.sh", import.meta.url));
  const binDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-install-bin-"));
  const dockerMarker = path.join(binDir, "docker-called");
  fs.writeFileSync(path.join(binDir, "docker"), `#!/bin/sh
: > "${dockerMarker}"
exit 77
`, { mode: 0o700 });
  try {
    for (const origin of rejectedOrigins) {
      fs.rmSync(dockerMarker, { force: true });
      const result = spawnSync("sh", [directInstallPath, "123", origin], {
        encoding: "utf8",
        env: { ...process.env, HOME: binDir, PATH: `${binDir}:${process.env.PATH}`, LAB_DEVICE_AGENT_CONSUMER_TOKEN: "" },
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
    const result = spawnSync("sh", [scriptPath, "123", "https://EXAMPLE.test:443", outputDir, "test-consumer-token"], {
      encoding: "utf8",
      env: { ...process.env, PATH: `${binDir}:${process.env.PATH}`, LAB_DEVICE_AGENT_CONSUMER_TOKEN: "" },
    });
    assert.equal(result.status, 0, result.stderr);
    assert.equal(fs.readFileSync(path.join(outputDir, "lab-device-agent.conf"), "utf8"), "123\nhttps://example.test\ntest-consumer-token\n");
  } finally {
    fs.rmSync(binDir, { recursive: true, force: true });
  }
});

function stubBundleBuildTools(binDir) {
  const dockerMarker = path.join(binDir, "docker-called");
  fs.writeFileSync(path.join(binDir, "docker"), `#!/bin/sh
: > "${dockerMarker}"
if [ "$2" = "cp" ]; then
  for last do :; done
  printf binary > "$last"
fi
`, { mode: 0o700 });
  fs.writeFileSync(path.join(binDir, "lipo"), `#!/bin/sh
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
  return dockerMarker;
}

test("bundle script requires an operator-supplied consumer token and does not generate one", () => {
  const source = fs.readFileSync(scriptPath, "utf8");
  assert.match(source, /LAB_DEVICE_AGENT_CONSUMER_TOKEN/);
  assert.doesNotMatch(source, /openssl\s+rand/);
});

test("bundle without consumer token fails before Docker", () => {
  const binDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-bundle-missing-token-"));
  const dockerMarker = stubBundleBuildTools(binDir);
  const outputDir = path.join(binDir, "bundle");
  try {
    for (const extra of [[], [""]]) {
      fs.rmSync(dockerMarker, { force: true });
      fs.rmSync(outputDir, { recursive: true, force: true });
      const result = spawnSync("sh", [scriptPath, "123", "https://clinic.example", outputDir, ...extra], {
        encoding: "utf8",
        env: { ...process.env, PATH: `${binDir}:${process.env.PATH}`, LAB_DEVICE_AGENT_CONSUMER_TOKEN: "" },
      });
      assert.equal(result.status, 2, result.stderr);
      assert.match(result.stderr, /LAB_DEVICE_AGENT_CONSUMER_TOKEN/);
      assert.equal(fs.existsSync(dockerMarker), false, "missing token reached Docker");
      assert.equal(fs.existsSync(outputDir), false, "missing token created output");
    }
  } finally {
    fs.rmSync(binDir, { recursive: true, force: true });
  }
});

test("bundle writes the operator-supplied consumer token from env or argument to conf L3", () => {
  const binDir = fs.mkdtempSync(path.join(os.tmpdir(), "lab-bundle-token-"));
  stubBundleBuildTools(binDir);
  try {
    const envOutputDir = path.join(binDir, "from-env");
    const envResult = spawnSync("sh", [scriptPath, "123", "https://clinic.example", envOutputDir], {
      encoding: "utf8",
      env: { ...process.env, PATH: `${binDir}:${process.env.PATH}`, LAB_DEVICE_AGENT_CONSUMER_TOKEN: "test-consumer-token" },
    });
    assert.equal(envResult.status, 0, envResult.stderr);
    assert.equal(fs.readFileSync(path.join(envOutputDir, "lab-device-agent.conf"), "utf8"), "123\nhttps://clinic.example\ntest-consumer-token\n");

    const argOutputDir = path.join(binDir, "from-arg");
    const argResult = spawnSync("sh", [scriptPath, "123", "https://clinic.example", argOutputDir, "test-consumer-token"], {
      encoding: "utf8",
      env: { ...process.env, PATH: `${binDir}:${process.env.PATH}`, LAB_DEVICE_AGENT_CONSUMER_TOKEN: "env-must-not-win" },
    });
    assert.equal(argResult.status, 0, argResult.stderr);
    assert.equal(fs.readFileSync(path.join(argOutputDir, "lab-device-agent.conf"), "utf8"), "123\nhttps://clinic.example\ntest-consumer-token\n");
  } finally {
    fs.rmSync(binDir, { recursive: true, force: true });
  }
});

function writeExecutable(pathname, contents) {
  fs.writeFileSync(pathname, contents, { mode: 0o700 });
}

function makeInstallerRollbackFixture(t, installerKind, activationFailure, recoveryFailure) {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), `lab-install-rollback-${installerKind}-`));
  const binDir = path.join(tempDir, "bin");
  const deviceDir = path.join(tempDir, "devices");
  const homeDir = path.join(tempDir, "home");
  const installDir = path.join(homeDir, "Library", "Application Support", "AnimalEkarte");
  const launchDir = path.join(homeDir, "Library", "LaunchAgents");
  const binaryPath = path.join(installDir, "lab-device-agent");
  const portsPath = path.join(installDir, "lab-device-agent-ports");
  const plistPath = path.join(launchDir, "com.animalekarte.lab-device-agent.plist");
  const launchLog = path.join(tempDir, "launchctl.log");
  fs.mkdirSync(binDir, { recursive: true });
  fs.mkdirSync(deviceDir, { recursive: true });
  fs.mkdirSync(installDir, { recursive: true });
  fs.mkdirSync(launchDir, { recursive: true });
  fs.writeFileSync(path.join(deviceDir, "cu.usbserial-one"), "");
  fs.writeFileSync(path.join(deviceDir, "cu.usbserial-two"), "");
  fs.writeFileSync(binaryPath, "old binary\n", { mode: 0o700 });
  fs.writeFileSync(portsPath, "old ports\n", { mode: 0o600 });
  fs.writeFileSync(plistPath, "old plist\n", { mode: 0o600 });
  writeExecutable(path.join(binDir, "uname"), "#!/bin/sh\necho arm64\n");
  writeExecutable(path.join(binDir, "lipo"), "#!/bin/sh\nexit 0\n");
  const activationMarker = path.join(tempDir, `activation-${activationFailure}-failed`);
  const recoveryMarker = path.join(tempDir, `recovery-${recoveryFailure ?? "none"}-failed`);
  writeExecutable(path.join(binDir, "launchctl"), `#!/bin/sh
printf '%s\n' "$*" >> '${launchLog}'
case "$1" in
  print) echo 'state = running'; exit 0 ;;
  bootstrap)
    if [ "${activationFailure}" = bootstrap ] && [ ! -e '${activationMarker}' ]; then : > '${activationMarker}'; exit 91; fi
    if [ "${recoveryFailure ?? "none"}" = bootstrap ] && [ -e '${activationMarker}' ] && [ ! -e '${recoveryMarker}' ]; then : > '${recoveryMarker}'; exit 93; fi
    ;;
  kickstart)
    if [ "${activationFailure}" = kickstart ] && [ ! -e '${activationMarker}' ]; then : > '${activationMarker}'; exit 92; fi
    if [ "${recoveryFailure ?? "none"}" = kickstart ] && [ -e '${activationMarker}' ] && [ ! -e '${recoveryMarker}' ]; then : > '${recoveryMarker}'; exit 94; fi
    ;;
esac
`);
  writeExecutable(path.join(binDir, "docker"), "#!/bin/sh\ncopy=0\nfor arg; do [ \"$arg\" = cp ] && copy=1; last=$arg; done\nif [ \"$copy\" = 1 ]; then printf 'new binary\\n' > \"$last\"; fi\nexit 0\n");

  let installerPath;
  let args;
  if (installerKind === "bundle") {
    const bundleDir = path.join(tempDir, "bundle");
    fs.mkdirSync(bundleDir);
    installerPath = path.join(bundleDir, "install.sh");
    fs.copyFileSync(fileURLToPath(new URL("../packaging/macos/install-lab-device-agent.sh", import.meta.url)), installerPath);
    fs.writeFileSync(path.join(bundleDir, "lab-device-agent"), "new binary\n");
    fs.writeFileSync(path.join(bundleDir, "lab-device-agent.conf"), "123\nhttps://clinic.example\ntest-consumer-token\n");
    fs.writeFileSync(path.join(bundleDir, "com.animalekarte.lab-device-agent.plist"), "template plist\n");
    writeExecutable(path.join(bundleDir, "configure-plist.sh"), "#!/bin/sh\nprintf 'new plist\\n' > \"$1\"\n");
    writeExecutable(path.join(binDir, "shasum"), "#!/bin/sh\nexit 0\n");
    args = [];
  } else {
    const fixtureRepo = path.join(tempDir, "repo");
    const fixtureScripts = path.join(fixtureRepo, "scripts");
    const fixtureMacos = path.join(fixtureRepo, "packaging", "macos");
    fs.mkdirSync(fixtureScripts, { recursive: true });
    fs.mkdirSync(fixtureMacos, { recursive: true });
    fs.copyFileSync(fileURLToPath(new URL("./install-lab-device-agent.sh", import.meta.url)), path.join(fixtureScripts, "install-lab-device-agent.sh"));
    fs.copyFileSync(originHelperPath, path.join(fixtureScripts, "canonicalize-lab-device-origin.mjs"));
    fs.copyFileSync(fileURLToPath(new URL("../packaging/macos/com.animalekarte.lab-device-agent.plist", import.meta.url)), path.join(fixtureMacos, "com.animalekarte.lab-device-agent.plist"));
    writeExecutable(path.join(fixtureMacos, "configure-lab-device-agent-plist.sh"), "#!/bin/sh\nprintf 'new plist\\n' > \"$1\"\n");
    installerPath = path.join(fixtureScripts, "install-lab-device-agent.sh");
    args = ["123", "https://clinic.example", "test-consumer-token"];
  }

  t.after(() => fs.rmSync(tempDir, { recursive: true, force: true }));
  const result = spawnSync("sh", [...(installerKind === "direct" ? ["-x"] : []), installerPath, ...args], {
    encoding: "utf8",
    env: {
      ...process.env,
      HOME: homeDir,
      PATH: `${binDir}:${process.env.PATH}`,
      LAB_DEVICE_AGENT_DEVICE_DIR: deviceDir,
      LAB_DEVICE_AGENT_CONSUMER_TOKEN: "test-consumer-token",
    },
  });
  assert.notEqual(result.status, 0, `${installerKind}/${activationFailure} unexpectedly succeeded`);
  assert.equal(fs.readFileSync(binaryPath, "utf8"), "old binary\n");
  assert.equal(fs.readFileSync(portsPath, "utf8"), "old ports\n");
  assert.equal(fs.readFileSync(plistPath, "utf8"), "old plist\n");
  assert.ok(fs.existsSync(launchLog), `status=${result.status}\nstdout=${result.stdout}\nstderr=${result.stderr}`);
  const launchCalls = fs.readFileSync(launchLog, "utf8");
  assert.match(launchCalls, /bootout gui\/\d+\/com\.animalekarte\.lab-device-agent/);
  assert.match(launchCalls, new RegExp(`bootstrap gui/\\d+ ${plistPath.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}`));
  if (recoveryFailure === "bootstrap") {
    assert.match(result.stderr, /Rollback failed: could not bootstrap the previous LaunchAgent/);
    assert.equal((launchCalls.match(/kickstart -k gui\/\d+\/com\.animalekarte\.lab-device-agent/g) ?? []).length, 1);
  } else {
    assert.match(launchCalls, /kickstart -k gui\/\d+\/com\.animalekarte\.lab-device-agent/);
  }
  if (recoveryFailure === "kickstart") {
    assert.match(result.stderr, /Rollback failed: could not restart the previous LaunchAgent/);
  }
  if (recoveryFailure) {
    assert.match(result.stderr, /Installation activation failed and the previous installation was not fully recovered\./);
  }
}

for (const installerKind of ["bundle", "direct"]) {
  for (const activationFailure of ["bootstrap", "kickstart"]) {
    test(`${installerKind} installer restores a running prior installation after ${activationFailure} fails`, (t) => {
      makeInstallerRollbackFixture(t, installerKind, activationFailure);
    });
  }
  for (const recoveryFailure of ["bootstrap", "kickstart"]) {
    test(`${installerKind} installer reports failed prior LaunchAgent ${recoveryFailure} recovery`, (t) => {
      makeInstallerRollbackFixture(t, installerKind, "kickstart", recoveryFailure);
    });
  }
}
