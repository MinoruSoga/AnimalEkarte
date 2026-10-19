import { act, useState, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { ManualArticle } from "@/lib/manual-index";

import { HelpChat } from "./HelpChat";
import { HelpChatHistoryProvider } from "./HelpChatHistoryProvider";
import type { SendSupportChatParams } from "../api/send-support-chat";
import type { SupportChatHistoryRecord, SupportChatResponse } from "../types";
import { useHelpChatHistory } from "../lib/help-chat-history";
import { computeManualRevision } from "../lib/manual-corpus";

interface HistoryQueryState {
  data: SupportChatHistoryRecord[];
  isSuccess: boolean;
  isError: boolean;
  isPending: boolean;
}

const EMPTY_HISTORY: HistoryQueryState = {
  data: [],
  isSuccess: true,
  isError: false,
  isPending: false,
};

const { mutateMock, clearMutateMock, historyState, pendingSends } = vi.hoisted(() => ({
  mutateMock: vi.fn(),
  clearMutateMock: vi.fn(),
  pendingSends: [] as Array<{
    resolve: (r: unknown) => void;
    reject: (e: unknown) => void;
  }>,
  historyState: {
    current: {
      data: [] as SupportChatHistoryRecord[],
      isSuccess: true,
      isError: false,
      isPending: false,
    },
  },
}));

const { hashControl } = vi.hoisted(() => ({
  hashControl: { pending: null as Promise<string> | null },
}));

vi.mock("../lib/manual-corpus", async (importOriginal) => {
  const original = await importOriginal<typeof import("../lib/manual-corpus")>();
  return {
    ...original,
    computeManualRevision: vi.fn((input: ManualArticle[]) => {
      const pending = hashControl.pending;
      if (pending !== null) {
        hashControl.pending = null;
        return pending;
      }
      return original.computeManualRevision(input);
    }),
  };
});

vi.mock("../api/send-support-chat", () => ({
  useSendSupportChat: () => ({
    mutate: mutateMock,
    mutateAsync: mutateMock,
    isPending: false,
  }),
}));

vi.mock("../api/get-support-chat-history", () => ({
  useGetSupportChatHistory: () => historyState.current,
}));

vi.mock("../api/clear-support-chat-history", () => ({
  useClearSupportChatHistory: () => ({
    mutate: clearMutateMock,
    isPending: false,
  }),
}));

const articles: ManualArticle[] = [
  {
    category: "screens",
    slug: "accounting",
    title: "画面別 会計",
    order: 1,
    section: "会計",
    content: "会計画面では会計処理を行います。レジ締めは締めボタンから実行します。",
    searchText: "画面別 会計 会計 レジ締め 会計画面では会計処理を行います 締めボタン",
  },
  {
    category: "screens",
    slug: "reception",
    title: "画面別 受付",
    order: 2,
    section: "受付",
    content: "受付画面では予約と診察の流れを管理します。",
    searchText: "画面別 受付 受付 受付画面では",
  },
];

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    );
  };
}

interface PersistentPanelProps {
  onRevisionConflict?: () => Promise<ManualArticle[]>;
}

function ClearProviderHistoryButton() {
  const { setTurns } = useHelpChatHistory();
  return (
    <button type="button" onClick={() => setTurns([])}>
      テスト用に履歴をクリア
    </button>
  );
}

function HistoryOrderObserver() {
  const { turns } = useHelpChatHistory();
  return (
    <ol aria-label="テスト用ターン順">
      {turns.map((turn, index) => (
        <li key={index}>
          {turn.role}:{turn.content}
        </li>
      ))}
    </ol>
  );
}

function PersistentPanelHarness({ onRevisionConflict }: PersistentPanelProps) {
  const [open, setOpen] = useState(true);
  return (
    <>
      <button type="button" onClick={() => setOpen((value) => !value)}>
        {open ? "チャットを閉じる" : "チャットを開く"}
      </button>
      <ClearProviderHistoryButton />
      {open ? (
        <HelpChat articles={articles} onClose={() => {}} onRevisionConflict={onRevisionConflict} />
      ) : null}
    </>
  );
}

function ProviderLifecycleHarness() {
  const [providerMounted, setProviderMounted] = useState(true);
  return (
    <>
      <button type="button" onClick={() => setProviderMounted((value) => !value)}>
        {providerMounted ? "履歴 provider を終了" : "履歴 provider を再開"}
      </button>
      {providerMounted ? (
        <HelpChatHistoryProvider>
          <PersistentPanelHarness />
        </HelpChatHistoryProvider>
      ) : null}
    </>
  );
}

type MutateOptions = {
  onSuccess?: (res: SupportChatResponse) => void;
  onError?: (err: unknown) => void;
};

type SendSupportChatVariables = SendSupportChatParams & { signal?: AbortSignal };

// mutateAsync 化に伴い、call 番目の pending send を解決/拒否するヘルパー。
function callOptions(callIndex = 0): MutateOptions {
  const send = pendingSends[callIndex];
  if (!send) throw new Error(`no pending send at index ${callIndex}`);
  return {
    onSuccess: (res) => send.resolve(res),
    onError: (err) => send.reject(err),
  };
}

type ClearOptions = {
  onSuccess?: () => void;
  onError?: (err: unknown) => void;
};

function clearCallOptions(callIndex = 0): ClearOptions {
  const [, options] = clearMutateMock.mock.calls[callIndex] as [undefined, ClearOptions];
  return options;
}

afterEach(async () => {
  // 前テストの未解決送信を片付け、global transition を残さない。
  await act(async () => {
    while (pendingSends.length > 0) {
      pendingSends.shift()?.reject(new Error("test-cleanup"));
    }
    await Promise.resolve();
  });
});

beforeEach(() => {
  mutateMock.mockReset();
  clearMutateMock.mockReset();
  vi.mocked(computeManualRevision).mockClear();
  hashControl.pending = null;
  pendingSends.length = 0;
  historyState.current = { ...EMPTY_HISTORY };
  // mutateAsync は deferred を pendingSends に積む。signal.abort で CanceledError を
  // reject する（実 axios と同挙動）。テストが解決/拒否を制御する。
  mutateMock.mockImplementation(
    (params: { signal?: AbortSignal }) =>
      new Promise((resolve, reject) => {
        pendingSends.push({ resolve, reject });
        const sig = params?.signal;
        const onAbort = () => {
          const err = new Error("canceled") as Error & { code?: string };
          err.name = "AbortError";
          err.code = "ERR_CANCELED";
          reject(err);
        };
        if (sig?.aborted) {
          onAbort();
          return;
        }
        sig?.addEventListener("abort", onAbort, { once: true });
      }),
  );
});

function deferredValue<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

function sendParams(index = 0): SendSupportChatVariables {
  const call = mutateMock.mock.calls[index];
  if (!call) throw new Error(`expected send call ${index}`);
  return call[0] as SendSupportChatVariables;
}

async function settleRejectedSend(index: number) {
  const result = mutateMock.mock.results[index];
  if (result?.type !== "return" || !(result.value instanceof Promise)) {
    throw new Error(`expected send promise ${index}`);
  }
  await act(async () => {
    await result.value.catch(() => undefined);
    await Promise.resolve();
  });
}

describe("HelpChat", () => {
  it("自然文の質問でもキーワード抽出により関連記事をコンテキストに含める", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    // 「の」「手順」「?」を除いた「レジ締め」で検索されること
    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
    const [params] = mutateMock.mock.calls[0] as [SendSupportChatParams, MutateOptions];
    expect(params.message).toBe("レジ締めの手順は？");
    expect(params).not.toHaveProperty("history");
    expect(params.keyword_candidates.length).toBeGreaterThan(0);
    expect(params.keyword_candidates[0].slug).toBe("accounting");
    expect(params.keyword_candidates.length).toBeLessThanOrEqual(20);
    expect(new Set(params.keyword_candidates.map((k) => `${k.category}/${k.slug}`)).size).toBe(
      params.keyword_candidates.length,
    );
    expect(params.manual_revision).toMatch(/^[0-9a-f]{64}$/);
    expect(JSON.stringify(params)).not.toContain("text");
  });

  it("回答と参照記事リンクを表示する", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    await act(async () => {
      callOptions().onSuccess?.({
        reply: "レジ締めは締めボタンから実行します。",
        sources: [{ title: "画面別 会計", category: "screens", slug: "accounting" }],
      });
    });

    expect(await screen.findByText("レジ締めは締めボタンから実行します。")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: /画面別 会計/ });
    expect(link).toHaveAttribute("href", "/manual/screens/accounting");
  });

  it("アシスタント回答のマークダウン記法はレンダリングされて表示される", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "予約の変更方法は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    await act(async () => {
      callOptions().onSuccess?.({
        reply: "**新規予約登録:**\n\n1. サイドバー「予約管理」をクリック\n2. 日時を選択",
        sources: [],
      });
    });

    // `**` は <strong> として描画され、生の記号文字は表示されない
    expect(await screen.findByText("新規予約登録:")).toBeInTheDocument();
    expect(screen.queryByText(/\*\*新規予約登録/)).not.toBeInTheDocument();
    // 番号付きリストが <ol>/<li> 構造になる
    expect(screen.getByRole("list")).toBeInTheDocument();
    expect(screen.getByText("サイドバー「予約管理」をクリック").closest("li")).not.toBeNull();
  });

  it("送信失敗時はインラインのエラーメッセージを表示する", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    await act(async () => {
      callOptions().onError?.(new Error("boom"));
    });

    expect(await screen.findByText(/送信に失敗しました/)).toBeInTheDocument();
  });

  it("空の状態では質問例チップを表示し、クリックでそのまま送信する", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.click(screen.getByRole("button", { name: "予約の変更方法は？" }));

    await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
    const [params] = mutateMock.mock.calls[0] as [SendSupportChatParams, MutateOptions];
    expect(params.message).toBe("予約の変更方法は？");
    expect(params).not.toHaveProperty("history");
  });

  it("Enter で送信し、Shift+Enter と IME 変換中の Enter では送信しない", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });
    const input = screen.getByLabelText("使い方を質問");

    await user.type(input, "レジ締めの手順は？");

    // Shift+Enter は改行（送信されない）
    fireEvent.keyDown(input, { key: "Enter", shiftKey: true });
    expect(mutateMock).not.toHaveBeenCalled();

    // IME 変換中の確定 Enter も送信されない
    const imeEvent = new KeyboardEvent("keydown", { key: "Enter", bubbles: true });
    Object.defineProperty(imeEvent, "isComposing", { value: true });
    fireEvent(input, imeEvent);
    expect(mutateMock).not.toHaveBeenCalled();

    await user.keyboard("{Enter}");
    await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  });

  it("送信失敗ターンの「もう一度送信」で同じ質問を再送できる", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    await act(async () => {
      callOptions().onError?.(new Error("boom"));
    });

    // transition 終了まで待ち、有効化直後の最新ノードで押す（stale node 回避）。
    await vi.waitFor(
      () => {
        expect(screen.getByRole("button", { name: /もう一度送信/ })).toBeEnabled();
      },
      { timeout: 5000, interval: 50 },
    );
    fireEvent.click(screen.getByRole("button", { name: /もう一度送信/ }));

    await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(2));
    const [retryParams] = mutateMock.mock.calls[1] as [SendSupportChatParams, MutateOptions];
    expect(retryParams.message).toBe("レジ締めの手順は？");
    // 画面のエラー表示は消える（会話履歴はサーバー側で管理される）
    expect(retryParams).not.toHaveProperty("history");
    expect(screen.queryByText(/送信に失敗しました/)).not.toBeInTheDocument();
  });

  it("「会話をリセット」で履歴を消して空状態に戻る", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "カルテの作り方を教えて");
    await user.click(screen.getByRole("button", { name: "送信" }));

    // 質問例チップは会話がある間は出ない
    expect(screen.queryByRole("button", { name: "予約の変更方法は？" })).not.toBeInTheDocument();

    // 送信中はリセット不可（isBusy）— 応答を解決してからリセットする。
    await act(async () => {
      callOptions(0).onSuccess?.({ reply: "回答", sources: [] });
    });
    await user.click(screen.getByRole("button", { name: "会話をリセット" }));

    // EMR-227: 破壊操作は確認ダイアログを挟む
    await user.click(await screen.findByRole("button", { name: "リセットする" }));

    // DELETE を呼び、成功後にローカルも消える
    expect(clearMutateMock).toHaveBeenCalledTimes(1);
    await act(async () => {
      clearCallOptions().onSuccess?.();
    });

    expect(screen.queryByText("カルテの作り方を教えて")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "予約の変更方法は？" })).toBeInTheDocument();
  });

  it("リセットの削除が失敗した場合は履歴を残してエラーを表示する", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "カルテの作り方を教えて");
    await user.click(screen.getByRole("button", { name: "送信" }));

    await act(async () => {
      callOptions(0).onSuccess?.({ reply: "回答", sources: [] });
    });
    await user.click(screen.getByRole("button", { name: "会話をリセット" }));
    await user.click(await screen.findByRole("button", { name: "リセットする" }));
    await act(async () => {
      clearCallOptions().onError?.(new Error("delete failed"));
    });

    expect(await screen.findByText(/履歴の削除に失敗しました/)).toBeInTheDocument();
    // ローカルの履歴は残る（見た目だけ消えると再ロードで復活してしまうため）
    expect(screen.getByText("カルテの作り方を教えて")).toBeInTheDocument();
  });

  it("サーバーに保存された会話履歴を復元して表示する", async () => {
    historyState.current = {
      data: [
        {
          id: 1,
          role: "user",
          content: "前回の質問です",
          created_at: "2025-01-01T09:00:00+09:00",
        },
        {
          id: 2,
          role: "assistant",
          content: "前回の回答です",
          sources: [{ title: "画面別 会計", category: "screens", slug: "accounting" }],
          created_at: "2025-01-01T09:00:05+09:00",
        },
      ],
      isSuccess: true,
      isError: false,
      isPending: false,
    };

    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    expect(await screen.findByText("前回の質問です")).toBeInTheDocument();
    expect(screen.getByText("前回の回答です")).toBeInTheDocument();
    // 保存済み sources の参照記事リンクも復元される
    const link = screen.getByRole("link", { name: /画面別 会計/ });
    expect(link).toHaveAttribute("href", "/manual/screens/accounting");
    // 復元履歴があるので質問例チップは出ない
    expect(screen.queryByRole("button", { name: "予約の変更方法は？" })).not.toBeInTheDocument();
  });

  it("履歴ロード中は読み込み表示になり送信できない", async () => {
    historyState.current = { data: [], isSuccess: false, isError: false, isPending: true };
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    expect(screen.getByText("履歴を読み込み中…")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "送信" })).toBeDisabled();
    // ロード完了前に質問例チップを押しても送信されない
    expect(screen.queryByRole("button", { name: "予約の変更方法は？" })).not.toBeInTheDocument();
    await user.keyboard("{Enter}");
    expect(mutateMock).not.toHaveBeenCalled();
  });

  it("履歴ロードに失敗しても空状態として送信できる", async () => {
    historyState.current = { data: [], isSuccess: false, isError: true, isPending: false };
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    // 質問例チップが使える
    await user.click(await screen.findByRole("button", { name: "予約の変更方法は？" }));
    expect(mutateMock).toHaveBeenCalledTimes(1);
  });

  it("provider 配下ではアンマウント→再マウント後も会話が残る", async () => {
    const user = userEvent.setup();

    function Harness() {
      const [mounted, setMounted] = useState(true);
      return (
        <HelpChatHistoryProvider>
          <button onClick={() => setMounted((v) => !v)}>パネル切替</button>
          {mounted ? <HelpChat articles={articles} onClose={() => {}} /> : null}
        </HelpChatHistoryProvider>
      );
    }

    render(<Harness />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    // パネルの開閉に相当するアンマウント→再マウント
    await user.click(screen.getByRole("button", { name: "パネル切替" }));
    await user.click(screen.getByRole("button", { name: "パネル切替" }));

    // 再マウント後もユーザーの質問ターンが残っている（質問例は出ない）
    expect(screen.getByText("レジ締めの手順は？")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "予約の変更方法は？" })).not.toBeInTheDocument();
  });

  it("provider 配下で再マウントしても保存済み履歴を再適用しない", async () => {
    historyState.current = {
      data: [
        {
          id: 1,
          role: "user",
          content: "前回の質問です",
          created_at: "2025-01-01T09:00:00+09:00",
        },
      ],
      isSuccess: true,
      isError: false,
      isPending: false,
    };
    const user = userEvent.setup();

    function Harness() {
      const [mounted, setMounted] = useState(true);
      return (
        <HelpChatHistoryProvider>
          <button onClick={() => setMounted((v) => !v)}>パネル切替</button>
          {mounted ? <HelpChat articles={articles} onClose={() => {}} /> : null}
        </HelpChatHistoryProvider>
      );
    }

    render(<Harness />, { wrapper: createWrapper() });

    // 初回マウントでサーバー履歴が復元される
    expect(await screen.findByText("前回の質問です")).toBeInTheDocument();

    // ローカルで新しい質問を追加（送信直後の状態）
    await user.type(screen.getByLabelText("使い方を質問"), "追加の質問");
    await user.click(screen.getByRole("button", { name: "送信" }));

    // アンマウント→再マウント
    await user.click(screen.getByRole("button", { name: "パネル切替" }));
    await user.click(screen.getByRole("button", { name: "パネル切替" }));

    // ローカルターンが残り、サーバー履歴が重複して挿入されない
    expect(screen.getByText("追加の質問")).toBeInTheDocument();
    expect(screen.getAllByText("前回の質問です")).toHaveLength(1);
  });
});

// ---- EMR-410 edge cases ----

function conflictError(): unknown {
  return { isAxiosError: true, response: { status: 409 } };
}

it("409 で一度だけ最新 snapshot で再送し、revision/candidate が更新される", async () => {
  const user = userEvent.setup();
  const fresh = articles.map((a) =>
    a.slug === "accounting" ? { ...a, slug: "fresh-accounting" } : a,
  );
  const onRevisionConflict = vi.fn(async () => fresh);
  render(
    <HelpChat articles={articles} onClose={() => {}} onRevisionConflict={onRevisionConflict} />,
    { wrapper: createWrapper() },
  );
  await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  const [first] = mutateMock.mock.calls[0] as [SendSupportChatParams];
  const firstRevision = first.manual_revision;

  await act(async () => {
    callOptions(0).onError?.(conflictError());
  });
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(2));
  expect(onRevisionConflict).toHaveBeenCalledTimes(1);
  const [second] = mutateMock.mock.calls[1] as [SendSupportChatParams];
  expect(second.message).toBe("レジ締めの手順は？");
  expect(second.manual_revision).not.toBe(firstRevision);
  expect(second.keyword_candidates.some((k) => k.slug === "fresh-accounting")).toBe(true);
  expect(second.keyword_candidates.some((k) => k.slug === "accounting")).toBe(false);
  expect(JSON.stringify(second)).not.toContain("text");
  expect(second).not.toHaveProperty("history");

  await act(async () => {
    callOptions(1).onSuccess?.({ reply: "回答です", sources: [] });
  });
  expect(await screen.findByText("回答です")).toBeInTheDocument();
});

it("二度目の 409 は再読込案内を表示し 3 回目の送信は行わない", async () => {
  const user = userEvent.setup();
  const onRevisionConflict = vi.fn(async () => articles);
  render(
    <HelpChat articles={articles} onClose={() => {}} onRevisionConflict={onRevisionConflict} />,
    { wrapper: createWrapper() },
  );
  await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  await act(async () => {
    callOptions(0).onError?.(conflictError());
  });
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(2));
  await act(async () => {
    callOptions(1).onError?.(conflictError());
  });

  expect(await screen.findByText(/再読み込み/)).toBeInTheDocument();
  expect(onRevisionConflict).toHaveBeenCalledTimes(1);
  expect(mutateMock).toHaveBeenCalledTimes(2);
  // ユーザーターンは 1 つだけ
  const userTurns = screen.getAllByText("レジ締めの手順は？");
  expect(userTurns.length).toBe(1);
});

it("409 refetch 中にアンマウントすると controller abort で再送しない", async () => {
  const user = userEvent.setup();
  let resolveConflict: ((a: ManualArticle[]) => void) | undefined;
  const onRevisionConflict = vi.fn(
    () =>
      new Promise<ManualArticle[]>((res) => {
        resolveConflict = res;
      }),
  );
  const { unmount } = render(
    <HelpChat articles={articles} onClose={() => {}} onRevisionConflict={onRevisionConflict} />,
    { wrapper: createWrapper() },
  );
  await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  const [firstParams] = mutateMock.mock.calls[0] as [{ signal?: AbortSignal }];
  const firstSignal = firstParams.signal;

  await act(async () => {
    callOptions(0).onError?.(conflictError());
  });
  await vi.waitFor(() => expect(onRevisionConflict).toHaveBeenCalledTimes(1));

  unmount();
  expect(firstSignal?.aborted).toBe(true);
  await act(async () => {
    resolveConflict?.(articles);
    await Promise.resolve();
  });
  // abort 済みのため 2 回目の HTTP は発行されない
  expect(mutateMock).toHaveBeenCalledTimes(1);
});

it("送信中に閉じても履歴 provider に中断ターンを残し、一度だけ再送できる", async () => {
  historyState.current = {
    data: [
      {
        id: 41,
        role: "user",
        content: "前回の質問",
        created_at: "2025-01-01T09:00:00+09:00",
      },
      {
        id: 42,
        role: "assistant",
        content: "前回の回答",
        created_at: "2025-01-01T09:00:05+09:00",
      },
    ],
    isSuccess: true,
    isError: false,
    isPending: false,
  };
  const user = userEvent.setup();
  const onRevisionConflict = vi.fn(async () => articles);
  render(
    <HelpChatHistoryProvider>
      <PersistentPanelHarness onRevisionConflict={onRevisionConflict} />
    </HelpChatHistoryProvider>,
    { wrapper: createWrapper() },
  );
  expect(await screen.findByText("前回の回答")).toBeInTheDocument();

  const question = "レジ締めの手順は？";
  await user.type(screen.getByLabelText("使い方を質問"), question);
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  const signal = sendParams().signal;

  await user.click(screen.getByRole("button", { name: "チャットを閉じる" }));
  expect(signal?.aborted).toBe(true);
  await user.click(screen.getByRole("button", { name: "チャットを開く" }));

  const canceled = await screen.findByRole("alert");
  expect(canceled).toHaveTextContent("送信が中断されました");
  expect(await screen.findByText("前回の回答")).toBeInTheDocument();
  expect(screen.getAllByText(question)).toHaveLength(1);
  expect(screen.getAllByRole("button", { name: "もう一度送信" })).toHaveLength(1);
  expect(onRevisionConflict).not.toHaveBeenCalled();

  await vi.waitFor(() =>
    expect(screen.getByRole("button", { name: "もう一度送信" })).toBeEnabled(),
  );
  await user.click(screen.getByRole("button", { name: "もう一度送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(2));
  expect(sendParams(1).message).toBe(question);
  expect(onRevisionConflict).not.toHaveBeenCalled();
  await act(async () => {
    callOptions(1).onSuccess?.({ reply: "再送の回答", sources: [] });
  });
  expect(await screen.findByText("再送の回答")).toBeInTheDocument();
});

it("revision計算中に閉じた送信はPOSTせず、再マウント後に再送できる", async () => {
  const hash = deferredValue<string>();
  hashControl.pending = hash.promise;
  const user = userEvent.setup();
  render(
    <HelpChatHistoryProvider>
      <PersistentPanelHarness />
    </HelpChatHistoryProvider>,
    { wrapper: createWrapper() },
  );
  const question = "レジ締めの手順は？";
  await user.type(screen.getByLabelText("使い方を質問"), question);
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(computeManualRevision).toHaveBeenCalledTimes(1));
  expect(mutateMock).not.toHaveBeenCalled();

  await user.click(screen.getByRole("button", { name: "チャットを閉じる" }));
  await act(async () => {
    hash.resolve("a".repeat(64));
    await hash.promise;
    await Promise.resolve();
  });
  expect(mutateMock).not.toHaveBeenCalled();
  await user.click(screen.getByRole("button", { name: "チャットを開く" }));

  expect(await screen.findByRole("alert")).toHaveTextContent(/送信が中断されました/);
  expect(screen.getAllByText(question)).toHaveLength(1);
  await vi.waitFor(() =>
    expect(screen.getByRole("button", { name: "もう一度送信" })).toBeEnabled(),
  );
  await user.click(screen.getByRole("button", { name: "もう一度送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  expect(sendParams().message).toBe(question);
  await act(async () => {
    callOptions(0).onSuccess?.({ reply: "再計算後の回答", sources: [] });
  });
  expect(await screen.findByText("再計算後の回答")).toBeInTheDocument();
});

it("遅延したQ1中断ターンをQ2の前に挿入し、Q2回答の順序を保つ", async () => {
  const firstHash = deferredValue<string>();
  hashControl.pending = firstHash.promise;
  const user = userEvent.setup();
  render(
    <HelpChatHistoryProvider>
      <>
        <HistoryOrderObserver />
        <PersistentPanelHarness />
      </>
    </HelpChatHistoryProvider>,
    { wrapper: createWrapper() },
  );

  const firstQuestion = "Q1の会計手順";
  await user.type(screen.getByLabelText("使い方を質問"), firstQuestion);
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(computeManualRevision).toHaveBeenCalledTimes(1));
  expect(mutateMock).not.toHaveBeenCalled();

  await user.click(screen.getByRole("button", { name: "チャットを閉じる" }));
  await user.click(screen.getByRole("button", { name: "チャットを開く" }));

  const secondQuestion = "Q2の予約変更";
  await user.type(screen.getByLabelText("使い方を質問"), secondQuestion);
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  expect(sendParams(0).message).toBe(secondQuestion);

  await act(async () => {
    firstHash.resolve("b".repeat(64));
    await firstHash.promise;
    await Promise.resolve();
  });
  const orderedTurns = () =>
    within(screen.getByRole("list", { name: "テスト用ターン順" }))
      .getAllByRole("listitem")
      .map((item) => item.textContent ?? "");
  await vi.waitFor(() =>
    expect(orderedTurns()).toEqual([
      `user:${firstQuestion}`,
      "assistant:送信が中断されました。必要であれば「もう一度送信」を選んでください。",
      `user:${secondQuestion}`,
    ]),
  );
  expect(mutateMock).toHaveBeenCalledTimes(1);

  await act(async () => {
    callOptions(0).onSuccess?.({ reply: "Q2の回答", sources: [] });
  });
  await vi.waitFor(() =>
    expect(orderedTurns()).toEqual([
      `user:${firstQuestion}`,
      "assistant:送信が中断されました。必要であれば「もう一度送信」を選んでください。",
      `user:${secondQuestion}`,
      "assistant:Q2の回答",
    ]),
  );
});

it("成功応答の後に閉じても中断エラーは追加されない", async () => {
  const user = userEvent.setup();
  render(
    <HelpChatHistoryProvider>
      <PersistentPanelHarness />
    </HelpChatHistoryProvider>,
    { wrapper: createWrapper() },
  );
  await user.type(screen.getByLabelText("使い方を質問"), "完了済みの質問");
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  await act(async () => {
    callOptions(0).onSuccess?.({ reply: "完了済みの回答", sources: [] });
  });
  expect(await screen.findByText("完了済みの回答")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "チャットを閉じる" }));
  await user.click(screen.getByRole("button", { name: "チャットを開く" }));
  expect(screen.getByText("完了済みの回答")).toBeInTheDocument();
  expect(screen.queryByText("送信が中断されました")).not.toBeInTheDocument();
  expect(screen.queryByRole("alert")).not.toBeInTheDocument();
});

it("履歴 reset と provider teardown 後に古い中断ターンを復元しない", async () => {
  const user = userEvent.setup();
  render(<ProviderLifecycleHarness />, { wrapper: createWrapper() });

  const resetQuestion = "reset対象の質問";
  await user.type(screen.getByLabelText("使い方を質問"), resetQuestion);
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(1));
  const resetSignal = sendParams(0).signal;
  await user.click(screen.getByRole("button", { name: "テスト用に履歴をクリア" }));
  await user.click(screen.getByRole("button", { name: "チャットを閉じる" }));
  expect(resetSignal?.aborted).toBe(true);
  await settleRejectedSend(0);
  await user.click(screen.getByRole("button", { name: "チャットを開く" }));
  expect(screen.queryByText(resetQuestion)).not.toBeInTheDocument();
  expect(screen.queryByText("送信が中断されました")).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "もう一度送信" })).not.toBeInTheDocument();

  const teardownQuestion = "provider終了対象の質問";
  await user.type(screen.getByLabelText("使い方を質問"), teardownQuestion);
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(mutateMock).toHaveBeenCalledTimes(2));
  const teardownSignal = sendParams(1).signal;
  await user.click(screen.getByRole("button", { name: "履歴 provider を終了" }));
  expect(teardownSignal?.aborted).toBe(true);
  await settleRejectedSend(1);
  await user.click(screen.getByRole("button", { name: "履歴 provider を再開" }));

  expect(await screen.findByLabelText("使い方を質問")).toBeInTheDocument();
  expect(screen.queryByText(teardownQuestion)).not.toBeInTheDocument();
  expect(screen.queryByText("送信が中断されました")).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "もう一度送信" })).not.toBeInTheDocument();
});
