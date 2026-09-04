import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Routes, Route } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { HospitalizationDetail } from "./HospitalizationDetail";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    hasPermission: () => true,
  }),
}));

const HOSP_ID = "42";

function createWrapper(initialPath: string) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/hospitalization/:id" element={children} />
          <Route path="/hospitalization" element={<div>list page</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

afterEach(() => {
  server.resetHandlers();
});

// ─────────────────────────────────────────────────────────────
// Issue #55: 入院詳細ページが正しくレンダリングされること
// ─────────────────────────────────────────────────────────────

describe("HospitalizationDetail — notFound/error ハンドリング (Issue #55)", () => {
  it("API が 404 を返すとき「入院情報が見つかりません」が表示される", async () => {
    server.use(
      http.get(`/api/v1/hospitalizations/${HOSP_ID}`, () =>
        HttpResponse.json({ error: "not found" }, { status: 404 }),
      ),
    );

    render(<HospitalizationDetail />, { wrapper: createWrapper(`/hospitalization/${HOSP_ID}`) });

    await waitFor(() => {
      expect(screen.getByText("入院情報が見つかりません")).toBeInTheDocument();
    });
    // スピナーや空白ではなく、エラーメッセージが表示されること
    expect(screen.queryByRole("progressbar")).not.toBeInTheDocument();
  });

  it("id パラメータが存在しないとき「入院情報が見つかりません」が表示される（空白にならない）", () => {
    // /hospitalization/:id ルートに id なしでマッチさせることはできないため、
    // id="" になる useParams を模倣する代わりに、
    // enabled: false のとき !id ガードが発動することを確認する。
    // このケースは router が直接 /hospitalization/ (trailing slash) に来た場合を想定。
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/hospitalization/"]}>
          <Routes>
            {/* :id がない route に HospitalizationDetail を置く */}
            <Route path="/hospitalization/" element={<HospitalizationDetail />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    // id が undefined → !id ガードが即座に発動 → "入院情報が見つかりません"
    expect(screen.getByText("入院情報が見つかりません")).toBeInTheDocument();
  });

  it("API がサーバーエラー (500) を返すとき汎用エラーが表示される（空白にならない）", async () => {
    server.use(
      http.get(`/api/v1/hospitalizations/${HOSP_ID}`, () =>
        HttpResponse.json({ error: "internal server error" }, { status: 500 }),
      ),
    );

    render(<HospitalizationDetail />, { wrapper: createWrapper(`/hospitalization/${HOSP_ID}`) });

    await waitFor(() => {
      expect(screen.getByText("データの取得に失敗しました")).toBeInTheDocument();
    });
  });
});
