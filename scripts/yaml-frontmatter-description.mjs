#!/usr/bin/env node
import { readFileSync } from "node:fs";

const source = readFileSync(process.argv[2] ?? "/dev/stdin", "utf8");
const frontmatter = source.match(/^---\r?\n([\s\S]*?)\r?\n---/);
if (!frontmatter) {
  console.error("missing yaml frontmatter");
  process.exit(1);
}

const line = frontmatter[1]
  .split(/\r?\n/)
  .find((entry) => entry.startsWith("description:"));
if (!line) {
  console.error("missing description");
  process.exit(1);
}

let value = line.slice("description:".length).trim();
if (value.startsWith('"') && value.endsWith('"')) {
  value = JSON.parse(value);
} else if (value.startsWith("'") && value.endsWith("'")) {
  value = value.slice(1, -1).replace(/''/g, "'");
}

process.stdout.write(JSON.stringify(value));
