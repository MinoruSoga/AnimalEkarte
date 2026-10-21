// mock / fixture が本番 import 経路へ入らないことを検査する（実 API 失敗時の mock fallback を構造的に防ぐ）。
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, resolve } from "node:path";
import { describe, expect, it } from "vitest";

const SRC = resolve(__dirname, "..");

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const full = join(dir, name);
    return statSync(full).isDirectory() ? walk(full) : [full];
  });
}

describe("revisit-intake-recovery fixtures", () => {
  it("test ファイル以外から import されない", () => {
    const offenders = walk(SRC)
      .filter((f) => /\.(ts|tsx)$/.test(f) && !/\.test\.(ts|tsx)$/.test(f))
      .filter((f) => !f.endsWith("revisit-intake-recovery-fixtures.ts"))
      .filter((f) => readFileSync(f, "utf8").includes("revisit-intake-recovery-fixtures"));
    expect(offenders).toEqual([]);
  });
});
