import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Routes, Route } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { HospitalizationDetail } from "./HospitalizationDetail";
import type { BackendHospitalization } from "../api/types";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    hasPermission: () => true,
  }),
}));

// 見出しの検証に不要な重い子ビューは境界でスタブ化する。
// PageLayout と HospitalizationDetail 本体はモックしない（見出しは PageLayout が描画する）。
vi.mock("../components/HospitalizationExpandedView", () => ({
  HospitalizationExpandedView: () => <div data-testid="hospitalization-expanded-view" />,
}));

vi.mock("../components/HospitalizationTabbedView", () => ({
  HospitalizationTabbedView: () => <div data-testid="hospitalization-tabbed-view" />,
}));

vi.mock("../components/HospitalizationDetailActions", () => ({
  HospitalizationDetailActions: () => <div data-testid="hospitalization-detail-actions" />,
}));

const HOSP_ID = "42";

/** 未解決（読込中/エラー/該当なし）時の区分非依存の見出し。 */
const NEUTRAL_TITLE = "入院・ホテル詳細・カルテ";
/** 保存済み「入院」記録の区分固有の見出し。 */
const INPATIENT_TITLE = "入院詳細・カルテ";
/** 保存済み「ホテル」記録の区分固有の見出し。 */
const HOTEL_TITLE = "ホテル詳細・カルテ";

/** GET /v1/hospitalizations/:id の最小 wire（必須項目のみ。transforms.test.ts と同じ形）。 */
const savedInpatient: BackendHospitalization = {
  id: 42,
  clinic_id: 1,
  owner_id: 20,
  pet_id: 10,
  hospitalization_type: "hospitalization",
  start_date: "2026-03-25T00:00:00Z",
  end_date: "2026-03-28T00:00:00Z",
  status: "admitted",
  memo: "",
  owner_request: "",
  staff_notes: "",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

const savedHotel: BackendHospitalization = { ...savedInpatient, hospitalization_type: "hotel" };

/** MSW 応答を手動で解放するゲート（未解決 Promise をテスト終了まで残さない）。 */
function createGate(): { gate: Promise<void>; release: () => void } {
  let release: (() => void) | undefined;
  const gate = new Promise<void>((resolve) => {
    release = () => resolve();
  });
  // Promise エグゼキュータは同期実行のため、この時点で release は設定済み。
  return { gate, release: () => release?.() };
}

/** 1 次見出し（PageLayout のタイトル）を名前で取得する。 */
function pageTitle(name: string) {
  return screen.getByRole("heading", { level: 1, name });
}

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
    // 該当なしも区分を断定しない中立見出し
    expect(pageTitle(NEUTRAL_TITLE)).toBeInTheDocument();
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
    // 該当なしも区分を断定しない中立見出し
    expect(pageTitle(NEUTRAL_TITLE)).toBeInTheDocument();
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
    // 取得失敗も区分を断定しない中立見出し
    expect(pageTitle(NEUTRAL_TITLE)).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// PR679: 詳細見出し — 未解決（読込中/エラー/該当なし）は区分非依存、保存済みは区分固有
// ─────────────────────────────────────────────────────────────

describe("HospitalizationDetail — 見出し (PR679)", () => {
  it("取得完了までの読込中は中立見出し「入院・ホテル詳細・カルテ」を表示する", async () => {
    const { gate, release } = createGate();
    server.use(
      http.get(`/api/v1/hospitalizations/${HOSP_ID}`, async () => {
        await gate;
        return HttpResponse.json(savedHotel);
      }),
    );

    render(<HospitalizationDetail />, { wrapper: createWrapper(`/hospitalization/${HOSP_ID}`) });

    // 応答が未解決の段階では区分を断定しない
    try {
      expect(pageTitle(NEUTRAL_TITLE)).toBeInTheDocument();
      expect(screen.getByRole("status")).toBeInTheDocument();
    } finally {
      // REDで見出しの検証に失敗しても、応答待ちを残さない。
      release();
    }
    await waitFor(() => {
      expect(pageTitle(HOTEL_TITLE)).toBeInTheDocument();
    });
    expect(screen.queryByRole("heading", { level: 1, name: NEUTRAL_TITLE })).toBeNull();
  });

  it("保存済み「入院」記録は区分固有の見出し「入院詳細・カルテ」を表示する", async () => {
    server.use(
      http.get(`/api/v1/hospitalizations/${HOSP_ID}`, () => HttpResponse.json(savedInpatient)),
    );

    render(<HospitalizationDetail />, { wrapper: createWrapper(`/hospitalization/${HOSP_ID}`) });

    await waitFor(() => {
      expect(screen.getByTestId("hospitalization-expanded-view")).toBeInTheDocument();
      expect(pageTitle(INPATIENT_TITLE)).toBeInTheDocument();
    });
    expect(screen.queryByRole("heading", { level: 1, name: NEUTRAL_TITLE })).toBeNull();
    expect(screen.getByTestId("hospitalization-expanded-view")).toBeInTheDocument();
  });

  it("保存済み「ホテル」記録は区分固有の見出し「ホテル詳細・カルテ」を表示する", async () => {
    server.use(
      http.get(`/api/v1/hospitalizations/${HOSP_ID}`, () => HttpResponse.json(savedHotel)),
    );

    render(<HospitalizationDetail />, { wrapper: createWrapper(`/hospitalization/${HOSP_ID}`) });

    await waitFor(() => {
      expect(pageTitle(HOTEL_TITLE)).toBeInTheDocument();
    });
    expect(screen.queryByRole("heading", { level: 1, name: NEUTRAL_TITLE })).toBeNull();
  });
});
