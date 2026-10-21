import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { act, render, screen } from "@testing-library/react";
import React from "react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";

import { ManualHelpTab } from "./ManualHelpTab";

const { axiosGetMock, axiosPostMock } = vi.hoisted(() => ({
  axiosGetMock: vi.fn(),
  axiosPostMock: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    get: axiosGetMock,
    post: axiosPostMock,
  },
}));

beforeEach(() => {
  axiosGetMock.mockReset();
  axiosPostMock.mockReset();
});

const authContext: AuthContextValue = {
  user: null,
  currentClinicId: "clinic-test-1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => false,
  refreshPermissions: async () => {},
};

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider value={authContext}>
          <MemoryRouter>{children}</MemoryRouter>
        </AuthContext.Provider>
      </QueryClientProvider>
    );
  };
}

describe("ManualHelpTab", () => {
  it("検索クエリに応じてマニュアル記事へのリンクを表示する", async () => {
    axiosGetMock.mockResolvedValue({ data: { data: [] } });
    const user = userEvent.setup();
    render(<ManualHelpTab onClose={() => {}} />, { wrapper: createWrapper() });

    await user.type(screen.getByLabelText("マニュアルを検索"), "会計");

    const results = await screen.findAllByRole("link");
    expect(results.length).toBeGreaterThan(0);
    expect(results[0]).toHaveAttribute("href", expect.stringContaining("/manual/"));
  });

  it("クエリが空のときは結果リストを出さない", () => {
    render(<ManualHelpTab onClose={() => {}} />, { wrapper: createWrapper() });

    expect(screen.queryByRole("list")).not.toBeInTheDocument();
  });

  it("チャットが有効な環境では会話 UI を表示する", async () => {
    axiosGetMock.mockImplementation((url: string) => {
      if (url === "/v1/support/chat/status") {
        return Promise.resolve({ data: { enabled: true } });
      }
      return Promise.resolve({ data: { data: [] } });
    });
    render(<ManualHelpTab onClose={() => {}} />, { wrapper: createWrapper() });

    expect(await screen.findByLabelText("使い方を質問")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "送信" })).toBeInTheDocument();
  });
});

// ---- EMR-410 strict-boundary integration ----

const strictAuth: AuthContextValue = { ...authContext, hasPermission: () => true };

function createPermWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider value={strictAuth}>
          <MemoryRouter>{children}</MemoryRouter>
        </AuthContext.Provider>
      </QueryClientProvider>
    );
  };
}

function chatGetHandler(url: string): Promise<{ data: unknown }> {
  if (url === "/v1/support/chat/status") {
    return Promise.resolve({ data: { enabled: true } });
  }
  return Promise.resolve({ data: { data: [] } });
}

it("権限ありで overrides 取得失敗 => 送信不能 + 通知、空 base へ fallback しない", async () => {
  axiosGetMock.mockImplementation((url: string) => {
    if (url === "/v1/manual/articles") return Promise.reject(new Error("boom"));
    return chatGetHandler(url);
  });
  axiosPostMock.mockResolvedValue({ data: { reply: "ok", sources: [] } });
  render(<ManualHelpTab onClose={() => {}} />, { wrapper: createPermWrapper() });

  const notice = await screen.findByRole("status");
  expect(notice).toHaveTextContent("送信できません");
  const input = screen.getByLabelText("使い方を質問");
  const user = userEvent.setup();
  await user.type(input, "会計の方法");
  const sendBtn = screen.getByRole("button", { name: "送信" });
  expect(sendBtn).toBeDisabled();
  await user.click(sendBtn);
  expect(axiosPostMock).not.toHaveBeenCalled();
});

it("pending POST 中に権限剥奪 => controller abort、assistant 応答は表示されない", async () => {
  let resolveGet: ((v: unknown) => void) | undefined;
  axiosGetMock.mockImplementation((url: string) => {
    if (url === "/v1/manual/articles") {
      return new Promise((res) => {
        resolveGet = res;
      });
    }
    return chatGetHandler(url);
  });
  let capturedSignal: AbortSignal | undefined;
  let resolvePost: (() => void) | undefined;
  axiosPostMock.mockImplementation((_u: string, _b: unknown, cfg?: { signal?: AbortSignal }) => {
    capturedSignal = cfg?.signal;
    // Deliberately resolve after cancellation to exercise the component's stale-response guard.
    return new Promise((resolve) => {
      resolvePost = () => resolve({ data: { reply: "権限変更前の回答", sources: [] } });
    });
  });
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  // 最初に権限 true でマウント → overrides resolve → canSubmit true
  function DynamicWrapper({ children }: { children: ReactNode }) {
    const [perm, setPerm] = React.useState(true);
    return (
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider value={{ ...strictAuth, hasPermission: () => perm }}>
          <MemoryRouter>{children}</MemoryRouter>
        </AuthContext.Provider>
        <button type="button" onClick={() => setPerm(false)}>
          テスト用権限解除
        </button>
      </QueryClientProvider>
    );
  }
  render(<ManualHelpTab onClose={() => {}} />, { wrapper: DynamicWrapper });
  await vi.waitFor(() => expect(resolveGet).toBeDefined());
  await act(async () => {
    resolveGet?.({
      data: {
        data: [
          {
            id: 1,
            category: "screens",
            slug: "test-permitted-guide",
            title: "会計テスト",
            order_value: 1,
            section: "テスト",
            body_markdown: "会計の方法を確認する。",
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
        ],
      },
    });
  });
  const user = userEvent.setup();
  await user.type(await screen.findByLabelText("使い方を質問"), "会計の方法");
  await user.click(screen.getByRole("button", { name: "送信" }));
  await vi.waitFor(() => expect(axiosPostMock).toHaveBeenCalledTimes(1));
  await user.click(screen.getByRole("button", { name: "テスト用権限解除" }));
  expect(capturedSignal?.aborted).toBe(true);
  await act(async () => {
    resolvePost?.();
  });
  expect(screen.queryByText("権限変更前の回答")).not.toBeInTheDocument();
  expect(axiosPostMock).toHaveBeenCalledTimes(1);
});

// ---- fix-11: 非 chat path の strict gate（bundled fallback 除去） ----

function chatOffGetHandler(url: string): Promise<{ data: unknown }> {
  if (url === "/v1/support/chat/status") {
    return Promise.resolve({ data: { enabled: false } });
  }
  return Promise.resolve({ data: { data: [] } });
}

describe("ManualHelpTab 非chat strict gate (fix-11)", () => {
  it("chat off + 権限あり + pending => 検索 UI 非表示・loading 表示", async () => {
    axiosGetMock.mockImplementation((url: string) => {
      if (url === "/v1/manual/articles") return new Promise(() => {});
      return chatOffGetHandler(url);
    });
    render(<ManualHelpTab onClose={() => {}} />, { wrapper: createPermWrapper() });
    expect(await screen.findByText("マニュアル記事を読み込んでいます")).toBeInTheDocument();
    expect(screen.queryByRole("searchbox")).toBeNull();
    expect(screen.queryByRole("button", { name: "再試行" })).toBeNull();
  });

  it("chat off + 権限あり + error => error+retry、bundled 検索結果なし、retry が refetch を呼ぶ", async () => {
    axiosGetMock.mockImplementation((url: string) => {
      if (url === "/v1/manual/articles") return Promise.reject(new Error("boom"));
      return chatOffGetHandler(url);
    });
    render(<ManualHelpTab onClose={() => {}} />, { wrapper: createPermWrapper() });
    expect(await screen.findByText("マニュアル記事の取得に失敗しました")).toBeInTheDocument();
    expect(screen.queryByRole("searchbox")).toBeNull();
    // bundled 記事の検索結果・説明文を出さない。
    expect(
      screen.queryByText("知りたい操作や画面名を入力すると、取扱説明書の関連項目を表示します。"),
    ).toBeNull();

    const user = userEvent.setup();
    const callsBefore = axiosGetMock.mock.calls.filter(([u]) => u === "/v1/manual/articles").length;
    await user.click(screen.getByRole("button", { name: "再試行" }));
    const callsAfter = axiosGetMock.mock.calls.filter(([u]) => u === "/v1/manual/articles").length;
    expect(callsAfter).toBeGreaterThan(callsBefore);
  });

  it("chat off + 権限あり + 200 data:[] => 検索 view を描画する", async () => {
    axiosGetMock.mockImplementation(chatOffGetHandler);
    render(<ManualHelpTab onClose={() => {}} />, { wrapper: createPermWrapper() });
    expect(await screen.findByRole("searchbox", { name: "マニュアルを検索" })).toBeInTheDocument();
    expect(screen.queryByText("マニュアル記事の取得に失敗しました")).toBeNull();
    expect(screen.queryByText("マニュアル記事を読み込んでいます")).toBeNull();
  });
});
