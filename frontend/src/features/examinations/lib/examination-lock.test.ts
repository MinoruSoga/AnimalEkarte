import { describe, expect, it } from "vitest";
import {
  isPersistedCompletedSeal,
  isPersistedConfirmedStatus,
  isPersistedResultsLocked,
} from "./examination-lock";

describe("examination-lock (BUG-033)", () => {
  it("locks confirmed always", () => {
    expect(isPersistedConfirmedStatus("確定")).toBe(true);
    expect(isPersistedResultsLocked("確定", undefined)).toBe(true);
    expect(isPersistedResultsLocked("確定", 2)).toBe(true);
  });

  it("locks first-pass completed without revision history", () => {
    expect(isPersistedCompletedSeal("完了", undefined)).toBe(true);
    expect(isPersistedCompletedSeal("完了", null)).toBe(true);
    expect(isPersistedResultsLocked("完了", undefined)).toBe(true);
  });

  it("does not lock post-unconfirm completed working copy", () => {
    expect(isPersistedCompletedSeal("完了", 1)).toBe(false);
    expect(isPersistedResultsLocked("完了", 1)).toBe(false);
  });

  it("does not lock earlier statuses", () => {
    for (const status of ["依頼中", "検査中", "結果入力済み"] as const) {
      expect(isPersistedResultsLocked(status, undefined)).toBe(false);
      expect(isPersistedCompletedSeal(status, undefined)).toBe(false);
    }
  });
});
