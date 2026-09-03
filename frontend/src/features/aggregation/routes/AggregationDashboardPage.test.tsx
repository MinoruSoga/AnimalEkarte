import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { AggregationDashboardPage } from "./AggregationDashboardPage";
import type { AggregationResponse } from "../api/get-aggregations";

const mockResponse: AggregationResponse = {
  owners: [
    {
      owner_id: "owner1",
      owner_name: "田中太郎",
      total_fee: 500000,
      total_visit_count: 20,
      annual_visit_count: 12,
      last_visit_date: "2026-04-20",
      first_visit_date: "2024-01-15",
      annual_amount: 150000,
      billing_count: 8,
      period_visit_count: 5,
      days_since_last_visit: 7,
      last_visit_bucket: "within_3m",
    },
    {
      owner_id: "owner2",
      owner_name: "鈴木花子",
      total_fee: 200000,
      total_visit_count: 8,
      annual_visit_count: 4,
      last_visit_date: "2026-01-10",
      first_visit_date: "2025-06-01",
      annual_amount: 80000,
      billing_count: 4,
      period_visit_count: 2,
      days_since_last_visit: 107,
      last_visit_bucket: "over_3m",
    },
  ],
  page: 1,
  per_page: 50,
  total: 2,
};

const createWrapper = (initialEntries: string[] = ["/aggregation"]) =>
  createTestWrapper({ initialEntries });

describe("AggregationDashboardPage", () => {
  beforeEach(() => {
    // Mock localStorage for clinic_id
    localStorage.setItem("auth_current_clinic:v1", "1");

    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", () => {
        return HttpResponse.json(mockResponse);
      }),
    );
    vi.clearAllMocks();
  });

  afterEach(() => {
    localStorage.removeItem("auth_current_clinic:v1");
  });

  it("should render revenue tab by default", async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByRole("tab", { name: "売上ランキング" })).toHaveAttribute(
      "data-state",
      "active",
    );
    expect(screen.getByText("田中太郎")).toBeInTheDocument();
  });

  it("should switch tabs and update URL state", async () => {
    const user = userEvent.setup();
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const tabs = screen.getAllByRole("tab");
    expect(tabs).toHaveLength(3);

    await user.click(tabs[1]);
    expect(screen.getByRole("tab", { name: "来院回数" })).toHaveAttribute("data-state", "active");
  });

  it("should restore active tab from URL ?tab=visit on initial render", async () => {
    render(<AggregationDashboardPage />, {
      wrapper: createWrapper(["/aggregation?tab=visit"]),
    });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByRole("tab", { name: "来院回数" })).toHaveAttribute("data-state", "active");
  });

  it("should restore active tab from URL ?tab=last_visit on initial render", async () => {
    render(<AggregationDashboardPage />, {
      wrapper: createWrapper(["/aggregation?tab=last_visit"]),
    });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByRole("tab", { name: "最終来院" })).toHaveAttribute("data-state", "active");
  });

  it("should fall back to revenue when URL has invalid tab value", async () => {
    render(<AggregationDashboardPage />, {
      wrapper: createWrapper(["/aggregation?tab=__not_a_tab__"]),
    });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByRole("tab", { name: "売上ランキング" })).toHaveAttribute(
      "data-state",
      "active",
    );
  });

  it("should reset selection when switching tabs", async () => {
    const user = userEvent.setup();
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("田中太郎")).toBeInTheDocument();
    });

    // revenue タブで 1 件選択
    const ownerCheckbox = screen.getByRole("checkbox", { name: /田中太郎を選択/ });
    await user.click(ownerCheckbox);
    expect(screen.getByRole("button", { name: /1件をCSV出力/ })).toBeInTheDocument();

    // visit タブに切り替え → 選択がリセットされる
    const visitTab = screen.getByRole("tab", { name: "来院回数" });
    await user.click(visitTab);

    await waitFor(() => {
      expect(screen.getByRole("tab", { name: "来院回数" })).toHaveAttribute("data-state", "active");
    });
    // 選択件数バッジが消え、CSV ボタンも disabled に戻る
    expect(screen.queryByText(/件選択中/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /CSV出力/ })).toBeDisabled();
  });

  it("should render exactly 3 tabs (revenue / visit / last_visit) with correct labels", async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const tabs = screen.getAllByRole("tab");
    expect(tabs).toHaveLength(3);
    expect(tabs[0]).toHaveTextContent("売上ランキング");
    expect(tabs[1]).toHaveTextContent("来院回数");
    expect(tabs[2]).toHaveTextContent("最終来院");
  });

  it("should display error message on API failure", async () => {
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", () => {
        return HttpResponse.json({ error: "Not found" }, { status: 404 });
      }),
    );

    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText(/読み込みに失敗しました/)).toBeInTheDocument();
    });
  });

  // 仕様書 §10.2「エラー時に失敗理由が分かること」: HTTP ステータスコードが画面に併記される
  it("should include HTTP status in error message for actionable diagnostics", async () => {
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", () => {
        return HttpResponse.json({ error: "internal" }, { status: 500 });
      }),
    );

    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText(/HTTP 500/)).toBeInTheDocument();
    });
    expect(screen.getByText(/読み込みに失敗しました/)).toBeInTheDocument();
  });

  it("should not render count badge or pagination when API fails", async () => {
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", () => {
        return HttpResponse.json({ error: "Not found" }, { status: 404 });
      }),
    );

    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText(/読み込みに失敗しました/)).toBeInTheDocument();
    });

    // 件数バッジが出ない (data が null のため)
    expect(screen.queryByText(/^\d+ 件$/)).not.toBeInTheDocument();
    // ページネーションのナビボタンも出ない
    expect(screen.queryByRole("button", { name: /次へ/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /前へ/ })).not.toBeInTheDocument();
  });

  it("should show empty state and disable header checkbox when owners is empty", async () => {
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", () => {
        return HttpResponse.json({ owners: [], page: 1, per_page: 50, total: 0 });
      }),
    );

    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("データが見つかりません")).toBeInTheDocument();
    });

    expect(screen.getByText("0 件")).toBeInTheDocument();
    const headerCheckbox = screen.getByRole("checkbox", { name: "全選択" });
    expect(headerCheckbox).toBeDisabled();

    // CSV ボタンも 0 件時には無効
    expect(screen.getByRole("button", { name: /CSV出力/ })).toBeDisabled();
  });

  it("should render pagination controls", async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.getByText("2 件")).toBeInTheDocument();
  });

  it("should display loading state initially", () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    expect(screen.getByText("読み込み中...")).toBeInTheDocument();
  });

  it("should render table with owner data", async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("田中太郎")).toBeInTheDocument();
    });

    expect(screen.getByText("鈴木花子")).toBeInTheDocument();
  });

  it("should disable CSV export button when no rows are selected", async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const csvButton = screen.getByRole("button", { name: /CSV出力/ });
    expect(csvButton).toBeDisabled();
    expect(csvButton).toHaveAttribute("title", "出力対象を選択してください");
  });

  it("should enable CSV export button when at least one row is selected", async () => {
    const user = userEvent.setup();
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("田中太郎")).toBeInTheDocument();
    });

    const ownerCheckbox = screen.getByRole("checkbox", { name: /田中太郎を選択/ });
    await user.click(ownerCheckbox);

    const csvButton = screen.getByRole("button", { name: /1件をCSV出力/ });
    expect(csvButton).not.toBeDisabled();
  });

  // 仕様書 §10.2: 各タブで期間・フィルタ変更時に page が1へ戻る
  it("should reset page to 1 when filter/sort changes", async () => {
    const user = userEvent.setup();
    const requestUrls: string[] = [];
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", ({ request }) => {
        requestUrls.push(request.url);
        return HttpResponse.json(mockResponse);
      }),
    );

    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // ソートを「期間内来院回数」に変更 → page=1 でリクエストされる
    const sortSelect = screen.getByRole("combobox", { name: /並び替え/ });
    await user.click(sortSelect);
    const periodVisitOption = await screen.findByRole("option", { name: "期間内来院回数" });
    await user.click(periodVisitOption);

    await waitFor(() => {
      // ISSUE-180: サマリーの人数リクエスト (cpm_stage 付き) を除外し、メインの一覧リクエストを見る。
      const mainUrls = requestUrls.filter((u) => !u.includes("cpm_stage="));
      const lastUrl = mainUrls[mainUrls.length - 1];
      expect(lastUrl).toContain("sort=period_visit_count");
      expect(lastUrl).toContain("page=1");
    });
  });

  it("should toggle order between asc and desc", async () => {
    const user = userEvent.setup();
    const requestUrls: string[] = [];
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", ({ request }) => {
        requestUrls.push(request.url);
        return HttpResponse.json(mockResponse);
      }),
    );

    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // 既定タブ revenue は order=desc。トグルで asc に変わる。
    const orderToggle = screen.getByRole("button", { name: /降順/ });
    await user.click(orderToggle);

    await waitFor(() => {
      // ISSUE-180: サマリーの人数リクエスト (cpm_stage 付き) を除外し、メインの一覧リクエストを見る。
      const mainUrls = requestUrls.filter((u) => !u.includes("cpm_stage="));
      const lastUrl = mainUrls[mainUrls.length - 1];
      expect(lastUrl).toContain("order=asc");
      expect(lastUrl).toContain("page=1");
    });
  });

  it("should show axis-specific sort options per tab", async () => {
    const user = userEvent.setup();
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // revenue タブ: 年間診療費 / 期間内来院回数 / 最終来院日 / 経過日数 / 飼主名
    let sortSelect = screen.getByRole("combobox", { name: /並び替え/ });
    await user.click(sortSelect);
    expect(await screen.findByRole("option", { name: "年間診療費" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "期間内来院回数" })).toBeInTheDocument();
    // visit / last_visit にのみ存在する選択肢は revenue では出さない (仕様 §4.1)
    // → revenue 専用の「年間診療費」が出ること自体で十分検証可。閉じる。
    await user.keyboard("{Escape}");

    // last_visit タブに切り替え → 経過日数が選べる
    const lastVisitTab = screen.getByRole("tab", { name: "最終来院" });
    await user.click(lastVisitTab);

    await waitFor(() => {
      expect(screen.getByRole("tab", { name: "最終来院" })).toHaveAttribute("data-state", "active");
    });

    sortSelect = screen.getByRole("combobox", { name: /並び替え/ });
    await user.click(sortSelect);
    expect(await screen.findByRole("option", { name: "経過日数" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "最終来院日" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "飼主名" })).toBeInTheDocument();
    // last_visit には「年間診療費」「期間内来院回数」を出さない (仕様 §4.3)
    expect(screen.queryByRole("option", { name: "年間診療費" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "期間内来院回数" })).not.toBeInTheDocument();
  });

  it('should render last_visit_bucket select with spec §4.3 label "3ヶ月未満"', async () => {
    const user = userEvent.setup();
    render(<AggregationDashboardPage />, {
      wrapper: createWrapper(["/aggregation?tab=last_visit"]),
    });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // 最終来院 select を開く
    const buckets = screen.getAllByRole("combobox");
    const bucketSelect = buckets.find((el) => el.textContent?.includes("3ヶ月以上"));
    expect(bucketSelect).toBeDefined();
    await user.click(bucketSelect!);

    expect(await screen.findByRole("option", { name: "3ヶ月未満" })).toBeInTheDocument();
    // 旧誤訳「3ヶ月以内」が残っていないこと
    expect(screen.queryByRole("option", { name: "3ヶ月以内" })).not.toBeInTheDocument();
  });
});
