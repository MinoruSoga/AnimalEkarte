import { describe, expect, it } from "vitest";
import { http, HttpResponse, delay } from "msw";
import { server } from "@/testing/mocks/node";
import {
  REVISIT_INTAKE_RECOVERY_RUNTIME_STATUS,
  RevisitIntakeRecoveryQueryError,
  classifyRecoveryHttpError,
  queryRevisitIntakeOperationRecoveryPage,
} from "./revisit-intake-recovery";
import {
  recoveryItemSchema,
  recoveryPageSchema,
  recoveryQueryRequestSchema,
} from "./revisit-intake-recovery-schemas";
import {
  DIGEST,
  KEY_A,
  KEY_B,
  RECOVERY_QUERY_PATH,
  closedItem,
  noStoreJson,
  recoveryError,
  recoveryHandler,
  redactedSuccessResult,
  selfPage,
  succeededItem,
  targetPage,
  unavailableItem,
  unresolvedItem,
  unresolvedResult,
} from "../testing/revisit-intake-recovery-fixtures";

const TARGET_REF = { appointment_id: "12001", pet_id: "801", target_generation: "3" };

async function expectKind(promise: Promise<unknown>, kind: string) {
  await expect(promise).rejects.toMatchObject({ name: "RevisitIntakeRecoveryQueryError", kind });
}

it("adapter は runtime 未実装として明示されている", () => {
  expect(REVISIT_INTAKE_RECOVERY_RUNTIME_STATUS).toBe("not-implemented");
});

describe("recoveryQueryRequestSchema（oneOf・closed）", () => {
  it("target / self を受理し cursor は required nullable", () => {
    expect(
      recoveryQueryRequestSchema.safeParse({ scope: "target", target_ref: TARGET_REF, cursor: null })
        .success,
    ).toBe(true);
    expect(recoveryQueryRequestSchema.safeParse({ scope: "self", cursor: "abc_-9" }).success).toBe(true);
    expect(recoveryQueryRequestSchema.safeParse({ scope: "self" }).success).toBe(false);
    // base64url の末尾 "=" padding は許容する
    expect(recoveryQueryRequestSchema.safeParse({ scope: "self", cursor: "YWJj_-8=" }).success).toBe(true);
    expect(recoveryQueryRequestSchema.safeParse({ scope: "self", cursor: "YWJj===" }).success).toBe(false);
  });

  it("未知 field・clinic/customer などの client 指定・非正規 ID・不正 cursor を拒否する", () => {
    const bad = [
      { scope: "self", cursor: null, clinic_id: "1" },
      { scope: "target", cursor: null },
      { scope: "target", target_ref: { ...TARGET_REF, extra: 1 }, cursor: null },
      { scope: "target", target_ref: { ...TARGET_REF, appointment_id: "012" }, cursor: null },
      { scope: "target", target_ref: { ...TARGET_REF, pet_id: 801 }, cursor: null },
      { scope: "target", target_ref: { ...TARGET_REF, target_generation: "9223372036854775808" }, cursor: null },
      { scope: "target", target_ref: { ...TARGET_REF, target_generation: "abc" }, cursor: null },
      { scope: "target", target_ref: { ...TARGET_REF, target_generation: "1.5" }, cursor: null },
      { scope: "self", cursor: "" },
      { scope: "self", cursor: "a+b/c=" },
    ];
    for (const body of bad) expect(recoveryQueryRequestSchema.safeParse(body).success).toBe(false);
  });
});

describe("recoveryItemSchema（状態の組）", () => {
  it("契約の組を受理する", () => {
    for (const item of [
      succeededItem(KEY_A),
      { ...succeededItem(KEY_A), result: redactedSuccessResult(KEY_A) },
      unresolvedItem(KEY_A),
      { ...unresolvedItem(KEY_A), result: null },
      closedItem(KEY_A),
      { ...closedItem(KEY_A), result: unresolvedResult(KEY_A, "result_unavailable") },
      unavailableItem(KEY_A),
    ]) {
      expect(recoveryItemSchema.safeParse(item).success).toBe(true);
    }
  });

  it("契約外の組・未知 field・key 不一致を拒否する", () => {
    const bad = [
      { ...succeededItem(KEY_A), receipt_digest: null },
      { ...succeededItem(KEY_A), receipt_digest: "A".repeat(64) },
      { ...succeededItem(KEY_A), receipt_digest: DIGEST.slice(1) },
      { ...succeededItem(KEY_A), result: null },
      { ...succeededItem(KEY_A), result: unresolvedResult(KEY_A, "claimed_uncommitted") },
      { ...unresolvedItem(KEY_A), receipt_digest: DIGEST },
      { ...closedItem(KEY_A), result: succeededItem(KEY_A).result },
      { ...unavailableItem(KEY_A), execution_state: "closed_without_effect" },
      { ...unavailableItem(KEY_A), result: unresolvedResult(KEY_A, "claimed_uncommitted") },
      { ...succeededItem(KEY_A), result: succeededItem(KEY_B).result },
      { ...succeededItem(KEY_A), extra: true },
      { ...succeededItem(KEY_A), result: { ...(succeededItem(KEY_A).result as object), target: 1 } },
      {
        ...succeededItem(KEY_A),
        result: {
          ...(succeededItem(KEY_A).result as { operation_result: object }),
          operation_result: { ...(succeededItem(KEY_A).result as { operation_result: object }).operation_result, answer: "x" },
        },
      },
      { ...succeededItem("00000000-0000-0000-0000-000000000000") },
      { ...succeededItem("00000000-0000-4000-8000-0000000000A1") },
      { ...succeededItem(KEY_A), intent: "other" },
      { ...succeededItem(KEY_A), claimed_at: "2026-02-30T00:00:00Z" },
      { ...succeededItem(KEY_A), discovery_state: "lost" },
    ];
    for (const item of bad) expect(recoveryItemSchema.safeParse(item).success).toBe(false);
  });
});

describe("recoveryPageSchema", () => {
  it("response key 集合は closed（欠落・未知を拒否）", () => {
    expect(recoveryPageSchema.safeParse(targetPage()).success).toBe(true);
    const missing: Record<string, unknown> = { ...targetPage() };
    delete missing.complete;
    expect(recoveryPageSchema.safeParse(missing).success).toBe(false);
    expect(recoveryPageSchema.safeParse({ ...targetPage(), extra: 1 }).success).toBe(false);
  });

  it("generation は正規 10 進文字列のみ・active は key 形式・重複 key / 51 件 / complete+cursor を拒否", () => {
    expect(recoveryPageSchema.safeParse(targetPage([], { recovery_generation: "01" })).success).toBe(false);
    expect(recoveryPageSchema.safeParse(targetPage([], { recovery_generation: 1 as unknown as string })).success).toBe(false);
    expect(
      recoveryPageSchema.safeParse(targetPage([], { recovery_generation: "9223372036854775808" })).success,
    ).toBe(false);
    for (const bad of ["abc", "1.5", "-1", "0", ""]) {
      expect(() => recoveryPageSchema.safeParse(targetPage([], { recovery_generation: bad }))).not.toThrow();
      expect(recoveryPageSchema.safeParse(targetPage([], { recovery_generation: bad })).success).toBe(false);
    }
    expect(recoveryPageSchema.safeParse(targetPage([], { active_operation_key: "x" })).success).toBe(false);
    expect(recoveryPageSchema.safeParse(selfPage([], { active_operation_key: KEY_A })).success).toBe(false);
    expect(recoveryPageSchema.safeParse(targetPage([succeededItem(KEY_A), succeededItem(KEY_A)])).success).toBe(false);
    expect(recoveryPageSchema.safeParse(targetPage([], { next_cursor: "abc", complete: true })).success).toBe(false);
    expect(recoveryPageSchema.safeParse(targetPage([], { next_cursor: "abc", complete: false })).success).toBe(true);
    const many = Array.from({ length: 51 }, (_, i) =>
      succeededItem(`00000000-0000-4000-8000-${String(i + 1).padStart(12, "0")}`),
    );
    expect(recoveryPageSchema.safeParse(targetPage(many)).success).toBe(false);
  });
});

describe("classifyRecoveryHttpError（fail-closed）", () => {
  it("契約の status と code の組だけを分類し、それ以外は unclassified", () => {
    const body = (code: string) => ({ error: "m", code });
    expect(classifyRecoveryHttpError(400, body("VALIDATION"))).toBe("validation");
    expect(classifyRecoveryHttpError(401, { anything: 1 })).toBe("unauthenticated");
    expect(classifyRecoveryHttpError(403, body("NOT_AUTHORIZED"))).toBe("not_authorized");
    expect(classifyRecoveryHttpError(404, body("SCOPE_NOT_FOUND"))).toBe("scope_not_found");
    expect(classifyRecoveryHttpError(409, body("RECOVERY_CONFLICT"))).toBe("recovery_conflict");
    expect(classifyRecoveryHttpError(409, body("RECOVERY_BUSY"))).toBe("recovery_conflict");
    expect(classifyRecoveryHttpError(503, body("RESULT_UNAVAILABLE"))).toBe("result_unavailable");
    expect(classifyRecoveryHttpError(404, body("NOT_FOUND"))).toBe("unclassified");
    expect(classifyRecoveryHttpError(404, { error: "m", code: "SCOPE_NOT_FOUND", extra: 1 })).toBe("unclassified");
    expect(classifyRecoveryHttpError(404, "<html>")).toBe("unclassified");
    expect(classifyRecoveryHttpError(500, body("RESULT_UNAVAILABLE"))).toBe("unclassified");
    expect(classifyRecoveryHttpError(429, body("VALIDATION"))).toBe("unclassified");
  });
});

describe("queryRevisitIntakeOperationRecoveryPage（adapter）", () => {
  const args = {
    clinicId: "1",
    idToken: "synthetic-token",
    request: { scope: "target" as const, target_ref: TARGET_REF, cursor: null },
  };

  it("idToken を Bearer で送り、body は request そのもの（clinic / customer を載せない）", async () => {
    const { handler, bodies, headers } = recoveryHandler(() => noStoreJson(targetPage([unresolvedItem(KEY_A)])));
    server.use(handler);
    const page = await queryRevisitIntakeOperationRecoveryPage(args);
    expect(page.items).toHaveLength(1);
    expect(bodies).toEqual([{ scope: "target", target_ref: TARGET_REF, cursor: null }]);
    expect(headers[0].get("authorization")).toBe("Bearer synthetic-token");
  });

  it("self scope の応答は generation / active が null でなければ契約違反", async () => {
    server.use(recoveryHandler(() => noStoreJson(selfPage())).handler);
    await expect(
      queryRevisitIntakeOperationRecoveryPage({ ...args, request: { scope: "self", cursor: null } }),
    ).resolves.toMatchObject({ recovery_generation: null });

    server.use(recoveryHandler(() => noStoreJson(targetPage())).handler);
    await expectKind(
      queryRevisitIntakeOperationRecoveryPage({ ...args, request: { scope: "self", cursor: null } }),
      "contract_violation",
    );
  });

  it("target scope の応答で generation が null なら契約違反", async () => {
    server.use(recoveryHandler(() => noStoreJson(selfPage())).handler);
    await expectKind(queryRevisitIntakeOperationRecoveryPage(args), "contract_violation");
  });

  it("no-store が無い 200 は契約違反（大文字小文字は区別しない）", async () => {
    server.use(http.post(RECOVERY_QUERY_PATH, () => HttpResponse.json(targetPage())));
    await expectKind(queryRevisitIntakeOperationRecoveryPage(args), "contract_violation");
    server.use(
      http.post(RECOVERY_QUERY_PATH, () =>
        HttpResponse.json(targetPage(), { headers: { "CACHE-CONTROL": "private, No-Store" } }),
      ),
    );
    await expect(queryRevisitIntakeOperationRecoveryPage(args)).resolves.toBeDefined();
  });

  it("HTTP エラーを kind に分類し、未知 field のある body は unclassified", async () => {
    const cases: Array<[number, string, string]> = [
      [400, "VALIDATION", "validation"],
      [403, "NOT_AUTHORIZED", "not_authorized"],
      [404, "SCOPE_NOT_FOUND", "scope_not_found"],
      [409, "RECOVERY_CONFLICT", "recovery_conflict"],
      [503, "RESULT_UNAVAILABLE", "result_unavailable"],
    ];
    for (const [status, code, kind] of cases) {
      server.use(recoveryHandler(() => recoveryError(status, code)).handler);
      await expectKind(queryRevisitIntakeOperationRecoveryPage(args), kind);
    }
    server.use(http.post(RECOVERY_QUERY_PATH, () => HttpResponse.json({ error: "x", code: "SCOPE_NOT_FOUND", z: 1 }, { status: 404 })));
    await expectKind(queryRevisitIntakeOperationRecoveryPage(args), "unclassified");
    server.use(http.post(RECOVERY_QUERY_PATH, () => HttpResponse.error()));
    await expectKind(queryRevisitIntakeOperationRecoveryPage(args), "network");
  });

  it("止まった接続は timeout で応答なし扱い（network）になり、永久に pending にならない", async () => {
    server.use(http.post(RECOVERY_QUERY_PATH, async () => { await delay("infinite"); return noStoreJson(targetPage()); }));
    await expectKind(queryRevisitIntakeOperationRecoveryPage({ ...args, timeoutMs: 50 }), "network");
  });

  it("不正な request は送信せず validation", async () => {
    let called = false;
    server.use(http.post(RECOVERY_QUERY_PATH, () => { called = true; return noStoreJson(targetPage()); }));
    await expectKind(
      queryRevisitIntakeOperationRecoveryPage({
        ...args,
        request: { scope: "self", cursor: "not base64url!" },
      }),
      "validation",
    );
    expect(called).toBe(false);
    expect(new RevisitIntakeRecoveryQueryError("validation", "x").kind).toBe("validation");
  });
});
