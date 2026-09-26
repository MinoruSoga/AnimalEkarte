import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";

import { ManualHelpTab } from "./ManualHelpTab";

const { axiosGetMock } = vi.hoisted(() => ({
  axiosGetMock: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    get: axiosGetMock,
  },
}));

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
