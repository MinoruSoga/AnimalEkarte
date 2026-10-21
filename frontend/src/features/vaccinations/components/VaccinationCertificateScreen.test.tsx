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
 * 契約モック(createMockCertificateApi)を明示注入して画面の状態を確認する。
 * 実 API・実 BE・実 renderer の検証ではない。PDF は合成の空白 1 ページ。
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
  const body: CertificateIssueRequest = {
    ...request,
    operation_key: "99999999-9999-4999-8999-999999999999",
    preview_fingerprint: preview.preview_fingerprint,
    preview_display_issue_date: preview.snapshot.display_issue_date,
  };
  await api.issue(101, body);
}

beforeEach(() => {
  permission.canEdit = true;
  printSpy.mockClear();
  URL.createObjectURL = vi.fn(() => "blob:mock");
  URL.revokeObjectURL = vi.fn();
});

describe("様式の選択（eligibility）", () => {
  it("狂犬病の対象外・未確認のワクチンでは通常様式だけを選べる", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await screen.findByRole("heading", { name: "証明書を発行する" });
    expect(screen.getByRole("radio", { name: "通常の予防接種証明書" })).toBeInTheDocument();
    expect(screen.queryByRole("radio", { name: "狂犬病予防注射済証" })).toBeNull();
    expect(screen.getByText(/狂犬病予防注射済証の対象として登録されていない/)).toBeInTheDocument();
  });

  it("対象のワクチンでは狂犬病様式も選べ、選ぶと 10 欄の入力が出る", async () => {
    const user = userEvent.setup();
    const { api } = createMockCertificateApi({ rabiesEligible: true });
    renderScreen(api);
    await user.click(await screen.findByRole("radio", { name: "狂犬病予防注射済証" }));
    expect(screen.getByLabelText("証明番号")).toBeInTheDocument();
    expect(screen.getByLabelText("注射済票費用")).toHaveValue(null);
    expect(screen.getByLabelText("合計")).toHaveValue(null);
  });

  it("通常様式は縦横を選べ、向きを変えるとプレビューを作り直す", async () => {
    const user = userEvent.setup();
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    await user.click(screen.getByRole("radio", { name: "通常（横）" }));
    expect(screen.queryByTitle("証明書の確認プレビュー")).toBeNull();
    expect(screen.queryByRole("button", { name: "この内容で発行する" })).toBeNull();
  });
});

describe("プレビューと発行", () => {
  it("プレビューに印刷ボタンは無く、発行後の保存版 PDF からだけ印刷できる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    expect(screen.queryByRole("button", { name: /印刷/ })).toBeNull();

    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    const printButton = await screen.findByRole("button", { name: "印刷する" });
    expect(callsOf(controls, "getEdition").length).toBeGreaterThan(0);
    expect(callsOf(controls, "getArtifactContent")).toHaveLength(1);
    await user.click(printButton);
    expect(printSpy).toHaveBeenCalledTimes(1);
  });

  it("狂犬病の料金は null と明示 0 を区別して送り、合計を自動計算しない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: true });
    renderScreen(api);
    await user.click(await screen.findByRole("radio", { name: "狂犬病予防注射済証" }));
    await user.type(screen.getByLabelText("注射料"), "3000");
    await user.type(screen.getByLabelText("注射済票費用"), "0");
    await user.type(screen.getByLabelText("合計"), "9999");
    expect(screen.getByLabelText("合計")).toHaveValue(9999); // 3000 + 0 に補正されない
    await createPreview(user);
    const sent = (callsOf(controls, "previewForIssue")[0].detail as { inputs: Record<string, unknown> }).inputs;
    expect(sent).toMatchObject({ registration_fee: null, injection_fee: 3000, tag_fee: 0, total: 9999 });
    expect(Object.keys(sent)).toHaveLength(10);
  });

  it("入力を変えるとプレビューと発行ボタンが消え、再プレビューが必要になる", async () => {
    const user = userEvent.setup();
    const { api } = createMockCertificateApi({ rabiesEligible: true });
    renderScreen(api);
    await user.click(await screen.findByRole("radio", { name: "狂犬病予防注射済証" }));
    await createPreview(user);
    await user.type(screen.getByLabelText("合計"), "1");
    expect(screen.queryByTitle("証明書の確認プレビュー")).toBeNull();
    expect(screen.queryByRole("button", { name: "この内容で発行する" })).toBeNull();
    expect(screen.getByRole("button", { name: "プレビューを作成" })).toBeInTheDocument();
  });

  it("整数でない料金はプレビュー前に欄の直下でエラーにする", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: true });
    renderScreen(api);
    await user.click(await screen.findByRole("radio", { name: "狂犬病予防注射済証" }));
    await user.type(screen.getByLabelText("注射料"), "1.5");
    await user.click(screen.getByRole("button", { name: "プレビューを作成" }));
    expect(await screen.findByText("整数（円）で入力してください")).toBeInTheDocument();
    expect(callsOf(controls, "previewForIssue")).toHaveLength(0);
  });
});

describe("結果不明・再送・前提変更", () => {
  it("応答が失われても同じ operation_key で照会し、成功なら発行済みとして扱う（再発行しない）", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    controls.failNext("issue", {
      error: new CertificateApiError("result-unavailable", { httpStatus: 503, code: "result-unavailable" }),
      commitFirst: true,
    });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    const issueCalls = callsOf(controls, "issue");
    const lookups = callsOf(controls, "lookupOperation");
    expect(issueCalls).toHaveLength(1);
    expect((lookups[0].detail as { operationKey: string }).operationKey).toBe(
      (issueCalls[0].detail as CertificateIssueRequest).operation_key,
    );
    expect(await screen.findByRole("button", { name: "印刷する" })).toBeInTheDocument();
  });

  it("処理中(operation-in-progress)は別発行を始めず、照会→未受付なら同じ key・同じ内容で再試行する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    controls.failNext("issue", {
      error: new CertificateApiError("operation-in-progress", { httpStatus: 409, code: "operation-in-progress" }),
    });
    controls.failNext("lookupOperation", {
      error: new CertificateApiError("operation-in-progress", { httpStatus: 409, code: "operation-in-progress" }),
    });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText("同じ発行の処理中です")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "この内容で発行する" })).toBeNull();

    await user.click(screen.getByRole("button", { name: "発行結果を照会する" }));
    expect(await screen.findByText("この操作は受け付けられていません")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "同じ内容で再試行" }));
    await screen.findByRole("heading", { name: "発行しました" });

    const issueCalls = callsOf(controls, "issue").map((c) => c.detail as CertificateIssueRequest);
    expect(issueCalls).toHaveLength(2);
    expect(issueCalls[1]).toEqual(issueCalls[0]);
  });

  it("結果不明の間は入力を編集できない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: true });
    renderScreen(api);
    await user.click(await screen.findByRole("radio", { name: "狂犬病予防注射済証" }));
    await createPreview(user);
    controls.failNext("issue", { error: new CertificateApiError("network") });
    controls.failNext("lookupOperation", { error: new CertificateApiError("network") });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText("発行結果を確認できません")).toBeInTheDocument();
    expect(screen.getByLabelText("証明番号")).toBeDisabled();
  });

  it("日跨ぎなどで前提が変わったら旧プレビューを捨て、再プレビュー後は新しい operation_key で発行する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    controls.setToday("2026-10-21");
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText("証明書の元情報が変わりました")).toBeInTheDocument();
    expect(screen.queryByTitle("証明書の確認プレビュー")).toBeNull();

    await createPreview(user);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    const keys = callsOf(controls, "issue").map((c) => (c.detail as CertificateIssueRequest).operation_key);
    expect(keys).toHaveLength(2);
    expect(keys[0]).not.toBe(keys[1]);
  });

  it("他端末の更新(version-conflict)では context を取り直してプレビューからやり直させる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    controls.bumpSourceVersion();
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText("他の端末で更新されました")).toBeInTheDocument();
    await waitFor(() => expect(callsOf(controls, "getContext").length).toBeGreaterThan(1));
  });

  it("発行直前に edit 権限が失われていたら API を呼ばず中止する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const view = renderScreen(api);
    await createPreview(user);
    permission.canEdit = false;
    view.rerender(<VaccinationCertificateScreen vaccinationId={101} api={api} />);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText("この操作を行う権限がありません")).toBeInTheDocument();
    expect(callsOf(controls, "issue")).toHaveLength(0);
  });

  it("context の can_issue が false なら閲覧のみで発行フォームを出さない", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: false, canIssue: false });
    renderScreen(api);
    expect(await screen.findByText(/発行する権限がありません/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "プレビューを作成" })).toBeNull();
  });
});

describe("PDF 取得（/content）", () => {
  it("取得に失敗しても別の PDF へ切り替えず、再取得できる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    controls.failNext("getArtifactContent", {
      error: new CertificateApiError("render-unavailable", { httpStatus: 503, code: "artifact_unavailable" }),
    });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText(/PDF を取得できません（保存された正本/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "印刷する" })).toBeNull();
    await user.click(screen.getByRole("button", { name: "PDF を再取得" }));
    expect(await screen.findByRole("button", { name: "印刷する" })).toBeInTheDocument();
  });
});

describe("履歴・保存版・再印刷", () => {
  it("現在有効な版は保存版の確認と再印刷ができ、再印刷も /content の PDF から印刷する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    await user.click(within(row).getByRole("button", { name: "保存版を確認" }));
    expect(await within(row).findByTitle("第1版の保存プレビュー")).toBeInTheDocument();
    expect(within(row).queryByRole("button", { name: "印刷する" })).toBeNull();

    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    await within(row).findByRole("heading", { name: "再印刷の準備ができました" });
    await user.click(await within(row).findByRole("button", { name: "印刷する" }));
    expect(printSpy).toHaveBeenCalledTimes(1);
    expect(callsOf(controls, "reprint")).toHaveLength(1);
  });

  it("再印刷の結果が不明の間は再実行も向き変更もできず、同じ key で照会する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    controls.failNext("reprint", { error: new CertificateApiError("network"), commitFirst: true });
    controls.failNext("lookupOperation", { error: new CertificateApiError("network") });
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    expect(await within(row).findByText("発行結果を確認できません")).toBeInTheDocument();
    expect(within(row).getByRole("button", { name: "再印刷の準備" })).toBeDisabled();
    expect(within(row).getByRole("combobox")).toBeDisabled();

    await user.click(within(row).getByRole("button", { name: "結果を照会する" }));
    await within(row).findByRole("heading", { name: "再印刷の準備ができました" });
    const reprintKey = (callsOf(controls, "reprint")[0].detail as { body: { operation_key: string } }).body.operation_key;
    const lookupKeys = callsOf(controls, "lookupOperation").map((c) => (c.detail as { operationKey: string }).operationKey);
    expect(callsOf(controls, "reprint")).toHaveLength(1);
    expect(lookupKeys.every((k) => k === reprintKey)).toBe(true);
  });

  it("発行済みの画面から「新しく発行する」で編集に戻り、プレビューと key を持ち越さない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    await user.click(screen.getByRole("button", { name: "新しく発行する" }));
    expect(await screen.findByRole("heading", { name: "証明書を発行する" })).toBeInTheDocument();
    expect(screen.queryByTitle("証明書の確認プレビュー")).toBeNull();
    expect(screen.queryByRole("button", { name: "この内容で発行する" })).toBeNull();
    await waitFor(() => expect(screen.getByRole("button", { name: "プレビューを作成" })).toBeInTheDocument());
    expect(callsOf(controls, "issue")).toHaveLength(1);
  });

  it("旧版(接種訂正後)は履歴参照のみで、再印刷・保存版確認の導線を出さない", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    controls.bumpSourceVersion();
    renderScreen(api);
    await screen.findByText("旧版のため印刷できません。履歴としての参照のみです。");
    expect(screen.getByText("接種が訂正済み（旧版）")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再印刷の準備" })).toBeNull();
    expect(screen.queryByRole("button", { name: "保存版を確認" })).toBeNull();
  });

  it("履歴が空のときは何が無いか・次の操作を示す", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    expect(await screen.findByText(/発行履歴はありません。上のフォームから発行すると/)).toBeInTheDocument();
  });
});

describe("再印刷の安全側の挙動", () => {
  it("再印刷の直前に edit 権限が失われていたら API を呼ばず中止する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    const view = renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    permission.canEdit = false;
    view.rerender(<VaccinationCertificateScreen vaccinationId={101} api={api} />);
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    expect(await within(row).findByText("この操作を行う権限がありません")).toBeInTheDocument();
    expect(callsOf(controls, "reprint")).toHaveLength(0);
  });

  it("context の can_issue が false なら再印刷の導線を無効にする", async () => {
    // seed は発行を伴うため、発行可能な状態で版を作ってから can_issue を落とす。
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    controls.setCanIssue(false);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    expect(within(row).getByRole("button", { name: "再印刷の準備" })).toBeDisabled();
  });

  it("「成功なし」が確定した再印刷は職員が破棄でき、向き変更と新しい key での再印刷へ進める", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    controls.failNext("reprint", { error: new CertificateApiError("network"), claimOnly: true });
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    await user.click(await within(row).findByRole("button", { name: "この再印刷をやめる" }));
    expect(within(row).getByRole("combobox")).toBeEnabled();
    expect(within(row).getByRole("button", { name: "再印刷の準備" })).toBeEnabled();

    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    await within(row).findByRole("heading", { name: "再印刷の準備ができました" });
    const keys = callsOf(controls, "reprint").map((c) => (c.detail as { body: { operation_key: string } }).body.operation_key);
    expect(keys).toHaveLength(2);
    expect(keys[0]).not.toBe(keys[1]);
  });

  it("結果不明の再印刷は履歴の様式タブを切り替えて行が作り直されても同じ key で照会する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: true });
    await seedNormalEdition(api);
    renderScreen(api);
    let row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    controls.failNext("reprint", { error: new CertificateApiError("network"), commitFirst: true });
    controls.failNext("lookupOperation", { error: new CertificateApiError("network") });
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    expect(await within(row).findByText("発行結果を確認できません")).toBeInTheDocument();

    const history = screen.getByRole("region", { name: "発行履歴" });
    await user.click(within(history).getByRole("button", { name: "狂犬病予防注射済証" }));
    await waitFor(() => expect(screen.queryByText("第1版")).toBeNull());
    await user.click(within(history).getByRole("button", { name: "通常の予防接種証明書" }));
    row = (await screen.findByText("第1版")).closest("li") as HTMLElement;

    // 実行した向き(縦)が、作り直された行でも表示される。
    expect(within(row).getByRole("combobox")).toHaveTextContent("縦");
    // 作り直された行でも、新しい key を作れず（再印刷は無効）、同じ key で照会できる。
    expect(within(row).getByRole("button", { name: "再印刷の準備" })).toBeDisabled();
    await user.click(within(row).getByRole("button", { name: "結果を照会する" }));
    await within(row).findByRole("heading", { name: "再印刷の準備ができました" });
    const reprintKey = (callsOf(controls, "reprint")[0].detail as { body: { operation_key: string } }).body.operation_key;
    const lookupKeys = callsOf(controls, "lookupOperation").map((c) => (c.detail as { operationKey: string }).operationKey);
    expect(callsOf(controls, "reprint")).toHaveLength(1);
    expect(lookupKeys.every((k) => k === reprintKey)).toBe(true);
  });
});

describe("再印刷の向きの表示", () => {
  it("行が作り直されても、保存版の確認は選択表示と同じ向き(再印刷で使った向き)で取得する", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: true });
    await seedNormalEdition(api);
    renderScreen(api);
    let row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    await user.click(within(row).getByRole("combobox"));
    await user.click(await screen.findByRole("option", { name: "通常（横）" }));
    controls.failNext("reprint", { error: new CertificateApiError("network"), commitFirst: true });
    controls.failNext("lookupOperation", { error: new CertificateApiError("network") });
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    expect(await within(row).findByText("発行結果を確認できません")).toBeInTheDocument();

    const history = screen.getByRole("region", { name: "発行履歴" });
    await user.click(within(history).getByRole("button", { name: "狂犬病予防注射済証" }));
    await waitFor(() => expect(screen.queryByText("第1版")).toBeNull());
    await user.click(within(history).getByRole("button", { name: "通常の予防接種証明書" }));
    row = (await screen.findByText("第1版")).closest("li") as HTMLElement;

    expect(within(row).getByRole("combobox")).toHaveTextContent("横");
    await user.click(within(row).getByRole("button", { name: "保存版を確認" }));
    await within(row).findByTitle("第1版の保存プレビュー");
    expect(callsOf(controls, "previewSaved").at(-1)?.detail).toMatchObject({ layout: "landscape" });
  });
});

describe("入力検証と背景再取得", () => {
  it("不正な料金があるときは最初の不正欄へフォーカスを移す", async () => {
    const user = userEvent.setup();
    const { api } = createMockCertificateApi({ rabiesEligible: true });
    renderScreen(api);
    await user.click(await screen.findByRole("radio", { name: "狂犬病予防注射済証" }));
    await user.type(screen.getByLabelText("注射料"), "1.5");
    await user.type(screen.getByLabelText("合計"), "2.5");
    await user.click(screen.getByRole("button", { name: "プレビューを作成" }));
    await waitFor(() => expect(screen.getByLabelText("注射料")).toHaveFocus());
  });

  it("発行後の context 再取得が失敗しても発行済みの画面を残し、更新失敗を知らせる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await createPreview(user);
    controls.failNext("getContext", { error: new CertificateApiError("result-unavailable", { httpStatus: 503 }) });
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByRole("heading", { name: "発行しました" })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: "再読み込み" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "発行しました" })).toBeInTheDocument();
  });
});

describe("context 読込の失敗", () => {
  it("403 は権限なしを示し再読み込みを出さない", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    controls.failNext("getContext", { error: new CertificateApiError("forbidden", { httpStatus: 403, code: "not-authorized" }) });
    renderScreen(api);
    expect(await screen.findByText(/閲覧する権限がありません/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "再読み込み" })).toBeNull();
  });

  it("503 は再読み込みを出し、モックの成功データへは切り替えない", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    controls.failNext("getContext", { error: new CertificateApiError("result-unavailable", { httpStatus: 503 }) });
    renderScreen(api);
    expect(await screen.findByText(/証明書の情報を取得できません/)).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "証明書を発行する" })).toBeNull();
    expect(screen.getByRole("button", { name: "再読み込み" })).toBeInTheDocument();
  });

  it("404 は接種記録が無いことを示す", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    controls.failNext("getContext", { error: new CertificateApiError("not-found", { httpStatus: 404, code: "not_found" }) });
    renderScreen(api);
    expect(await screen.findByText(/接種記録が見つかりません/)).toBeInTheDocument();
  });

  it("409 は接種の訂正・削除を示す", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    controls.failNext("getContext", { error: new CertificateApiError("dependency-changed", { httpStatus: 409, code: "dependency-changed" }) });
    renderScreen(api);
    expect(await screen.findByText(/訂正または削除されたため/)).toBeInTheDocument();
  });
});

describe("応答の対象束縛（形が正しいだけの別対象を採用しない）", () => {
  const OTHER_KEY = "22222222-2222-4222-8222-222222222222";

  it("別の接種の preview は表示せず、発行ボタンも出さない", async () => {
    const user = userEvent.setup();
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    const mixedUp: CertificateApi = {
      ...api,
      previewForIssue: async (vaccinationId, body, signal) => {
        const preview = await api.previewForIssue(vaccinationId, body, signal);
        return { ...preview, snapshot: { ...preview.snapshot, source: { ...preview.snapshot.source, vaccination_id: 999 } } };
      },
    };
    renderScreen(mixedUp);
    await user.click(await screen.findByRole("button", { name: "プレビューを作成" }));
    expect(await screen.findByText("プレビューを作成できませんでした")).toBeInTheDocument();
    expect(screen.queryByTitle("証明書の確認プレビュー")).toBeNull();
    expect(screen.queryByRole("button", { name: "この内容で発行する" })).toBeNull();
  });

  it("別の操作の発行結果は印刷対象にせず、結果不明として同じ key で照会できる", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const mixedUp: CertificateApi = {
      ...api,
      issue: async (vaccinationId, body) => ({ ...(await api.issue(vaccinationId, body)), operation_key: OTHER_KEY }),
    };
    renderScreen(mixedUp);
    await createPreview(user);
    await user.click(screen.getByRole("button", { name: "この内容で発行する" }));
    expect(await screen.findByText("発行結果の参照が不足しています")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "発行しました" })).toBeNull();
    expect(callsOf(controls, "getArtifactContent")).toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "発行結果を照会する" }));
    await screen.findByRole("heading", { name: "発行しました" });
    expect(callsOf(controls, "issue")).toHaveLength(1);
  });

  it("別の版を指す再印刷結果は印刷対象にしない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    const mixedUp: CertificateApi = {
      ...api,
      reprint: async (editionId, body) => {
        const result = await api.reprint(editionId, body);
        return result.status === "not-succeeded"
          ? result
          : { ...result, result_refs: { ...result.result_refs, edition_id: "999" } };
      },
    };
    renderScreen(mixedUp);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    await user.click(within(row).getByRole("button", { name: "再印刷の準備" }));
    expect(await within(row).findByText("発行結果の参照が不足しています")).toBeInTheDocument();
    expect(within(row).queryByRole("heading", { name: "再印刷の準備ができました" })).toBeNull();
    expect(callsOf(controls, "getArtifactContent")).toHaveLength(0);
  });
});
