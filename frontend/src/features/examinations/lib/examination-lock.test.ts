import { describe, expect, it } from "vitest";
import { transformExamination } from "@/lib/transforms/examination";
import type { Examination as BackendExamination } from "@/types/generated/models";
import {
  isPersistedCompletedSeal,
  isPersistedConfirmedStatus,
  isPersistedResultsLocked,
  normalizeExaminationLockStatus,
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

  it("accepts BE EN enums as well as FE JA labels", () => {
    expect(normalizeExaminationLockStatus("completed")).toBe("完了");
    expect(normalizeExaminationLockStatus("confirmed")).toBe("確定");
    expect(isPersistedCompletedSeal("completed", null)).toBe(true);
    expect(isPersistedCompletedSeal("completed", undefined)).toBe(true);
    expect(isPersistedResultsLocked("completed", null)).toBe(true);
    expect(isPersistedResultsLocked("confirmed", 9)).toBe(true);
    // post-unconfirm working copy still editable even with EN status
    expect(isPersistedCompletedSeal("completed", 2)).toBe(false);
  });

  it("1014565-like API payload: completed + omitted revision → sealed after transform", () => {
    // UAT residual note: id 1014563 is soft-deleted pending (API 404).
    // Live completed fixture is 1014565 (status=completed, current_revision_version omitted).
    const payload = {
      id: 1014565,
      clinic_id: 1,
      exam_type_id: 1,
      date: "2026-08-12T09:00:00+09:00",
      result_summary: "",
      machine: "",
      status: "completed",
      created_at: "2026-08-12T02:27:45.057479+09:00",
      updated_at: "2026-08-12T02:28:06.71831+09:00",
    } satisfies BackendExamination;

    const record = transformExamination(payload);
    expect(record.status).toBe("完了");
    expect(record.currentRevisionVersion).toBeUndefined();
    expect(
      isPersistedCompletedSeal(record.status, record.currentRevisionVersion),
    ).toBe(true);
    expect(
      isPersistedResultsLocked(record.status, record.currentRevisionVersion),
    ).toBe(true);
  });
});
