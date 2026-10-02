import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { SupportChatExchange } from "../types";

let exchangesMock: SupportChatExchange[] = [];
let loadingMock = false;
let errorMock = false;

vi.mock("../api/get-support-chat-exchanges", () => ({
  useGetSupportChatExchanges: () => ({
    data: exchangesMock,
    isLoading: loadingMock,
    isError: errorMock,
  }),
}));

import { ChatHistoryPage } from "./ChatHistoryPage";

function makeExchange(overrides: Partial<SupportChatExchange> = {}): SupportChatExchange {
  return {
    id: 1,
    clinic_name: "さくら動物病院",
    staff_name: "田中",
    question: "予約のキャンセル方法は？",
    answer: "予約詳細からキャンセルできます",
    created_at: "2026-10-01T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  exchangesMock = [makeExchange()];
  loadingMock = false;
  errorMock = false;
});

describe("ChatHistoryPage 一覧", () => {
  it("全医院共有ボードとして医院・スタッフ列を表示し、権限チェックなしで描画する", () => {
    // 権限ゲートは意図的に無い — useAuth/usePermission をモックせず描画される
    render(<ChatHistoryPage />);

    expect(screen.getByRole("columnheader", { name: "医院" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "スタッフ" })).toBeInTheDocument();
    expect(screen.getByText("さくら動物病院")).toBeInTheDocument();
    expect(screen.getByText("田中")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "詳細: 予約のキャンセル方法は？" }),
    ).toBeInTheDocument();
  });

  it("質問と回答を1行で表示する", () => {
    render(<ChatHistoryPage />);

    expect(screen.getByText("予約のキャンセル方法は？")).toBeInTheDocument();
    expect(screen.getByText("予約詳細からキャンセルできます")).toBeInTheDocument();
  });

  it("履歴がない場合は空状態を表示する", () => {
    exchangesMock = [];
    render(<ChatHistoryPage />);

    expect(screen.getByText("チャット履歴はまだありません")).toBeInTheDocument();
  });

  it("取得失敗時はエラー表示にする", () => {
    errorMock = true;
    render(<ChatHistoryPage />);

    expect(screen.getByText(/取得に失敗しました/)).toBeInTheDocument();
  });
});

describe("ChatHistoryPage 詳細ダイアログ", () => {
  it("行クリックで質問・回答全文と参照記事を表示する", async () => {
    const user = userEvent.setup();
    exchangesMock = [
      makeExchange({
        sources: [{ title: "予約操作マニュアル", category: "予約", slug: "cancel" }],
      }),
    ];
    render(<ChatHistoryPage />);

    await user.click(screen.getByRole("button", { name: "詳細: 予約のキャンセル方法は？" }));

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveTextContent("予約のキャンセル方法は？");
    expect(dialog).toHaveTextContent("予約詳細からキャンセルできます");
    expect(dialog).toHaveTextContent("予約操作マニュアル");
    expect(dialog).toHaveTextContent("さくら動物病院");
    expect(dialog).toHaveTextContent("田中");
  });
});
