#!/usr/bin/env node
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdirSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const ROOT = join(dirname(fileURLToPath(import.meta.url)), "..");
const EXTRACT = join(ROOT, "scripts/yaml-frontmatter-description.mjs");
const SYNC = join(ROOT, ".claude/scripts/sync-agents-skills.sh");

function extract(file) {
  const result = spawnSync(process.execPath, [EXTRACT, file], { encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr);
  return result.stdout;
}

test("quoted command descriptions stay parseable after extraction", () => {
  const dir = mkdtempSync(join(tmpdir(), "yaml-desc-"));
  const file = join(dir, "implement.md");
  writeFileSync(
    file,
    '---\ndescription: "Linear Issue実装 → セルフレビュー \\"quoted\\""\n---\n\n# implement\n',
  );
  assert.equal(JSON.parse(extract(file)), 'Linear Issue実装 → セルフレビュー "quoted"');
});

test("sync-agents-skills.sh emits parseable frontmatter for quoted command descriptions", () => {
  const workspace = mkdtempSync(join(tmpdir(), "sync-skills-"));
  mkdirSync(join(workspace, ".claude/skills/demo-skill"), { recursive: true });
  mkdirSync(join(workspace, ".claude/commands"), { recursive: true });
  mkdirSync(join(workspace, ".claude/rules"), { recursive: true });
  mkdirSync(join(workspace, ".claude/scripts"), { recursive: true });
  mkdirSync(join(workspace, "scripts"), { recursive: true });

  writeFileSync(
    join(workspace, ".claude/skills/demo-skill/SKILL.md"),
    "---\nname: demo-skill\ndescription: demo\n---\n\n# demo\n",
  );
  writeFileSync(
    join(workspace, ".claude/commands/implement.md"),
    '---\ndescription: "Linear Issue実装 → セルフレビュー \\"quoted\\""\n---\n\n# implement\n',
  );
  writeFileSync(join(workspace, ".claude/rules/example.md"), "# rule\n");
  writeFileSync(
    join(workspace, "scripts/yaml-frontmatter-description.mjs"),
    readFileSync(EXTRACT),
  );
  writeFileSync(
    join(workspace, ".claude/scripts/sync-agents-skills.sh"),
    readFileSync(SYNC),
    { mode: 0o755 },
  );

  const result = spawnSync("bash", [join(workspace, ".claude/scripts/sync-agents-skills.sh")], {
    encoding: "utf8",
    cwd: workspace,
  });
  assert.equal(result.status, 0, result.stderr + result.stdout);

  const generated = readFileSync(
    join(workspace, ".agents/skills/source-command-implement/SKILL.md"),
    "utf8",
  );
  const description = generated.match(/^description: (".*")$/m)?.[1];
  assert.ok(description, "generated description missing");
  assert.equal(JSON.parse(description), 'Linear Issue実装 → セルフレビュー "quoted"');
});
