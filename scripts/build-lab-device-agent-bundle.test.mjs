import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";

const script = fs.readFileSync(new URL("./build-lab-device-agent-bundle.sh", import.meta.url), "utf8");
const plist = '$output_dir/com.animalekarte.lab-device-agent.plist';

test("bundle plutil replacements always target the copied plist", () => {
  assert.match(script, new RegExp(`plutil -replace ProgramArguments\\.2 -string .* "\\${plist}"`));
  assert.match(script, new RegExp(`plutil -replace ProgramArguments\\.6 -string .* "\\${plist}"`));
});
