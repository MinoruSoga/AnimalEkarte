import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";

import type { SupportChatRoom } from "../types";

let roomsMock: SupportChatRoom[] = [];
let loadingMock = false;
let errorMock = false;

vi.mock("../api/get-support-chat-rooms", () => ({
  useGetSupportChatRooms: () => ({
    data: roomsMock,
    isLoading: loadingMock,
    isError: errorMock,
  }),
}));

import { ChatHistoryPage } from "./ChatHistoryPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <ChatHistoryPage />
    </MemoryRouter>,
  );
}

function makeRoom(overrides: Partial<SupportChatRoom> = {}): SupportChatRoom {
  return {
    id: "1-1-1",
    clinic_name: "さくら動物病院",
    staff_name: "田中",
    started_at: "2026-10-01T00:00:00Z",
    last_message_at: "2026-10-01T00:05:00Z",
    turns: [
      {
        id: 1,
        role: "user",
        content: "予約のキャンセル方法は？",
        created_at: "2026-10-01T00:00:00Z",
      },
      {
        id: 2,
        role: "assistant",
        content: "予約詳細からキャンセルできます",
        created_at: "2026-10-01T00:05:00Z",
      },
    ],
    ...overrides,
  };
}

beforeEach(() => {
  roomsMock = [makeRoom()];
  loadingMock = false;
  errorMock = false;
});

describe("ChatHistoryPage ルーム一覧", () => {
  it("サイドバーに医院・スタッフ・先頭質問を持つルームを表示し、権限チェックなしで描画する", () => {
    // 権限ゲートは意図的に無い — useAuth/usePermission をモックせず描画される
    renderPage();

    expect(screen.getByRole("navigation", { name: "チャットルーム一覧" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /予約のキャンセル方法は？/ })).toBeInTheDocument();
    expect(screen.getAllByText(/さくら動物病院/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/田中/).length).toBeGreaterThan(0);
  });

  it("最新ルームが初期選択され、質問と回答が吹き出しで表示される", () => {
    renderPage();

    expect(screen.getByRole("button", { name: /予約のキャンセル方法は？/ })).toHaveAttribute(
      "aria-current",
      "true",
    );
    const conversation = screen.getByRole("region", { name: "選択中の会話" });
    expect(conversation).toHaveTextContent("予約のキャンセル方法は？");
    expect(conversation).toHaveTextContent("予約詳細からキャンセルできます");
  });

  it("ルームをクリックすると表示会話が切り替わる", async () => {
    const user = userEvent.setup();
    roomsMock = [
      makeRoom(),
      makeRoom({
        id: "1-1-3",
        clinic_name: "ひまわり動物病院",
        staff_name: "佐藤",
        turns: [
          {
            id: 3,
            role: "user",
            content: "ワクチン記録の削除方法は？",
            created_at: "2026-09-30T01:00:00Z",
          },
          {
            id: 4,
            role: "assistant",
            content: "ワクチン一覧から削除できます",
            created_at: "2026-09-30T01:01:00Z",
          },
        ],
      }),
    ];
    renderPage();

    await user.click(screen.getByRole("button", { name: /ワクチン記録の削除方法は？/ }));

    const conversation = screen.getByRole("region", { name: "選択中の会話" });
    expect(conversation).toHaveTextContent("ワクチン記録の削除方法は？");
    expect(conversation).toHaveTextContent("ワクチン一覧から削除できます");
    expect(conversation).not.toHaveTextContent("予約のキャンセル方法は？");
    expect(conversation).toHaveTextContent("ひまわり動物病院");
    expect(conversation).toHaveTextContent("佐藤");
  });

  it("回答に参照記事がある場合はリンクとして表示する", () => {
    roomsMock = [
      makeRoom({
        turns: [
          {
            id: 1,
            role: "user",
            content: "予約のキャンセル方法は？",
            created_at: "2026-10-01T00:00:00Z",
          },
          {
            id: 2,
            role: "assistant",
            content: "予約詳細からキャンセルできます",
            sources: [{ title: "予約操作マニュアル", category: "screens", slug: "cancel" }],
            created_at: "2026-10-01T00:05:00Z",
          },
        ],
      }),
    ];
    renderPage();

    expect(screen.getByRole("link", { name: "予約操作マニュアル" })).toBeInTheDocument();
  });

  it("履歴がない場合は空状態を表示する", () => {
    roomsMock = [];
    renderPage();

    expect(screen.getByText("チャット履歴はまだありません")).toBeInTheDocument();
  });

  it("取得失敗時はエラー表示にする", () => {
    errorMock = true;
    renderPage();

    expect(screen.getByText(/取得に失敗しました/)).toBeInTheDocument();
  });
});
