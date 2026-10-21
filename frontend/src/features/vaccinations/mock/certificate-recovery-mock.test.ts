import { describe, expect, it } from "vitest";
import type { CertificateIssueRequest } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import { certificateRecoveryQueryResponseSchema } from "../api/certificate-recovery-schemas";
import { discoverCertificateOperations } from "../lib/certificate-recovery-discovery";
import { createMockCertificateApi } from "./certificate-mock-api";

const KEY = "11111111-1111-4111-8111-111111111111";
const issueReq = { target_ref: { vaccination_id: 101, edition_id: null }, intent: "issue", cursor: null } as const;

async function issueNormal(api: ReturnType<typeof createMockCertificateApi>["api"], key = KEY) {
  const base = { kind: "normal", layout: "portrait", inputs: {}, preconditions: { source_version: 7, latest_edition_id: null } } as const;
  const preview = await api.previewForIssue(101, base);
  const body: CertificateIssueRequest = {
    ...base,
    operation_key: key,
    preview_fingerprint: preview.preview_fingerprint,
    preview_display_issue_date: preview.snapshot.display_issue_date,
  };
  return api.issue(101, body);
}

describe("契約モック: 回復 discovery（実 API ではない）", () => {
  it("応答は契約の strict zod を通り、成功は digest 付き・未成功は unresolved で返る", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await issueNormal(api);
    const key2 = "22222222-2222-4222-8222-222222222222";
    controls.failNext("issue", { error: new CertificateApiError("network"), claimOnly: true });
    await api.issue(101, { kind: "normal", layout: "portrait", inputs: {}, preconditions: { source_version: 7, latest_edition_id: 301 }, operation_key: key2, preview_fingerprint: "x".repeat(64), preview_display_issue_date: "2026-10-20" }).catch(() => undefined);

    const page = await api.queryOperationRecovery(issueReq);
    expect(certificateRecoveryQueryResponseSchema.safeParse(page).success).toBe(true);
    expect(page.items.map((i) => [i.operation_key, i.execution_state])).toEqual([
      [key2, "unresolved"],
      [KEY, "succeeded"],
    ]);
    expect(page.complete).toBe(true);
    expect(page.next_cursor).toBeNull();
  });

  it("120 件は 50/50/20 に分かれ、cursor を辿ると重複なく claimed_at DESC, key DESC で全件返る", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    controls.recovery.seedUnresolved(120, "issue", null);
    const first = await api.queryOperationRecovery(issueReq);
    expect(first.items).toHaveLength(50);
    expect(first.complete).toBe(false);
    expect(first.next_cursor).not.toBeNull();

    const all = await discoverCertificateOperations(api, { intent: "issue", vaccinationId: 101 });
    expect(all.items).toHaveLength(120);
    expect(new Set(all.items.map((i) => i.operation_key)).size).toBe(120);
    const times = all.items.map((i) => i.claimed_at);
    expect([...times].sort().reverse()).toEqual(times);
  });

  it("改ざんした cursor・契約違反の要求は 400 VALIDATION、別接種/未知の版は 404 SCOPE_NOT_FOUND", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    await expect(api.queryOperationRecovery({ ...issueReq, cursor: "bm90LWpzb24" })).rejects.toMatchObject({ httpStatus: 400, code: "VALIDATION" });
    await expect(
      api.queryOperationRecovery({ target_ref: { vaccination_id: 101, edition_id: null }, intent: "reprint", cursor: null } as never),
    ).rejects.toMatchObject({ httpStatus: 400 });
    await expect(
      api.queryOperationRecovery({ ...issueReq, target_ref: { vaccination_id: 999, edition_id: null } }),
    ).rejects.toMatchObject({ kind: "recovery-scope-not-found", httpStatus: 404, code: "SCOPE_NOT_FOUND" });
    await expect(
      api.queryOperationRecovery({ target_ref: { vaccination_id: 101, edition_id: 12345 }, intent: "reprint", cursor: null }),
    ).rejects.toMatchObject({ kind: "recovery-scope-not-found" });
  });

  it("unavailable は unresolved + result null、closed は closed_without_effect で返る", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const [a, b] = controls.recovery.seedUnresolved(2, "issue", null);
    controls.recovery.markUnavailable(a);
    controls.recovery.closeWithoutEffect(b);
    const page = await api.queryOperationRecovery(issueReq);
    const byKey = Object.fromEntries(page.items.map((i) => [i.operation_key, i]));
    expect(byKey[a]).toMatchObject({ discovery_state: "unavailable", execution_state: "unresolved", result: null });
    expect(byKey[b]).toMatchObject({ discovery_state: "result", execution_state: "closed_without_effect" });
  });
});
