import { describe, expect, it } from "vitest";
import {
  RECOVERY_KEY_A,
  RECOVERY_KEY_B,
  closedItem,
  recoveryPage,
  succeededItem,
  unresolvedItem,
} from "../mock/certificate-recovery-fixtures";
import {
  certificateRecoveryItemSchema,
  certificateRecoveryQueryRequestSchema,
  certificateRecoveryQueryResponseSchema,
  isCanonicalRecoveryGeneration,
} from "./certificate-recovery-schemas";

describe("回復 discovery の要求（送信前の契約強制）", () => {
  it("issue は edition_id=null、reprint は正の edition_id だけ受理する", () => {
    const issue = { target_ref: { vaccination_id: 101, edition_id: null }, intent: "issue", cursor: null };
    const reprint = { target_ref: { vaccination_id: 101, edition_id: 301 }, intent: "reprint", cursor: "abc_-" };
    expect(certificateRecoveryQueryRequestSchema.safeParse(issue).success).toBe(true);
    expect(certificateRecoveryQueryRequestSchema.safeParse(reprint).success).toBe(true);
  });

  it("reprint + null / issue + 非 null を拒否する（両方向）", () => {
    const reprintNull = { target_ref: { vaccination_id: 101, edition_id: null }, intent: "reprint", cursor: null };
    const issueWith = { target_ref: { vaccination_id: 101, edition_id: 301 }, intent: "issue", cursor: null };
    expect(certificateRecoveryQueryRequestSchema.safeParse(reprintNull).success).toBe(false);
    expect(certificateRecoveryQueryRequestSchema.safeParse(issueWith).success).toBe(false);
  });

  it("安全でない整数・0・負・小数・未知 field・空 cursor を拒否する", () => {
    const ok = { target_ref: { vaccination_id: 101, edition_id: null }, intent: "issue", cursor: null };
    const bad = [
      { ...ok, target_ref: { vaccination_id: 2 ** 53, edition_id: null } },
      { ...ok, target_ref: { vaccination_id: 0, edition_id: null } },
      { ...ok, target_ref: { vaccination_id: -1, edition_id: null } },
      { ...ok, target_ref: { vaccination_id: 1.5, edition_id: null } },
      { ...ok, extra: 1 },
      { ...ok, target_ref: { ...ok.target_ref, clinic_id: 1 } },
      { ...ok, cursor: "" },
      { ...ok, cursor: undefined },
    ];
    for (const value of bad) expect(certificateRecoveryQueryRequestSchema.safeParse(value).success).toBe(false);
  });
});

describe("recovery_generation（canonical 十進文字列）", () => {
  it("1..int64 最大だけを受理し、Number へ丸めない", () => {
    expect(isCanonicalRecoveryGeneration("1")).toBe(true);
    expect(isCanonicalRecoveryGeneration("9223372036854775807")).toBe(true);
    for (const v of ["0", "01", "-1", "1.0", "", "9223372036854775808", "10000000000000000000", " 1", "1e3"]) {
      expect(isCanonicalRecoveryGeneration(v)).toBe(false);
    }
  });
});

describe("回復 discovery の応答（closed・組合せ検証）", () => {
  it("3 種の item と 5 キーの応答を受理する", () => {
    const page = recoveryPage([succeededItem(RECOVERY_KEY_A), unresolvedItem(RECOVERY_KEY_B), closedItem("cccccccc-cccc-4ccc-8ccc-cccccccccccc")]);
    expect(certificateRecoveryQueryResponseSchema.safeParse(page).success).toBe(true);
    const unavailable = unresolvedItem(RECOVERY_KEY_B, { unavailable: true });
    expect(certificateRecoveryItemSchema.safeParse(unavailable).success).toBe(true);
  });

  it("契約外の組合せを拒否する（succeeded に digest なし・非 succeeded に digest・unavailable に result など）", () => {
    const good = succeededItem(RECOVERY_KEY_A);
    const unresolved = unresolvedItem(RECOVERY_KEY_B);
    const cases: unknown[] = [
      { ...good, receipt_digest: null },
      { ...good, receipt_digest: "ABC" },
      { ...good, receipt_digest: "0".repeat(63) },
      { ...good, result: null },
      { ...good, result: { status: "not-succeeded", operation_key: RECOVERY_KEY_A } },
      { ...unresolved, receipt_digest: "0".repeat(64) },
      { ...unresolved, result: succeededItem(RECOVERY_KEY_B).result },
      { ...unresolved, discovery_state: "unavailable" }, // unavailable なのに result が非 null
      { ...unresolved, discovery_state: "unavailable", result: null, execution_state: "closed_without_effect" },
      { ...unresolved, execution_state: "succeeded" },
      { ...good, intent: "refund" },
      { ...good, claimed_at: "yesterday" },
      { ...good, operation_key: "00000000-0000-0000-0000-000000000000" },
      { ...good, extra: true },
      { ...good, result: { ...good.result, extra: true } },
      { ...good, result: { ...good.result, operation_key: RECOVERY_KEY_B } }, // item と result の key 不一致
    ];
    for (const value of cases) expect(certificateRecoveryItemSchema.safeParse(value).success).toBe(false);
  });

  it("5 キー以外・件数超過・世代不正・complete と next_cursor の矛盾を拒否する", () => {
    const page = recoveryPage([]);
    const withoutComplete: Record<string, unknown> = { ...page };
    delete withoutComplete.complete;
    const tooMany = recoveryPage(
      Array.from({ length: 51 }, (_, i) => unresolvedItem(`00000000-0000-4000-8000-${String(i + 1).padStart(12, "0")}`)),
    );
    const cases: unknown[] = [
      withoutComplete,
      { ...page, extra: 1 },
      tooMany,
      { ...page, recovery_generation: "0" },
      { ...page, recovery_generation: null },
      { ...page, complete: true, next_cursor: "abc" },
      { ...page, complete: false, next_cursor: null },
      { ...page, active_operation_key: "not-a-uuid" },
    ];
    for (const value of cases) expect(certificateRecoveryQueryResponseSchema.safeParse(value).success).toBe(false);
    expect(certificateRecoveryQueryResponseSchema.safeParse(recoveryPage([], { nextCursor: "abc" })).success).toBe(true);
  });
});
