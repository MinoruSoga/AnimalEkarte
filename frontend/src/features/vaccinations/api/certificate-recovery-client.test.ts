import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "@/testing/mocks/node";
import {
  RECOVERY_KEY_A,
  RECOVERY_KEY_B,
  recoveryPage,
  succeededItem,
  unresolvedItem,
} from "../mock/certificate-recovery-fixtures";
import { CertificateApiError, classifyCertificateError, classifyRecoveryError } from "./certificate-errors";
import { createHttpCertificateApi } from "./certificate-http-client";
import type { CertificateRecoveryQueryRequest } from "./certificate-recovery-schemas";

/**
 * 回復 discovery の実 API クライアント（契約モックではない）。契約は selected / runtime 未実装のため
 * 実 BE との結合は未検証。MSW はこのテスト内の応答を差し替えるだけ。
 */
const ENDPOINT = "/api/v1/vaccination-certificate-operation-recovery/query";
const api = createHttpCertificateApi();

const issueRequest: CertificateRecoveryQueryRequest = {
  target_ref: { vaccination_id: 101, edition_id: null },
  intent: "issue",
  cursor: null,
};
const reprintRequest: CertificateRecoveryQueryRequest = {
  target_ref: { vaccination_id: 101, edition_id: 301 },
  intent: "reprint",
  cursor: null,
};

describe("queryOperationRecovery（POST）", () => {
  it("閉じた body を POST し、契約どおりの応答を返す", async () => {
    let received: unknown;
    server.use(
      http.post(ENDPOINT, async ({ request }) => {
        received = await request.json();
        return HttpResponse.json(recoveryPage([succeededItem(RECOVERY_KEY_A), unresolvedItem(RECOVERY_KEY_B)], { active: RECOVERY_KEY_A }));
      }),
    );
    const page = await api.queryOperationRecovery(issueRequest);
    expect(received).toEqual({ target_ref: { vaccination_id: 101, edition_id: null }, intent: "issue", cursor: null });
    expect(page.items).toHaveLength(2);
    expect(page.recovery_generation).toBe("1");
  });

  it("契約違反の要求（reprint+null / issue+非 null / 非安全整数）は送信せず validation で拒否する", async () => {
    let sent = 0;
    server.use(
      http.post(ENDPOINT, () => {
        sent += 1;
        return HttpResponse.json(recoveryPage([]));
      }),
    );
    const bad = [
      { target_ref: { vaccination_id: 101, edition_id: null }, intent: "reprint", cursor: null },
      { target_ref: { vaccination_id: 101, edition_id: 301 }, intent: "issue", cursor: null },
      { target_ref: { vaccination_id: 2 ** 53, edition_id: null }, intent: "issue", cursor: null },
    ] as unknown as CertificateRecoveryQueryRequest[];
    for (const request of bad) {
      await expect(api.queryOperationRecovery(request)).rejects.toMatchObject({ kind: "validation" });
    }
    expect(sent).toBe(0);
  });

  it("未知 field・組合せ不正・世代不正の応答は malformed（保留側）", async () => {
    const bodies: unknown[] = [
      { ...recoveryPage([]), extra: 1 },
      recoveryPage([], { generation: "0" }),
      { ...recoveryPage([]), items: [{ ...unresolvedItem(RECOVERY_KEY_A), receipt_digest: "0".repeat(64) }] },
    ];
    for (const body of bodies) {
      server.use(http.post(ENDPOINT, () => HttpResponse.json(body)));
      await expect(api.queryOperationRecovery(issueRequest)).rejects.toMatchObject({ kind: "malformed" });
    }
  });

  it("別 intent の item・別接種/別版を指す成功は、形が正しくても採用しない（要求へ束縛）", async () => {
    server.use(
      http.post(ENDPOINT, () => HttpResponse.json(recoveryPage([unresolvedItem(RECOVERY_KEY_A, { intent: "reprint" })]))),
    );
    await expect(api.queryOperationRecovery(issueRequest)).rejects.toMatchObject({ kind: "malformed" });

    server.use(
      http.post(ENDPOINT, () =>
        HttpResponse.json(recoveryPage([succeededItem(RECOVERY_KEY_A, { refs: { vaccinationId: 999 } })])),
      ),
    );
    await expect(api.queryOperationRecovery(issueRequest)).rejects.toMatchObject({ kind: "malformed" });

    // reprint は対象の版を指す成功だけ。別の版(302)の成功は採用しない。
    server.use(
      http.post(ENDPOINT, () =>
        HttpResponse.json(
          recoveryPage([succeededItem(RECOVERY_KEY_A, { intent: "reprint", refs: { editionId: 302 } })]),
        ),
      ),
    );
    await expect(api.queryOperationRecovery(reprintRequest)).rejects.toMatchObject({ kind: "malformed" });
    server.use(
      http.post(ENDPOINT, () =>
        HttpResponse.json(
          recoveryPage([succeededItem(RECOVERY_KEY_A, { intent: "reprint", refs: { editionId: 301 } })]),
        ),
      ),
    );
    await expect(api.queryOperationRecovery(reprintRequest)).resolves.toMatchObject({ items: [{ operation_key: RECOVERY_KEY_A }] });
  });

  it("新エラー形 {error,code} を保留側に分類する（404 は存在を開示せず『未 claim』にしない）", async () => {
    const cases: [number, unknown, string][] = [
      [400, { error: "x", code: "VALIDATION" }, "validation"],
      [403, { error: "x", code: "NOT_AUTHORIZED" }, "forbidden"],
      [404, { error: "x", code: "SCOPE_NOT_FOUND" }, "recovery-scope-not-found"],
      [404, { error: "x" }, "server"], // 未実装 route など契約外の 404
      [409, { error: "x", code: "RECOVERY_CONFLICT" }, "recovery-conflict"],
      [503, { error: "x", code: "RESULT_UNAVAILABLE" }, "recovery-unavailable"],
      [500, { error: "x" }, "server"],
    ];
    for (const [status, body, kind] of cases) {
      server.use(http.post(ENDPOINT, () => HttpResponse.json(body as Record<string, unknown>, { status })));
      const error = await api.queryOperationRecovery(issueRequest).catch((e: unknown) => e);
      expect(error).toBeInstanceOf(CertificateApiError);
      expect(error).toMatchObject({ kind, httpStatus: status });
    }
  });

  it("通信断は network（保留）", async () => {
    server.use(http.post(ENDPOINT, () => HttpResponse.error()));
    await expect(api.queryOperationRecovery(issueRequest)).rejects.toMatchObject({ kind: "network" });
  });
});

describe("既存 C06 の分類は変えない", () => {
  it("401 は 認証切れ（共有 axios の refresh interceptor を避けるため分類関数で確認）", () => {
    expect(classifyRecoveryError(401, { error: "x" }).kind).toBe("unauthenticated");
    expect(classifyRecoveryError(undefined, null).kind).toBe("network");
  });

  it("lookup の 404 scope-not-found は従来どおり『未 claim』、回復 discovery の 404 は別 kind", () => {
    expect(classifyCertificateError(404, { code: "scope-not-found" }).kind).toBe("scope-not-found");
    expect(classifyCertificateError(409, { code: "operation-in-progress" }).kind).toBe("operation-in-progress");
    expect(classifyCertificateError(503, { code: "result-unavailable" }).kind).toBe("result-unavailable");
    expect(classifyRecoveryError(404, { code: "scope-not-found" }).kind).toBe("server");
    expect(classifyRecoveryError(404, { code: "SCOPE_NOT_FOUND" }).kind).toBe("recovery-scope-not-found");
  });
});
