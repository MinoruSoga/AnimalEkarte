// actionlint 1.7.12 predates GitHub's queue:max concurrency property.
// Validate that property strictly, then lint every other field unchanged.
// https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency
import { spawnSync } from "node:child_process";
import { readFileSync, readdirSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const PLANE_WORKFLOW = ".github/workflows/plane-merge-state.yml";

export function lintInput(source) {
  const lines = source.split("\n");
  const start = lines.indexOf("concurrency:");
  const queues = lines.flatMap((line, index) => /^\s*queue\s*:/.test(line) ? [index] : []);
  const end = lines.findIndex((line, index) => index > start && /^[^\s#]/.test(line));
  const boundary = end < 0 ? lines.length : end;
  if (start < 0 || queues.length !== 1 || queues[0] <= start || queues[0] >= boundary ||
      lines[queues[0]] !== "  queue: max" ||
      lines.slice(start, boundary).filter(line => /^\s*cancel-in-progress\s*:/.test(line)).join("\n") !==
        "  cancel-in-progress: false") {
    throw new Error("Plane concurrency must have exactly one queue: max and cancel-in-progress: false");
  }
  // Keep the original line numbers for diagnostics; no other source is rewritten.
  lines[queues[0]] = "  # queue: max validated by actionlint-queue-compat.mjs";
  return lines.join("\n");
}

export function lintWorkflows(binary, run = spawnSync) {
  const input = lintInput(readFileSync(PLANE_WORKFLOW, "utf8"));
  const paths = readdirSync(".github/workflows").filter(name => /\.ya?ml$/.test(name))
    .map(name => `.github/workflows/${name}`).filter(path => path !== PLANE_WORKFLOW).sort();
  const other = run(binary, paths, { stdio: "inherit" });
  if (other.status !== 0) return other.status ?? 1;
  const plane = run(binary, ["-stdin-filename", PLANE_WORKFLOW, "-"], {
    input, stdio: ["pipe", "inherit", "inherit"],
  });
  return plane.status ?? 1;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    process.exitCode = lintWorkflows(process.argv[2] ?? "./actionlint");
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
