import { act, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { ManualArticle } from "@/lib/manual-index";

import { HelpChat } from "./HelpChat";
import type { SendSupportChatParams } from "../api/send-support-chat";
import type { SupportChatResponse } from "../types";

const { mutateMock } = vi.hoisted(() => ({
  mutateMock: vi.fn(),
}));

vi.mock("../api/send-support-chat", () => ({
  useSendSupportChat: () => ({
    mutate: mutateMock,
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

beforeEach(() => {
  mutateMock.mockReset();
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

    const [, options] = mutateMock.mock.calls[0] as [SendSupportChatParams, MutateOptions];
    await act(async () => {
      options.onSuccess?.({
        reply: "レジ締めは締めボタンから実行します。",
        sources: [{ title: "画面別 会計", category: "screens", slug: "accounting" }],
      });
    });

    expect(await screen.findByText("レジ締めは締めボタンから実行します。")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: /画面別 会計/ });
    expect(link).toHaveAttribute("href", "/manual/screens/accounting");
  });

  it("送信失敗時はインラインのエラーメッセージを表示する", async () => {
    const user = userEvent.setup();
    render(<HelpChat articles={articles} onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("使い方を質問"), "レジ締めの手順は？");
    await user.click(screen.getByRole("button", { name: "送信" }));

    const [, options] = mutateMock.mock.calls[0] as [SendSupportChatParams, MutateOptions];
    await act(async () => {
      options.onError?.(new Error("boom"));
    });

    expect(await screen.findByText(/送信に失敗しました/)).toBeInTheDocument();
  });
});
