/**
 * ManualEditor.test.tsx — 既存記事編集フローのユーザ視点テスト（EMR-513）
 *
 * 契約wire adapter（../api/manual-article-edit）をmockし、
 * 編集元取得 → 編集 → 保存（receipt照合） → 409分岐 → unknown照会停止 → 権限喪失
 * の振る舞いを検証する。wire自体の契約は api/manual-article-edit.test.ts（MSW契約mock）で検証する。
 */

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import {
  ManualArticleSaveError,
  ManualEditConflictError,
  ManualReceiptMismatchError,
  type ManualArticleEditView,
  type ManualEditContentVersion,
  type ManualEditReceipt,
} from "../api/manual-article-edit";
import type { ManualArticle } from "@/lib/manual-index";
import { ManualEditor } from "./ManualEditor";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { editViewMock, saveMock, operationsMock } = vi.hoisted(() => ({
  editViewMock: vi.fn(),
  saveMock: vi.fn(),
  operationsMock: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
}));

// adapterの関数だけを差し替え、エラークラス等は実物を使う（instanceof分岐の同一性を保つ）
vi.mock("../api/manual-article-edit", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/manual-article-edit")>();
  return {
    ...actual,
    getManualArticleEditView: editViewMock,
    saveManualArticleEdit: saveMock,
    getManualEditOperationResult: operationsMock,
  };
});

vi.mock("./ManualContent", () => ({
  ManualContent: () => <div>preview</div>,
}));

const ARTICLE_ID = "9223372036854775807";
const CONTENT_VERSION_ID = "4100000000000000123";
const NEW_CONTENT_VERSION_ID = "4100000000000000124";
const SHA_DIGEST = "3f79bb7b435b05321651daefd374cdc681dc06faa65e374e38337b88ca4c6a1c";
const OPERATION_KEY = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee";

const article: ManualArticle = {
  category: "screens",
  slug: "00-overview",
  title: "概要",
  order: 0,
  section: "基本",
  content: "初期本文",
  searchText: "概要 初期本文",
};

function currentFixture(
  overrides: Partial<ManualEditContentVersion> = {},
): ManualEditContentVersion {
  return {
    id: CONTENT_VERSION_ID,
    article_id: ARTICLE_ID,
    content_revision: 7,
    content_digest: SHA_DIGEST,
    source_kind: "edited",
    restored_from_id: null,
    content: {
      title: "概要",
      order_value: 12.34,
      section: "基本",
      body_markdown: "現在の本文",
    },
    edited_by_staff_id: 12,
    edited_at: "2026-10-19T09:00:00Z",
    ...overrides,
  };
}

function editViewFixture(overrides: Partial<ManualArticleEditView> = {}): ManualArticleEditView {
  return {
    id: ARTICLE_ID,
    category: "screens",
    slug: "00-overview",
    version: 41,
    current: currentFixture(),
    published_version_id: "4100000000000000100",
    hidden: false,
    ...overrides,
  };
}

function receiptFixture(overrides: Partial<ManualEditReceipt> = {}): ManualEditReceipt {
  return {
    operation_key: OPERATION_KEY,
    article_id: ARTICLE_ID,
    intent: "edit",
    previous_version: 41,
    resulting_version: 42,
    content_version_id: NEW_CONTENT_VERSION_ID,
    content_digest: SHA_DIGEST,
    published_version_id: "4100000000000000100",
    ...overrides,
  };
}

/** axiosの応答だけを持つ擬似AxiosError（extractApiErrorMessageの分岐用） */
function fakeAxiosError(status: number, body: Record<string, unknown> = {}): unknown {
  return Object.assign(new Error(`Request failed with status code ${status}`), {
    isAxiosError: true,
    response: { status, data: body },
  });
}

/** 編集元の取得完了を待ってから本文を書き換える */
async function typeBodyAfterLoad(text: string): Promise<void> {
  const textarea = await screen.findByLabelText("マニュアル本文編集");
  await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
  await userEvent.clear(textarea);
  await userEvent.type(textarea, text);
}

function renderEditor(props: { canEdit?: boolean; onClose?: () => void } = {}) {
  const onClose = props.onClose ?? vi.fn();
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <ManualEditor article={article} onClose={onClose} canEdit={props.canEdit} />
    </QueryClientProvider>,
  );
  return { onClose };
}

describe("ManualEditor 既存記事編集フロー（EMR-513）", () => {
  beforeEach(() => {
    editViewMock.mockReset();
    saveMock.mockReset();
    operationsMock.mockReset();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.info).mockClear();
  });

  it("編集元のcurrent内容版を表示し、保存成功（receipt照合一致）で閉じる", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    saveMock.mockResolvedValue({
      operationKey: OPERATION_KEY,
      receipt: receiptFixture(),
    });
    const { onClose } = renderEditor({ canEdit: true });

    // バンドル版ではなくcurrent内容版を編集対象にする
    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));

    await userEvent.clear(textarea);
    await userEvent.type(textarea, "更新本文");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => expect(saveMock).toHaveBeenCalled());
    // mutate(variables, options) の2引数呼び出しのため1引数目（保存要求変数）を検証する
    expect(saveMock.mock.calls.at(-1)?.[0]).toEqual(
      expect.objectContaining({
        category: "screens",
        slug: "00-overview",
        articleId: ARTICLE_ID,
        expectedVersion: 41,
        expectedContentVersionId: CONTENT_VERSION_ID,
        content: expect.objectContaining({
          title: "概要",
          order_value: 12.34,
          section: "基本",
          body_markdown: "更新本文",
        }),
      }),
    );
    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("マニュアルを保存しました");
      expect(onClose).toHaveBeenCalled();
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("current=null は架空補完せず元内容確認で保存停止する", async () => {
    editViewMock.mockResolvedValue(editViewFixture({ current: null }));
    renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    // 編集対象が解決するまでは閲覧（readOnly）のみ
    await waitFor(() => expect(textarea).toHaveProperty("readOnly", true));
    expect(textarea).toHaveValue("初期本文");

    expect(await screen.findByRole("alert")).toHaveTextContent("元内容を確認");
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    expect(saveMock).not.toHaveBeenCalled();
  });

  it.each([
    [
      "legacy_unknown",
      editViewFixture({ current: currentFixture({ source_kind: "legacy_unknown" }) }),
    ],
    ["digest欠落", editViewFixture({ current: currentFixture({ content_digest: null }) })],
  ] as const)("%s でも保存を停止する", async (_name, view) => {
    editViewMock.mockResolvedValue(view);
    renderEditor({ canEdit: true });

    expect(await screen.findByRole("alert")).toHaveTextContent("元内容を確認");
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    expect(saveMock).not.toHaveBeenCalled();
  });

  it("409 version_conflict: 入力を保持し最新edit viewを別取得して比較・再確認を促す（自動再送しない）", async () => {
    const freshView = editViewFixture({
      version: 42,
      current: currentFixture({
        id: "4100000000000000125",
        content_revision: 8,
        content: {
          title: "概要",
          order_value: 12.34,
          section: "基本",
          body_markdown: "最新の本文",
        },
      }),
    });
    editViewMock
      .mockResolvedValue(freshView)
      .mockResolvedValueOnce(editViewFixture())
      .mockResolvedValueOnce(freshView);
    saveMock
      .mockRejectedValueOnce(
        new ManualEditConflictError({
          operationKey: OPERATION_KEY,
          conflict: { error: "conflict", code: "version_conflict" },
        }),
      )
      .mockResolvedValueOnce({ operationKey: OPERATION_KEY, receipt: receiptFixture() });
    const { onClose } = renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "競合時の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    // 入力保持・別取得・比較メッセージ（自動再送しないのでsaveMockは1回）
    const notice = await screen.findByRole("alert");
    await waitFor(() => expect(editViewMock).toHaveBeenCalledTimes(2));
    expect(notice).toHaveTextContent("version_conflict");
    expect(notice).toHaveTextContent("自動再送はしません");
    expect(notice).toHaveTextContent("本文が更新されています");
    expect(textarea).toHaveValue("競合時の入力");
    expect(saveMock).toHaveBeenCalledTimes(1);
    expect(onClose).not.toHaveBeenCalled();

    // 再確認後の保存は取得した最新CAS/内容版を使う
    await userEvent.click(screen.getByRole("button", { name: "保存" }));
    await waitFor(() => expect(saveMock).toHaveBeenCalledTimes(2));
    // mutate(variables, options) の2引数呼び出しのため1引数目（保存要求変数）を検証する
    expect(saveMock.mock.calls.at(-1)?.[0]).toEqual(
      expect.objectContaining({
        expectedVersion: 42,
        expectedContentVersionId: "4100000000000000125",
        content: expect.objectContaining({ body_markdown: "競合時の入力" }),
      }),
    );
  });

  it("409 source_unresolved: 入力を保持して保存を停止する", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    saveMock.mockRejectedValue(
      new ManualEditConflictError({
        operationKey: OPERATION_KEY,
        conflict: { error: "unresolved", code: "source_unresolved" },
      }),
    );
    renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "停止確認用の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("source_unresolved");
    expect(textarea).toHaveValue("停止確認用の入力");
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    expect(saveMock).toHaveBeenCalledTimes(1);
  });

  it("通信断→operations unknown: 同一operation_keyで照会し、入力保持で停止する（新keyでの再保存をしない）", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    saveMock.mockRejectedValue(
      new ManualArticleSaveError({
        operationKey: OPERATION_KEY,
        message: "manual article save failed: no response",
      }),
    );
    operationsMock.mockResolvedValue({
      operation_key: OPERATION_KEY,
      state: "unknown",
      receipt: null,
    });
    renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "unknown確認用の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(operationsMock).toHaveBeenCalledWith("screens", "00-overview", OPERATION_KEY);
    });
    expect(await screen.findByRole("alert")).toHaveTextContent("unknown");
    expect(textarea).toHaveValue("unknown確認用の入力");
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    // 新しいoperation_keyでの再保存はしない
    expect(saveMock).toHaveBeenCalledTimes(1);
  });

  it("通信断→operations succeeded: 照会結果もreceipt照合と最新edit view比較を通して成功扱いにする", async () => {
    // 照合成功後の「現在状態を別取得して比較」で、保存後のview（version進行・新内容版）を返す
    editViewMock.mockResolvedValueOnce(editViewFixture()).mockResolvedValueOnce(
      editViewFixture({
        version: 42,
        current: currentFixture({ id: NEW_CONTENT_VERSION_ID, content_revision: 8 }),
      }),
    );
    saveMock.mockRejectedValue(
      new ManualArticleSaveError({
        operationKey: OPERATION_KEY,
        message: "manual article save failed: no response",
      }),
    );
    operationsMock.mockResolvedValue({
      operation_key: OPERATION_KEY,
      state: "succeeded",
      receipt: receiptFixture(),
    });
    const { onClose } = renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "照会成功の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(operationsMock).toHaveBeenCalledWith("screens", "00-overview", OPERATION_KEY);
      // 初回取得 + 照合後の「現在状態を別取得して比較」（成功後のinvalidate再取得は対象外）
      expect(editViewMock.mock.calls.length).toBeGreaterThanOrEqual(2);
      expect(toast.success).toHaveBeenCalledWith("マニュアルを保存しました");
      expect(onClose).toHaveBeenCalled();
    });
  });

  it("operations succeededでも最新edit viewが不一致なら、確定した成功を伝え追加情報を出す（成功扱い）", async () => {
    // receipt照合は通過＝保存確定。2回目の取得も古いviewのまま（その後さらに編集が進んだ等）でも
    // 確定した成功を否定せず、追加情報（toast.info）を出して閉じる
    editViewMock.mockResolvedValue(editViewFixture());
    saveMock.mockRejectedValue(
      new ManualArticleSaveError({
        operationKey: OPERATION_KEY,
        message: "manual article save failed: no response",
      }),
    );
    operationsMock.mockResolvedValue({
      operation_key: OPERATION_KEY,
      state: "succeeded",
      receipt: receiptFixture(),
    });
    const { onClose } = renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "古い照会結果の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("マニュアルを保存しました");
      expect(toast.info).toHaveBeenCalledWith(
        expect.stringContaining("ほかの操作でこの記事が更新されています"),
      );
      expect(onClose).toHaveBeenCalled();
    });
  });

  it("unknown停止後に同じoperation_keyで再照会できる（新key保存はしない）", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    saveMock.mockRejectedValue(
      new ManualArticleSaveError({
        operationKey: OPERATION_KEY,
        message: "manual article save failed: no response",
      }),
    );
    operationsMock
      .mockResolvedValueOnce({ operation_key: OPERATION_KEY, state: "unknown", receipt: null })
      .mockResolvedValueOnce({
        operation_key: OPERATION_KEY,
        state: "succeeded",
        receipt: receiptFixture(),
      });
    const { onClose } = renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "再照会の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    // 1回目の照会はunknownで停止。入力保持・保存停止
    const alert = await screen.findByRole("alert");
    await waitFor(() => expect(alert).toHaveTextContent("unknown"));
    expect(saveMock).toHaveBeenCalledTimes(1);
    expect(onClose).not.toHaveBeenCalled();

    // 再照会ボタンで同じkeyをもう一度問い合わせる（照会中・停止中もtextareaは凍結）
    expect(textarea).toHaveAttribute("readonly");
    await userEvent.click(screen.getByRole("button", { name: "保存結果を再照会" }));
    expect(operationsMock).toHaveBeenCalledTimes(2);
    expect(operationsMock).toHaveBeenLastCalledWith("screens", "00-overview", OPERATION_KEY);
  });

  it("ネットワーク再接続でも編集元を背景refetchしない（保存基準を無言で差し替えない）", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    renderEditor({ canEdit: true });

    await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(editViewMock).toHaveBeenCalledTimes(1));

    // offline→online を疑似的に発生させる（TanStackのreconnect契約）
    window.dispatchEvent(new Event("offline"));
    window.dispatchEvent(new Event("online"));
    await new Promise((resolve) => setTimeout(resolve, 50));

    expect(editViewMock).toHaveBeenCalledTimes(1);
  });

  it("409後に最新edit viewが未解決でも入力を保持し、閉じる際に離脱確認が出る", async () => {
    // 競合後に取得した最新viewがcurrent=nullでも、draftは「未編集」扱いにしない（codex review C4）
    editViewMock
      .mockResolvedValueOnce(editViewFixture())
      .mockResolvedValueOnce(editViewFixture({ current: null }));
    saveMock.mockRejectedValue(
      new ManualEditConflictError({
        operationKey: OPERATION_KEY,
        conflict: { error: "", code: "version_conflict" },
      }),
    );
    const { onClose } = renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "競合後未解決の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    // 競合通知（action notice）と未解決source通知の両方がalertとして出る
    const alerts = await screen.findAllByRole("alert");
    await waitFor(() => {
      const text = alerts.map((a) => a.textContent).join(" ");
      expect(text).toContain("version_conflict");
      expect(text).toContain("未解決");
    });
    expect(onClose).not.toHaveBeenCalled();

    // 入力があるので閉じる操作では離脱確認が出る（静かな破棄をしない）
    await userEvent.click(screen.getByRole("button", { name: "編集を終了" }));
    expect(screen.getByText("編集内容が保存されていません")).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("保存中はtextareaを編集不可にする（送信後の編集消失を防ぐ）", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    let resolveSave: (value: {
      operationKey: string;
      receipt: ReturnType<typeof receiptFixture>;
    }) => void = () => {};
    saveMock.mockReturnValue(
      new Promise((resolve) => {
        resolveSave = resolve;
      }),
    );
    const { onClose } = renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "保存中の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    // 保存が進行中の間は追記できない（送信済み内容との乖離と消失を防ぐ）
    await waitFor(() => expect(textarea).toHaveAttribute("readonly"));
    expect(onClose).not.toHaveBeenCalled();

    resolveSave({ operationKey: OPERATION_KEY, receipt: receiptFixture() });
    await waitFor(() => expect(toast.success).toHaveBeenCalledWith("マニュアルを保存しました"));
    expect(onClose).toHaveBeenCalled();
  });

  it("receipt不一致は成功表示せず停止する", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    saveMock.mockRejectedValue(new ManualReceiptMismatchError("article_id mismatch"));
    const { onClose } = renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "照合失敗の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("照合に失敗");
    expect(toast.success).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
  });

  it("403 は権限メッセージを表示し、入力を保持して自動再試行しない", async () => {
    editViewMock.mockResolvedValue(editViewFixture());
    saveMock.mockRejectedValue(
      new ManualArticleSaveError({
        operationKey: OPERATION_KEY,
        status: 403,
        message: "manual article save failed: 403",
        requestError: fakeAxiosError(403),
      }),
    );
    renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "権限確認用の入力");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("権限がありません");
    expect(textarea).toHaveValue("権限確認用の入力");
    expect(saveMock).toHaveBeenCalledTimes(1);
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("title空文字はクライアントでも保存しない（trim・文字数上限は追加しない）", async () => {
    editViewMock.mockResolvedValue(
      editViewFixture({
        current: currentFixture({
          content: {
            title: "",
            order_value: 12.34,
            section: "基本",
            body_markdown: "現在の本文",
          },
        }),
      }),
    );
    renderEditor({ canEdit: true });

    const textarea = await screen.findByLabelText("マニュアル本文編集");
    await waitFor(() => expect(textarea).toHaveValue("現在の本文"));
    await userEvent.clear(textarea);
    await userEvent.type(textarea, "空titleの保存");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "タイトルとセクションが空のため保存できません。元内容を確認してください。",
      );
    });
    expect(saveMock).not.toHaveBeenCalled();
  });

  describe("保存権限の再確認（fail-closed）", () => {
    it("canEdit=true なら dirty 保存で保存要求を呼ぶ", async () => {
      editViewMock.mockResolvedValue(editViewFixture());
      saveMock.mockResolvedValue({ operationKey: OPERATION_KEY, receipt: receiptFixture() });
      const { onClose } = renderEditor({ canEdit: true });

      await typeBodyAfterLoad("更新本文");
      await userEvent.click(screen.getByRole("button", { name: "保存" }));

      await waitFor(() => expect(saveMock).toHaveBeenCalled());
      expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
      expect(onClose).toHaveBeenCalled();
    });

    it("canEdit=false では保存要求を呼ばず toast.error する", async () => {
      editViewMock.mockResolvedValue(editViewFixture());
      const { onClose } = renderEditor({ canEdit: false });

      await typeBodyAfterLoad("権限失効後の本文");
      await userEvent.click(screen.getByRole("button", { name: "保存" }));

      await waitFor(() => {
        expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
      });
      expect(saveMock).not.toHaveBeenCalled();
      expect(onClose).not.toHaveBeenCalled();
    });

    it("canEdit 省略時は fail-closed で保存要求を呼ばず toast.error する", async () => {
      editViewMock.mockResolvedValue(editViewFixture());
      renderEditor();

      await typeBodyAfterLoad("省略時本文");
      await userEvent.click(screen.getByRole("button", { name: "保存" }));

      await waitFor(() => {
        expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
      });
      expect(saveMock).not.toHaveBeenCalled();
    });
  });
});
