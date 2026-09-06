#!/usr/bin/env node
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

const REQUIRED_METRICS = [
  "http_reqs",
  "iterations",
  "checks",
  "successful_logins",
];

function numeric(value) {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

export function metricActivity(metric) {
  if (!metric || typeof metric !== "object") {
    return 0;
  }

  const flatCount = numeric(metric.count);
  if (flatCount !== null) {
    return flatCount;
  }

  const flatPasses = numeric(metric.passes);
  const flatFails = numeric(metric.fails);
  if (flatPasses !== null || flatFails !== null) {
    return (flatPasses ?? 0) + (flatFails ?? 0);
  }

  const nested = metric.values;
  if (!nested || typeof nested !== "object") {
    return 0;
  }

  const nestedCount = numeric(nested.count);
  if (nestedCount !== null) {
    return nestedCount;
  }

  const nestedPasses = numeric(nested.passes);
  const nestedFails = numeric(nested.fails);
  if (nestedPasses !== null || nestedFails !== null) {
    return (nestedPasses ?? 0) + (nestedFails ?? 0);
  }

  return 0;
}

export function validateSummary(data) {
  const metrics = data?.metrics && typeof data.metrics === "object" ? data.metrics : {};
  const report = [];
  let failed = false;

  for (const name of REQUIRED_METRICS) {
    const activity = metricActivity(metrics[name]);
    report.push(`${name}=${activity}`);
    if (activity <= 0) {
      failed = true;
    }
  }

  return { failed, report };
}

function validatePath(path) {
  if (!existsSync(path)) {
    console.error(`missing summary export: ${path}`);
    process.exit(1);
  }

  let data;
  try {
    data = JSON.parse(readFileSync(path, "utf8"));
  } catch {
    console.error(`invalid JSON summary: ${path}`);
    process.exit(1);
  }

  const { failed, report } = validateSummary(data);
  console.log(`${path}: ${report.join(", ")}`);
  if (failed) {
    console.error(`fail-closed: zero or missing aggregate metrics in ${path}`);
    process.exit(1);
  }
}

const isMain =
  Boolean(process.argv[1]) &&
  fileURLToPath(import.meta.url) === resolve(process.argv[1]);
if (isMain) {
  const paths = process.argv.slice(2);
  if (paths.length === 0) {
    console.error("usage: node scripts/validate-k6-summary.mjs <summary.json>...");
    process.exit(1);
  }
  for (const path of paths) {
    validatePath(path);
  }
  console.log("k6 summary aggregates validated");
}
