import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
import test from "node:test";
import { lintInput, lintWorkflows, PLANE_WORKFLOW } from "./actionlint-queue-compat.mjs";

const source = readFileSync(PLANE_WORKFLOW, "utf8");

test("queue compatibility changes only the validated line and preserves diagnostic positions", () => {
  const expected = source.replace("  queue: max", "  # queue: max validated by actionlint-queue-compat.mjs");
  assert.equal(lintInput(source), expected);
  assert.equal(lintInput(source).split("\n").length, source.split("\n").length);
});

test("queue compatibility rejects invalid, missing, duplicate and misplaced settings", () => {
  for (const broken of [
    source.replace("queue: max", "queue: arbitrary"),
    source.replace("  queue: max\n", ""),
    source.replace("queue: max", "queue: max\n  queue: max"),
    source.replace("cancel-in-progress: false", "cancel-in-progress: true"),
    source.replace("cancel-in-progress: false", "cancel-in-progress: false\n  cancel-in-progress: true"),
    source.replace("  queue: max", "    queue: max"),
    source.replace("  queue: max\n", "") + "\nother:\n  queue: max\n",
  ]) assert.throws(() => lintInput(broken));
});

test("every workflow reaches actionlint and failures propagate", () => {
  for (const statuses of [[0, 0], [1], [0, 1], [null]]) {
    const calls = [];
    const result = lintWorkflows("actionlint", (binary, args, options) => {
      calls.push({ binary, args, options });
      return { status: statuses[calls.length - 1] };
    });
    assert.equal(result, statuses.every(status => status === 0) ? 0 : 1);
    const expected = readdirSync(".github/workflows").filter(name => /\.ya?ml$/.test(name))
      .map(name => `.github/workflows/${name}`).filter(path => path !== PLANE_WORKFLOW).sort();
    assert.deepEqual(calls[0].args, expected);
    if (statuses.length === 2) {
      assert.deepEqual(calls[1].args, ["-stdin-filename", PLANE_WORKFLOW, "-"]);
      assert.equal(calls[1].options.input, lintInput(source));
    }
  }
});
