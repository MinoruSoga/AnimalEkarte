import { act, useState, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { ManualArticle } from "@/lib/manual-index";

import { HelpChat } from "./HelpChat";
import { HelpChatHistoryProvider } from "./HelpChatHistoryProvider";
import type { SendSupportChatParams } from "../api/send-support-chat";
import type { SupportChatHistoryRecord, SupportChatResponse } from "../types";

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

const { mutateMock, clearMutateMock, historyState } = vi.hoisted(() => ({
  mutateMock: vi.fn(),
  clearMutateMock: vi.fn(),
  historyState: {
    current: {
      data: [] as SupportChatHistoryRecord[],
      isSuccess: true,
      isError: false,
      isPending: false,
    },
  },
}));

vi.mock("../api/send-support-chat", () => ({
  useSendSupportChat: () => ({
    mutate: mutateMock,
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

type MutateOptions = {
  onSuccess?: (res: SupportChatResponse) => void;
  onError?: (err: unknown) => void;
};

function callOptions(callIndex = 0): MutateOptions {
  const [, options] = mutateMock.mock.calls[callIndex] as [SendSupportChatParams, MutateOptions];
  return options;
}

type ClearOptions = {
  onSuccess?: () => void;
  onError?: (err: unknown) => void;
};

function clearCallOptions(callIndex = 0): ClearOptions {
  const [, options] = clearMutateMock.mock.calls[callIndex] as [undefined, ClearOptions];
  return options;
}

beforeEach(() => {
  mutateMock.mockReset();
  clearMutateMock.mockReset();
  historyState.current = { ...EMPTY_HISTORY };
});

describe("HelpChat", () => {
  it("自然文の質問でもキーワード抽出により関連記事をコンテキストに含める", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    // 「の」「手順」「?」を除いた「レジ締め」で検索されること
    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    expect(mutateMock).toHaveBeenCalledTimes(1);
    const [params] = mutateMock.mock.calls[0] as [SendSupportChatParams, MutateOptions];
    expect(params.message).toBe("レジ締めの手順は？");
    expect(params.history).toEqual([]);
    expect(params.context.length).toBeGreaterThan(0);
    expect(params.context[0].title).toBe("画面別 会計");
    expect(params.context[0].slug).toBe("accounting");
    expect(params.context[0].text.length).toBeGreaterThan(0);
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

    expect(mutateMock).toHaveBeenCalledTimes(1);
    const [params] = mutateMock.mock.calls[0] as [SendSupportChatParams, MutateOptions];
    expect(params.message).toBe("予約の変更方法は？");
    expect(params.history).toEqual([]);
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
    expect(mutateMock).toHaveBeenCalledTimes(1);
  });

  it("送信失敗ターンの「もう一度送信」で同じ質問を再送できる", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    await act(async () => {
      callOptions().onError?.(new Error("boom"));
    });

    await user.click(await screen.findByRole("button", { name: /もう一度送信/ }));

    expect(mutateMock).toHaveBeenCalledTimes(2);
    const [retryParams] = mutateMock.mock.calls[1] as [SendSupportChatParams, MutateOptions];
    expect(retryParams.message).toBe("レジ締めの手順は？");
    // 失敗したやり取りは history に含めず、画面のエラー表示は消える
    expect(retryParams.history).toEqual([]);
    expect(screen.queryByText(/送信に失敗しました/)).not.toBeInTheDocument();
  });

  it("「会話をリセット」で履歴を消して空状態に戻る", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "カルテの作り方を教えて");
    await user.click(screen.getByRole("button", { name: "送信" }));

    // 質問例チップは会話がある間は出ない
    expect(screen.queryByRole("button", { name: "予約の変更方法は？" })).not.toBeInTheDocument();

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
