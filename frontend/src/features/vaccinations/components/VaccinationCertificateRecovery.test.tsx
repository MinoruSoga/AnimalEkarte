import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createTestWrapper } from "@/testing/TestUtils";
import type { CertificateApi, CertificateIssueRequest } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import { createMockCertificateApi, type MockCertificateControls } from "../mock/certificate-mock-api";
import { VaccinationCertificateScreen } from "./VaccinationCertificateScreen";

const permission = vi.hoisted(() => ({ canEdit: true }));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: permission.canEdit, canDelete: false }),
}));
const printSpy = vi.hoisted(() => vi.fn(async () => {}));
vi.mock("../lib/pdf", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../lib/pdf")>()),
  printPdfBlob: printSpy,
}));

/**
 * EMR-487 Slice A: reload 後の未確定操作の発見・保留・再照会。
 * 「reload」は unmount して、同じ契約モック(サーバ側の状態)に対し新しい QueryClient で描画し直して再現する
 * （in-memory の pending key 登録簿は QueryClient 単位なので空になる）。実 API・実 BE の検証ではない。
 */
function renderScreen(api: CertificateApi) {
  return render(<VaccinationCertificateScreen vaccinationId={101} api={api} />, {
    wrapper: createTestWrapper(),
  });
}

function callsOf(controls: MockCertificateControls, operation: string) {
  return controls.calls.filter((c) => c.operation === operation);
}

async function createPreview(user: ReturnType<typeof userEvent.setup>) {
  await user.click(await screen.findByRole("button", { name: "プレビューを作成" }));
  await screen.findByTitle("証明書の確認プレビュー");
}

async function seedNormalEdition(api: CertificateApi) {
  const request = {
    kind: "normal" as const,
    layout: "portrait" as const,
    inputs: {},
    preconditions: { source_version: 7, latest_edition_id: null },
  };
  const preview = await api.previewForIssue(101, request);
  await api.issue(101, {
    ...request,
    operation_key: "99999999-9999-4999-8999-999999999999",
    preview_fingerprint: preview.preview_fingerprint,
    preview_display_issue_date: preview.snapshot.display_issue_date,
  });
}

const HOLD_ISSUE = "結果が確定していない発行操作があります";
const HOLD_REPRINT = "結果が確定していない再印刷操作があります";

beforeEach(() => {
  permission.canEdit = true;
  printSpy.mockClear();
  window.localStorage.clear();
  window.sessionStorage.clear();
  URL.createObjectURL = vi.fn(() => "blob:mock");
  URL.revokeObjectURL = vi.fn();
});

describe("reload 後の発見（発行）", () => {
  it("発見は issue / edition_id=null で行い、空でも保留なし・key は発行操作の時にだけ作る", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await screen.findByRole("heading", { name: "証明書を発行する" });
    await waitFor(() => expect(callsOf(controls, "queryOperationRecovery").length).toBeGreaterThan(0));
    expect(callsOf(controls, "queryOperationRecovery")[0].detail).toEqual({
      target_ref: { vaccination_id: 101, edition_id: null },
      intent: "issue",
      cursor: null,
    });
    expect(screen.queryByText(HOLD_ISSUE)).toBeNull();
    expect(callsOf(controls, "issue")).toHaveLength(0);

    // 空の発見は新しい発行を妨げない（既存の初回フローは従来どおり動く）。
    await createPreview(user);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    expect(callsOf(controls, "issue")).toHaveLength(1);
  });

  it("完了済みの発行は reload 後に結果を表示でき、保存版 PDF からだけ印刷する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const first = renderScreen(api);
    await createPreview(user);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    first.unmount();

    renderScreen(api);
    const summary = await screen.findByText(/確認できた過去の発行操作（1件）/);
    await user.click(summary);
    expect(screen.queryByText(HOLD_ISSUE)).toBeNull();
    await user.click(await screen.findByRole("button", { name: "結果を表示", hidden: true }));
    const printButton = await screen.findByRole("button", { name: "印刷する", hidden: true });
    expect(callsOf(controls, "getArtifactContent").length).toBeGreaterThan(0);
    await user.click(printButton);
    expect(printSpy).toHaveBeenCalledTimes(1);
    expect(callsOf(controls, "issue")).toHaveLength(1); // reload 後に再送していない
  });

  it("未確定の発行は保留: 新しい発行を止め、同じ key の既存 lookup で再照会し、新しい key を作らない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const first = renderScreen(api);
    await createPreview(user);
    controls.failNext("issue", { error: new CertificateApiError("network"), claimOnly: true });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByText("この操作では発行されていません");
    const sent = callsOf(controls, "issue")[0].detail as CertificateIssueRequest;
    first.unmount();

    renderScreen(api);
    expect(await screen.findByText(HOLD_ISSUE)).toBeInTheDocument();
    expect(screen.getByText("結果未確認")).toBeInTheDocument();
    await createPreview(user);
    const issueButton = screen.getByRole("button", { name: "この内容で発行する" });
    expect(issueButton).toBeDisabled();
    await user.click(issueButton);
    expect(callsOf(controls, "issue")).toHaveLength(1); // 新しい key の送信はない

    const lookupsBefore = callsOf(controls, "lookupOperation").length;
    await user.click(screen.getByRole("button", { name: "この操作の結果を再照会" }));
    expect(await screen.findByText(/現時点では成功を確認できません/)).toBeInTheDocument();
    const added = callsOf(controls, "lookupOperation").slice(lookupsBefore);
    expect(added.map((c) => c.detail)).toEqual([{ operationKey: sent.operation_key, intent: "issue" }]);
    expect(screen.getByText(HOLD_ISSUE)).toBeInTheDocument(); // 成功なしでは保留のまま
    expect(screen.getByRole("button", { name: "この内容で発行する" })).toBeDisabled();
    // 再送・再試行・やり直しの導線は発見した操作にはない（fence / 新 key 再開は Slice B）。
    expect(screen.queryByRole("button", { name: "同じ内容で再試行" })).toBeNull();
    expect(screen.queryByRole("button", { name: "内容を直してやり直す" })).toBeNull();
    expect(window.localStorage.length).toBe(0);
    expect(window.sessionStorage.length).toBe(0);
  });

  it("再照会で成功を確認できた操作だけが保留を外れ、その結果を表示できる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const first = renderScreen(api);
    await createPreview(user);
    controls.failNext("issue", { error: new CertificateApiError("network"), claimOnly: true });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByText("この操作では発行されていません");
    const sent = callsOf(controls, "issue")[0].detail as CertificateIssueRequest;
    first.unmount();

    renderScreen(api);
    await screen.findByText(HOLD_ISSUE);
    await createPreview(user);
    expect(screen.getByRole("button", { name: "この内容で発行する" })).toBeDisabled();

    await api.issue(101, sent); // 遅れて到着した旧送信が成功した状況
    await user.click(screen.getByRole("button", { name: "この操作の結果を再照会" }));
    expect(await screen.findByText(/再照会で完了を確認しました/)).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText(HOLD_ISSUE)).toBeNull());
    await user.click(screen.getByRole("button", { name: "結果を表示" }));
    await screen.findByRole("button", { name: "印刷する" });
  });

  it("lookup 不調の操作(unavailable)は保留。再照会しても成功を確認できない限り外れない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const [key] = controls.recovery.seedUnresolved(1, "issue", null);
    controls.recovery.markUnavailable(key);
    renderScreen(api);
    expect(await screen.findByText(HOLD_ISSUE)).toBeInTheDocument();
    expect(screen.getByText("結果を取得できません")).toBeInTheDocument();

    controls.failNext("lookupOperation", { error: new CertificateApiError("network") });
    await user.click(screen.getByRole("button", { name: "この操作の結果を再照会" }));
    expect(await screen.findByText(/結果を確認できませんでした/)).toBeInTheDocument();
    expect(screen.getByText(HOLD_ISSUE)).toBeInTheDocument();
  });

  it("効果なしで終結済み(closed_without_effect)は参考表示のみで、保留にしない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const [key] = controls.recovery.seedUnresolved(1, "issue", null);
    controls.recovery.closeWithoutEffect(key);
    renderScreen(api);
    await user.click(await screen.findByText(/確認できた過去の発行操作（1件）/));
    expect(screen.getByText(/効果なしで終結済みの操作です/)).toBeInTheDocument();
    expect(screen.queryByText(HOLD_ISSUE)).toBeNull();
    expect(screen.queryByRole("button", { name: "この操作の結果を再照会" })).toBeNull();
  });

  it("51 件以上でもページを辿って全件を確認し、未確定が 2 ページ目にあっても保留する", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const keys = controls.recovery.seedUnresolved(60, "issue", null);
    for (const key of keys.slice(10)) controls.recovery.closeWithoutEffect(key); // 新しい 50 件は終結済み・古い 10 件が未確定
    renderScreen(api);
    expect(await screen.findByText(HOLD_ISSUE)).toBeInTheDocument();
    const queries = callsOf(controls, "queryOperationRecovery");
    expect(queries.length).toBe(2);
    expect((queries[1].detail as { cursor: string | null }).cursor).not.toBeNull();
  });
});

describe("発見に失敗したとき（fail-closed）", () => {
  it("発見できない間は新しい発行を始められず、再確認で取り直せる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    controls.failNext("queryOperationRecovery", {
      error: new CertificateApiError("recovery-unavailable", { httpStatus: 503, code: "RESULT_UNAVAILABLE" }),
    });
    renderScreen(api);
    expect(await screen.findByText("未確定の操作を確認できません")).toBeInTheDocument();
    await createPreview(user);
    expect(screen.getByRole("button", { name: "この内容で発行する" })).toBeDisabled();
    expect(callsOf(controls, "issue")).toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "再確認" }));
    await waitFor(() => expect(screen.queryByText("未確定の操作を確認できません")).toBeNull());
    await waitFor(() => expect(screen.getByRole("button", { name: "この内容で発行する" })).toBeEnabled());
  });

  it("未実装の実 route（契約外の 404）を『対象なし』と読み替えず、発行は止まる", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    const notImplemented: CertificateApi = {
      ...api,
      queryOperationRecovery: async () => {
        throw new CertificateApiError("server", { httpStatus: 404 });
      },
    };
    renderScreen(notImplemented);
    expect(await screen.findByText("未確定の操作を確認できません")).toBeInTheDocument();
  });
});

describe("発見の完了待ちと関門", () => {
  function gatedRecovery(api: CertificateApi) {
    let release!: () => void;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const gated: CertificateApi = {
      ...api,
      queryOperationRecovery: async (request, signal) => {
        await gate;
        return api.queryOperationRecovery(request, signal);
      },
    };
    return { gated, release };
  }

  it("発見の途中で発行を押しても、完了を待ってから進む（空なら発行される）", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const { gated, release } = gatedRecovery(api);
    renderScreen(gated);
    await createPreview(user);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(callsOf(controls, "issue")).toHaveLength(0);
    release();
    await screen.findByRole("heading", { name: "発行しました" });
    expect(callsOf(controls, "issue")).toHaveLength(1);
  });

  it("発見の途中で発行を押しても、完了後に未確定が見つかれば key を作らない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    controls.recovery.seedUnresolved(1, "issue", null);
    const { gated, release } = gatedRecovery(api);
    renderScreen(gated);
    await createPreview(user);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    release();
    expect(await screen.findByText(HOLD_ISSUE)).toBeInTheDocument();
    expect(callsOf(controls, "issue")).toHaveLength(0);
    expect(screen.getByRole("button", { name: "この内容で発行する" })).toBeDisabled();
  });
});

describe("reload 後の発見（再印刷）", () => {
  it("版ごとに reprint / その版の edition_id で発見し、未確定があればその行の再印刷だけを止める", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    controls.recovery.seedUnresolved(1, "reprint", 301);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    expect(await within(row).findByText(HOLD_REPRINT)).toBeInTheDocument();
    expect(callsOf(controls, "queryOperationRecovery").some((c) => {
      const d = c.detail as { intent: string; target_ref: { edition_id: number | null } };
      return d.intent === "reprint" && d.target_ref.edition_id === 301;
    })).toBe(true);

    const reprintButton = within(row).getByRole("button", { name: "再印刷の準備" });
    expect(reprintButton).toBeDisabled();
    await user.click(reprintButton);
    expect(callsOf(controls, "reprint")).toHaveLength(0);
    // 発行側の保留には影響しない（intent ごとに分かれる）。
    expect(screen.queryByText(HOLD_ISSUE)).toBeNull();
  });

  it("発行側の未確定は再印刷を止めず、再印刷側の未確定は発行を止めない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    controls.recovery.seedUnresolved(1, "issue", null);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    expect(await screen.findByText(HOLD_ISSUE)).toBeInTheDocument();
    await waitFor(() => expect(within(row).getByRole("button", { name: "再印刷の準備" })).toBeEnabled());
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    await within(row).findByRole("heading", { name: "再印刷の準備ができました" });
    expect(callsOf(controls, "reprint")).toHaveLength(1);
  });

  it("再印刷の再照会は同じ key の既存 lookup（intent=reprint）で、成功なしでは保留のまま", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    const [key] = controls.recovery.seedUnresolved(1, "reprint", 301);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    await user.click(await within(row).findByRole("button", { name: "この操作の結果を再照会" }));
    expect(await within(row).findByText(/現時点では成功を確認できません/)).toBeInTheDocument();
    expect(callsOf(controls, "lookupOperation").at(-1)?.detail).toEqual({ operationKey: key, intent: "reprint" });
    expect(within(row).getByText(HOLD_REPRINT)).toBeInTheDocument();
  });
});

describe("同じ operation key を既存フローも持つとき（二重の所有者を作らない）", () => {
  it("発行: 既存フローが結果不明として持つ key は、発見側が成功表示を出さず既存の照会で確定する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const view = renderScreen(api);
    await createPreview(user);
    controls.failNext("issue", { error: new CertificateApiError("network"), commitFirst: true });
    controls.failNext("lookupOperation", { error: new CertificateApiError("network") });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText("発行結果を確認できません")).toBeInTheDocument();

    view.rerender(<div />); // 画面を離れる
    view.rerender(<VaccinationCertificateScreen vaccinationId={101} api={api} />); // 同じ QueryClient で戻る
    expect(await screen.findByText("発行結果を確認できません")).toBeInTheDocument();
    await user.click(await screen.findByText(/確認できた過去の発行操作（1件）/));
    expect(await screen.findByText(/この画面で進行中の操作です/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "結果を表示", hidden: true })).toBeNull();

    await user.click(screen.getByRole("button", { name: "発行結果を照会する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    expect(callsOf(controls, "issue")).toHaveLength(1);
  });

  it("再印刷: 既存フローが持つ key は、その行の発見側が成功表示を出さない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    const view = renderScreen(api);
    let row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    controls.failNext("reprint", { error: new CertificateApiError("network"), commitFirst: true });
    controls.failNext("lookupOperation", { error: new CertificateApiError("network") });
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    expect(await within(row).findByText("発行結果を確認できません")).toBeInTheDocument();

    view.rerender(<div />);
    view.rerender(<VaccinationCertificateScreen vaccinationId={101} api={api} />);
    row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    expect(await within(row).findByText("発行結果を確認できません")).toBeInTheDocument();
    await user.click(await within(row).findByText(/確認できた過去の再印刷操作（1件）/));
    expect(await within(row).findByText(/この画面で進行中の操作です/)).toBeInTheDocument();
    expect(within(row).queryByRole("button", { name: "結果を表示", hidden: true })).toBeNull();
    expect(callsOf(controls, "reprint")).toHaveLength(1);
  });
});

describe("自分のフローが保留中の key の成功を確認したとき", () => {
  it("発行: 同じ内容で再試行して成功すると保留が外れ、新しい発行へ進める", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const view = renderScreen(api);
    await createPreview(user);
    controls.failNext("issue", { error: new CertificateApiError("network"), claimOnly: true });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByText("この操作では発行されていません");
    view.rerender(<div />);
    view.rerender(<VaccinationCertificateScreen vaccinationId={101} api={api} />);
    expect(await screen.findByText(HOLD_ISSUE)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "発行結果を照会する" }));
    await screen.findByText("この操作では発行されていません");
    expect(screen.getByText(HOLD_ISSUE)).toBeInTheDocument(); // 成功なしでは保留のまま
    await user.click(screen.getByRole("button", { name: "同じ内容で再試行" }));
    await screen.findByRole("heading", { name: "発行しました" });

    await user.click(screen.getByRole("button", { name: "新しく発行する" }));
    await createPreview(user);
    expect(screen.queryByText(HOLD_ISSUE)).toBeNull();
    expect(screen.queryByText(/この画面で進行中の操作です/)).toBeNull();
    expect(screen.getByRole("button", { name: "この内容で発行する" })).toBeEnabled();
    expect(callsOf(controls, "issue")).toHaveLength(2); // 同じ key の再試行のみ（新 key は未送信）
  });

  it("再印刷: 同じ内容で再試行して成功すると、その行の保留が外れる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    const view = renderScreen(api);
    let row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    controls.failNext("reprint", { error: new CertificateApiError("network"), claimOnly: true });
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    await within(row).findByRole("button", { name: "同じ内容で再試行" });
    view.rerender(<div />);
    view.rerender(<VaccinationCertificateScreen vaccinationId={101} api={api} />);
    row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    expect(await within(row).findByText(HOLD_REPRINT)).toBeInTheDocument();

    await user.click(within(row).getByRole("button", { name: "同じ内容で再試行" }));
    await within(row).findByRole("heading", { name: "再印刷の準備ができました" });
    expect(within(row).queryByText(HOLD_REPRINT)).toBeNull();
    expect(within(row).queryByText(/この画面で進行中の操作です/)).toBeNull();
    expect(within(row).getByRole("button", { name: "再印刷の準備" })).toBeEnabled();
    expect(callsOf(controls, "reprint")).toHaveLength(2);
  });
});
