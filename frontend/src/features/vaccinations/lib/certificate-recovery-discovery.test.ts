import { describe, expect, it, vi } from "vitest";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import type { CertificateRecoveryQueryRequest, CertificateRecoveryQueryResponse } from "../api/certificate-recovery-schemas";
import {
  RECOVERY_KEY_A,
  RECOVERY_KEY_B,
  RECOVERY_KEY_C,
  recoveryPage,
  succeededItem,
  unresolvedItem,
} from "../mock/certificate-recovery-fixtures";
import {
  MAX_RECOVERY_PAGES,
  discoverCertificateOperations,
  recoveryRequestFor,
} from "./certificate-recovery-discovery";

/** queryOperationRecovery だけを差し替えた api（他の操作は呼ばれない想定）。 */
function apiReturning(pages: (CertificateRecoveryQueryResponse | Error)[]) {
  const requests: CertificateRecoveryQueryRequest[] = [];
  const query = vi.fn(async (request: CertificateRecoveryQueryRequest) => {
    requests.push(request);
    const next = pages[Math.min(requests.length - 1, pages.length - 1)];
    if (next instanceof Error) throw next;
    return next;
  });
  return { api: { queryOperationRecovery: query } as unknown as CertificateApi, requests };
}

const issueTarget = { intent: "issue", vaccinationId: 101 } as const;
const reprintTarget = { intent: "reprint", vaccinationId: 101, editionId: 301 } as const;

describe("要求の組み立て（intent と edition_id の対応）", () => {
  it("issue は edition_id=null、reprint は版 ID を載せる", () => {
    expect(recoveryRequestFor(issueTarget, null)).toEqual({
      target_ref: { vaccination_id: 101, edition_id: null },
      intent: "issue",
      cursor: null,
    });
    expect(recoveryRequestFor(reprintTarget, "c1")).toEqual({
      target_ref: { vaccination_id: 101, edition_id: 301 },
      intent: "reprint",
      cursor: "c1",
    });
  });
});

describe("ページ送り", () => {
  it("next_cursor を辿って全ページを集め、cursor を順に送る（0 件・complete は許可を意味しない）", async () => {
    const { api, requests } = apiReturning([
      recoveryPage([unresolvedItem(RECOVERY_KEY_A, { claimedAt: "2026-10-20T14:00:03Z" })], { nextCursor: "p2" }),
      recoveryPage([succeededItem(RECOVERY_KEY_B, { claimedAt: "2026-10-20T14:00:02Z" })], { nextCursor: "p3" }),
      recoveryPage([], { nextCursor: null }),
    ]);
    const result = await discoverCertificateOperations(api, issueTarget);
    expect(requests.map((r) => r.cursor)).toEqual([null, "p2", "p3"]);
    expect(result.items.map((i) => i.operation_key)).toEqual([RECOVERY_KEY_A, RECOVERY_KEY_B]);
    expect(result.scanEnded).toBe(true);
    expect(result.truncated).toBe(false);
    expect(result.recoveryGeneration).toBe("1");
  });

  it("空の complete を返しても、結果は『見えた事実』だけで許可の語彙を持たない", async () => {
    const { api } = apiReturning([recoveryPage([])]);
    const result = await discoverCertificateOperations(api, issueTarget);
    expect(result.items).toEqual([]);
    expect(Object.keys(result)).not.toContain("canStart");
    expect(Object.keys(result)).not.toContain("safe");
  });

  it("cursor の循環（同じ cursor の再提示）は malformed で止める", async () => {
    const { api } = apiReturning([
      recoveryPage([unresolvedItem(RECOVERY_KEY_A)], { nextCursor: "loop" }),
      recoveryPage([unresolvedItem(RECOVERY_KEY_B)], { nextCursor: "loop" }),
    ]);
    await expect(discoverCertificateOperations(api, issueTarget)).rejects.toMatchObject({ kind: "malformed" });
  });

  it("ページ上限で打ち切り、完了扱いにせず truncated を返す（無制限に辿らない）", async () => {
    let n = 0;
    const query = vi.fn(async () => {
      n += 1;
      return recoveryPage([unresolvedItem(`00000000-0000-4000-8000-${String(n).padStart(12, "0")}`)], {
        nextCursor: `page${n}`,
      });
    });
    const api = { queryOperationRecovery: query } as unknown as CertificateApi;
    const result = await discoverCertificateOperations(api, issueTarget);
    expect(query).toHaveBeenCalledTimes(MAX_RECOVERY_PAGES);
    expect(result.truncated).toBe(true);
    expect(result.scanEnded).toBe(false);
  });

  it("重複 key・ページ間の世代/active のずれ・別 intent は malformed", async () => {
    const dup = apiReturning([
      recoveryPage([unresolvedItem(RECOVERY_KEY_A)], { nextCursor: "p2" }),
      recoveryPage([unresolvedItem(RECOVERY_KEY_A)]),
    ]);
    await expect(discoverCertificateOperations(dup.api, issueTarget)).rejects.toMatchObject({ kind: "malformed" });

    const generation = apiReturning([
      recoveryPage([unresolvedItem(RECOVERY_KEY_A)], { nextCursor: "p2", generation: "1" }),
      recoveryPage([unresolvedItem(RECOVERY_KEY_B)], { generation: "2" }),
    ]);
    await expect(discoverCertificateOperations(generation.api, issueTarget)).rejects.toMatchObject({ kind: "malformed" });

    const active = apiReturning([
      recoveryPage([unresolvedItem(RECOVERY_KEY_A)], { nextCursor: "p2", active: RECOVERY_KEY_A }),
      recoveryPage([unresolvedItem(RECOVERY_KEY_B)], { active: RECOVERY_KEY_C }),
    ]);
    await expect(discoverCertificateOperations(active.api, issueTarget)).rejects.toMatchObject({ kind: "malformed" });

    const intent = apiReturning([recoveryPage([unresolvedItem(RECOVERY_KEY_A, { intent: "reprint" })])]);
    await expect(discoverCertificateOperations(intent.api, issueTarget)).rejects.toMatchObject({ kind: "malformed" });
  });

  it("adapter が strict 検査を通さず返した壊れたページも、ここで malformed にする", async () => {
    const broken = { ...recoveryPage([]), extra: true } as unknown as CertificateRecoveryQueryResponse;
    const { api } = apiReturning([broken]);
    await expect(discoverCertificateOperations(api, issueTarget)).rejects.toMatchObject({ kind: "malformed" });
  });

  it("adapter の失敗は握りつぶさず伝える（呼び出し側は保留にする）", async () => {
    const { api } = apiReturning([new CertificateApiError("recovery-unavailable", { httpStatus: 503 })]);
    await expect(discoverCertificateOperations(api, issueTarget)).rejects.toMatchObject({ kind: "recovery-unavailable" });
  });

  it("中止済みの signal では要求を送らない", async () => {
    const controller = new AbortController();
    controller.abort();
    const { api, requests } = apiReturning([recoveryPage([])]);
    await expect(discoverCertificateOperations(api, issueTarget, controller.signal)).rejects.toBeInstanceOf(CertificateApiError);
    expect(requests).toHaveLength(0);
  });
});
