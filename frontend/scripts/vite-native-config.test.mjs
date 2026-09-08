import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";
import test from "node:test";

// Native loading must work without Vite writing a bundled config into the
// read-only source/dependency mounts used by the offline verification runner.
const { default: config } = await import("../vite.config.ts");

test("native config loading preserves the source alias and all app entries", () => {
  assert.equal(config.resolve.alias["@"], fileURLToPath(new URL("../src", import.meta.url)));
  assert.deepEqual(config.build.rollupOptions.input, {
    main: fileURLToPath(new URL("../index.html", import.meta.url)),
    "line-reserve": fileURLToPath(new URL("../line-reserve/index.html", import.meta.url)),
    liff: fileURLToPath(new URL("../liff/index.html", import.meta.url)),
  });
});

test("native config retains the project test setup and coverage exclusions", () => {
  assert.equal(config.test.environment, "jsdom");
  assert.equal(config.test.setupFiles, "./src/testing/setup.ts");
  assert.ok(config.test.exclude.includes("scripts/**"));
  assert.ok(config.test.coverage.exclude.includes("src/types/generated/**"));
});
